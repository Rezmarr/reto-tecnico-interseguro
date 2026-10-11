package nodeapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"example.com/reto-tecnico/api-go/integrations/nodeapi/contracts"
)

var (
	ErrUnavailable = errors.New("la API de Node no está disponible")
	ErrTimeout     = errors.New("la API de Node agotó el tiempo de espera")
	ErrBadResponse = errors.New("la API de Node devolvió una respuesta inválida")
)

type Client struct {
	endpoint   string
	serviceKey string
	http       *http.Client
}

func New(baseURL string, timeout time.Duration, serviceKey string) (*Client, error) {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("NODE_API_URL debe ser una URL absoluta HTTP o HTTPS")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/api/v1/statistics"
	return &Client{
		endpoint: parsed.String(), serviceKey: serviceKey,
		http: &http.Client{Timeout: timeout},
	}, nil
}

func (client *Client) CalculateStatistics(ctx context.Context, requestID, userToken string, q, r [][]float64) (contracts.StatisticsResponse, error) {
	body, err := json.Marshal(contracts.StatisticsRequest{
		Matrices: map[string][][]float64{"q": q, "r": r},
	})
	if err != nil {
		return contracts.StatisticsResponse{}, fmt.Errorf("error al codificar la solicitud para la API de Node: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.endpoint, bytes.NewReader(body))
	if err != nil {
		return contracts.StatisticsResponse{}, fmt.Errorf("error al crear la solicitud para la API de Node: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-Request-ID", requestID)
	request.Header.Set("Authorization", "Bearer "+userToken)
	request.Header.Set("X-Service-Token", client.serviceKey)

	response, err := client.http.Do(request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return contracts.StatisticsResponse{}, ErrTimeout
		}
		var netErr interface{ Timeout() bool }
		if errors.As(err, &netErr) && netErr.Timeout() {
			return contracts.StatisticsResponse{}, ErrTimeout
		}
		return contracts.StatisticsResponse{}, ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return contracts.StatisticsResponse{}, ErrUnavailable
	}

	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	var result contracts.StatisticsResponse
	if err := decoder.Decode(&result); err != nil {
		return contracts.StatisticsResponse{}, ErrBadResponse
	}
	expectedCount := len(q)*len(q) + len(r)*len(r[0])
	if err := validateResult(result, expectedCount, len(r) == len(r[0])); err != nil {
		return contracts.StatisticsResponse{}, ErrBadResponse
	}
	return result, nil
}

func validateResult(result contracts.StatisticsResponse, expectedCount int, rIsSquare bool) error {
	for _, value := range []float64{result.Minimum, result.Maximum, result.Sum, result.Average, result.Tolerance} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
		return errors.New("las estadísticas contienen valores no finitos")
		}
	}
	if result.Count != expectedCount || result.Count == 0 || result.Minimum > result.Maximum || result.Tolerance < 0 {
		return errors.New("las estadísticas son inconsistentes")
	}
	if !validDiagonalCheck(result.Diagonal.Q) || !validDiagonalCheck(result.Diagonal.R) {
		return errors.New("el resultado de diagonalidad es inconsistente")
	}
	if !result.Diagonal.Q.Applicable || result.Diagonal.R.Applicable != rIsSquare {
		return errors.New("la aplicabilidad de la diagonalidad no coincide con las dimensiones de la matriz")
	}
	anyDiagonal := (result.Diagonal.Q.Applicable && *result.Diagonal.Q.IsDiagonal) ||
		(result.Diagonal.R.Applicable && *result.Diagonal.R.IsDiagonal)
	if result.Diagonal.AnyMatrixDiagonal != anyDiagonal || result.Average < result.Minimum || result.Average > result.Maximum {
		return errors.New("las estadísticas agregadas son inconsistentes")
	}
	return nil
}

func validDiagonalCheck(check contracts.DiagonalCheck) bool {
	if check.Applicable {
		return check.IsDiagonal != nil
	}
	return check.IsDiagonal == nil
}
