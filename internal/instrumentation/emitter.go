package instrumentation

import (
	"fmt"
	"os"
	"sync"
)

const channelBufferSize = 512

// EventEmitter buffers events and writes them asynchronously via an EventWriter.
type EventEmitter struct {
	ch    chan Event
	store EventWriter
	done  chan struct{}
	wg    sync.WaitGroup
}

// New creates a new EventEmitter backed by the given EventWriter.
func New(s EventWriter) *EventEmitter {
	return &EventEmitter{
		ch:    make(chan Event, channelBufferSize),
		store: s,
		done:  make(chan struct{}),
	}
}

// Start launches the background writer goroutine. Call once before Emit.
func (e *EventEmitter) Start() {
	e.wg.Add(1)
	go e.loop()
}

// loop reads events from the channel and writes them to the store.
// It drains remaining events after done is closed before returning.
func (e *EventEmitter) loop() {
	defer e.wg.Done()
	for {
		select {
		case ev := <-e.ch:
			e.write(ev)
		case <-e.done:
			// Drain any buffered events before exiting.
			for {
				select {
				case ev := <-e.ch:
					e.write(ev)
				default:
					return
				}
			}
		}
	}
}

// write calls the store and logs errors to stderr without panicking.
func (e *EventEmitter) write(ev Event) {
	if err := e.store.WriteEvent(ev); err != nil {
		fmt.Fprintf(os.Stderr, "instrumentation: write event %s: %v\n", ev.EventID, err)
	}
}

// Emit sends an event to the buffer. Non-blocking: if the buffer is full the
// event is dropped and a warning is written to stderr.
func (e *EventEmitter) Emit(ev Event) {
	select {
	case e.ch <- ev:
	default:
		fmt.Fprintf(os.Stderr,
			"instrumentation: channel full, dropping event %s (type=%s)\n",
			ev.EventID, ev.EventType,
		)
	}
}

// Stop signals the writer goroutine to stop, waits for it to drain remaining
// events, and blocks until it exits cleanly.
func (e *EventEmitter) Stop() {
	close(e.done)
	e.wg.Wait()
}
