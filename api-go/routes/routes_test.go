package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"example.com/reto-tecnico/api-go/config"
	"example.com/reto-tecnico/api-go/integrations/nodeapi"
	nodecontracts "example.com/reto-tecnico/api-go/integrations/nodeapi/contracts"
	"example.com/reto-tecnico/api-go/modules/auth/contracts"
)

func TestRutaDeAutenticacionYFactorizacionSeIntegraConNode(t *testing.T) {
	const serviceKey = "integration-service-key-with-at-least-32-bytes"
	const username, password = "demo-user", "demo-password"
	const tokenSecret = "01234567890123456789012345678901"
	trueValue := true
	var nodeCallCount int

	nodeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nodeCallCount++
		if r.URL.Path != "/api/v1/statistics" {
			t.Errorf("ruta de Node = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" || r.Header.Get("X-Service-Token") != serviceKey {
			t.Errorf("Go no reenvió las credenciales de usuario y servicio")
		}
		var input nodecontracts.StatisticsRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Errorf("error al decodificar la solicitud de Node: %v", err)
		}
		if len(input.Matrices["q"]) != 2 || len(input.Matrices["r"]) != 2 || len(input.Matrices["r"][0]) != 3 {
			t.Errorf("se enviaron matrices inesperadas a Node: %#v", input.Matrices)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(nodecontracts.StatisticsResponse{
			Minimum: 0, Maximum: 1, Sum: 4, Average: .4, Count: 10, Tolerance: 1e-10,
			Diagonal: nodecontracts.DiagonalStatistics{
				Q: nodecontracts.DiagonalCheck{Applicable: true, IsDiagonal: &trueValue},
				R: nodecontracts.DiagonalCheck{Applicable: false, IsDiagonal: nil}, AnyMatrixDiagonal: true,
			},
		})
	}))
	defer nodeServer.Close()

	client, err := nodeapi.New(nodeServer.URL, 2*time.Second, serviceKey)
	if err != nil {
		t.Fatalf("error al crear el cliente de Node: %v", err)
	}
	cfg := config.Config{
		Port: "8080", NodeTimeout: 2 * time.Second, MaxRows: 100, MaxColumns: 100, MaxElements: 10000,
		BodyLimit: 1 << 20, JWTSecret: []byte(tokenSecret), JWTIssuer: "test-issuer", JWTTTL: time.Minute,
		DemoUsername: username, DemoPassword: password, ServiceAPIKey: serviceKey,
	}
	app := New(cfg, client)

	badLogin := postJSON(t, app, "/api/v1/auth/token", `{"username":"demo-user","password":"wrong"}`, "")
	if badLogin.Code != http.StatusUnauthorized {
		t.Fatalf("estado con credenciales incorrectas = %d; se esperaba 401", badLogin.Code)
	}

	login := postJSON(t, app, "/api/v1/auth/token", `{"username":"demo-user","password":"demo-password"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("estado del inicio de sesión = %d, cuerpo: %s", login.Code, login.Body)
	}
	var token contracts.TokenResponse
	if err := json.Unmarshal(login.Body.Bytes(), &token); err != nil {
		t.Fatalf("error al interpretar la respuesta del token: %v", err)
	}
	if token.AccessToken == "" || token.TokenType != "Bearer" || token.ExpiresIn != 60 {
		t.Fatalf("respuesta inesperada del token: %+v", token)
	}

	unauthorized := postJSON(t, app, "/api/v1/factorizations", `{"matrix":[[1,2,3],[4,5,6]]}`, "")
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("estado sin token = %d; se esperaba 401", unauthorized.Code)
	}
	invalidToken := postJSON(t, app, "/api/v1/factorizations", `{"matrix":[[1,2,3],[4,5,6]]}`, "not-a-jwt")
	if invalidToken.Code != http.StatusUnauthorized {
		t.Fatalf("estado con token inválido = %d; se esperaba 401", invalidToken.Code)
	}
	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: username, Issuer: cfg.JWTIssuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
	})
	expiredToken, err := expired.SignedString(cfg.JWTSecret)
	if err != nil {
		t.Fatalf("error al firmar el token vencido: %v", err)
	}
	expiredResponse := postJSON(t, app, "/api/v1/factorizations", `{"matrix":[[1,2,3],[4,5,6]]}`, expiredToken)
	if expiredResponse.Code != http.StatusUnauthorized {
		t.Fatalf("estado con token vencido = %d; se esperaba 401", expiredResponse.Code)
	}

	response := postJSON(t, app, "/api/v1/factorizations", `{"matrix":[[1,2,3],[4,5,6]]}`, token.AccessToken)
	if response.Code != http.StatusOK {
		t.Fatalf("estado de factorización = %d, cuerpo: %s", response.Code, response.Body)
	}
	var payload struct {
		Shape struct {
			Rows    int `json:"rows"`
			Columns int `json:"columns"`
		} `json:"shape"`
		Q          [][]float64 `json:"q"`
		R          [][]float64 `json:"r"`
		Statistics struct {
			Count int `json:"count"`
		} `json:"statistics"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("error al interpretar la respuesta de factorización: %v", err)
	}
	if payload.Shape.Rows != 2 || payload.Shape.Columns != 3 || len(payload.Q) != 2 || len(payload.R) != 2 || len(payload.R[0]) != 3 || payload.Statistics.Count != 10 {
		t.Fatalf("respuesta de API inesperada: %+v", payload)
	}
	if nodeCallCount != 1 {
		t.Fatalf("se llamó a Node %d veces; se esperaba una", nodeCallCount)
	}
}

func postJSON(t *testing.T, app *fiber.App, path, body, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("error al enviar la solicitud de prueba: %v", err)
	}
	defer resp.Body.Close()
	recorder := httptest.NewRecorder()
	recorder.Code = resp.StatusCode
	_, _ = recorder.Body.ReadFrom(resp.Body)
	return recorder
}
