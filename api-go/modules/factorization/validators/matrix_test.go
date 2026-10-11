package validators

import (
	"errors"
	"testing"
)

func TestParseMatrixAcceptsRectangularAndScientificValues(t *testing.T) {
	got, err := ParseMatrix([]byte(`{"matrix":[[1,-2.5],[3e2,4]]}`), 10, 10, 100)
	if err != nil {
		t.Fatalf("ParseMatrix() error = %v", err)
	}
	if got[1][0] != 300 || got[0][1] != -2.5 {
		t.Fatalf("unexpected parsed matrix: %#v", got)
	}
}

func TestParseMatrixRejectsInvalidPayloadsAndLimits(t *testing.T) {
	cases := []struct {
		name, body              string
		rows, columns, elements int
		wantCode                string
	}{
		{"malformed JSON", `{"matrix":`, 10, 10, 100, "INVALID_PAYLOAD"},
		{"unknown property", `{"matrix":[[1]],"extra":true}`, 10, 10, 100, "INVALID_PAYLOAD"},
		{"trailing JSON", `{"matrix":[[1]]}{}`, 10, 10, 100, "INVALID_PAYLOAD"},
		{"missing matrix", `{}`, 10, 10, 100, "INVALID_PAYLOAD"},
		{"empty matrix", `{"matrix":[]}`, 10, 10, 100, "INVALID_MATRIX"},
		{"empty row", `{"matrix":[[]]}`, 10, 10, 100, "INVALID_MATRIX"},
		{"ragged rows", `{"matrix":[[1],[2,3]]}`, 10, 10, 100, "INVALID_MATRIX"},
		{"non numeric", `{"matrix":[["1"]]}`, 10, 10, 100, "INVALID_MATRIX"},
		{"null", `{"matrix":[[null]]}`, 10, 10, 100, "INVALID_MATRIX"},
		{"row cap", `{"matrix":[[1],[2]]}`, 1, 10, 100, "INVALID_MATRIX"},
		{"column cap", `{"matrix":[[1,2]]}`, 10, 1, 100, "INVALID_MATRIX"},
		{"element cap", `{"matrix":[[1,2],[3,4]]}`, 10, 10, 3, "INVALID_MATRIX"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseMatrix([]byte(tc.body), tc.rows, tc.columns, tc.elements)
			var inputErr *InputError
			if !errors.As(err, &inputErr) {
				t.Fatalf("want InputError, got %v", err)
			}
			if inputErr.Code != tc.wantCode {
				t.Fatalf("code = %q; want %q", inputErr.Code, tc.wantCode)
			}
		})
	}
}
