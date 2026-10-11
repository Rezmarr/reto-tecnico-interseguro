package contracts

import "encoding/json"

type Request struct {
	Matrix json.RawMessage `json:"matrix"`
}
type MatrixShape struct {
	Rows    int `json:"rows"`
	Columns int `json:"columns"`
}
type Response struct {
	Shape      MatrixShape `json:"shape"`
	Q          [][]float64 `json:"q"`
	R          [][]float64 `json:"r"`
	Statistics Statistics  `json:"statistics"`
}
