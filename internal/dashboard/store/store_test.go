package store_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/store"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// --- helpers ----------------------------------------------------------------

func migrationsPath(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", "..", "migrations"))
	if err != nil {
		t.Fatalf("resolver ruta de migraciones: %v", err)
	}
	return abs
}

func newTestStore(t *testing.T, maxRuns int) *store.TestableStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.New(dbPath, migrationsPath(t), maxRuns)
	if err != nil {
		t.Fatalf("crear test store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return store.NewTestableStore(s)
}

func mustNewEvent(t *testing.T, runID, eventType string, payload any) instrumentation.Event {
	t.Helper()
	ev, err := instrumentation.NewEvent(runID, eventType, payload)
	if err != nil {
		t.Fatalf("crear evento %s: %v", eventType, err)
	}
	return ev
}

func mustNewRunID(t *testing.T) string {
	t.Helper()
	id, err := instrumentation.NewRunID()
	if err != nil {
		t.Fatalf("crear run ID: %v", err)
	}
	return id
}

func writeRunStart(t *testing.T, s *store.TestableStore, runID string) {
	t.Helper()
	ev := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:          "task-001",
		TaskDescription: "test task",
		Complexity:      "medium",
		Provider:        "anthropic",
		AgentsPlanned:   []string{"developer", "tester"},
		TriggeredBy:     "user",
	})
	if err := s.WriteEvent(ev); err != nil {
		t.Fatalf("escribir run.start: %v", err)
	}
}

func writeAgentStart(t *testing.T, s *store.TestableStore, runID, agentID string) {
	t.Helper()
	ev := mustNewEvent(t, runID, instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID:   agentID,
		AgentRole: "developer",
		Sequence:  1,
		DependsOn: []string{},
		Model:     "claude-sonnet-4-6",
	})
	if err := s.WriteEvent(ev); err != nil {
		t.Fatalf("escribir agent.start: %v", err)
	}
}

// --- tests ------------------------------------------------------------------

func Test_New(t *testing.T) {
	t.Run("creates all tables", func(t *testing.T) {
		s := newTestStore(t, 5)

		tables := []string{"runs", "agents", "files", "events"}
		for _, table := range tables {
			var name string
			err := s.DB().QueryRow(
				"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
			).Scan(&name)
			if err != nil {
				t.Errorf("tabla %q no encontrada en sqlite_master: %v", table, err)
			}
		}
	})

	t.Run("defaults maxRuns to 500 when zero", func(t *testing.T) {
		s := newTestStore(t, 0)
		if got := s.MaxRuns(); got != 500 {
			t.Errorf("maxRuns: esperado 500, obtuvo %d", got)
		}
	})
}

