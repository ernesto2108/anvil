//go:build dashboard

package store

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Metrics es el agregado retornado por GetMetrics al paquete dashboard.
// Usa tipos Go nativos (no nullables de database/sql) para que los conversores
// de dtos.go operen sin traducción extra.
type Metrics struct {
	RunsCount     int
	HasEnoughData bool

	TotalTokens   int
	AvgDurationMs int64
	SuccessRate   float64

	// Deltas nil-able: nil == sin comparación disponible.
	TotalTokensDeltaPct *float64
	AvgDurationDeltaMs  *int64
	SuccessRateDelta    *float64

	TokensPerRun    []TokensPerRunRow // orden cronológico ASC
	AvgTimePerAgent []AgentTimeRow    // orden DESC por AvgDurationMs
}

// TokensPerRunRow es una fila del chart "Tokens per run".
type TokensPerRunRow struct {
	RunID     string
	StartedAt time.Time
	Tokens    int
}

// AgentTimeRow es una fila del chart "Avg time per agent".
type AgentTimeRow struct {
	Role          string
	AvgDurationMs int64
	RunsIncluded  int
}

// runRecord es un registro interno para el cálculo de métricas.
type runRecord struct {
	id          string
	startedAt   time.Time
	durationMs  int64
	totalTokens int
	status      string
}

// GetMetrics calcula los agregados sobre los últimos 30 runs terminados
// (status IN ('success','failed')). Usa runs 31-60 para los deltas "vs anterior".
// Si no hay runs terminados, retorna Metrics{RunsCount:0, HasEnoughData:false,
// TokensPerRun:[], AvgTimePerAgent:[]} sin error.
func (s *SQLiteStore) GetMetrics(ctx context.Context) (Metrics, error) {
	empty := Metrics{
		TokensPerRun:    []TokensPerRunRow{},
		AvgTimePerAgent: []AgentTimeRow{},
	}

	// Q_CURRENT: los últimos 30 runs terminados ordenados DESC.
	const qCurrent = `
		SELECT id, started_at, duration_ms, total_tokens, status
		FROM runs
		WHERE status IN ('success', 'failed')
		ORDER BY started_at DESC
		LIMIT 30`

	currentRows, err := s.db.QueryContext(ctx, qCurrent)
	if err != nil {
		return empty, fmt.Errorf("dashboard/store: consultar runs actuales: %w", err)
	}
	defer currentRows.Close() //nolint:errcheck

	current, err := scanRunRecords(currentRows)
	if err != nil {
		return empty, fmt.Errorf("dashboard/store: escanear runs actuales: %w", err)
	}

	if len(current) == 0 {
		return empty, nil
	}

	// Q_PREVIOUS: los 30 runs terminados anteriores al más antiguo del set actual.
	oldestStartedAt := current[len(current)-1].startedAt.Format(time.RFC3339Nano)

	const qPrevious = `
		SELECT id, started_at, duration_ms, total_tokens, status
		FROM runs
		WHERE status IN ('success', 'failed')
		  AND started_at < ?
		ORDER BY started_at DESC
		LIMIT 30`

	prevRows, err := s.db.QueryContext(ctx, qPrevious, oldestStartedAt)
	if err != nil {
		return empty, fmt.Errorf("dashboard/store: consultar runs anteriores: %w", err)
	}
	defer prevRows.Close() //nolint:errcheck

	previous, err := scanRunRecords(prevRows)
	if err != nil {
		return empty, fmt.Errorf("dashboard/store: escanear runs anteriores: %w", err)
	}

	// Calcular métricas del período actual.
	curStats := calcStats(current)

	// Calcular métricas del período anterior (para deltas).
	var prevStats *periodStats
	if len(previous) > 0 {
		ps := calcStats(previous)
		prevStats = &ps
	}

	// Calcular deltas.
	totalTokensDeltaPct := calcDeltaPct(curStats.totalTokens, prevStats)
	avgDurationDeltaMs := calcDeltaDurationMs(curStats.avgDurationMs, prevStats)
	successRateDelta := calcDeltaSuccessRate(curStats.successRate, prevStats)

	// Q_AVG_PER_AGENT: promedio duration_ms por agent_role sobre agents del set actual.
	avgTimePerAgent, err := s.queryAvgTimePerAgent(ctx, current)
	if err != nil {
		return empty, err
	}

	// TokensPerRun: derivar de current invirtiendo el orden a ASC.
	tokensPerRun := make([]TokensPerRunRow, len(current))
	for i, r := range current {
		// current está en orden DESC, invertimos para ASC.
		tokensPerRun[len(current)-1-i] = TokensPerRunRow{
			RunID:     r.id,
			StartedAt: r.startedAt,
			Tokens:    r.totalTokens,
		}
	}

	hasEnoughData := len(current) >= 5

	return Metrics{
		RunsCount:           len(current),
		HasEnoughData:       hasEnoughData,
		TotalTokens:         curStats.totalTokens,
		AvgDurationMs:       curStats.avgDurationMs,
		SuccessRate:         curStats.successRate,
		TotalTokensDeltaPct: totalTokensDeltaPct,
		AvgDurationDeltaMs:  avgDurationDeltaMs,
		SuccessRateDelta:    successRateDelta,
		TokensPerRun:        tokensPerRun,
		AvgTimePerAgent:     avgTimePerAgent,
	}, nil
}

