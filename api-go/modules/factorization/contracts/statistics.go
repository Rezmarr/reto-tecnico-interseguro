package contracts

type Statistics struct {
	Minimum   float64            `json:"minimum"`
	Maximum   float64            `json:"maximum"`
	Sum       float64            `json:"sum"`
	Average   float64            `json:"average"`
	Count     int                `json:"count"`
	Diagonal  DiagonalStatistics `json:"diagonal"`
	Tolerance float64            `json:"tolerance"`
}

type DiagonalCheck struct {
	Applicable bool  `json:"applicable"`
	IsDiagonal *bool `json:"isDiagonal"`
}

type DiagonalStatistics struct {
	Q                 DiagonalCheck `json:"q"`
	R                 DiagonalCheck `json:"r"`
	AnyMatrixDiagonal bool          `json:"anyMatrixDiagonal"`
}