func Test_WriteEvent(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, s *store.TestableStore, runID string)
		event     func(t *testing.T, runID string) instrumentation.Event
		verify    func(t *testing.T, db *sql.DB, runID string)
	}{
		{
			name: "run.start inserts into runs",
			event: func(t *testing.T, runID string) instrumentation.Event {
				return mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
					TaskID:          "task-001",
					TaskDescription: "test task",
					Complexity:      "medium",
					Provider:        "anthropic",
					AgentsPlanned:   []string{"developer"},
					TriggeredBy:     "user",
				})
			},
			verify: func(t *testing.T, db *sql.DB, runID string) {
				var status, taskID string
				err := db.QueryRow("SELECT status, task_id FROM runs WHERE id = ?", runID).Scan(&status, &taskID)
				if err != nil {
					t.Fatalf("consultar runs: %v", err)
				}
				if status != "running" {
					t.Errorf("status: esperado %q, obtuvo %q", "running", status)
				}
				if taskID != "task-001" {
					t.Errorf("task_id: esperado %q, obtuvo %q", "task-001", taskID)
				}
			},
		},
		{
			name: "run.end updates runs",
			setup: func(t *testing.T, s *store.TestableStore, runID string) {
				writeRunStart(t, s, runID)
			},
			event: func(t *testing.T, runID string) instrumentation.Event {
				return mustNewEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
					Status:          "success",
					DurationMs:      1500,
					TotalTokens:     800,
					TotalFiles:      3,
					AgentsCompleted: 2,
					AgentsFailed:    0,
				})
			},
			verify: func(t *testing.T, db *sql.DB, runID string) {
				var status string
				var durationMs int64
				var totalTokens, agentsCount int
				err := db.QueryRow(
					"SELECT status, duration_ms, total_tokens, agents_count FROM runs WHERE id = ?", runID,
				).Scan(&status, &durationMs, &totalTokens, &agentsCount)
				if err != nil {
					t.Fatalf("consultar runs: %v", err)
				}
				if status != "success" {
					t.Errorf("status: esperado %q, obtuvo %q", "success", status)
				}
				if durationMs != 1500 {
					t.Errorf("duration_ms: esperado 1500, obtuvo %d", durationMs)
				}
				if totalTokens != 800 {
					t.Errorf("total_tokens: esperado 800, obtuvo %d", totalTokens)
				}
				if agentsCount != 2 {
					t.Errorf("agents_count: esperado 2, obtuvo %d", agentsCount)
				}
			},
		},
		{
			name: "agent.start inserts into agents",
			setup: func(t *testing.T, s *store.TestableStore, runID string) {
				writeRunStart(t, s, runID)
			},
			event: func(t *testing.T, runID string) instrumentation.Event {
				return mustNewEvent(t, runID, instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
					AgentID:   "agent-abc",
					AgentRole: "developer",
					Sequence:  1,
					DependsOn: []string{},
					Model:     "claude-sonnet-4-6",
				})
			},
			verify: func(t *testing.T, db *sql.DB, runID string) {
				var role, status string
				err := db.QueryRow(
					"SELECT agent_role, status FROM agents WHERE run_id = ? AND agent_id = ?", runID, "agent-abc",
				).Scan(&role, &status)
				if err != nil {
					t.Fatalf("consultar agents: %v", err)
				}
				if role != "developer" {
					t.Errorf("agent_role: esperado %q, obtuvo %q", "developer", role)
				}
				if status != "running" {
					t.Errorf("status: esperado %q, obtuvo %q", "running", status)
				}
			},
		},
		{
			name: "agent.end updates agents and inserts files",
			setup: func(t *testing.T, s *store.TestableStore, runID string) {
				writeRunStart(t, s, runID)
				writeAgentStart(t, s, runID, "agent-xyz")
			},
			event: func(t *testing.T, runID string) instrumentation.Event {
				return mustNewEvent(t, runID, instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
					AgentID:      "agent-xyz",
					Status:       "success",
					DurationMs:   500,
					TokensInput:  100,
					TokensOutput: 200,
					TokensTotal:  300,
					FilesTouched: []string{"/src/main.go", "/src/util.go"},
					ExitCode:     0,
				})
			},
			verify: func(t *testing.T, db *sql.DB, runID string) {
				var status string
				var tokensTotal int
				err := db.QueryRow(
					"SELECT status, tokens_total FROM agents WHERE run_id = ? AND agent_id = ?", runID, "agent-xyz",
				).Scan(&status, &tokensTotal)
				if err != nil {
					t.Fatalf("consultar agents: %v", err)
				}
				if status != "success" {
					t.Errorf("status: esperado %q, obtuvo %q", "success", status)
				}
				if tokensTotal != 300 {
					t.Errorf("tokens_total: esperado 300, obtuvo %d", tokensTotal)
				}

				var fileCount int
				if err := db.QueryRow("SELECT COUNT(*) FROM files WHERE run_id = ? AND agent_id = ?", runID, "agent-xyz").Scan(&fileCount); err != nil {
					t.Fatalf("contar files: %v", err)
				}
				if fileCount != 2 {
					t.Errorf("files: esperado 2, obtuvo %d", fileCount)
				}
			},
		},
		{
			name: "agent.error marks agent as failed",
			setup: func(t *testing.T, s *store.TestableStore, runID string) {
				writeRunStart(t, s, runID)
				writeAgentStart(t, s, runID, "agent-err")
			},
			event: func(t *testing.T, runID string) instrumentation.Event {
				return mustNewEvent(t, runID, instrumentation.EventAgentError, instrumentation.AgentErrorPayload{
					AgentID:    "agent-err",
					Error:      "timeout after 30s",
					StderrTail: "fatal: process killed",
					DurationMs: 30000,
				})
			},
			verify: func(t *testing.T, db *sql.DB, runID string) {
				var status, errorMsg string
				err := db.QueryRow(
					"SELECT status, error_msg FROM agents WHERE run_id = ? AND agent_id = ?", runID, "agent-err",
				).Scan(&status, &errorMsg)
				if err != nil {
					t.Fatalf("consultar agents: %v", err)
				}
				if status != "failed" {
					t.Errorf("status: esperado %q, obtuvo %q", "failed", status)
				}
				if errorMsg != "timeout after 30s" {
					t.Errorf("error_msg: esperado %q, obtuvo %q", "timeout after 30s", errorMsg)
				}
			},
		},
		{
			name: "file.touched inserts into files",
			setup: func(t *testing.T, s *store.TestableStore, runID string) {
				writeRunStart(t, s, runID)
			},
			event: func(t *testing.T, runID string) instrumentation.Event {
				return mustNewEvent(t, runID, instrumentation.EventFileTouched, instrumentation.FileTouchedPayload{
					AgentID:   "agent-file",
					Path:      "/internal/api/handler.go",
					Operation: "write",
				})
			},
			verify: func(t *testing.T, db *sql.DB, runID string) {
				var path, operation string
				err := db.QueryRow("SELECT path, operation FROM files WHERE run_id = ?", runID).Scan(&path, &operation)
				if err != nil {
					t.Fatalf("consultar files: %v", err)
				}
				if path != "/internal/api/handler.go" {
					t.Errorf("path: esperado %q, obtuvo %q", "/internal/api/handler.go", path)
				}
				if operation != "write" {
					t.Errorf("operation: esperado %q, obtuvo %q", "write", operation)
				}
			},
		},
		{
			name: "qa.score updates runs",
			setup: func(t *testing.T, s *store.TestableStore, runID string) {
				writeRunStart(t, s, runID)
			},
			event: func(t *testing.T, runID string) instrumentation.Event {
				return mustNewEvent(t, runID, instrumentation.EventQAScore, instrumentation.QAScorePayload{
					AgentID:  "agent-qa",
					Score:    8.5,
					MaxScore: 10,
					Criteria: map[string]float64{"coverage": 9.0, "style": 8.0},
				})
			},
			verify: func(t *testing.T, db *sql.DB, runID string) {
				var qaScore float64
				err := db.QueryRow("SELECT qa_score FROM runs WHERE id = ?", runID).Scan(&qaScore)
				if err != nil {
					t.Fatalf("consultar qa_score: %v", err)
				}
				if qaScore != 8.5 {
					t.Errorf("qa_score: esperado 8.5, obtuvo %f", qaScore)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStore(t, 500)
			runID := mustNewRunID(t)

			if tt.setup != nil {
				tt.setup(t, s, runID)
			}

			ev := tt.event(t, runID)
			if err := s.WriteEvent(ev); err != nil {
				t.Fatalf("WriteEvent: %v", err)
			}

			tt.verify(t, s.DB(), runID)

			// Every event must also appear in the raw audit trail.
			var count int
			if err := s.DB().QueryRow("SELECT COUNT(*) FROM events WHERE event_id = ?", ev.EventID).Scan(&count); err != nil {
				t.Fatalf("consultar events: %v", err)
			}
			if count != 1 {
				t.Errorf("audit trail: esperado 1 row para event_id %q, obtuvo %d", ev.EventID, count)
			}
		})
	}
}

