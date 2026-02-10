package main_test

import (
	"net/http"
	"os"
	"testing"
)

var baseURL = "http://localhost:3000"

func init() {
	if u := os.Getenv("API_URL"); u != "" {
		baseURL = u
	}
}

func apiAvailable(t *testing.T) bool {
	resp, err := http.Get(baseURL + "/alive")
	if err != nil {
		t.Skipf("API not available at %s: %v (start API with: air)", baseURL, err)
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func TestGetOrders(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	t.Run("returns 200", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/orders")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})
}
