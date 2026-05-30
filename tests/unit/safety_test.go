package safety_test

import (
	"testing"

	"github.com/readall/risinggo/internal/safety"
)

func TestValidateReadOnlyQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantValid bool
	}{
		{"simple select", "SELECT * FROM users", true},
		{"select with limit", "SELECT id, name FROM users LIMIT 10", true},
		{"with clause", "WITH cte AS (SELECT 1) SELECT * FROM cte", true},
		{"explain", "EXPLAIN SELECT * FROM users", true},
		{"show tables", "SHOW TABLES", true},
		{"insert rejected", "INSERT INTO users VALUES (1)", false},
		{"update rejected", "UPDATE users SET name = 'x'", false},
		{"delete rejected", "DELETE FROM users", false},
		{"drop rejected", "DROP TABLE users", false},
		{"create rejected", "CREATE TABLE test (id int)", false},
		{"alter rejected", "ALTER TABLE users ADD col int", false},
		{"grant rejected", "GRANT SELECT ON users TO public", false},
		{"multiple statements rejected", "SELECT * FROM users; DROP TABLE users", false},
		{"comment rejected", "SELECT * FROM users -- DROP TABLE users", false},
		{"empty query rejected", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safety.ValidateReadOnlyQuery(tt.query, 10000, 30)
			if result.Valid != tt.wantValid {
				t.Errorf("ValidateReadOnlyQuery() valid = %v, want %v (reason: %s)", result.Valid, tt.wantValid, result.Reason)
			}
		})
	}
}

func TestContainsMutation(t *testing.T) {
	tests := []struct {
		query string
		want  bool
	}{
		{"SELECT * FROM users", false},
		{"INSERT INTO users VALUES (1)", true},
		{"UPDATE users SET name = 'x'", true},
		{"DELETE FROM users", true},
		{"DROP TABLE users", true},
		{"CREATE TABLE test (id int)", true},
		{"ALTER TABLE users ADD col int", true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			if got := safety.ContainsMutation(tt.query); got != tt.want {
				t.Errorf("ContainsMutation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		wantErr  bool
	}{
		{"users", false},
		{"table_name", false},
		{"TableName", false},
		{"public.users", false},
		{"schema.table_name", false},
		{"", true},
		{"123invalid", true},
		{"table-name", true},
		{"table name", true},
		{"pg_catalog.foo", true},
		{"information_schema.columns", true},
		{"pg_something", true},
		{"very_long_identifier_that_exceeds_the_sixty_three_character_limit_12345", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := safety.SanitizeIdentifier(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeIdentifier(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}