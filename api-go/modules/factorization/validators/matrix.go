package validators

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"strconv"
	"strings"
)

type InputError struct{ Code, Message string }

func (e *InputError) Error() string { return e.Message }
func IsJSON(contentType string) bool {
	media, _, err := mime.ParseMediaType(contentType)
	return err == nil && strings.EqualFold(media, "application/json")
}
func ParseMatrix(body []byte, maxRows, maxColumns, maxElements int) ([][]float64, error) {
	var input struct {
		Matrix json.RawMessage `json:"matrix"`
	}
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if d.Decode(&input) != nil {
		return nil, &InputError{"INVALID_PAYLOAD", "El cuerpo debe contener un objeto JSON con matrix."}
	}
	if err := d.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, &InputError{"INVALID_PAYLOAD", "El cuerpo debe contener un único objeto JSON."}
	}
	if len(input.Matrix) == 0 {
		return nil, &InputError{"INVALID_PAYLOAD", "Falta el campo matrix."}
	}
	var raw [][]json.RawMessage
	if json.Unmarshal(input.Matrix, &raw) != nil || len(raw) == 0 {
		return nil, invalid("matrix debe ser una matriz rectangular no vacía.")
	}
	cols := len(raw[0])
	if cols == 0 {
		return nil, invalid("Las filas de matrix no pueden estar vacías.")
	}
	if len(raw) > maxRows {
		return nil, invalid(fmt.Sprintf("La matriz supera el máximo de %d filas.", maxRows))
	}
	if cols > maxColumns {
		return nil, invalid(fmt.Sprintf("La matriz supera el máximo de %d columnas.", maxColumns))
	}
	if len(raw) > maxElements/cols {
		return nil, invalid(fmt.Sprintf("La matriz supera el máximo de %d elementos.", maxElements))
	}
	values := make([][]float64, len(raw))
	for i, row := range raw {
		if len(row) != cols {
			return nil, invalid("matrix debe ser rectangular.")
		}
		values[i] = make([]float64, cols)
		for j, item := range row {
			v, err := strconv.ParseFloat(string(item), 64)
			if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, invalid(fmt.Sprintf("El elemento de fila %d y columna %d debe ser un número finito.", i, j))
			}
			values[i][j] = v
		}
	}
	return values, nil
}
func invalid(message string) *InputError { return &InputError{"INVALID_MATRIX", message} }