func Test_WriteEvent_DuplicateEventID(t *testing.T) {
	s := newTestStore(t, 500)
	runID := mustNewRunID(t)

	ev := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:      "dup-test",
		Complexity:  "trivial",
		Provider:    "test",
		TriggeredBy: "test",
	})

	if err := s.WriteEvent(ev); err != nil {
		t.Fatalf("primer WriteEvent: %v", err)
	}

	err := s.WriteEvent(ev)
	if err == nil {
		t.Fatal("esperaba error al escribir evento duplicado, obtuvo nil")
	}
	if !strings.Contains(err.Error(), "UNIQUE constraint") {
		t.Errorf("esperaba UNIQUE constraint error, obtuvo: %v", err)
	}
}

func Test_WriteEvent_RunEndOrphan(t *testing.T) {
	s := newTestStore(t, 500)
	runID := mustNewRunID(t)

	ev := mustNewEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
		Status:     "success",
		DurationMs: 1000,
	})

	err := s.WriteEvent(ev)
	if err == nil {
		t.Fatal("esperaba error al escribir run.end sin run.start previo, obtuvo nil")
	}
	if !strings.Contains(err.Error(), "no encontrado") {
		t.Errorf("esperaba error 'no encontrado', obtuvo: %v", err)
	}
}

