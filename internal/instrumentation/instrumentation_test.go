package instrumentation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Mock EventWriter
// ---------------------------------------------------------------------------

// mockWriter is a thread-safe in-memory EventWriter used across test cases.
type mockWriter struct {
	mu     sync.Mutex
	events []Event
	err    error // if non-nil, WriteEvent returns this error
	delay  time.Duration
}

func (m *mockWriter) WriteEvent(ev Event) error {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, ev)
	return nil
}

func (m *mockWriter) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

func (m *mockWriter) all() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Event, len(m.events))
	copy(out, m.events)
	return out
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

var uuidV4Re = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

var runIDRe = regexp.MustCompile(`^r_\d{8}_\d{6}_[0-9a-f]{4}$`)

func mustNewEvent(t *testing.T, runID, eventType string, payload any) Event {
	t.Helper()
	ev, err := NewEvent(runID, eventType, payload)
	if err != nil {
		t.Fatalf("NewEvent() unexpected error: %v", err)
	}
	return ev
}

// captureStderr redirects os.Stderr for the duration of fn and returns what
// was written.  Not safe for parallel sub-tests that also redirect stderr.
func captureStderr(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stderr
	os.Stderr = w

	fn()

	_ = w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// ---------------------------------------------------------------------------
// newUUIDv4 tests
// ---------------------------------------------------------------------------

func TestNewUUIDv4_Format(t *testing.T) {
	for i := 0; i < 20; i++ {
		id, err := newUUIDv4()
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if !uuidV4Re.MatchString(id) {
			t.Errorf("iteration %d: %q does not match UUID v4 pattern", i, id)
		}
	}
}

func TestNewUUIDv4_Unique(t *testing.T) {
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		id, err := newUUIDv4()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[id] {
			t.Errorf("duplicate UUID generated: %s", id)
		}
		seen[id] = true
	}
}

// ---------------------------------------------------------------------------
// NewRunID tests
// ---------------------------------------------------------------------------

func TestNewRunID_Format(t *testing.T) {
	id, err := NewRunID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !runIDRe.MatchString(id) {
		t.Errorf("%q does not match r_YYYYMMDD_HHmmss_XXXX pattern", id)
	}
}

func TestNewRunID_Unique(t *testing.T) {
	// Two consecutive calls within the same second share the timestamp part
	// but differ in their random suffix.
	a, err := NewRunID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := NewRunID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == b {
		t.Errorf("expected different run IDs, got identical: %s", a)
	}
}

func TestNewRunID_DateIsUTC(t *testing.T) {
	before := time.Now().UTC()
	id, err := NewRunID()
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Extract date part r_20060102_...
	parts := strings.Split(id, "_")
	if len(parts) < 2 {
		t.Fatalf("unexpected run ID format: %s", id)
	}
	datePart := parts[1] // YYYYMMDD
	timePart := parts[2] // HHmmss
	embedded, parseErr := time.ParseInLocation("20060102_150405",
		datePart+"_"+timePart, time.UTC)
	if parseErr != nil {
		t.Fatalf("failed to parse embedded timestamp %s_%s: %v", datePart, timePart, parseErr)
	}
	// Allow ±1 second tolerance.
	if embedded.Before(before.Add(-time.Second)) || embedded.After(after.Add(time.Second)) {
		t.Errorf("embedded timestamp %v outside expected range [%v, %v]", embedded, before, after)
	}
}

// ---------------------------------------------------------------------------
// NewEvent tests
// ---------------------------------------------------------------------------

func TestNewEvent_SchemaVersion(t *testing.T) {
	ev := mustNewEvent(t, "r_test", EventRunStart, RunStartPayload{TaskID: "t1"})
	if ev.SchemaVersion != SchemaVersion {
		t.Errorf("SchemaVersion = %q, want %q", ev.SchemaVersion, SchemaVersion)
	}
	if SchemaVersion != "1" {
		t.Errorf("SchemaVersion constant = %q, want \"1\"", SchemaVersion)
	}
}

func TestNewEvent_EventIDIsUUIDv4(t *testing.T) {
	ev := mustNewEvent(t, "r_test", EventRunStart, RunStartPayload{})
	if !uuidV4Re.MatchString(ev.EventID) {
		t.Errorf("EventID %q is not a valid UUID v4", ev.EventID)
	}
}

