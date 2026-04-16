package ollama

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// EnsureRunning checks if Ollama is reachable. If not, starts it in the background
// and waits until it responds. Sets OLLAMA_KEEP_ALIVE so models unload after idle.
// Returns nil if Ollama is ready, or an error if it couldn't be started.
func EnsureRunning(ctx context.Context, keepAlive string) error {
	if isHealthy(ctx) {
		return nil
	}

	if keepAlive == "" {
		keepAlive = "5m"
	}

	cmd := exec.Command("ollama", "serve")
	cmd.Env = append(os.Environ(), "OLLAMA_KEEP_ALIVE="+keepAlive)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ollama: start serve: %w", err)
	}

	// Detach — let the process outlive Anvil.
	go func() { _ = cmd.Wait() }()

	// Poll until ready or context cancelled.
	deadline := time.After(15 * time.Second)
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("ollama: timed out waiting for serve to start")
		case <-tick.C:
			if isHealthy(ctx) {
				return nil
			}
		}
	}
}

func isHealthy(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:11434", nil)
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
