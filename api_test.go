package main_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var baseURL = "http://localhost:3000"

func init() {
	if u := os.Getenv("API_URL"); u != "" {
		baseURL = u
	}
}

func apiAvailable(t *testing.T) bool {
	t.Helper()
	resp, err := http.Get(baseURL + "/alive")
	if err != nil {
		t.Skipf("API not available at %s: %v (start API with: air)", baseURL, err)
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

// Response types for test assertions.

type ordersResponse struct {
	Orders     []orderItem `json:"orders"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

type orderItem struct {
	ID           string  `json:"id"`
	CustomerName string  `json:"customer_name"`
	Status       string  `json:"status"`
	Total        float64 `json:"total"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Helpers.

// totalOrders fetches the unfiltered total count from the API.
func totalOrders(t *testing.T) int {
	t.Helper()
	_, result := getOrders(t, "?limit=1")
	return result.Total
}

// getOrders performs a GET /orders request with the given query string and
// decodes a successful (200) response into an ordersResponse.
func getOrders(t *testing.T, query string) (*http.Response, ordersResponse) {
	t.Helper()
	resp, err := http.Get(baseURL + "/orders" + query)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	var result ordersResponse
	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, &result); err != nil {
			t.Fatalf("failed to decode response: %v\nbody: %s", err, body)
		}
	}
	return resp, result
}

// getOrdersError performs a GET /orders request expected to fail and decodes
// the response body into an errorResponse.
func getOrdersError(t *testing.T, query string) (*http.Response, errorResponse) {
	t.Helper()
	resp, err := http.Get(baseURL + "/orders" + query)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	var result errorResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to decode error response: %v\nbody: %s", err, body)
	}
	return resp, result
}

// --- Pagination Tests ---

// TestGetOrders_DefaultPagination verifies that calling GET /orders with no
// query parameters returns the first page with default limit=20 and correct
// total/total_pages metadata.
func TestGetOrders_DefaultPagination(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	total := totalOrders(t)
	resp, result := getOrders(t, "")

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result.Total != total {
		t.Errorf("expected total=%d, got %d", total, result.Total)
	}
	expectedPages := (total + 19) / 20 // ceil(total / 20)
	if result.TotalPages != expectedPages {
		t.Errorf("expected total_pages=%d, got %d", expectedPages, result.TotalPages)
	}
	expectedLen := 20
	if total < 20 {
		expectedLen = total
	}
	if len(result.Orders) != expectedLen {
		t.Errorf("expected %d orders, got %d", expectedLen, len(result.Orders))
	}
}

// TestGetOrders_CustomPageAndLimit verifies that explicit page and limit
// parameters are respected: page=2&limit=5 should return exactly 5 orders
// from the second page and matching pagination metadata.
func TestGetOrders_CustomPageAndLimit(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	total := totalOrders(t)
	resp, result := getOrders(t, "?page=2&limit=5")

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(result.Orders) != 5 {
		t.Errorf("expected 5 orders, got %d", len(result.Orders))
	}
	if result.Total != total {
		t.Errorf("expected total=%d, got %d", total, result.Total)
	}
	expectedPages := (total + 4) / 5 // ceil(total / 5)
	if result.TotalPages != expectedPages {
		t.Errorf("expected total_pages=%d, got %d", expectedPages, result.TotalPages)
	}
}

// TestGetOrders_LastPagePartial requests the last page with limit=15 and
// verifies it returns only the remaining orders (fewer than limit).
func TestGetOrders_LastPagePartial(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	total := totalOrders(t)
	const limit = 15
	expectedPages := (total + limit - 1) / limit // ceil(total / limit)
	lastPageCount := total - (expectedPages-1)*limit

	resp, result := getOrders(t, fmt.Sprintf("?page=%d&limit=%d", expectedPages, limit))

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result.TotalPages != expectedPages {
		t.Errorf("expected total_pages=%d, got %d", expectedPages, result.TotalPages)
	}
	if len(result.Orders) != lastPageCount {
		t.Errorf("expected %d orders on last page, got %d", lastPageCount, len(result.Orders))
	}
}

// TestGetOrders_PageBeyondTotal requests a page far beyond the last one
// (page=999). Expects an empty orders list but correct total and total_pages
// reflecting the full dataset.
func TestGetOrders_PageBeyondTotal(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	total := totalOrders(t)
	resp, result := getOrders(t, "?page=999&limit=20")

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(result.Orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(result.Orders))
	}
	// Total and total_pages should still reflect the full dataset.
	if result.Total != total {
		t.Errorf("expected total=%d, got %d", total, result.Total)
	}
	expectedPages := (total + 19) / 20 // ceil(total / 20)
	if result.TotalPages != expectedPages {
		t.Errorf("expected total_pages=%d, got %d", expectedPages, result.TotalPages)
	}
}

// --- Filter Tests ---