// queryAvgTimePerAgent ejecuta Q_AVG_PER_AGENT sobre los IDs de runs del set actual.
func (s *SQLiteStore) queryAvgTimePerAgent(ctx context.Context, runs []runRecord) ([]AgentTimeRow, error) {
	if len(runs) == 0 {
		return []AgentTimeRow{}, nil
	}

	// Construir placeholders dinámicamente: ?, ?, ...
	placeholders := strings.Repeat("?,", len(runs))
	placeholders = placeholders[:len(placeholders)-1] // quitar la última coma

	q := fmt.Sprintf(`
		SELECT agent_role,
		       AVG(duration_ms) AS avg_ms,
		       COUNT(*)         AS runs_included
		FROM agents
		WHERE run_id IN (%s)
		  AND duration_ms IS NOT NULL
		GROUP BY agent_role
		ORDER BY avg_ms DESC
		LIMIT 8`, placeholders)

	// Construir slice de args.
	args := make([]any, len(runs))
	for i, r := range runs {
		args[i] = r.id
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: consultar avg time per agent: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	result := make([]AgentTimeRow, 0)
	for rows.Next() {
		var (
			role         string
			avgMs        float64
			runsIncluded int
		)
		if err := rows.Scan(&role, &avgMs, &runsIncluded); err != nil {
			return nil, fmt.Errorf("dashboard/store: escanear avg time per agent: %w", err)
		}
		result = append(result, AgentTimeRow{
			Role:          role,
			AvgDurationMs: int64(avgMs),
			RunsIncluded:  runsIncluded,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard/store: iterar avg time per agent: %w", err)
	}

	return result, nil
}

// scanRunRecords escanea filas de runs en una lista de runRecord.
func scanRunRecords(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]runRecord, error) {
	var result []runRecord
	for rows.Next() {
		var (
			id           string
			startedAtRaw string
			durationMs   int64
			totalTokens  int
			status       string
		)
		if err := rows.Scan(&id, &startedAtRaw, &durationMs, &totalTokens, &status); err != nil {
			return nil, fmt.Errorf("escanear fila de runs: %w", err)
		}
		startedAt, err := time.Parse(time.RFC3339Nano, startedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("parsear started_at %q: %w", startedAtRaw, err)
		}
		result = append(result, runRecord{
			id:          id,
			startedAt:   startedAt,
			durationMs:  durationMs,
			totalTokens: totalTokens,
			status:      status,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar filas de runs: %w", err)
	}
	return result, nil
}

// periodStats contiene los agregados calculados sobre un período de runs.
type periodStats struct {
	totalTokens   int
	avgDurationMs int64
	successRate   float64
}

// calcStats calcula los agregados sobre un slice de runs.
func calcStats(runs []runRecord) periodStats {
	if len(runs) == 0 {
		return periodStats{}
	}

	var totalTokens int
	var totalDurationMs int64
	var successCount, finishedCount int

	for _, r := range runs {
		totalTokens += r.totalTokens
		totalDurationMs += r.durationMs
		finishedCount++
		if r.status == "success" {
			successCount++
		}
	}

	avgDurationMs := totalDurationMs / int64(len(runs))

	var successRate float64
	if finishedCount > 0 {
		successRate = float64(successCount) / float64(finishedCount)
	}

	return periodStats{
		totalTokens:   totalTokens,
		avgDurationMs: avgDurationMs,
		successRate:   successRate,
	}
}

// calcDeltaPct calcula el cambio porcentual de tokens vs período anterior.
// Retorna nil si no hay período anterior o si prev.total==0.
func calcDeltaPct(curTotal int, prev *periodStats) *float64 {
	if prev == nil || prev.totalTokens == 0 {
		return nil
	}
	delta := float64(curTotal-prev.totalTokens) / float64(prev.totalTokens)
	return &delta
}

// calcDeltaDurationMs calcula la diferencia absoluta en ms vs período anterior.
// Retorna nil si no hay período anterior.
func calcDeltaDurationMs(curAvg int64, prev *periodStats) *int64 {
	if prev == nil {
		return nil
	}
	delta := curAvg - prev.avgDurationMs
	return &delta
}

// calcDeltaSuccessRate calcula la diferencia de tasa de éxito vs período anterior.
// Retorna nil si no hay período anterior.
func calcDeltaSuccessRate(curRate float64, prev *periodStats) *float64 {
	if prev == nil {
		return nil
	}
	delta := curRate - prev.successRate
	return &delta
}
