package logic

import (
	"math"
	"testing"
)

func TestFactorizacionReconstruyeEntradaYDevuelveQOrtogonal(t *testing.T) {
	cases := []struct {
		name  string
		input [][]float64
	}{
		{"cuadrada", [][]float64{{2, 1}, {1, 3}}},
		{"alta", [][]float64{{1, 0}, {0, 1}, {1, 1}}},
		{"ancha", [][]float64{{1, 2, 3}, {4, 5, 6}}},
		{"una fila", [][]float64{{2, -1, 4}}},
		{"una columna", [][]float64{{2}, {-3}, {4}}},
		{"matriz cero", [][]float64{{0, 0}, {0, 0}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q, r, err := NewService().Factorize(tc.input)
			if err != nil {
				t.Fatalf("error en Factorize(): %v", err)
			}
			m, n := len(tc.input), len(tc.input[0])
			if len(q) != m || len(r) != m {
				t.Fatalf("Q tiene %d filas y R tiene %d; se esperaban %d", len(q), len(r), m)
			}
			for i := 0; i < m; i++ {
				if len(q[i]) != m || len(r[i]) != n {
					t.Fatalf("dimensiones de fila inesperadas: Q %d, R %d", len(q[i]), len(r[i]))
				}
				for j := 0; j < n; j++ {
					var reconstructed float64
					for k := 0; k < m; k++ {
						reconstructed += q[i][k] * r[k][j]
					}
					if math.Abs(reconstructed-tc.input[i][j]) > 1e-10 {
						t.Errorf("QR[%d,%d] = %.12g; se esperaba %.12g", i, j, reconstructed, tc.input[i][j])
					}
				}
				for j := 0; j < i && j < n; j++ {
					if math.Abs(r[i][j]) > 1e-10 {
						t.Errorf("R[%d,%d] debajo de la diagonal no es cero", i, j)
					}
				}
			}
			for i := 0; i < m; i++ {
				for j := 0; j < m; j++ {
					var dot float64
					for k := 0; k < m; k++ {
						dot += q[k][i] * q[k][j]
					}
					want := 0.0
					if i == j {
						want = 1
					}
					if math.Abs(dot-want) > 1e-10 {
						t.Errorf("QᵀQ[%d,%d] = %.12g; se esperaba %.12g", i, j, dot, want)
					}
				}
			}
		})
	}
}
