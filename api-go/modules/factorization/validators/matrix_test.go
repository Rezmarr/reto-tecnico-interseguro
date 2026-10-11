package validators

import (
	"errors"
	"testing"
)

func TestParseMatrixAceptaMatricesRectangularesYValoresCientificos(t *testing.T) {
	got, err := ParseMatrix([]byte(`{"matrix":[[1,-2.5],[3e2,4]]}`), 10, 10, 100)
	if err != nil {
		t.Fatalf("error en ParseMatrix(): %v", err)
	}
	if got[1][0] != 300 || got[0][1] != -2.5 {
		t.Fatalf("matriz interpretada inesperada: %#v", got)
	}
}

func TestParseMatrixRechazaSolicitudesInvalidasYExcesosDeLimites(t *testing.T) {
	cases := []struct {
		name, body              string
		rows, columns, elements int
		wantCode                string
	}{
		{"JSON malformado", `{"matrix":`, 10, 10, 100, "INVALID_PAYLOAD"},
		{"propiedad desconocida", `{"matrix":[[1]],"extra":true}`, 10, 10, 100, "INVALID_PAYLOAD"},
		{"JSON sobrante", `{"matrix":[[1]]}{}`, 10, 10, 100, "INVALID_PAYLOAD"},
		{"falta la matriz", `{}`, 10, 10, 100, "INVALID_PAYLOAD"},
		{"matriz vacía", `{"matrix":[]}`, 10, 10, 100, "INVALID_MATRIX"},
		{"fila vacía", `{"matrix":[[]]}`, 10, 10, 100, "INVALID_MATRIX"},
		{"filas irregulares", `{"matrix":[[1],[2,3]]}`, 10, 10, 100, "INVALID_MATRIX"},
		{"valor no numérico", `{"matrix":[["1"]]}`, 10, 10, 100, "INVALID_MATRIX"},
		{"null", `{"matrix":[[null]]}`, 10, 10, 100, "INVALID_MATRIX"},
		{"límite de filas", `{"matrix":[[1],[2]]}`, 1, 10, 100, "INVALID_MATRIX"},
		{"límite de columnas", `{"matrix":[[1,2]]}`, 10, 1, 100, "INVALID_MATRIX"},
		{"límite de elementos", `{"matrix":[[1,2],[3,4]]}`, 10, 10, 3, "INVALID_MATRIX"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseMatrix([]byte(tc.body), tc.rows, tc.columns, tc.elements)
			var inputErr *InputError
			if !errors.As(err, &inputErr) {
				t.Fatalf("se esperaba InputError, se obtuvo %v", err)
			}
			if inputErr.Code != tc.wantCode {
				t.Fatalf("código = %q; se esperaba %q", inputErr.Code, tc.wantCode)
			}
		})
	}
}