func Test_Rotate(t *testing.T) {
	t.Run("deletes oldest runs beyond maxRuns", func(t *testing.T) {
		s := newTestStore(t, 5)

		var runIDs []string
		for i := 0; i < 7; i++ {
			runID := mustNewRunID(t)
			runIDs = append(runIDs, runID)

			startEv := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
				TaskID:   "task-rotate",
				Provider: "anthropic",
			})
			startEv.Timestamp = time.Date(2024, 1, 1, 0, 0, i, 0, time.UTC)
			if err := s.WriteEvent(startEv); err != nil {
				t.Fatalf("run.start [%d]: %v", i, err)
			}

			endEv := mustNewEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
				Status:     "success",
				DurationMs: 100,
			})
			endEv.Timestamp = time.Date(2024, 1, 1, 0, 0, i, 500000000, time.UTC)
			if err := s.WriteEvent(endEv); err != nil {
				t.Fatalf("run.end [%d]: %v", i, err)
			}
		}

		var count int
		if err := s.DB().QueryRow("SELECT COUNT(*) FROM runs").Scan(&count); err != nil {
			t.Fatalf("contar runs: %v", err)
		}
		if count != 5 {
			t.Errorf("runs después de rotación: esperado 5, obtuvo %d", count)
		}

		for _, id := range runIDs[:2] {
			var n int
			_ = s.DB().QueryRow("SELECT COUNT(*) FROM runs WHERE id = ?", id).Scan(&n)
			if n != 0 {
				t.Errorf("run %q debería haber sido eliminado", id)
			}
		}
	})

	t.Run("cascades to agents files and events", func(t *testing.T) {
		s := newTestStore(t, 2)

		var runIDs []string
		for i := 0; i < 3; i++ {
			runID := mustNewRunID(t)
			runIDs = append(runIDs, runID)

			startEv := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
				TaskID:   "task-cascade",
				Provider: "anthropic",
			})
			startEv.Timestamp = time.Date(2024, 1, 1, 0, 0, i, 0, time.UTC)
			if err := s.WriteEvent(startEv); err != nil {
				t.Fatalf("run.start [%d]: %v", i, err)
			}

			agentEv := mustNewEvent(t, runID, instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
				AgentID:   "agent-cascade",
				AgentRole: "developer",
				Sequence:  1,
				DependsOn: []string{},
				Model:     "claude-sonnet-4-6",
			})
			if err := s.WriteEvent(agentEv); err != nil {
				t.Fatalf("agent.start [%d]: %v", i, err)
			}

			agentEndEv := mustNewEvent(t, runID, instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
				AgentID:      "agent-cascade",
				Status:       "success",
				DurationMs:   200,
				FilesTouched: []string{"/src/file.go"},
			})
			if err := s.WriteEvent(agentEndEv); err != nil {
				t.Fatalf("agent.end [%d]: %v", i, err)
			}

			endEv := mustNewEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
				Status:          "success",
				DurationMs:      300,
				AgentsCompleted: 1,
			})
			endEv.Timestamp = time.Date(2024, 1, 1, 0, 0, i, 300000000, time.UTC)
			if err := s.WriteEvent(endEv); err != nil {
				t.Fatalf("run.end [%d]: %v", i, err)
			}
		}

		deleted := runIDs[0]
		for _, table := range []string{"agents", "files", "events"} {
			var n int
			if err := s.DB().QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE run_id = ?", table), deleted).Scan(&n); err != nil {
				t.Fatalf("contar %s: %v", table, err)
			}
			if n != 0 {
				t.Errorf("%s del run eliminado: esperado 0, obtuvo %d", table, n)
			}
		}
	})
}

