package orchestrator

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func Test_CLIGateHandler_RequestApproval(t *testing.T) {
	t.Parallel()

	node := Node{ID: "gate-node", Role: "reviewer"}

	tests := []struct {
		name         string
		input        string
		wantDecision GateDecision
		wantErr      bool
	}{
		{
			name:         "input approve → GateApproved",
			input:        "approve\n",
			wantDecision: GateApproved,
		},
		{
			name:         "input reject → GateRejected",
			input:        "reject\n",
			wantDecision: GateRejected,
		},
		{
			name:         "input skip → GateSkipped",
			input:        "skip\n",
			wantDecision: GateSkipped,
		},
		{
			name:         "uppercase input is normalised → GateApproved",
			input:        "APPROVE\n",
			wantDecision: GateApproved,
		},
		{
			name:         "whitespace-padded input is normalised → GateSkipped",
			input:        "  skip  \n",
			wantDecision: GateSkipped,
		},
		{
			name:         "invalid input then valid input → eventually approved",
			input:        "yes\napprove\n",
			wantDecision: GateApproved,
		},
		{
			name:    "EOF without valid answer → error",
			input:   "",
			wantErr: true,
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			in := strings.NewReader(tc.input)
			out := &bytes.Buffer{}
			h := NewCLIGateHandler(in, out)

			decision, err := h.RequestApproval(context.Background(), node)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if decision != tc.wantDecision {
				t.Errorf("RequestApproval() = %q, want %q", decision, tc.wantDecision)
			}
		})
	}
}

// blockingReader is an io.Reader whose Read method blocks until Close is called.
// This lets us test context cancellation without os.Pipe.
type blockingReader struct {
	closed chan struct{}
}

func newBlockingReader() *blockingReader {
	return &blockingReader{closed: make(chan struct{})}
}

func (r *blockingReader) Read(_ []byte) (int, error) {
	<-r.closed
	return 0, io.EOF
}

func (r *blockingReader) Close() {
	close(r.closed)
}

func Test_CLIGateHandler_ContextCancellation_BlockingReader(t *testing.T) {
	t.Parallel()

	br := newBlockingReader()
	out := &bytes.Buffer{}
	h := NewCLIGateHandler(br, out)
	node := Node{ID: "blocked", Role: "tester"}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		_, err := h.RequestApproval(ctx, node)
		done <- err
	}()

	cancel()
	br.Close()

	err := <-done
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}