// TestGetOrders_FilterByStatusPending filters by status=pending and verifies
// every returned order has status "pending" and that at least one exists.
func TestGetOrders_FilterByStatusPending(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	resp, result := getOrders(t, "?status=pending&limit=100")

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result.Total == 0 {
		t.Fatal("expected pending orders, got 0")
	}
	for i, o := range result.Orders {
		if o.Status != "pending" {
			t.Errorf("order[%d] status=%q, expected pending", i, o.Status)
		}
	}
}

// TestGetOrders_FilterByStatusCancelled filters by status=cancelled, checks
// all returned orders have that status, and confirms the count is a strict
// subset of the full dataset.
func TestGetOrders_FilterByStatusCancelled(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	resp, result := getOrders(t, "?status=cancelled&limit=100")

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result.Total == 0 {
		t.Fatal("expected cancelled orders, got 0")
	}
	for i, o := range result.Orders {
		if o.Status != "cancelled" {
			t.Errorf("order[%d] status=%q, expected cancelled", i, o.Status)
		}
	}
	// Cancelled orders are fewer than the full dataset.
	total := totalOrders(t)
	if result.Total >= total {
		t.Errorf("expected fewer than %d cancelled orders, got %d", total, result.Total)
	}
}

// TestGetOrders_FilterByAmountRange filters with amount_min=100 and
// amount_max=200. Verifies every returned order's total falls within the
// range and the result is a subset of all orders.
func TestGetOrders_FilterByAmountRange(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	resp, result := getOrders(t, "?amount_min=100&amount_max=200&limit=100")

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result.Total == 0 {
		t.Fatal("expected orders in amount range, got 0")
	}
	for i, o := range result.Orders {
		if o.Total < 100 || o.Total > 200 {
			t.Errorf("order[%d] total=%.2f, expected between 100 and 200", i, o.Total)
		}
	}
	// Filtered count should be less than all orders.
	total := totalOrders(t)
	if result.Total >= total {
		t.Errorf("expected fewer than %d orders in range, got %d", total, result.Total)
	}
}

// TestGetOrders_FilterByDateRange filters orders by a 5-day window
// (NOW-10d to NOW-5d). Verifies that results are non-empty and represent
// a subset of the full dataset.
func TestGetOrders_FilterByDateRange(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	// Seed data spans NOW()-1d to NOW()-30d. Pick a window in the middle.
	now := time.Now()
	dateFrom := now.AddDate(0, 0, -10).Format("2006-01-02")
	dateTo := now.AddDate(0, 0, -5).Format("2006-01-02")

	query := fmt.Sprintf("?date_from=%s&date_to=%s&limit=100", dateFrom, dateTo)
	resp, result := getOrders(t, query)

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result.Total == 0 {
		t.Fatal("expected orders in date range, got 0")
	}
	// Should be a subset, not the full dataset.
	total := totalOrders(t)
	if result.Total >= total {
		t.Errorf("expected fewer than %d orders in date range, got %d", total, result.Total)
	}
}

// TestGetOrders_CombinedFilters applies status=pending, amount_min=50, and
// limit=5 simultaneously. Verifies all returned orders satisfy both filters,
// the page size is respected, and total >= returned count.
func TestGetOrders_CombinedFilters(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	resp, result := getOrders(t, "?status=pending&amount_min=50&limit=5")

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	for i, o := range result.Orders {
		if o.Status != "pending" {
			t.Errorf("order[%d] status=%q, expected pending", i, o.Status)
		}
		if o.Total < 50 {
			t.Errorf("order[%d] total=%.2f, expected >= 50", i, o.Total)
		}
	}
	if len(result.Orders) > 5 {
		t.Errorf("expected at most 5 orders (limit), got %d", len(result.Orders))
	}
	// Pagination metadata should reflect the filtered total, not the page size.
	if result.Total < len(result.Orders) {
		t.Errorf("total (%d) must be >= returned orders (%d)", result.Total, len(result.Orders))
	}
}

// --- Invalid Input Tests ---

// TestGetOrders_InvalidStatus sends an unrecognized status value ("shipped")
// and expects a 400 Bad Request with an error message.
func TestGetOrders_InvalidStatus(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	resp, errResp := getOrdersError(t, "?status=shipped")

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response")
	}
}

// TestGetOrders_InvalidPage sends a non-numeric page value ("abc") and
// expects a 400 Bad Request with an error message.
func TestGetOrders_InvalidPage(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	resp, errResp := getOrdersError(t, "?page=abc")

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response")
	}
}

// TestGetOrders_InvalidAmountMin sends a non-numeric amount_min ("notanumber")
// and expects a 400 Bad Request with an error message.
func TestGetOrders_InvalidAmountMin(t *testing.T) {
	if !apiAvailable(t) {
		return
	}

	resp, errResp := getOrdersError(t, "?amount_min=notanumber")

	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if errResp.Error == "" {
		t.Error("expected error message in response")
	}
}
