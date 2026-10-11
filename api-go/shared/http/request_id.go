package http

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/gofiber/fiber/v3"
	"strconv"
	"time"
)

const requestIDKey = "requestId"

func RequestIDMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := NewRequestID()
		c.Locals(requestIDKey, id)
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}
func RequestID(c fiber.Ctx) string {
	if id, ok := c.Locals(requestIDKey).(string); ok && id != "" {
		return id
	}
	return NewRequestID()
}
func NewRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err == nil {
		return hex.EncodeToString(value[:])
	}
	return strconv.FormatInt(time.Now().UnixNano(), 16)
}
