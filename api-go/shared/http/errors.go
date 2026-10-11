package http

import (
	"errors"
	sharedcontracts "example.com/reto-tecnico/api-go/shared/contracts"
	"github.com/gofiber/fiber/v3"
)

func WriteError(c fiber.Ctx, status int, code, message string) error {
	id := RequestID(c)
	c.Set("X-Request-ID", id)
	return c.Status(status).JSON(sharedcontracts.ErrorResponse{Error: sharedcontracts.ErrorDetail{Code: code, Message: message, RequestID: id}})
}
func ErrorHandler(c fiber.Ctx, err error) error {
	var fe *fiber.Error
	if !errors.As(err, &fe) || fe == nil {
		return WriteError(c, 500, "INTERNAL_ERROR", "Ocurrió un error interno.")
	}
	switch fe.Code {
	case 413:
		return WriteError(c, 413, "PAYLOAD_TOO_LARGE", "El cuerpo de la petición excede el límite permitido.")
	case 404:
		return WriteError(c, 404, "NOT_FOUND", "La ruta solicitada no existe.")
	case 405:
		return WriteError(c, 405, "METHOD_NOT_ALLOWED", "El método HTTP no está permitido para esta ruta.")
	case 400:
		return WriteError(c, 400, "INVALID_PAYLOAD", "La petición no tiene un formato válido.")
	default:
		if fe.Code >= 400 && fe.Code < 500 {
			return WriteError(c, fe.Code, "INVALID_PAYLOAD", "La petición no pudo procesarse.")
		}
		return WriteError(c, 500, "INTERNAL_ERROR", "Ocurrió un error interno.")
	}
}
