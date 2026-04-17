package cli

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/ernesto2108/anvil/internal/instrumentation"
	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/internal/orchestrator"
	"github.com/ernesto2108/anvil/pkg/output"
)

// digestCheckpointer is an orchestrator.EventSink decorator that triggers an
// incremental digest UPSERT after every agent.end event.
//
// Purpose: if the Claude conversation or the run process dies mid-pipeline,
// the last checkpoint survives in the digests table. A new session can then
// recover context by searching digests (via embeddings) without needing a
// handoff file or TASK-ID.
//
// Concurrency model:
//   - Emit is called from the orchestrator's main goroutine (one at a time,
//     per applyWorkerResult). It is non-blocking — digest work happens in a
//     worker goroutine.
//   - At most one checkpoint goroutine runs at any time. Emits that arrive
//     while a checkpoint is in flight set a pending flag; the worker loops
//     once more after finishing instead of queueing.
//   - Wait() blocks until the worker goroutine is idle. Call it before the
//     terminal digest generation to avoid racing on the same digests row.
type digestCheckpointer struct {
	inner      orchestrator.EventSink
	db         *sql.DB
	summarizer memory.Summarizer
	embedder   memory.Embedder
	modelUsed  string
	runID      string
	project    string
	parentCtx  context.Context // cancelled by Ctrl+C via the caller's signal context

	mu      sync.Mutex
	running bool
	pending bool
	wg      sync.WaitGroup
}

// newDigestCheckpointer wraps the given inner sink with incremental digest
// triggers. When summarizer is nil the wrapper is a pass-through — Emit still
// forwards to inner but no digest work is scheduled. This lets callers build
// the wrapper unconditionally and rely on pickSummarizer's availability
// signal at construction time.
//
// parentCtx should be the signal-cancellable context of the run. When the
// user hits Ctrl+C the in-flight checkpoint aborts instead of blocking the
// process for the full per-checkpoint timeout.
func newDigestCheckpointer(
	parentCtx context.Context,
	inner orchestrator.EventSink,
	db *sql.DB,
	summarizer memory.Summarizer,
	embedder memory.Embedder,
	modelUsed, runID, project string,
) *digestCheckpointer {
	return &digestCheckpointer{
		inner:      inner,
		db:         db,
		summarizer: summarizer,
		embedder:   embedder,
		modelUsed:  modelUsed,
		runID:      runID,
		project:    project,
		parentCtx:  parentCtx,
	}
}

// Emit forwards the event to the wrapped sink and, for agent.end events,
// triggers an incremental digest checkpoint. Never blocks the caller.
func (d *digestCheckpointer) Emit(ev instrumentation.Event) {
	d.inner.Emit(ev)

	if d.summarizer == nil {
		return
	}
	if ev.EventType != instrumentation.EventAgentEnd {
		return
	}

	d.mu.Lock()
	if d.running {
		// A checkpoint is already running. Mark pending so the worker
		// runs once more after it finishes, picking up the newly
		// arrived agent output.
		d.pending = true
		d.mu.Unlock()
		return
	}
	d.running = true
	d.wg.Add(1)
	d.mu.Unlock()

	go d.worker()
}

// worker runs checkpoints until no more are pending.
func (d *digestCheckpointer) worker() {
	defer d.wg.Done()

	for {
		d.runOne()

		d.mu.Lock()
		if !d.pending {
			d.running = false
			d.mu.Unlock()
			return
		}
		d.pending = false
		d.mu.Unlock()
	}
}

// runOne executes a single checkpoint with a bounded timeout so a slow
// summarizer cannot hold the worker indefinitely. Derives from parentCtx so
// Ctrl+C propagates and the worker exits promptly.
func (d *digestCheckpointer) runOne() {
	ctx, cancel := context.WithTimeout(d.parentCtx, 2*time.Minute)
	defer cancel()

	if _, err := summarizeAndUpsert(ctx, d.db, d.summarizer, d.embedder, d.modelUsed, d.runID, d.project); err != nil {
		// Suppress cancellation noise — expected on Ctrl+C.
		if ctx.Err() == nil {
			output.Warn("checkpoint digest: %s", err)
		}
	}
}

// Wait blocks until the worker goroutine is idle. Call this before running
// the terminal digest to avoid a race on the same digests row.
func (d *digestCheckpointer) Wait() {
	d.wg.Wait()
}
