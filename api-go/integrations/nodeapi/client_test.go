package nodeapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCalculateStatisticsTraduceFallosDeNode(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		timeout time.Duration
		want    error
	}{
		{"error del servidor", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }, time.Second, ErrUnavailable},
		{"JSON inválido", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("not-json")) }, time.Second, ErrBadResponse},
		{"tiempo de espera agotado", func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(50 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}, 5 * time.Millisecond, ErrTimeout},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			client, err := New(server.URL, tc.timeout, "service-key")
			if err != nil {
				t.Fatalf("error al crear el cliente HTTP: %v", err)
			}
			_, err = client.CalculateStatistics(context.Background(), "request-id", "user-token", [][]float64{{1, 0}, {0, 1}}, [][]float64{{1, 0}, {0, 1}})
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v; se esperaba %v", err, tc.want)
			}
		})
	}
}
