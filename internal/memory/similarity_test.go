package memory

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name    string
		a       []float32
		b       []float32
		want    float64
		approx  bool
		epsilon float64
	}{
		{
			name: "identical vectors",
			a:    []float32{1, 0},
			b:    []float32{1, 0},
			want: 1.0,
		},
		{
			name: "opposite vectors",
			a:    []float32{1, 0},
			b:    []float32{-1, 0},
			want: -1.0,
		},
		{
			name: "orthogonal vectors",
			a:    []float32{1, 0},
			b:    []float32{0, 1},
			want: 0.0,
		},
		{
			name:    "arbitrary angle",
			a:       []float32{1, 1},
			b:       []float32{1, 0},
			want:    1.0 / math.Sqrt2,
			approx:  true,
			epsilon: 1e-6,
		},
		{
			name: "zero vector a",
			a:    []float32{0, 0},
			b:    []float32{1, 0},
			want: 0.0,
		},
		{
			name: "zero vector b",
			a:    []float32{1, 0},
			b:    []float32{0, 0},
			want: 0.0,
		},
		{
			name: "empty slices",
			a:    []float32{},
			b:    []float32{},
			want: 0.0,
		},
		{
			name: "length mismatch",
			a:    []float32{1, 0},
			b:    []float32{1, 0, 0},
			want: 0.0,
		},
		{
			name: "single element same",
			a:    []float32{3},
			b:    []float32{3},
			want: 1.0,
		},
		{
			name: "single element opposite",
			a:    []float32{2},
			b:    []float32{-2},
			want: -1.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CosineSimilarity(tc.a, tc.b)
			if tc.approx {
				if math.Abs(got-tc.want) > tc.epsilon {
					t.Errorf("CosineSimilarity(%v, %v) = %v, want ≈ %v (ε=%v)",
						tc.a, tc.b, got, tc.want, tc.epsilon)
				}
			} else {
				if got != tc.want {
					t.Errorf("CosineSimilarity(%v, %v) = %v, want %v",
						tc.a, tc.b, got, tc.want)
				}
			}
		})
	}
}
