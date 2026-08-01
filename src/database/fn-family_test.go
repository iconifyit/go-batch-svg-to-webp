package database

import (
	"testing"

	"gorm.io/gorm"
)

// TestGetFamilyById_QueryShape verifies the primary-key lookup queries the
// families table by id with a single-row limit.
func TestGetFamilyById_QueryShape(t *testing.T) {
	// Scenario: fetch product family 7 by primary key.
	svc, capture := newCaptureService(t)

	if _, err := svc.GetFamilyById(7); err != nil {
		t.Fatalf("GetFamilyById() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "families"`, `"id" = $`, "LIMIT")
	if !containsVar(capture.lastVars(t), 7) {
		t.Errorf("bind variables %v do not include id 7", capture.lastVars(t))
	}
}

// TestGetFamily_AppliesFiltersAndOrder verifies filters and the Order param
// shape the single-row families query.
func TestGetFamily_AppliesFiltersAndOrder(t *testing.T) {
	// Scenario: fetch the diversity-avatars family by unique_id, newest first.
	svc, capture := newCaptureService(t)

	_, err := svc.GetFamily(&QueryParams{
		Filters: []func(tx *gorm.DB) *gorm.DB{Where("unique_id", "2C11DB2D5F79")},
		Order:   "created_at desc",
	})
	if err != nil {
		t.Fatalf("GetFamily() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "families"`, "unique_id = $", "ORDER BY created_at desc", "LIMIT")
	if !containsVar(capture.lastVars(t), "2C11DB2D5F79") {
		t.Errorf("bind variables %v do not include the family unique id", capture.lastVars(t))
	}
}

// TestGetFamilies_LimitAndOffset verifies Order, Limit, and Offset shape the
// multi-row families query.
func TestGetFamilies_LimitAndOffset(t *testing.T) {
	// Scenario: second page of 10 families for a contributor listing.
	svc, capture := newCaptureService(t)

	_, err := svc.GetFamilies(QueryParams{
		Order:  "name asc",
		Limit:  10,
		Offset: 10,
	})
	if err != nil {
		t.Fatalf("GetFamilies() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "families"`, "ORDER BY name asc", "LIMIT", "OFFSET")
	vars := capture.lastVars(t)
	if !containsVar(vars, 10) {
		t.Errorf("bind variables %v do not include limit/offset 10", vars)
	}
}
