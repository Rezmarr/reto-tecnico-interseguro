package controllers

import (
	"context"
	"errors"
	"example.com/reto-tecnico/api-go/config"
	"example.com/reto-tecnico/api-go/integrations/nodeapi"
	nodecontracts "example.com/reto-tecnico/api-go/integrations/nodeapi/contracts"
	"example.com/reto-tecnico/api-go/modules/factorization/contracts"
	"example.com/reto-tecnico/api-go/modules/factorization/logic"
	"example.com/reto-tecnico/api-go/modules/factorization/validators"
	sharedhttp "example.com/reto-tecnico/api-go/shared/http"
	"github.com/gofiber/fiber/v3"
)

type Controller struct {
	config     config.Config
	statistics *nodeapi.Client
	service    *logic.Service
}

func New(cfg config.Config, client *nodeapi.Client) *Controller {
	return &Controller{config: cfg, statistics: client, service: logic.NewService()}
}
func (h *Controller) Create(c fiber.Ctx) error {
	if !validators.IsJSON(c.Get("Content-Type")) {
		return sharedhttp.WriteError(c, 400, "INVALID_PAYLOAD", "El cuerpo debe ser JSON.")
	}
	values, err := validators.ParseMatrix(c.Body(), h.config.MaxRows, h.config.MaxColumns, h.config.MaxElements)
	if err != nil {
		var input *validators.InputError
		if errors.As(err, &input) {
			return sharedhttp.WriteError(c, 400, input.Code, input.Message)
		}
		return sharedhttp.WriteError(c, 400, "INVALID_MATRIX", err.Error())
	}
	q, r, err := h.service.Factorize(values)
	if errors.Is(err, logic.ErrNumericalFailure) {
		return sharedhttp.WriteError(c, 422, "NUMERICAL_FAILURE", "La matriz produjo valores fuera del rango numérico soportado.")
	}
	if err != nil {
		return sharedhttp.WriteError(c, 500, "INTERNAL_ERROR", "No se pudo factorizar la matriz.")
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.config.NodeTimeout)
	defer cancel()
	userToken, _ := c.Locals("userToken").(string)
	stats, err := h.statistics.CalculateStatistics(ctx, sharedhttp.RequestID(c), userToken, q, r)
	if err != nil {
		if errors.Is(err, nodeapi.ErrTimeout) {
			return sharedhttp.WriteError(c, 504, "STATISTICS_SERVICE_TIMEOUT", "La API de estadísticas no respondió a tiempo.")
		}
		return sharedhttp.WriteError(c, 502, "STATISTICS_SERVICE_UNAVAILABLE", "No se pudieron calcular las estadísticas en la API Node.")
	}
	return c.JSON(contracts.Response{Shape: contracts.MatrixShape{Rows: len(values), Columns: len(values[0])}, Q: q, R: r, Statistics: mapStatistics(stats)})
}

func mapStatistics(value nodecontracts.StatisticsResponse) contracts.Statistics {
	return contracts.Statistics{
		Minimum: value.Minimum, Maximum: value.Maximum, Sum: value.Sum,
		Average: value.Average, Count: value.Count, Tolerance: value.Tolerance,
		Diagonal: contracts.DiagonalStatistics{
			Q:                 contracts.DiagonalCheck{Applicable: value.Diagonal.Q.Applicable, IsDiagonal: value.Diagonal.Q.IsDiagonal},
			R:                 contracts.DiagonalCheck{Applicable: value.Diagonal.R.Applicable, IsDiagonal: value.Diagonal.R.IsDiagonal},
			AnyMatrixDiagonal: value.Diagonal.AnyMatrixDiagonal,
		},
	}
}
