package database

import (
	"testing"

	"gorm.io/gorm"
)

// TestGetSetById_QueryShape verifies the primary-key lookup queries the sets
// table by id with a single-row limit.
func TestGetSetById_QueryShape(t *testing.T) {
	// Scenario: fetch icon set 12 by primary key.
	svc, capture := newCaptureService(t)

	if _, err := svc.GetSetById(12); err != nil {
		t.Fatalf("GetSetById() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "sets"`, `"id" = $`, "LIMIT")
	if !containsVar(capture.lastVars(t), 12) {
		t.Errorf("bind variables %v do not include id 12", capture.lastVars(t))
	}
}

// TestGetSet_AppliesFilters verifies custom filters land in the WHERE clause
// of the single-row sets query.
func TestGetSet_AppliesFilters(t *testing.T) {
	// Scenario: fetch a set by its 12-character unique id.
	svc, capture := newCaptureService(t)

	_, err := svc.GetSet(&QueryParams{
		Filters: []func(tx *gorm.DB) *gorm.DB{Where("unique_id", "B24091F3DF3E")},
	})
	if err != nil {
		t.Fatalf("GetSet() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "sets"`, "unique_id = $", "LIMIT")
	if !containsVar(capture.lastVars(t), "B24091F3DF3E") {
		t.Errorf("bind variables %v do not include the set unique id", capture.lastVars(t))
	}
}

// TestGetSets_FiltersAndPagination verifies filters combine with Limit and
// Offset in the multi-row sets query.
func TestGetSets_FiltersAndPagination(t *testing.T) {
	// Scenario: first 20 sets belonging to family 7.
	svc, capture := newCaptureService(t)

	_, err := svc.GetSets(&QueryParams{
		Filters: []func(tx *gorm.DB) *gorm.DB{Where("family_id", 7)},
		Limit:   20,
	})
	if err != nil {
		t.Fatalf("GetSets() error = %v", err)
	}

	sql := capture.last(t)
	mustContain(t, sql, `FROM "sets"`, "family_id = $", "LIMIT")
	vars := capture.lastVars(t)
	if !containsVar(vars, 7) || !containsVar(vars, 20) {
		t.Errorf("bind variables %v do not include family id 7 and limit 20", vars)
	}
}
