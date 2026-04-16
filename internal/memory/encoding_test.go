package memory

import (
	"math"
	"testing"
)

func TestEncodeDecodeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		v    []float32
	}{
		{name: "empty slice", v: []float32{}},
		{name: "single zero", v: []float32{0}},
		{name: "single positive", v: []float32{1.5}},
		{name: "single negative", v: []float32{-1.5}},
		{name: "multiple values", v: []float32{0.1, 0.2, 0.3, -0.4, 99.9}},
		{name: "special: positive infinity", v: []float32{float32(math.Inf(1))}},
		{name: "special: negative infinity", v: []float32{float32(math.Inf(-1))}},
		{name: "special: NaN", v: []float32{float32(math.NaN())}},
		{name: "max float32", v: []float32{math.MaxFloat32}},
		{name: "smallest positive nonzero", v: []float32{math.SmallestNonzeroFloat32}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded := EncodeEmbedding(tc.v)

			// byte length must be exactly 4 * len(v)
			if got, want := len(encoded), len(tc.v)*4; got != want {
				t.Fatalf("encoded length = %d, want %d", got, want)
			}

			decoded, err := DecodeEmbedding(encoded)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(decoded) != len(tc.v) {
				t.Fatalf("decoded length = %d, want %d", len(decoded), len(tc.v))
			}

			for i, want := range tc.v {
				got := decoded[i]
				// NaN != NaN, so compare bits directly.
				if math.IsNaN(float64(want)) {
					if !math.IsNaN(float64(got)) {
						t.Errorf("index %d: got %v, want NaN", i, got)
					}
					continue
				}
				if got != want {
					t.Errorf("index %d: got %v, want %v", i, got, want)
				}
			}
		})
	}
}

func TestEncodeEmbedding_ByteOrder(t *testing.T) {
	// 1.0 as IEEE 754 float32 = 0x3F800000 → little-endian bytes: 00 00 80 3F
	v := []float32{1.0}
	got := EncodeEmbedding(v)
	want := []byte{0x00, 0x00, 0x80, 0x3F}

	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("byte[%d]: got %#x, want %#x", i, got[i], want[i])
		}
	}
}

func TestDecodeEmbedding_InvalidLength(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
	}{
		{name: "one byte", b: []byte{0x01}},
		{name: "two bytes", b: []byte{0x01, 0x02}},
		{name: "three bytes", b: []byte{0x01, 0x02, 0x03}},
		{name: "five bytes", b: []byte{0x01, 0x02, 0x03, 0x04, 0x05}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeEmbedding(tc.b)
			if err == nil {
				t.Fatal("expected error for non-multiple-of-4 blob, got nil")
			}
		})
	}
}

func TestDecodeEmbedding_EmptySlice(t *testing.T) {
	decoded, err := DecodeEmbedding([]byte{})
	if err != nil {
		t.Fatalf("unexpected error for empty blob: %v", err)
	}
	if len(decoded) != 0 {
		t.Fatalf("expected empty slice, got length %d", len(decoded))
	}
}
