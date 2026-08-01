package database

import (
	"testing"

	"gorm.io/gorm"
)

// TestGetUserById_QueryShape verifies the primary-key lookup queries the
// users table by id with a single-row limit.
func TestGetUserById_QueryShape(t *testing.T) {
	// Scenario: fetch contributor account 42 by primary key.
	svc, capture := newCaptureService(t)

	if _, err := svc.GetUserById(42); err != nil {
		t.Fatalf("GetUserById() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "users"`, `"users"."id" = $`, "LIMIT")
	if !containsVar(capture.lastVars(t), 42) {
		t.Errorf("bind variables %v do not include id 42", capture.lastVars(t))
	}
}

// TestGetUser_AppliesFilters verifies custom filters land in the WHERE
// clause of a single-row users query.
func TestGetUser_AppliesFilters(t *testing.T) {
	// Scenario: contributor validation looks up the user named iconify.
	svc, capture := newCaptureService(t)

	_, err := svc.GetUser(&QueryParams{
		Filters: []func(tx *gorm.DB) *gorm.DB{Where("username", "iconify")},
	})
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "users"`, "username = $", "LIMIT")
	if !containsVar(capture.lastVars(t), "iconify") {
		t.Errorf("bind variables %v do not include username iconify", capture.lastVars(t))
	}
}

// TestGetUsers_SortAndPagination verifies SortBy/SortOrder, Limit, and
// Offset all shape the users query.
func TestGetUsers_SortAndPagination(t *testing.T) {
	// Scenario: page 3 of contributors sorted by newest username first.
	svc, capture := newCaptureService(t)

	_, err := svc.GetUsers(&QueryParams{
		SortBy:    "username",
		SortOrder: "desc",
		Limit:     25,
		Offset:    50,
	})
	if err != nil {
		t.Fatalf("GetUsers() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "users"`, "ORDER BY username desc", "LIMIT", "OFFSET")
	vars := capture.lastVars(t)
	if !containsVar(vars, 25) || !containsVar(vars, 50) {
		t.Errorf("bind variables %v do not include limit 25 and offset 50", vars)
	}
}
