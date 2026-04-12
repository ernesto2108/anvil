package orchestrator

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

// fprintfIgnoreErr writes to w and intentionally discards the error; prompts
// to a writer are best-effort (e.g. a closed terminal should not abort the run).
func fprintfIgnoreErr(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

// GateHandler is implemented by anything that can request human (or automated)
// approval before a gated node runs.
type GateHandler interface {
	RequestApproval(ctx context.Context, node Node) (GateDecision, error)
}

// CLIGateHandler implements GateHandler by prompting on a writer and reading
// the response from a reader. Using io.Reader/io.Writer makes it testable
// without touching os.Stdin/os.Stdout.
type CLIGateHandler struct {
	in  io.Reader
	out io.Writer
}

// NewCLIGateHandler creates a CLIGateHandler that reads from in and writes to out.
func NewCLIGateHandler(in io.Reader, out io.Writer) *CLIGateHandler {
	return &CLIGateHandler{in: in, out: out}
}

// RequestApproval prints a prompt and waits for the user to type one of
// "approve", "reject", or "skip". It respects ctx.Done() so it can be
// cancelled by the orchestrator if the run is aborted.
func (h *CLIGateHandler) RequestApproval(ctx context.Context, node Node) (GateDecision, error) {
	fprintfIgnoreErr(h.out, "\n[GATE] Node %q (role: %s) requires approval.\n", node.ID, node.Role)
	fprintfIgnoreErr(h.out, "  Enter decision [approve/reject/skip]: ")

	type result struct {
		decision GateDecision
		err      error
	}

	ch := make(chan result, 1)

	go func() {
		scanner := bufio.NewScanner(h.in)
		for scanner.Scan() {
			line := strings.TrimSpace(strings.ToLower(scanner.Text()))
			switch line {
			case "approve":
				ch <- result{GateApproved, nil}
				return
			case "reject":
				ch <- result{GateRejected, nil}
				return
			case "skip":
				ch <- result{GateSkipped, nil}
				return
			default:
				fprintfIgnoreErr(h.out, "  Unknown input %q. Enter [approve/reject/skip]: ", line)
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- result{"", fmt.Errorf("orchestrator/gate: read input: %w", err)}
			return
		}
		// EOF reached without a valid answer.
		ch <- result{"", fmt.Errorf("orchestrator/gate: input closed before decision for node %q", node.ID)}
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("orchestrator/gate: context cancelled waiting for approval of node %q: %w", node.ID, ctx.Err())
	case r := <-ch:
		return r.decision, r.err
	}
}

// AutoApproveHandler is a GateHandler that approves every gate automatically.
// Useful for non-interactive environments and testing.
type AutoApproveHandler struct {
	out io.Writer
}

// NewAutoApproveHandler creates an AutoApproveHandler that logs approvals to out.
func NewAutoApproveHandler(out io.Writer) *AutoApproveHandler {
	return &AutoApproveHandler{out: out}
}

// RequestApproval always returns GateApproved.
func (h *AutoApproveHandler) RequestApproval(_ context.Context, node Node) (GateDecision, error) {
	fprintfIgnoreErr(h.out, "[GATE] Node %q (role: %s) — auto-approved\n", node.ID, node.Role)
	return GateApproved, nil
}
