//go:build dashboard

package dashboard

import (
	"testing"
)

func Test_normalizeDate(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		endOfDay  bool
		expected  string
	}{
		{
			name:     "fecha válida sin endOfDay produce medianoche UTC",
			raw:      "2026-04-08",
			endOfDay: false,
			expected: "2026-04-08T00:00:00Z",
		},
		{
			name:     "fecha válida con endOfDay produce 23:59:59.999999999 UTC",
			raw:      "2026-04-08",
			endOfDay: true,
			expected: "2026-04-08T23:59:59.999999999Z",
		},
		{
			name:     "string vacío retorna string vacío",
			raw:      "",
			endOfDay: false,
			expected: "",
		},
		{
			name:     "string vacío con endOfDay retorna string vacío",
			raw:      "",
			endOfDay: true,
			expected: "",
		},
		{
			name:     "formato inválido retorna string vacío",
			raw:      "invalid",
			endOfDay: false,
			expected: "",
		},
		{
			name:     "formato inválido con endOfDay retorna string vacío",
			raw:      "invalid-date",
			endOfDay: true,
			expected: "",
		},
		{
			name:     "formato RFC3339 completo no es aceptado como YYYY-MM-DD retorna vacío",
			raw:      "2026-04-08T10:00:00Z",
			endOfDay: false,
			expected: "",
		},
		{
			name:     "fecha borde inicio de año sin endOfDay",
			raw:      "2026-01-01",
			endOfDay: false,
			expected: "2026-01-01T00:00:00Z",
		},
		{
			name:     "fecha borde fin de año con endOfDay",
			raw:      "2026-12-31",
			endOfDay: true,
			expected: "2026-12-31T23:59:59.999999999Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeDate(tt.raw, tt.endOfDay)
			if got != tt.expected {
				t.Errorf("normalizeDate(%q, %v): esperado %q, obtuvo %q",
					tt.raw, tt.endOfDay, tt.expected, got)
			}
		})
	}
}
