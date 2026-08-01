package database

import (
	"net/url"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newDryRunDB returns a gorm DB that generates SQL without connecting to a
// database (no ping, dry-run sessions only). This lets the Where* filter
// helpers be verified against the SQL they actually produce.
func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open("host=localhost user=none dbname=none"), &gorm.Config{
		DisableAutomaticPing: true,
		DryRun:               true,
	})
	if err != nil {
		t.Fatalf("failed to open dry-run gorm DB: %v", err)
	}
	return db
}

// buildSQL applies a filter scope to a users query and returns the generated
// SQL with variables inlined.
func buildSQL(t *testing.T, filter func(tx *gorm.DB) *gorm.DB) string {
	t.Helper()
	db := newDryRunDB(t)
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var users []map[string]interface{}
		return tx.Table("users").Scopes(filter).Find(&users)
	})
	if sql == "" {
		t.Fatal("ToSQL returned empty SQL")
	}
	return sql
}

// TestBuildDSN verifies the connection URL is assembled from the POSTGRES_*
// env vars with credentials escaped, so passwords containing special
// characters survive the round trip.
func TestBuildDSN(t *testing.T) {
	// Scenario: a dev database with a password containing '+' and '@'.
	t.Setenv("POSTGRES_HOST", "db-dev.example.net")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "vectopus")
	t.Setenv("POSTGRES_PASS", "s3cr3t+pa@ss")
	t.Setenv("POSTGRES_DB", "postgres")

	got := buildDSN()
	want := "postgres://vectopus:s3cr3t+pa%40ss@db-dev.example.net:5432/postgres?sslmode=disable"
	if got != want {
		t.Errorf("buildDSN() = %q, want %q", got, want)
	}
}

// TestBuildDSN_SpaceInPassword verifies a password containing a space is
// percent-encoded (%20), not query-encoded as '+', which userinfo parsing
// would treat as a literal plus.
func TestBuildDSN_SpaceInPassword(t *testing.T) {
	// Scenario: an operator password with an embedded space.
	t.Setenv("POSTGRES_HOST", "db-dev.example.net")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "vectopus")
	t.Setenv("POSTGRES_PASS", "my secret pass")
	t.Setenv("POSTGRES_DB", "postgres")

	got := buildDSN()
	want := "postgres://vectopus:my%20secret%20pass@db-dev.example.net:5432/postgres?sslmode=disable"
	if got != want {
		t.Errorf("buildDSN() = %q, want %q", got, want)
	}
}

// TestBuildDSN_TrimsWhitespace verifies stray whitespace in env values -
// common in hand-edited .env files and fatal in URLs - is stripped.
func TestBuildDSN_TrimsWhitespace(t *testing.T) {
	// Scenario: the host value carries a trailing space from the .env file.
	t.Setenv("POSTGRES_HOST", "db-dev.example.net ")
	t.Setenv("POSTGRES_PORT", " 5432")
	t.Setenv("POSTGRES_USER", "vectopus")
	t.Setenv("POSTGRES_PASS", "s3cr3t")
	t.Setenv("POSTGRES_DB", "postgres")

	got := buildDSN()
	want := "postgres://vectopus:s3cr3t@db-dev.example.net:5432/postgres?sslmode=disable"
	if got != want {
		t.Errorf("buildDSN() = %q, want %q", got, want)
	}
}

// TestNewDatabaseService_RedactsCredentialsInErrors verifies a connection
// error never carries the password, since driver parse errors echo the
// full DSN.
func TestNewDatabaseService_RedactsCredentialsInErrors(t *testing.T) {
	// Scenario: an unparseable host forces a driver-level DSN error.
	t.Setenv("POSTGRES_HOST", "bad host.example.net")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "vectopus")
	t.Setenv("POSTGRES_PASS", "s3cr3t+topsecret")
	t.Setenv("POSTGRES_DB", "postgres")

	_, err := NewDatabaseService()
	if err == nil {
		t.Fatal("NewDatabaseService() with invalid host: expected error, got nil")
	}
	if strings.Contains(err.Error(), "topsecret") || strings.Contains(err.Error(), url.QueryEscape("s3cr3t+topsecret")) {
		t.Errorf("connection error leaks the password: %q", err.Error())
	}
}

// TestWhere verifies the equality filter produces a WHERE column = value
// clause.
func TestWhere(t *testing.T) {
	// Scenario: contributor lookup by username.
	sql := buildSQL(t, Where("username", "iconify"))
	if !strings.Contains(sql, `username = 'iconify'`) {
		t.Errorf("Where() SQL = %q, want it to contain username = 'iconify'", sql)
	}
}

// TestWhereNot verifies the negation filter excludes the given value.
func TestWhereNot(t *testing.T) {
	// Scenario: exclude a specific contributor account. gorm renders
	// map-based Not as a <> comparison.
	sql := buildSQL(t, WhereNot("username", "iconify"))
	if !strings.Contains(sql, `"username" <> 'iconify'`) {
		t.Errorf("WhereNot() SQL = %q, want username <> 'iconify'", sql)
	}
}

// TestWhereIn verifies the IN filter includes every listed value.
func TestWhereIn(t *testing.T) {
	// Scenario: fetch a batch of contributors by username.
	sql := buildSQL(t, WhereIn("username", []interface{}{"iconify", "vectopus"}))
	if !strings.Contains(sql, "username IN") {
		t.Errorf("WhereIn() SQL = %q, want a username IN clause", sql)
	}
	if !strings.Contains(sql, "'iconify'") || !strings.Contains(sql, "'vectopus'") {
		t.Errorf("WhereIn() SQL = %q, want both usernames inlined", sql)
	}
}

// TestWhereLike verifies the LIKE filter passes the pattern through.
func TestWhereLike(t *testing.T) {
	// Scenario: prefix search over contributor emails.
	sql := buildSQL(t, WhereLike("email", "scott@%"))
	if !strings.Contains(sql, "email LIKE 'scott@%'") {
		t.Errorf("WhereLike() SQL = %q, want email LIKE 'scott@%%'", sql)
	}
}

// TestWhereCustom verifies an arbitrary condition and its arguments are
// rendered into the query.
func TestWhereCustom(t *testing.T) {
	// Scenario: contributors created after a cutoff, by raw condition.
	sql := buildSQL(t, WhereCustom("created_at > ?", "2024-01-15"))
	if !strings.Contains(sql, "created_at > '2024-01-15'") {
		t.Errorf("WhereCustom() SQL = %q, want created_at > '2024-01-15'", sql)
	}
}
