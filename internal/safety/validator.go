package safety

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	mutationKeywords = []string{
		"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE",
		"GRANT", "REVOKE", "ROLLBACK", "COMMIT", "SET TRANSACTION",
		"CREATE DATABASE", "DROP DATABASE", "CREATE SCHEMA", "DROP SCHEMA",
		"CREATE ROLE", "DROP ROLE", "CREATE USER", "DROP USER",
		"EXECUTE", "CALL", "DO", "DECLARE", "OPEN", "FETCH", "CLOSE",
	}
	dangerousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i);\s*(drop|delete|insert|update|alter|create)`),
		regexp.MustCompile(`(?i)--`),
		regexp.MustCompile(`(?i)/\*`),
	}
	allowedQueryTypes = map[string]bool{
		"SELECT":  true,
		"WITH":    true,
		"EXPLAIN": true,
		"SHOW":    true,
		"DESC":    true,
		"DESCRIBE": true,
	}
)

type ValidationResult struct {
	Valid       bool
	Reason      string
	QueryType   string
}

func ValidateReadOnlyQuery(query string, maxRows int, timeoutSeconds int) ValidationResult {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return ValidationResult{Valid: false, Reason: "empty query"}
	}

	upper := strings.ToUpper(trimmed)
	for _, kw := range mutationKeywords {
		if strings.Contains(upper, kw) {
			return ValidationResult{
				Valid:     false,
				Reason:    fmt.Sprintf("mutation keyword detected: %s", kw),
				QueryType: "mutation",
			}
		}
	}

	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(upper) {
			return ValidationResult{
				Valid:     false,
				Reason:    "dangerous pattern detected (comment or multiple statements)",
				QueryType: "dangerous",
			}
		}
	}

	firstWord := strings.Fields(strings.TrimSpace(upper))[0]
	if !allowedQueryTypes[firstWord] && !strings.HasPrefix(firstWord, "WITH") {
		return ValidationResult{
			Valid:     false,
			Reason:    fmt.Sprintf("query type '%s' not allowed", firstWord),
			QueryType: "unknown",
		}
	}

	return ValidationResult{
		Valid:     true,
		Reason:    "query validated",
		QueryType: firstWord,
	}
}

func ContainsMutation(query string) bool {
	upper := strings.ToUpper(query)
	for _, kw := range mutationKeywords {
		if strings.Contains(upper, kw) {
			return true
		}
	}
	return false
}

func SanitizeIdentifier(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("identifier cannot be empty")
	}
	if len(name) > 63 {
		return "", fmt.Errorf("identifier too long (max 63 chars)")
	}

	// Support schema.table style (validate each segment)
	parts := strings.Split(name, ".")
	for _, part := range parts {
		if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(part) {
			return "", fmt.Errorf("invalid identifier format: %s", name)
		}
	}

	// Basic reserved word protection (common dangerous ones for read-only context)
	lower := strings.ToLower(name)
	reserved := []string{"pg_", "information_schema", "pg_catalog"}
	for _, r := range reserved {
		if strings.HasPrefix(lower, r) {
			return "", fmt.Errorf("identifier uses reserved prefix: %s", name)
		}
	}

	return name, nil
}