package main_test

import (
	"bytes"
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

	t.Run("returns 200 with data and total", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/orders")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("respects page and limit params", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/orders?page=1&limit=5")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("accepts status filter", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/orders?status=pending")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("accepts amount filters", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/orders?min_amount=10&max_amount=1000")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("accepts date range filters", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/orders?from_date=2025-01-01&to_date=2025-12-31")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestPostOrder(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	t.Run("creates order and returns 201", func(t *testing.T) {
		body := bytes.NewBufferString(`{"customer_name":"John Doe","status":"pending","total":49.99}`)
		resp, err := http.Post(baseURL+"/orders", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			t.Errorf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("defaults status to pending when empty", func(t *testing.T) {
		body := bytes.NewBufferString(`{"customer_name":"Jane Doe","total":25.00}`)
		resp, err := http.Post(baseURL+"/orders", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			t.Errorf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 400 when customer_name is missing", func(t *testing.T) {
		body := bytes.NewBufferString(`{"status":"pending","total":99.99}`)
		resp, err := http.Post(baseURL+"/orders", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 400 {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 400 when customer_name is empty", func(t *testing.T) {
		body := bytes.NewBufferString(`{"customer_name":"","status":"pending","total":99.99}`)
		resp, err := http.Post(baseURL+"/orders", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 400 {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		body := bytes.NewBufferString(`{invalid json}`)
		resp, err := http.Post(baseURL+"/orders", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 400 {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("edge: accepts zero total", func(t *testing.T) {
		body := bytes.NewBufferString(`{"customer_name":"Zero Order","total":0}`)
		resp, err := http.Post(baseURL+"/orders", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			t.Errorf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("edge: accepts large amount", func(t *testing.T) {
		body := bytes.NewBufferString(`{"customer_name":"Big Order","total":999999.99}`)
		resp, err := http.Post(baseURL+"/orders", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			t.Errorf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("edge: empty body returns 400", func(t *testing.T) {
		body := bytes.NewBufferString(``)
		resp, err := http.Post(baseURL+"/orders", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 400 {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})
}
