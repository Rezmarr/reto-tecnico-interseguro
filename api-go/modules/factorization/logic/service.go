package logic

import (
	"errors"
	"math"

	"gonum.org/v1/gonum/blas/blas64"
	"gonum.org/v1/gonum/lapack/lapack64"
)

var ErrNumericalFailure = errors.New("QR factorization produced a non-finite value")

type Service struct{}

func NewService() *Service { return &Service{} }
func (service *Service) Factorize(input [][]float64) (q, r [][]float64, err error) {
	m, n := len(input), len(input[0])
	data := make([]float64, m*n)
	for i := range input {
		copy(data[i*n:(i+1)*n], input[i])
	}
	a := blas64.General{Rows: m, Cols: n, Stride: n, Data: data}
	tau := make([]float64, min(m, n))
	work := []float64{0}
	lapack64.Geqrf(a, tau, work, -1)
	work = make([]float64, max(1, int(work[0])))
	lapack64.Geqrf(a, tau, work, len(work))

	qData := make([]float64, m*m)
	for i := 0; i < m; i++ {
		copy(qData[i*m:(i+1)*m], data[i*n:i*n+min(m, n)])
	}
	qMatrix := blas64.General{Rows: m, Cols: m, Stride: m, Data: qData}
	qWork := []float64{0}
	lapack64.Orgqr(qMatrix, tau, qWork, -1)
	qWork = make([]float64, max(1, int(qWork[0])))
	lapack64.Orgqr(qMatrix, tau, qWork, len(qWork))

	q = make([][]float64, m)
	r = make([][]float64, m)
	for i := 0; i < m; i++ {
		q[i] = append([]float64(nil), qData[i*m:(i+1)*m]...)
		r[i] = make([]float64, n)
		for j := i; j < n; j++ {
			r[i][j] = data[i*n+j]
		}
	}
	if !finite(q) || !finite(r) {
		return nil, nil, ErrNumericalFailure
	}
	return q, r, nil
}

func finite(matrix [][]float64) bool {
	for _, row := range matrix {
		for _, value := range row {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return false
			}
		}
	}
	return true
}
