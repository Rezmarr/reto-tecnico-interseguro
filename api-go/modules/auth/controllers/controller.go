package controllers

import (
	"crypto/sha256"
	"crypto/subtle"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"example.com/reto-tecnico/api-go/config"
	"example.com/reto-tecnico/api-go/modules/auth/contracts"
	sharedhttp "example.com/reto-tecnico/api-go/shared/http"
)

type Controller struct{ cfg config.Config }

func New(cfg config.Config) *Controller { return &Controller{cfg: cfg} }
func (h *Controller) Issue(c fiber.Ctx) error {
	var input contracts.TokenRequest
	if err := c.Bind().JSON(&input); err != nil {
		return sharedhttp.WriteError(c, 400, "INVALID_PAYLOAD", "El cuerpo debe contener usuario y contraseña en JSON.")
	}
	if !secureEqual(input.Username, h.cfg.DemoUsername) || !secureEqual(input.Password, h.cfg.DemoPassword) {
		return sharedhttp.WriteError(c, 401, "UNAUTHORIZED", "Las credenciales no son válidas.")
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{Subject: input.Username, Issuer: h.cfg.JWTIssuer, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(h.cfg.JWTTTL))}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.cfg.JWTSecret)
	if err != nil {
		return sharedhttp.WriteError(c, 500, "INTERNAL_ERROR", "No se pudo emitir el token.")
	}
	expiresAt := now.Add(h.cfg.JWTTTL)
	return c.JSON(contracts.TokenResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: int64(h.cfg.JWTTTL.Seconds()), ExpiresAt: expiresAt.UTC().Format(time.RFC3339)})
}

func secureEqual(a, b string) bool {
	ah, bh := sha256.Sum256([]byte(a)), sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(ah[:], bh[:]) == 1
}

func (h *Controller) RequireToken(c fiber.Ctx) error {
	value := c.Get("Authorization")
	if len(value) < 8 || value[:7] != "Bearer " || value[7:] == "" {
		return sharedhttp.WriteError(c, 401, "UNAUTHORIZED", "Se requiere un token Bearer válido.")
	}
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(value[7:], claims, func(token *jwt.Token) (any, error) { return h.cfg.JWTSecret, nil }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(h.cfg.JWTIssuer), jwt.WithExpirationRequired())
	if err != nil || claims.Subject == "" {
		return sharedhttp.WriteError(c, 401, "UNAUTHORIZED", "El token no es válido o expiró.")
	}
	c.Locals("userToken", value[7:])
	return c.Next()
}
