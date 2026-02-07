package dbservice

import (
	"testing"
	"time"

	"orders-management-api/models"
)

func TestBuildWhereClause(t *testing.T) {
	t.Run("no filters returns empty where", func(t *testing.T) {
		filters := models.OrderFilters{}
		where, args := buildWhereClause(filters)
		if where != "" {
			t.Errorf("expected empty where, got %q", where)
		}
		if len(args) != 0 {
			t.Errorf("expected no args, got %d", len(args))
		}
	})

	t.Run("status filter", func(t *testing.T) {
		s := "pending"
		filters := models.OrderFilters{Status: &s}
		where, args := buildWhereClause(filters)
		if where != " WHERE status = $1" {
			t.Errorf("unexpected where: %q", where)
		}
		if len(args) != 1 || args[0] != "pending" {
			t.Errorf("unexpected args: %v", args)
		}
	})

	t.Run("min_amount filter", func(t *testing.T) {
		v := 50.0
		filters := models.OrderFilters{MinAmount: &v}
		where, args := buildWhereClause(filters)
		if where != " WHERE total >= $1" {
			t.Errorf("unexpected where: %q", where)
		}
		if len(args) != 1 || args[0] != 50.0 {
			t.Errorf("unexpected args: %v", args)
		}
	})

	t.Run("max_amount filter", func(t *testing.T) {
		v := 200.0
		filters := models.OrderFilters{MaxAmount: &v}
		where, args := buildWhereClause(filters)
		if where != " WHERE total <= $1" {
			t.Errorf("unexpected where: %q", where)
		}
		if len(args) != 1 || args[0] != 200.0 {
			t.Errorf("unexpected args: %v", args)
		}
	})

	t.Run("from_date filter", func(t *testing.T) {
		date := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
		filters := models.OrderFilters{FromDate: &date}
		where, args := buildWhereClause(filters)
		if where != " WHERE created_at >= $1" {
			t.Errorf("unexpected where: %q", where)
		}
		if len(args) != 1 {
			t.Errorf("unexpected args len: %d", len(args))
		}
	})

	t.Run("to_date filter", func(t *testing.T) {
		date := time.Date(2025, 2, 7, 23, 59, 59, 0, time.UTC)
		filters := models.OrderFilters{ToDate: &date}
		where, args := buildWhereClause(filters)
		if where != " WHERE created_at <= $1" {
			t.Errorf("unexpected where: %q", where)
		}
		if len(args) != 1 {
			t.Errorf("unexpected args len: %d", len(args))
		}
	})

	t.Run("status and min_amount combined", func(t *testing.T) {
		s := "completed"
		v := 100.0
		filters := models.OrderFilters{Status: &s, MinAmount: &v}
		where, args := buildWhereClause(filters)
		if where != " WHERE status = $1 AND total >= $2" {
			t.Errorf("unexpected where: %q", where)
		}
		if len(args) != 2 {
			t.Errorf("expected 2 args, got %d", len(args))
		}
	})

	t.Run("min_amount and max_amount range", func(t *testing.T) {
		min := 50.0
		max := 150.0
		filters := models.OrderFilters{MinAmount: &min, MaxAmount: &max}
		where, args := buildWhereClause(filters)
		if where != " WHERE total >= $1 AND total <= $2" {
			t.Errorf("unexpected where: %q", where)
		}
		if len(args) != 2 {
			t.Errorf("expected 2 args, got %d", len(args))
		}
	})

	t.Run("from_date and to_date range", func(t *testing.T) {
		from := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 2, 7, 23, 59, 59, 0, time.UTC)
		filters := models.OrderFilters{FromDate: &from, ToDate: &to}
		where, args := buildWhereClause(filters)
		if where != " WHERE created_at >= $1 AND created_at <= $2" {
			t.Errorf("unexpected where: %q", where)
		}
		if len(args) != 2 {
			t.Errorf("expected 2 args, got %d", len(args))
		}
	})

	t.Run("all filters combined", func(t *testing.T) {
		s := "pending"
		min := 10.0
		max := 500.0
		from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 2, 28, 23, 59, 59, 0, time.UTC)
		filters := models.OrderFilters{
			Status: &s, MinAmount: &min, MaxAmount: &max,
			FromDate: &from, ToDate: &to,
		}
		where, args := buildWhereClause(filters)
		expected := " WHERE status = $1 AND total >= $2 AND total <= $3 AND created_at >= $4 AND created_at <= $5"
		if where != expected {
			t.Errorf("unexpected where: %q", where)
		}
		if len(args) != 5 {
			t.Errorf("expected 5 args, got %d", len(args))
		}
	})

	t.Run("empty status is ignored", func(t *testing.T) {
		s := ""
		filters := models.OrderFilters{Status: &s}
		where, args := buildWhereClause(filters)
		if where != "" {
			t.Errorf("empty status should be ignored, got where: %q", where)
		}
		if len(args) != 0 {
			t.Errorf("expected no args for empty status, got %d", len(args))
		}
	})

	// Edge cases
	t.Run("edge: zero min_amount", func(t *testing.T) {
		v := 0.0
		filters := models.OrderFilters{MinAmount: &v}
		where, args := buildWhereClause(filters)
		if where != " WHERE total >= $1" {
			t.Errorf("zero min_amount should be valid, got %q", where)
		}
		if len(args) != 1 || args[0] != 0.0 {
			t.Errorf("unexpected args: %v", args)
		}
	})

	t.Run("edge: negative max_amount", func(t *testing.T) {
		v := -10.0
		filters := models.OrderFilters{MaxAmount: &v}
		where, args := buildWhereClause(filters)
		if where != " WHERE total <= $1" {
			t.Errorf("negative max_amount should be accepted, got %q", where)
		}
		if args[0] != -10.0 {
			t.Errorf("unexpected args: %v", args)
		}
	})

	t.Run("edge: nil filters struct", func(t *testing.T) {
		var filters models.OrderFilters
		where, args := buildWhereClause(filters)
		if where != "" || len(args) != 0 {
			t.Errorf("nil filters should return empty, got where=%q args=%v", where, args)
		}
	})
}
