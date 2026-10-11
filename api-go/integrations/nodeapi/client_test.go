package nodeapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCalculateStatisticsMapsNodeFailures(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		timeout time.Duration
		want    error
	}{
		{"server error", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }, time.Second, ErrUnavailable},
		{"invalid JSON", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("not-json")) }, time.Second, ErrBadResponse},
		{"timeout", func(w http.ResponseWriter, _ *http.Request) {
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
				t.Fatal(err)
			}
			_, err = client.CalculateStatistics(context.Background(), "request-id", "user-token", [][]float64{{1, 0}, {0, 1}}, [][]float64{{1, 0}, {0, 1}})
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v; want %v", err, tc.want)
			}
		})
	}
}
