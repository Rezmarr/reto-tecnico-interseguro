package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const maxRequestBytes = 1 << 20

type Config struct {
	Port            string
	NodeAPIURL      string
	FrontendOrigins []string
	NodeTimeout     time.Duration
	MaxRows         int
	MaxColumns      int
	MaxElements     int
	BodyLimit       int
	JWTSecret       []byte
	JWTIssuer       string
	JWTTTL          time.Duration
	DemoUsername    string
	DemoPassword    string
	ServiceAPIKey   string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("error al cargar el archivo .env: %w", err)
	}

	port := envString("PORT", "8080")
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return Config{}, fmt.Errorf("PORT debe ser un entero entre 1 y 65535")
	}

	timeout, err := time.ParseDuration(envString("NODE_API_TIMEOUT", "90s"))
	if err != nil || timeout <= 0 {
		return Config{}, fmt.Errorf("NODE_API_TIMEOUT debe ser una duración positiva de Go")
	}

	maxRows, err := envPositiveInt("MAX_MATRIX_ROWS", 100)
	if err != nil {
		return Config{}, err
	}
	maxColumns, err := envPositiveInt("MAX_MATRIX_COLUMNS", 100)
	if err != nil {
		return Config{}, err
	}
	maxElements, err := envPositiveInt("MAX_MATRIX_ELEMENTS", 10000)
	if err != nil {
		return Config{}, err
	}
	jwtSecretEncoded := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if len(jwtSecretEncoded) >= 2 && (jwtSecretEncoded[0] == '\'' && jwtSecretEncoded[len(jwtSecretEncoded)-1] == '\'' || jwtSecretEncoded[0] == '"' && jwtSecretEncoded[len(jwtSecretEncoded)-1] == '"') {
		jwtSecretEncoded = jwtSecretEncoded[1 : len(jwtSecretEncoded)-1]
	}
	jwtSecret, decodeErr := base64.StdEncoding.Strict().DecodeString(jwtSecretEncoded)
	serviceKey := os.Getenv("SERVICE_API_KEY")
	username, password := os.Getenv("DEMO_USERNAME"), os.Getenv("DEMO_PASSWORD")
	if decodeErr != nil {
		return Config{}, fmt.Errorf("JWT_SECRET debe ser Base64 estándar válido; quita las comillas adicionales si las hay")
	}
	if len(jwtSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET debe decodificarse a por lo menos 32 bytes (actualmente tiene %d)", len(jwtSecret))
	}
	if len(serviceKey) < 32 {
		return Config{}, fmt.Errorf("SERVICE_API_KEY debe contener por lo menos 32 bytes")
	}
	if username == "" || password == "" {
		return Config{}, fmt.Errorf("DEMO_USERNAME y DEMO_PASSWORD son obligatorios")
	}
	ttl, err := time.ParseDuration(envString("JWT_TTL", "15m"))
	if err != nil || ttl <= 0 || ttl > 24*time.Hour {
		return Config{}, fmt.Errorf("JWT_TTL debe ser una duración positiva de máximo 24h")
	}

	return Config{
		Port:            port,
		NodeAPIURL:      envString("NODE_API_URL", "http://localhost:3000"),
		FrontendOrigins: splitList(envString("FRONTEND_ORIGINS", "http://localhost:4200,http://127.0.0.1:4200")),
		NodeTimeout:     timeout,
		MaxRows:         maxRows,
		MaxColumns:      maxColumns,
		MaxElements:     maxElements,
		BodyLimit:       maxRequestBytes,
		JWTSecret:       jwtSecret, JWTIssuer: envString("JWT_ISSUER", "reto-tecnico"), JWTTTL: ttl,
		DemoUsername: username, DemoPassword: password, ServiceAPIKey: serviceKey,
	}, nil
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func splitList(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}

func envPositiveInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s debe ser un entero positivo", key)
	}
	return parsed, nil
}
