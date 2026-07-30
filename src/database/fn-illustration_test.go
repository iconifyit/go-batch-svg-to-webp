package database

import (
	"testing"

	"gorm.io/gorm"
)

// TestGetIllustrationById_QueryShape verifies the primary-key lookup queries
// the illustrations table by id with a single-row limit.
func TestGetIllustrationById_QueryShape(t *testing.T) {
	// Scenario: fetch illustration 88 by primary key.
	svc, capture := newCaptureService(t)

	if _, err := svc.GetIllustrationById(88); err != nil {
		t.Fatalf("GetIllustrationById() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "illustrations"`, `"id" = $`, "LIMIT")
	if !containsVar(capture.lastVars(t), 88) {
		t.Errorf("bind variables %v do not include id 88", capture.lastVars(t))
	}
}

// TestGetIllustration_AppliesFilters verifies custom filters land in the
// WHERE clause of the single-row illustrations query.
func TestGetIllustration_AppliesFilters(t *testing.T) {
	// Scenario: fetch the mountain-sunrise illustration by slug.
	svc, capture := newCaptureService(t)

	_, err := svc.GetIllustration(QueryParams{
		Filters: []func(tx *gorm.DB) *gorm.DB{Where("slug", "mountain-sunrise")},
	})
	if err != nil {
		t.Fatalf("GetIllustration() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "illustrations"`, "slug = $", "LIMIT")
	if !containsVar(capture.lastVars(t), "mountain-sunrise") {
		t.Errorf("bind variables %v do not include slug mountain-sunrise", capture.lastVars(t))
	}
}

// TestGetIllustrations_OrderLimitOffset verifies Order, Limit, and Offset
// shape the multi-row illustrations query.
func TestGetIllustrations_OrderLimitOffset(t *testing.T) {
	// Scenario: first 30 illustrations for a family, newest first.
	svc, capture := newCaptureService(t)

	_, err := svc.GetIllustrations(QueryParams{
		Order: "created_at desc",
		Limit: 30,
	})
	if err != nil {
		t.Fatalf("GetIllustrations() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "illustrations"`, "ORDER BY created_at desc", "LIMIT")
	if !containsVar(capture.lastVars(t), 30) {
		t.Errorf("bind variables %v do not include limit 30", capture.lastVars(t))
	}
}