func TestNewEvent_TimestampIsUTCAndRecent(t *testing.T) {
	before := time.Now().UTC()
	ev := mustNewEvent(t, "r_test", EventRunStart, RunStartPayload{})
	after := time.Now().UTC()

	if ev.Timestamp.Location() != time.UTC {
		t.Errorf("Timestamp location = %v, want UTC", ev.Timestamp.Location())
	}
	if ev.Timestamp.Before(before) || ev.Timestamp.After(after) {
		t.Errorf("Timestamp %v is not within [%v, %v]", ev.Timestamp, before, after)
	}
}

func TestNewEvent_RunIDAndTypePassedThrough(t *testing.T) {
	const wantRunID = "r_20240101_120000_abcd"
	const wantType = EventAgentStart
	ev := mustNewEvent(t, wantRunID, wantType, AgentStartPayload{AgentID: "a1"})
	if ev.RunID != wantRunID {
		t.Errorf("RunID = %q, want %q", ev.RunID, wantRunID)
	}
	if ev.EventType != wantType {
		t.Errorf("EventType = %q, want %q", ev.EventType, wantType)
	}
}

func TestNewEvent_PayloadMarshaledAsJSON(t *testing.T) {
	p := AgentEndPayload{
		AgentID:      "agent-1",
		Status:       "ok",
		DurationMs:   42,
		TokensInput:  100,
		TokensOutput: 200,
		TokensTotal:  300,
		FilesTouched: []string{"main.go"},
		ExitCode:     0,
		Output:       "some output",
	}
	ev := mustNewEvent(t, "r_x", EventAgentEnd, p)

	var got AgentEndPayload
	if err := json.Unmarshal(ev.Payload, &got); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	wantBytes, _ := json.Marshal(p)
	gotBytes, _ := json.Marshal(got)
	if string(wantBytes) != string(gotBytes) {
		t.Errorf("payload round-trip mismatch:\n  got  %s\n  want %s", gotBytes, wantBytes)
	}
}

