// Package mcp implements an MCP (Model Context Protocol) server over stdio.
// The server exposes Anvil's internals — agents, skills, pipelines, runs,
// memories, and project context — as JSON-RPC 2.0 tools that any MCP-capable
// AI tool (Claude Code, Cursor, Gemini CLI) can invoke.
package mcp

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ernesto2108/anvil/pkg/config"
)

// request is a JSON-RPC 2.0 request (method call or notification).
// Notifications have no ID field (omitted or null).
type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// response is a JSON-RPC 2.0 response.
type response struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Result  any              `json:"result,omitempty"`
	Error   *rpcError        `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// callParams holds the params for a tools/call request.
type callParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Server holds runtime state shared across all tool handlers.
type Server struct {
	cfg      *config.App
	db       *sql.DB
	vaultDir string // path to Obsidian/knowledge-base vault (optional)
}

// Serve runs the MCP server, reading JSON-RPC requests from stdin and writing
// responses to stdout. It blocks until ctx is cancelled or stdin is closed.
func Serve(ctx context.Context, cfg *config.App, db *sql.DB) error {
	vaultDir := os.Getenv("ANVIL_VAULT_PATH")

	srv := &Server{cfg: cfg, db: db, vaultDir: vaultDir}
	tools := srv.buildRegistry()

	in := bufio.NewReader(os.Stdin)
	enc := json.NewEncoder(os.Stdout)

	send := func(r response) {
		_ = enc.Encode(r)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := in.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("mcp: read stdin: %w", err)
		}
		if len(line) == 0 {
			continue
		}

		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			// Malformed JSON — send parse error without an ID.
			send(response{
				JSONRPC: "2.0",
				Error:   &rpcError{Code: -32700, Message: "parse error"},
			})
			continue
		}

		// Notifications (no ID) — don't respond.
		if req.ID == nil {
			continue
		}

		var result any
		var rpcErr *rpcError

		switch req.Method {
		case "initialize":
			result = map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]any{"tools": map[string]any{}},
				"serverInfo":      map[string]any{"name": "anvil", "version": "1.0"},
			}

		case "tools/list":
			defs := make([]toolDef, 0, len(tools))
			for _, t := range tools {
				defs = append(defs, t.def)
			}
			result = map[string]any{"tools": defs}

		case "tools/call":
			var p callParams
			if err := json.Unmarshal(req.Params, &p); err != nil {
				rpcErr = &rpcError{Code: -32602, Message: "invalid params"}
				break
			}
			t, ok := tools[p.Name]
			if !ok {
				rpcErr = &rpcError{Code: -32601, Message: fmt.Sprintf("tool not found: %s", p.Name)}
				break
			}
			text, callErr := t.handler(ctx, p.Arguments)
			if callErr != nil {
				rpcErr = &rpcError{Code: -32603, Message: callErr.Error()}
			} else {
				result = map[string]any{
					"content": []map[string]any{
						{"type": "text", "text": text},
					},
				}
			}

		default:
			rpcErr = &rpcError{Code: -32601, Message: fmt.Sprintf("method not found: %s", req.Method)}
		}

		send(response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
			Error:   rpcErr,
		})
	}
}