func Test_Performance_1000Runs(t *testing.T) {
	if testing.Short() {
		t.Skip("omitiendo test de performance en modo short")
	}

	s := newTestStore(t, 2000)

	start := time.Now()
	for i := 0; i < 1000; i++ {
		runID := fmt.Sprintf("r_perf_%04d", i)

		startEv := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
			TaskID:   "task-perf",
			Provider: "anthropic",
		})
		if err := s.WriteEvent(startEv); err != nil {
			t.Fatalf("run.start [%d]: %v", i, err)
		}

		endEv := mustNewEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
			Status:     "success",
			DurationMs: 50,
		})
		if err := s.WriteEvent(endEv); err != nil {
			t.Fatalf("run.end [%d]: %v", i, err)
		}
	}
	t.Logf("insertar 1000 runs: %v", time.Since(start))

	queryStart := time.Now()
	var count int
	if err := s.DB().QueryRow("SELECT COUNT(*) FROM runs").Scan(&count); err != nil {
		t.Fatalf("SELECT COUNT(*): %v", err)
	}
	queryDuration := time.Since(queryStart)

	if count != 1000 {
		t.Errorf("COUNT(*): esperado 1000, obtuvo %d", count)
	}
	if queryDuration > 500*time.Millisecond {
		t.Errorf("query tardó %v, debe ser < 500ms", queryDuration)
	}
}

func Test_GetRunSummary(t *testing.T) {
	t.Run("run existente retorna el summary correcto", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)
		writeRunStart(t, s, runID)

		got, err := s.GetRunSummary(context.Background(), runID)
		if err != nil {
			t.Fatalf("GetRunSummary: %v", err)
		}
		if got == nil {
			t.Fatal("GetRunSummary: esperaba RunSummary no nil")
		}
		if got.ID != runID {
			t.Errorf("ID: esperado %q, obtuvo %q", runID, got.ID)
		}
		if got.Status != "running" {
			t.Errorf("Status: esperado %q, obtuvo %q", "running", got.Status)
		}
	})

	t.Run("run inexistente retorna (nil, nil)", func(t *testing.T) {
		s := newTestStore(t, 500)

		got, err := s.GetRunSummary(context.Background(), "no-existe")
		if err != nil {
			t.Fatalf("GetRunSummary: esperaba nil error, obtuvo: %v", err)
		}
		if got != nil {
			t.Errorf("GetRunSummary: esperaba nil, obtuvo %+v", got)
		}
	})

	t.Run("run con qa_score retorna QAScore correcto", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)
		writeRunStart(t, s, runID)
		ev := mustNewEvent(t, runID, instrumentation.EventQAScore, instrumentation.QAScorePayload{
			AgentID:  "agent-qa",
			Score:    9.5,
			MaxScore: 10,
		})
		if err := s.WriteEvent(ev); err != nil {
			t.Fatalf("WriteEvent qa.score: %v", err)
		}

		got, err := s.GetRunSummary(context.Background(), runID)
		if err != nil {
			t.Fatalf("GetRunSummary: %v", err)
		}
		if got == nil {
			t.Fatal("GetRunSummary: esperaba RunSummary no nil")
		}
		if got.QAScore == nil {
			t.Fatal("QAScore: esperaba valor no nil")
		}
		if *got.QAScore != 9.5 {
			t.Errorf("QAScore: esperado 9.5, obtuvo %f", *got.QAScore)
		}
	})
}