func TestNewEvent_UnmarshalablePayloadReturnsError(t *testing.T) {
	// Channels cannot be marshaled to JSON.
	_, err := NewEvent("r_x", EventRunStart, make(chan int))
	if err == nil {
		t.Fatal("expected error for un-marshalable payload, got nil")
	}
	if !strings.Contains(err.Error(), "marshal payload") {
		t.Errorf("error message %q does not mention 'marshal payload'", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Payload round-trip tests (table-driven)
// ---------------------------------------------------------------------------

func TestPayloadRoundTrips(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		payload   any
		unmarshal func([]byte) (any, error)
	}{
		{
			name:      "RunStartPayload",
			eventType: EventRunStart,
			payload: RunStartPayload{
				TaskID:          "task-1",
				TaskDescription: "do something",
				Complexity:      "medium",
				AgentsPlanned:   []string{"pm", "dev"},
				Provider:        "openai",
				TriggeredBy:     "user",
			},
			unmarshal: func(b []byte) (any, error) {
				var v RunStartPayload
				return v, json.Unmarshal(b, &v)
			},
		},
		{
			name:      "RunEndPayload",
			eventType: EventRunEnd,
			payload: RunEndPayload{
				Status:          "success",
				DurationMs:      5000,
				TotalTokens:     1234,
				TotalFiles:      3,
				AgentsCompleted: 2,
				AgentsFailed:    0,
			},
			unmarshal: func(b []byte) (any, error) {
				var v RunEndPayload
				return v, json.Unmarshal(b, &v)
			},
		},
		{
			name:      "AgentStartPayload",
			eventType: EventAgentStart,
			payload: AgentStartPayload{
				AgentID:   "a-1",
				AgentRole: "developer",
				Sequence:  1,
				DependsOn: []string{"pm"},
				Model:     "gpt-4o",
			},
			unmarshal: func(b []byte) (any, error) {
				var v AgentStartPayload
				return v, json.Unmarshal(b, &v)
			},
		},
		{
			name:      "AgentEndPayload",
			eventType: EventAgentEnd,
			payload: AgentEndPayload{
				AgentID:      "a-1",
				Status:       "success",
				DurationMs:   1500,
				TokensInput:  400,
				TokensOutput: 600,
				TokensTotal:  1000,
				FilesTouched: []string{"pkg/foo.go"},
				ExitCode:     0,
				Output:       "some output",
			},
			unmarshal: func(b []byte) (any, error) {
				var v AgentEndPayload
				return v, json.Unmarshal(b, &v)
			},
		},
		{
			name:      "AgentErrorPayload",
			eventType: EventAgentError,
			payload: AgentErrorPayload{
				AgentID:    "a-2",
				Error:      "exit status 1",
				StderrTail: "panic: nil pointer",
				DurationMs: 200,
			},
			unmarshal: func(b []byte) (any, error) {
				var v AgentErrorPayload
				return v, json.Unmarshal(b, &v)
			},
		},
		{
			name:      "FileTouchedPayload",
			eventType: EventFileTouched,
			payload: FileTouchedPayload{
				AgentID:   "a-1",
				Path:      "internal/foo/bar.go",
				Operation: "create",
			},
			unmarshal: func(b []byte) (any, error) {
				var v FileTouchedPayload
				return v, json.Unmarshal(b, &v)
			},
		},
		{
			name:      "QAScorePayload",
			eventType: EventQAScore,
			payload: QAScorePayload{
				AgentID:  "qa",
				Score:    8.5,
				MaxScore: 10,
				Criteria: map[string]float64{"coverage": 4.5, "style": 4.0},
				Notes:    "good coverage",
			},
			unmarshal: func(b []byte) (any, error) {
				var v QAScorePayload
				return v, json.Unmarshal(b, &v)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev, err := NewEvent("r_test", tc.eventType, tc.payload)
			if err != nil {
				t.Fatalf("NewEvent() error: %v", err)
			}
			got, err := tc.unmarshal(ev.Payload)
			if err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			// Re-marshal both sides and compare JSON bytes to avoid reflect.DeepEqual
			// issues with nil vs empty slice, etc.
			wantBytes, _ := json.Marshal(tc.payload)
			gotBytes, _ := json.Marshal(got)
			if string(wantBytes) != string(gotBytes) {
				t.Errorf("round-trip mismatch:\n  got  %s\n  want %s", gotBytes, wantBytes)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// EventEmitter tests
// ---------------------------------------------------------------------------

func TestEventEmitter_SingleEventWritten(t *testing.T) {
	store := &mockWriter{}
	em := New(store)
	em.Start()

	ev := mustNewEvent(t, "r_x", EventRunStart, RunStartPayload{TaskID: "t1"})
	em.Emit(ev)
	em.Stop()

	if store.count() != 1 {
		t.Errorf("expected 1 event written, got %d", store.count())
	}
	if store.all()[0].EventID != ev.EventID {
		t.Errorf("written event ID mismatch")
	}
}

func TestEventEmitter_MultipleEventsAllWritten(t *testing.T) {
	const n = 50
	store := &mockWriter{}
	em := New(store)
	em.Start()

	for i := 0; i < n; i++ {
		ev := mustNewEvent(t, "r_x", EventAgentEnd,
			AgentEndPayload{AgentID: fmt.Sprintf("agent-%d", i)})
		em.Emit(ev)
	}
	em.Stop()

	if store.count() != n {
		t.Errorf("expected %d events, got %d", n, store.count())
	}
}

func TestEventEmitter_EmitIsNonBlocking(t *testing.T) {
	// Use a slow store to keep the worker goroutine busy draining slowly.
	store := &mockWriter{delay: 10 * time.Millisecond}
	em := New(store)
	em.Start()

	ev := mustNewEvent(t, "r_x", EventRunStart, RunStartPayload{})

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Emit many events; none should block the caller for longer than a short
		// timeout regardless of store speed (up to buffer capacity).
		for i := 0; i < 100; i++ {
			em.Emit(ev)
		}
	}()

	select {
	case <-done:
		// success: Emit returned promptly
	case <-time.After(500 * time.Millisecond):
		t.Error("Emit blocked the caller — expected non-blocking behavior")
	}

	em.Stop()
}

func TestEventEmitter_ChannelFull_DropsAndWarnsStderr(t *testing.T) {
	// Slow store + buffer overflow → some events must be dropped, stderr warning expected.
	store := &mockWriter{delay: 5 * time.Millisecond}
	em := New(store)
	em.Start()

	ev := mustNewEvent(t, "r_x", EventRunStart, RunStartPayload{})

	stderr := captureStderr(func() {
		// Flood beyond buffer capacity (512) while the worker is slow.
		for i := 0; i < channelBufferSize+50; i++ {
			em.Emit(ev)
		}
	})

	em.Stop()

	if !strings.Contains(stderr, "channel full") {
		t.Errorf("expected 'channel full' warning on stderr, got: %q", stderr)
	}
}

func TestEventEmitter_StopDrainsBufferedEvents(t *testing.T) {
	// A store with no delay — events accumulate in channel while worker is idle.
	store := &mockWriter{}
	em := New(store)
	em.Start()

	const n = 200
	for i := 0; i < n; i++ {
		ev := mustNewEvent(t, "r_x", EventAgentEnd,
			AgentEndPayload{AgentID: fmt.Sprintf("agent-%d", i)})
		em.Emit(ev)
	}
	em.Stop()

	if store.count() != n {
		t.Errorf("after Stop, expected %d events drained, got %d", n, store.count())
	}
}

func TestEventEmitter_StopBlocksUntilDrained(t *testing.T) {
	// Store with a small delay per event — Stop must not return before all
	// enqueued events are flushed.
	const n = 10
	store := &mockWriter{delay: 5 * time.Millisecond}
	em := New(store)
	em.Start()

	for i := 0; i < n; i++ {
		ev := mustNewEvent(t, "r_x", EventAgentStart, AgentStartPayload{AgentID: fmt.Sprintf("a%d", i)})
		em.Emit(ev)
	}

	stopDone := make(chan struct{})
	go func() {
		em.Stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
		// Stop returned
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not return within 2 seconds")
	}

	// After Stop, all events should be written.
	if store.count() != n {
		t.Errorf("expected %d events after Stop, got %d", n, store.count())
	}
}

func TestEventEmitter_StoreWriteError_ContinuesWithoutCrash(t *testing.T) {
	// Simulates "disco lleno" — store returns an error; emitter must log to
	// stderr and continue without panicking.
	store := &mockWriter{err: errors.New("no space left on device")}
	em := New(store)
	em.Start()

	ev := mustNewEvent(t, "r_x", EventFileTouched, FileTouchedPayload{
		AgentID: "a1", Path: "/tmp/foo", Operation: "write",
	})

	stderr := captureStderr(func() {
		em.Emit(ev)
		em.Stop()
	})

	// Emitter must still be alive (Stop returned) and must have logged the error.
	if !strings.Contains(stderr, "write event") {
		t.Errorf("expected 'write event' in stderr warning, got: %q", stderr)
	}
}

func TestEventEmitter_StoreWriteError_PipelineContinues(t *testing.T) {
	// After a write error the emitter must keep accepting and attempting to write
	// subsequent events (acceptance criterion 3 — pipeline continues).
	store := &errorAfterNWriter{n: 2}
	em := New(store)
	em.Start()

	for i := 0; i < 5; i++ {
		ev := mustNewEvent(t, "r_x", EventAgentEnd,
			AgentEndPayload{AgentID: fmt.Sprintf("a%d", i)})
		em.Emit(ev)
	}
	em.Stop()

	// At minimum, the emitter attempted all 5 events (some may have failed).
	if store.attempts < 5 {
		t.Errorf("expected 5 write attempts, got %d", store.attempts)
	}
}

// errorAfterNWriter returns an error for the first n writes, then succeeds.
type errorAfterNWriter struct {
	mu       sync.Mutex
	attempts int
	n        int
}

func (w *errorAfterNWriter) WriteEvent(_ Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.attempts++
	if w.attempts <= w.n {
		return errors.New("simulated write error")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Concurrency / race-detector tests
// ---------------------------------------------------------------------------

func TestEventEmitter_ConcurrentEmitters(t *testing.T) {
	store := &mockWriter{}
	em := New(store)
	em.Start()

	const goroutines = 10
	const eventsPerGoroutine = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {

		go func() {
			defer wg.Done()
			for i := 0; i < eventsPerGoroutine; i++ {
				ev := mustNewEvent(t, "r_race", EventAgentEnd,
					AgentEndPayload{AgentID: fmt.Sprintf("g%d-ev%d", g, i)})
				em.Emit(ev)
			}
		}()
	}

	wg.Wait()
	em.Stop()

	total := goroutines * eventsPerGoroutine
	if store.count() != total {
		t.Errorf("expected %d events from concurrent emitters, got %d", total, store.count())
	}
}

// ---------------------------------------------------------------------------
// Overhead / latency test (acceptance criterion 4 — ≤200 ms added latency)
// ---------------------------------------------------------------------------

func TestEventEmitter_Overhead(t *testing.T) {
	store := &mockWriter{}
	em := New(store)
	em.Start()

	const n = 1000
	start := time.Now()
	for i := 0; i < n; i++ {
		ev := mustNewEvent(t, "r_perf", EventAgentEnd,
			AgentEndPayload{AgentID: fmt.Sprintf("a%d", i)})
		em.Emit(ev)
	}
	elapsed := time.Since(start)

	em.Stop()

	// Emitting 1000 events must take well under 200 ms for the caller.
	if elapsed > 200*time.Millisecond {
		t.Errorf("Emit overhead too high: %v for %d events (want < 200ms)", elapsed, n)
	}
}

// ---------------------------------------------------------------------------
// NoopEmitter tests
// ---------------------------------------------------------------------------

func TestNoopEmitter_ImplementsEmitter(t *testing.T) {
	var _ Emitter = &NoopEmitter{}
}

func TestNoopEmitter_NoPanics(t *testing.T) {
	n := &NoopEmitter{}
	n.Start()

	ev := mustNewEvent(t, "r_noop", EventAgentEnd, AgentEndPayload{AgentID: "a1"})
	n.Emit(ev)
	n.Stop()
}

func TestNoopEmitter_EmitManyNoPanic(t *testing.T) {
	n := &NoopEmitter{}
	n.Start()
	for i := 0; i < 1000; i++ {
		ev := mustNewEvent(t, "r_noop", EventFileTouched,
			FileTouchedPayload{AgentID: "a", Path: "/tmp/x", Operation: "read"})
		n.Emit(ev)
	}
	n.Stop()
}

// ---------------------------------------------------------------------------
// Event type constant tests
// ---------------------------------------------------------------------------

func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"EventRunStart", EventRunStart, "run.start"},
		{"EventRunEnd", EventRunEnd, "run.end"},
		{"EventAgentStart", EventAgentStart, "agent.start"},
		{"EventAgentEnd", EventAgentEnd, "agent.end"},
		{"EventAgentError", EventAgentError, "agent.error"},
		{"EventFileTouched", EventFileTouched, "file.touched"},
		{"EventQAScore", EventQAScore, "qa.score"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value != tc.want {
				t.Errorf("%s = %q, want %q", tc.name, tc.value, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AgentError "failed" status scenario (acceptance criterion 2)
// ---------------------------------------------------------------------------

func TestAgentError_FailedStatus(t *testing.T) {
	store := &mockWriter{}
	em := New(store)
	em.Start()

	const errMsg = "timeout after 30s"
	ev, err := NewEvent("r_x", EventAgentError, AgentErrorPayload{
		AgentID:    "developer",
		Error:      errMsg,
		StderrTail: "context deadline exceeded",
		DurationMs: 30000,
	})
	if err != nil {
		t.Fatalf("NewEvent error: %v", err)
	}

	em.Emit(ev)
	em.Stop()

	events := store.all()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	var payload AgentErrorPayload
	if err := json.Unmarshal(events[0].Payload, &payload); err != nil {
		t.Fatalf("unmarshal AgentErrorPayload: %v", err)
	}
	if payload.Error != errMsg {
		t.Errorf("Error field = %q, want %q", payload.Error, errMsg)
	}
	if events[0].EventType != EventAgentError {
		t.Errorf("EventType = %q, want %q", events[0].EventType, EventAgentError)
	}
}

// ---------------------------------------------------------------------------
// Event JSON serialization — required fields (acceptance criterion 1)
// ---------------------------------------------------------------------------

func TestEvent_JSONSerialization_RequiredFields(t *testing.T) {
	runID, err := NewRunID()
	if err != nil {
		t.Fatalf("NewRunID: %v", err)
	}
	ev, err := NewEvent(runID, EventRunStart, RunStartPayload{
		TaskID:          "TASK-001",
		TaskDescription: "test json serialization",
		Complexity:      "small",
		Provider:        "anthropic",
		TriggeredBy:     "user",
	})
	if err != nil {
		t.Fatalf("NewEvent error: %v", err)
	}

	raw, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	requiredFields := []string{
		"schema_version",
		"event_id",
		"run_id",
		"event_type",
		"timestamp",
		"payload",
	}
	for _, field := range requiredFields {
		val, ok := m[field]
		if !ok {
			t.Errorf("required field %q missing from JSON", field)
			continue
		}
		switch v := val.(type) {
		case string:
			if v == "" {
				t.Errorf("required field %q is an empty string", field)
			}
		case nil:
			t.Errorf("required field %q is null", field)
		}
	}

	// Timestamp must be parseable as RFC3339 / ISO 8601.
	tsStr, _ := m["timestamp"].(string)
	if tsStr == "" {
		t.Fatal("timestamp field is missing or not a string")
	}
	ts, err := time.Parse(time.RFC3339Nano, tsStr)
	if err != nil {
		// Fallback: try plain RFC3339 without sub-seconds.
		ts, err = time.Parse(time.RFC3339, tsStr)
		if err != nil {
			t.Errorf("timestamp %q is not valid RFC3339: %v", tsStr, err)
		}
	}
	if !ts.IsZero() && ts.Year() < 2020 {
		t.Errorf("timestamp year %d looks implausible", ts.Year())
	}
}

// ---------------------------------------------------------------------------
// AgentEnd with "failed" status (acceptance criterion 2)
// ---------------------------------------------------------------------------

func TestAgentEnd_FailedStatus_IncludesErrorInfo(t *testing.T) {
	store := &mockWriter{}
	em := New(store)
	em.Start()

	const wantStatus = "failed"
	const wantOutput = "panic: runtime error: index out of range"

	ev, err := NewEvent("r_agentend", EventAgentEnd, AgentEndPayload{
		AgentID:    "developer",
		Status:     wantStatus,
		ExitCode:   1,
		Output:     wantOutput,
		DurationMs: 500,
	})
	if err != nil {
		t.Fatalf("NewEvent error: %v", err)
	}

	em.Emit(ev)
	em.Stop()

	events := store.all()
	if len(events) != 1 {
		t.Fatalf("expected 1 stored event, got %d", len(events))
	}

	stored := events[0]
	if stored.EventType != EventAgentEnd {
		t.Errorf("EventType = %q, want %q", stored.EventType, EventAgentEnd)
	}

	var payload AgentEndPayload
	if err := json.Unmarshal(stored.Payload, &payload); err != nil {
		t.Fatalf("unmarshal AgentEndPayload: %v", err)
	}
	if payload.Status != wantStatus {
		t.Errorf("Status = %q, want %q", payload.Status, wantStatus)
	}
	if payload.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", payload.ExitCode)
	}
	if !strings.Contains(payload.Output, "panic") {
		t.Errorf("Output = %q, want it to contain error info", payload.Output)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks (acceptance criterion 4: < 200 ms overhead)
// ---------------------------------------------------------------------------

// fastWriter is a no-op EventWriter with zero latency for benchmark use.
type fastWriter struct{}

func (f *fastWriter) WriteEvent(_ Event) error { return nil }

func BenchmarkNewEvent(b *testing.B) {
	runID, err := NewRunID()
	if err != nil {
		b.Fatalf("NewRunID: %v", err)
	}
	payload := AgentStartPayload{
		AgentID:   "developer",
		AgentRole: "developer",
		Model:     "claude-opus-4-5",
		Sequence:  1,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := NewEvent(runID, EventAgentStart, payload); err != nil {
			b.Fatalf("NewEvent: %v", err)
		}
	}
}

func BenchmarkEmit(b *testing.B) {
	em := New(&fastWriter{})
	em.Start()
	b.Cleanup(func() { em.Stop() })

	runID, err := NewRunID()
	if err != nil {
		b.Fatalf("NewRunID: %v", err)
	}
	ev, err := NewEvent(runID, EventAgentStart, AgentStartPayload{
		AgentID:   "developer",
		AgentRole: "developer",
		Model:     "claude-opus-4-5",
		Sequence:  1,
	})
	if err != nil {
		b.Fatalf("NewEvent: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.Emit(ev)
	}
}

func BenchmarkNewRunID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewRunID()
	}
}
