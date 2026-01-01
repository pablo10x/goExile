package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// =====================================================
// SQL IDENTIFIER VALIDATION
// =====================================================

var validIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// GenerateRandomKey generates a cryptographically secure random key of specified byte length (hex encoded will be 2x length).
func GenerateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func ValidateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("identifier too long")
	}
	if !validIdentifierRegex.MatchString(name) {
		return fmt.Errorf("invalid identifier")
	}
	reserved := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE", "GRANT", "REVOKE"}
	upper := strings.ToUpper(name)
	for _, r := range reserved {
		if upper == r {
			return fmt.Errorf("reserved word")
		}
	}
	return nil
}

func ValidateSchemaName(schema string) error {
	if err := ValidateIdentifier(schema); err != nil {
		return err
	}
	forbidden := []string{"pg_catalog", "information_schema", "pg_toast", "pg_temp"}
	lower := strings.ToLower(schema)
	for _, f := range forbidden {
		if lower == f {
			return fmt.Errorf("forbidden schema")
		}
	}
	return nil
}

func ValidateTableName(table string) error {
	return ValidateIdentifier(table)
}

func ValidateColumnName(column string) error {
	return ValidateIdentifier(column)
}

func ValidateFunctionName(name string) error {
	return ValidateIdentifier(name)
}

func ValidateRoleName(name string) error {
	if err := ValidateIdentifier(name); err != nil {
		return err
	}
	forbidden := []string{"postgres"}
	lower := strings.ToLower(name)
	for _, f := range forbidden {
		if lower == f {
			return fmt.Errorf("forbidden role")
		}
	}
	return nil
}

var allowedSQLTypes = map[string]bool{
	"integer": true, "text": true, "boolean": true, "timestamp": true, "jsonb": true,
}

func ValidateSQLType(sqlType string) error {
	normalized := strings.ToLower(strings.TrimSpace(sqlType))
	if !allowedSQLTypes[normalized] {
		if len(normalized) > 0 && len(normalized) < 50 {
			return nil
		}
		return fmt.Errorf("invalid type")
	}
	return nil
}

func ValidateFunctionLanguage(lang string) error {
	l := strings.ToLower(lang)
	if l != "sql" && l != "plpgsql" {
		return fmt.Errorf("invalid language")
	}
	return nil
}

func ValidateVolatility(vol string) error {
	l := strings.ToLower(vol)
	if l != "volatile" && l != "stable" && l != "immutable" {
		return fmt.Errorf("invalid volatility")
	}
	return nil
}

func ValidateDefaultValue(val string) error {
	if len(val) > 100 {
		return fmt.Errorf("too long")
	}
	return nil
}

func ValidateFunctionBody(body string) error {
	if len(body) == 0 {
		return fmt.Errorf("empty body")
	}
	return nil
}

func ValidateFunctionArguments(args string) error {
	return nil
}

func ValidateFunctionReturnType(ret string) error {
	return nil
}

func ValidateFilename(filename string) error {
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, `\`) {
		return fmt.Errorf("invalid filename")
	}
	return nil
}

func GetClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func ValidateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid ip")
	}
	return nil
}

func BlockIPSystem(ip string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("ufw is only supported on Linux")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP")
	}
	ipStr := parsedIP.String()
	if err := ValidateIP(ipStr); err != nil {
		return err
	}
	cmd := exec.Command("ufw", "deny", "from", ipStr, "to", "any")
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "root") || strings.Contains(string(output), "permission") {
			cmd = exec.Command("sudo", "ufw", "deny", "from", ipStr, "to", "any")
			if _, err := cmd.CombinedOutput(); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func UnblockIPSystem(ip string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("ufw is only supported on Linux")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP")
	}
	ipStr := parsedIP.String()
	cmd := exec.Command("ufw", "delete", "deny", "from", ipStr, "to", "any")
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "root") || strings.Contains(string(output), "permission") {
			cmd = exec.Command("sudo", "ufw", "delete", "deny", "from", ipStr, "to", "any")
			if _, err := cmd.CombinedOutput(); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func IsSafeReadOnlyQuery(query string) error {
	q := strings.ToLower(query)
	if strings.HasPrefix(q, "select") || strings.HasPrefix(q, "with") {
		return nil
	}
	return fmt.Errorf("not readonly")
}

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("too short")
	}
	return nil
}

func SanitizeDBError(err error) string {
	if err == nil {
		return ""
	}
	return "database error"
}

// Basic regex for column definition: "name TYPE [constraints]"
// Allow alphanumeric, spaces, and some special chars for types like VARCHAR(255)
var validColumnDefRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+(\s+[a-zA-Z0-9_()\[\]]+)+(\s+(NOT NULL|NULL|DEFAULT|PRIMARY KEY|UNIQUE|REFERENCES|CHECK|AUTOINCREMENT))*$`)

func ValidateColumnDefinition(def string) error {
	// Simple validation to prevent obvious injection
	// This is not a full SQL parser, but it blocks dangerous characters
	if strings.ContainsAny(def, ";--") {
		return fmt.Errorf("invalid characters in column definition")
	}
	
	// Ensure it starts with a valid identifier
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return fmt.Errorf("invalid column definition format")
	}
	
	if err := ValidateIdentifier(parts[0]); err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}

	return nil
}
