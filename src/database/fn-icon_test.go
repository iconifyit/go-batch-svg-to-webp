package database

import (
	"testing"

	"gorm.io/gorm"
)

// TestGetIconById_QueryShape verifies the primary-key lookup queries the
// icons table by id with a single-row limit.
func TestGetIconById_QueryShape(t *testing.T) {
	// Scenario: fetch icon 314 by primary key.
	svc, capture := newCaptureService(t)

	if _, err := svc.GetIconById(314); err != nil {
		t.Fatalf("GetIconById() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "icons"`, `"id" = $`, "LIMIT")
	if !containsVar(capture.lastVars(t), 314) {
		t.Errorf("bind variables %v do not include id 314", capture.lastVars(t))
	}
}

// TestGetIcon_AppliesFilters verifies custom filters land in the WHERE
// clause of the single-row icons query.
func TestGetIcon_AppliesFilters(t *testing.T) {
	// Scenario: fetch the coffee-cup icon by slug.
	svc, capture := newCaptureService(t)

	_, err := svc.GetIcon(QueryParams{
		Filters: []func(tx *gorm.DB) *gorm.DB{Where("slug", "coffee-cup")},
	})
	if err != nil {
		t.Fatalf("GetIcon() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "icons"`, "slug = $", "LIMIT")
	if !containsVar(capture.lastVars(t), "coffee-cup") {
		t.Errorf("bind variables %v do not include slug coffee-cup", capture.lastVars(t))
	}
}

// TestGetIcons_OrderLimitOffset verifies Order, Limit, and Offset shape the
// multi-row icons query.
func TestGetIcons_OrderLimitOffset(t *testing.T) {
	// Scenario: page 2 of 50 icons in a set, ordered by name.
	svc, capture := newCaptureService(t)

	_, err := svc.GetIcons(QueryParams{
		Order:  "name asc",
		Limit:  50,
		Offset: 50,
	})
	if err != nil {
		t.Fatalf("GetIcons() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "icons"`, "ORDER BY name asc", "LIMIT", "OFFSET")
	if !containsVar(capture.lastVars(t), 50) {
		t.Errorf("bind variables %v do not include limit/offset 50", capture.lastVars(t))
	}
}
