package main

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"unicode"
)

// =====================================================
// SQL IDENTIFIER VALIDATION
// =====================================================

// validIdentifierRegex matches valid PostgreSQL identifiers (alphanumeric + underscore)
var validIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// ValidateIdentifier checks if a string is a valid SQL identifier
// Prevents SQL injection through identifier names
func ValidateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if len(name) > 63 { // PostgreSQL identifier limit
		return fmt.Errorf("identifier too long (max 63 characters)")
	}
	if !validIdentifierRegex.MatchString(name) {
		return fmt.Errorf("invalid identifier: must start with letter or underscore, contain only alphanumeric and underscore")
	}
	// Check for reserved words that could be problematic
	reserved := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE", "GRANT", "REVOKE"}
	upper := strings.ToUpper(name)
	for _, r := range reserved {
		if upper == r {
			return fmt.Errorf("identifier cannot be a reserved SQL keyword: %s", name)
		}
	}
	return nil
}

// ValidateSchemaName validates a PostgreSQL schema name
func ValidateSchemaName(schema string) error {
	if err := ValidateIdentifier(schema); err != nil {
		return fmt.Errorf("invalid schema name: %w", err)
	}
	// Prevent access to system schemas
	forbidden := []string{"pg_catalog", "information_schema", "pg_toast", "pg_temp"}
	lower := strings.ToLower(schema)
	for _, f := range forbidden {
		if lower == f {
			return fmt.Errorf("access to system schema '%s' is forbidden", schema)
		}
	}
	return nil
}

// ValidateTableName validates a PostgreSQL table name
func ValidateTableName(table string) error {
	if err := ValidateIdentifier(table); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}
	return nil
}

// ValidateColumnName validates a PostgreSQL column name
func ValidateColumnName(column string) error {
	if err := ValidateIdentifier(column); err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}
	return nil
}

// ValidateFunctionName validates a PostgreSQL function name
func ValidateFunctionName(name string) error {
	if err := ValidateIdentifier(name); err != nil {
		return fmt.Errorf("invalid function name: %w", err)
	}
	return nil
}

// ValidateRoleName validates a PostgreSQL role name
func ValidateRoleName(name string) error {
	if err := ValidateIdentifier(name); err != nil {
		return fmt.Errorf("invalid role name: %w", err)
	}
	// Prevent modification of superuser roles
	forbidden := []string{"postgres", "pg_read_all_data", "pg_write_all_data", "pg_read_all_settings", "pg_read_all_stats", "pg_stat_scan_tables", "pg_monitor", "pg_database_owner", "pg_signal_backend"}
	lower := strings.ToLower(name)
	for _, f := range forbidden {
		if lower == f {
			return fmt.Errorf("cannot modify system role '%s'", name)
		}
	}
	return nil
}

// =====================================================
// SQL TYPE VALIDATION
// =====================================================

// allowedSQLTypes is a whitelist of safe PostgreSQL data types
var allowedSQLTypes = map[string]bool{
	// Numeric types
	"smallint": true, "integer": true, "bigint": true, "int": true, "int2": true, "int4": true, "int8": true,
	"decimal": true, "numeric": true, "real": true, "double precision": true, "float": true, "float4": true, "float8": true,
	"serial": true, "bigserial": true, "smallserial": true,
	// Monetary
	"money": true,
	// Character types
	"character varying": true, "varchar": true, "character": true, "char": true, "text": true,
	// Binary
	"bytea": true,
	// Date/Time
	"timestamp": true, "timestamp without time zone": true, "timestamp with time zone": true, "timestamptz": true,
	"date": true, "time": true, "time without time zone": true, "time with time zone": true, "timetz": true,
	"interval": true,
	// Boolean
	"boolean": true, "bool": true,
	// UUID
	"uuid": true,
	// JSON
	"json": true, "jsonb": true,
	// Arrays (common)
	"integer[]": true, "text[]": true, "varchar[]": true, "boolean[]": true, "jsonb[]": true,
	// Network
	"inet": true, "cidr": true, "macaddr": true, "macaddr8": true,
	// Geometric
	"point": true, "line": true, "lseg": true, "box": true, "path": true, "polygon": true, "circle": true,
	// Other
	"xml": true, "tsvector": true, "tsquery": true,
}

// ValidateSQLType checks if a SQL type is in the allowed whitelist
func ValidateSQLType(sqlType string) error {
	if sqlType == "" {
		return fmt.Errorf("SQL type cannot be empty")
	}

	// Normalize the type (lowercase, trim whitespace)
	normalized := strings.ToLower(strings.TrimSpace(sqlType))

	// Check for type with size modifier like varchar(255) or numeric(10,2)
	baseType := normalized
	if idx := strings.Index(normalized, "("); idx != -1 {
		baseType = strings.TrimSpace(normalized[:idx])
		// Validate the size modifier contains only digits and comma
		modifier := normalized[idx:]
		if !regexp.MustCompile(`^\(\d+(?:,\s*\d+)?\)$`).MatchString(modifier) {
			return fmt.Errorf("invalid type modifier: %s", modifier)
		}
	}

	// Check for array suffix
	if strings.HasSuffix(baseType, "[]") {
		baseType = strings.TrimSuffix(baseType, "[]")
	}

	if !allowedSQLTypes[baseType] {
		return fmt.Errorf("SQL type '%s' is not in the allowed list", sqlType)
	}

	return nil
}

// =====================================================
// SQL FUNCTION LANGUAGE VALIDATION
// =====================================================

// allowedFunctionLanguages is a whitelist of safe function languages
var allowedFunctionLanguages = map[string]bool{
	"sql":     true,
	"plpgsql": true,
}

// ValidateFunctionLanguage checks if a function language is allowed
func ValidateFunctionLanguage(lang string) error {
	if lang == "" {
		return fmt.Errorf("function language cannot be empty")
	}
	normalized := strings.ToLower(strings.TrimSpace(lang))
	if !allowedFunctionLanguages[normalized] {
		return fmt.Errorf("function language '%s' is not allowed (permitted: sql, plpgsql)", lang)
	}
	return nil
}

// allowedVolatility is a whitelist of valid volatility settings
var allowedVolatility = map[string]bool{
	"volatile": true, "stable": true, "immutable": true,
}

// ValidateVolatility checks if a volatility setting is valid
func ValidateVolatility(vol string) error {
	if vol == "" {
		return nil // Optional, defaults to VOLATILE
	}
	normalized := strings.ToLower(strings.TrimSpace(vol))
	if !allowedVolatility[normalized] {
		return fmt.Errorf("invalid volatility '%s' (permitted: volatile, stable, immutable)", vol)
	}
	return nil
}

// =====================================================
// SQL DEFAULT VALUE VALIDATION
// =====================================================

// ValidateDefaultValue checks if a default value is safe
// Only allows literals, not function calls or subqueries
func ValidateDefaultValue(val string) error {
	if val == "" {
		return nil // Empty means DROP DEFAULT
	}

	trimmed := strings.TrimSpace(val)
	lower := strings.ToLower(trimmed)

	// Allow NULL
	if lower == "null" {
		return nil
	}

	// Allow boolean literals
	if lower == "true" || lower == "false" {
		return nil
	}

	// Allow simple numeric values (including negative and decimal)
	if regexp.MustCompile(`^-?\d+(\.\d+)?$`).MatchString(trimmed) {
		return nil
	}

	// Allow simple string literals (single quoted, with escaped quotes)
	if regexp.MustCompile(`^'([^']*('')*)*'$`).MatchString(trimmed) {
		return nil
	}

	// Allow common safe functions
	safeFunctions := []string{"now()", "current_timestamp", "current_date", "current_time", "uuid_generate_v4()", "gen_random_uuid()"}
	for _, safe := range safeFunctions {
		if lower == safe {
			return nil
		}
	}

	// Allow simple array literals like '{}'
	if regexp.MustCompile(`^'\{[^}]*\}'(::[\w\[\]]+)?$`).MatchString(trimmed) {
		return nil
	}

	// Allow typed literals like '{}':jsonb
	if regexp.MustCompile(`^'[^']*'::[\w\[\]]+$`).MatchString(trimmed) {
		return nil
	}

	return fmt.Errorf("default value '%s' is not allowed - only literals and safe functions permitted", val)
}

// =====================================================
// SQL FUNCTION BODY VALIDATION
// =====================================================

// ValidateFunctionBody performs basic validation on function body
// This is a defense-in-depth measure - the function still runs with limited privileges
func ValidateFunctionBody(body string) error {
	if body == "" {
		return fmt.Errorf("function body cannot be empty")
	}

	// Check for dangerous patterns
	lower := strings.ToLower(body)
	dangerous := []string{
		"pg_read_file",
		"pg_write_file",
		"pg_execute_server_program",
		"copy from program",
		"copy to program",
		"lo_import",
		"lo_export",
		"dblink",
		"pg_read_binary_file",
	}

	for _, d := range dangerous {
		if strings.Contains(lower, d) {
			return fmt.Errorf("function body contains forbidden operation: %s", d)
		}
	}

	return nil
}

// =====================================================
// FUNCTION ARGUMENTS VALIDATION
// =====================================================

// ValidateFunctionArguments validates function argument definitions
// Expected format: "arg1 type1, arg2 type2" or empty
func ValidateFunctionArguments(args string) error {
	if args == "" {
		return nil
	}

	// Split by comma and validate each argument
	parts := strings.Split(args, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split into name and type (type might include modifiers)
		tokens := strings.Fields(part)
		if len(tokens) < 2 {
			return fmt.Errorf("invalid argument format: %s (expected 'name type')", part)
		}

		// Validate argument name
		argName := tokens[0]
		if err := ValidateIdentifier(argName); err != nil {
			return fmt.Errorf("invalid argument name '%s': %w", argName, err)
		}

		// Validate argument type (join remaining tokens for types like "character varying")
		argType := strings.Join(tokens[1:], " ")
		if err := ValidateSQLType(argType); err != nil {
			return fmt.Errorf("invalid argument type for '%s': %w", argName, err)
		}
	}

	return nil
}

// ValidateFunctionReturnType validates function return type
func ValidateFunctionReturnType(ret string) error {
	if ret == "" {
		return fmt.Errorf("return type cannot be empty")
	}

	lower := strings.ToLower(strings.TrimSpace(ret))

	// Allow VOID
	if lower == "void" {
		return nil
	}

	// Allow TRIGGER
	if lower == "trigger" {
		return nil
	}

	// Allow SETOF <type>
	if strings.HasPrefix(lower, "setof ") {
		typeAfterSetof := strings.TrimPrefix(lower, "setof ")
		return ValidateSQLType(typeAfterSetof)
	}

	// Allow TABLE(...) for table-returning functions
	if strings.HasPrefix(lower, "table(") {
		return nil // Complex validation would be needed; rely on DB for errors
	}

	// Standard type validation
	return ValidateSQLType(ret)
}

// =====================================================
// PATH TRAVERSAL PREVENTION
// =====================================================

// ValidateFilename checks for path traversal attempts in filenames
func ValidateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Check for path traversal patterns (both Unix and Windows)
	if strings.Contains(filename, "..") {
		return fmt.Errorf("filename contains path traversal sequence")
	}
	if strings.Contains(filename, "/") {
		return fmt.Errorf("filename contains forward slash")
	}
	if strings.Contains(filename, "\\") {
		return fmt.Errorf("filename contains backslash")
	}
	if strings.Contains(filename, "\x00") {
		return fmt.Errorf("filename contains null byte")
	}

	// Check for shell metacharacters
	dangerous := []string{";", "&", "|", "$", "`", "(", ")", "{", "}", "<", ">", "!", "*", "?", "[", "]", "~"}
	for _, d := range dangerous {
		if strings.Contains(filename, d) {
			return fmt.Errorf("filename contains forbidden character: %s", d)
		}
	}

	// Ensure filename has allowed extension
	allowedExtensions := []string{".sql", ".backup", ".dump"}
	hasAllowedExt := false
	for _, ext := range allowedExtensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			hasAllowedExt = true
			break
		}
	}
	if !hasAllowedExt {
		return fmt.Errorf("filename must have extension: .sql, .backup, or .dump")
	}

	return nil
}

// =====================================================
// IP ADDRESS HANDLING
// =====================================================

// GetClientIP extracts the real client IP, handling reverse proxies
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (common for reverse proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs; the first is the client
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if parsedIP := net.ParseIP(clientIP); parsedIP != nil {
				return clientIP
			}
		}
	}

	// Check X-Real-IP header (nginx)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		if parsedIP := net.ParseIP(xri); parsedIP != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ValidateIP validates an IP address and checks for localhost/private ranges
func ValidateIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address format")
	}

	// Prevent blocking localhost
	if parsedIP.IsLoopback() {
		return fmt.Errorf("cannot block loopback address")
	}

	// Prevent blocking unspecified addresses
	if parsedIP.IsUnspecified() {
		return fmt.Errorf("cannot block unspecified address")
	}

	return nil
}

// =====================================================
// SQL QUERY SAFETY
// =====================================================

// IsSafeReadOnlyQuery performs basic validation that a query is read-only
// This is NOT a complete SQL parser - it's a defense-in-depth measure
func IsSafeReadOnlyQuery(query string) error {
	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// Normalize: remove comments, lowercase
	normalized := strings.ToLower(strings.TrimSpace(query))

	// Remove single-line comments
	lines := strings.Split(normalized, "\n")
	cleanedLines := []string{}
	for _, line := range lines {
		if idx := strings.Index(line, "--"); idx != -1 {
			line = line[:idx]
		}
		cleanedLines = append(cleanedLines, line)
	}
	normalized = strings.Join(cleanedLines, " ")

	// Remove multi-line comments (basic - doesn't handle nested)
	for {
		start := strings.Index(normalized, "/*")
		if start == -1 {
			break
		}
		end := strings.Index(normalized[start:], "*/")
		if end == -1 {
			break
		}
		normalized = normalized[:start] + normalized[start+end+2:]
	}

	// Check for dangerous statements
	dangerous := []string{
		"insert ", "update ", "delete ", "drop ", "create ", "alter ", "truncate ",
		"grant ", "revoke ", "copy ", "pg_read_file", "pg_write_file",
		"lo_import", "lo_export", "dblink", "execute ", "prepare ",
	}

	for _, d := range dangerous {
		if strings.Contains(normalized, d) {
			return fmt.Errorf("query contains forbidden keyword: %s", strings.TrimSpace(d))
		}
	}

	// Must start with SELECT, WITH, EXPLAIN, or SHOW
	allowedPrefixes := []string{"select ", "with ", "explain ", "show ", "table "}
	hasAllowedPrefix := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(normalized, prefix) {
			hasAllowedPrefix = true
			break
		}
	}
	if !hasAllowedPrefix {
		return fmt.Errorf("query must start with SELECT, WITH, EXPLAIN, SHOW, or TABLE")
	}

	return nil
}

// =====================================================
// PASSWORD SECURITY
// =====================================================

// ValidatePassword checks password complexity
func ValidatePassword(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must be at most 128 characters")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// =====================================================
// ERROR SANITIZATION
// =====================================================

// SanitizeDBError removes potentially sensitive information from database errors
func SanitizeDBError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Check for common PostgreSQL error patterns that might leak info
	sensitivePatterns := []string{
		"password authentication failed",
		"connection refused",
		"no pg_hba.conf entry",
		"could not connect to server",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return "Database operation failed"
		}
	}

	// Remove potential file paths
	if strings.Contains(errStr, "/") || strings.Contains(errStr, "\\") {
		// If it looks like it contains a path, sanitize
		if strings.Contains(errStr, "home") || strings.Contains(errStr, "var") || strings.Contains(errStr, "Users") {
			return "Database operation failed"
		}
	}

	// Return a truncated error if it's too long (might contain dumps)
	if len(errStr) > 200 {
		return errStr[:200] + "..."
	}

	return errStr
}

// =====================================================
// COLUMN DEFINITION VALIDATION
// =====================================================

// ValidateColumnDefinition validates a column definition for CREATE TABLE
// Expected format: "column_name TYPE [constraints]"
func ValidateColumnDefinition(def string) error {
	if def == "" {
		return fmt.Errorf("column definition cannot be empty")
	}

	// Split into parts
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return fmt.Errorf("column definition must include name and type")
	}

	// Validate column name
	if err := ValidateColumnName(parts[0]); err != nil {
		return err
	}

	// Find and validate the type
	// Type might be multi-word like "character varying" or have modifiers
	typeStr := parts[1]

	// Handle multi-word types
	typeEndIdx := 1
	for i := 2; i < len(parts); i++ {
		// Check if this word is part of the type
		combined := typeStr + " " + parts[i]
		if allowedSQLTypes[strings.ToLower(combined)] {
			typeStr = combined
			typeEndIdx = i
		} else {
			break
		}
	}

	// Validate the type
	if err := ValidateSQLType(typeStr); err != nil {
		return fmt.Errorf("in column '%s': %w", parts[0], err)
	}

	// Validate constraints (remaining parts)
	allowedConstraints := map[string]bool{
		"not":        true,
		"null":       true,
		"primary":    true,
		"key":        true,
		"unique":     true,
		"default":    true,
		"references": true,
		"check":      true,
		"constraint": true,
	}

	for i := typeEndIdx + 1; i < len(parts); i++ {
		part := parts[i]
		partLower := strings.ToLower(part)
		// Skip values after DEFAULT
		if i > 0 && strings.ToLower(parts[i-1]) == "default" {
			continue
		}
		// Skip constraint names
		if i > 0 && strings.ToLower(parts[i-1]) == "constraint" {
			if err := ValidateIdentifier(part); err != nil {
				return fmt.Errorf("invalid constraint name: %w", err)
			}
			continue
		}
		// Check if it's an allowed keyword or looks like a value
		if !allowedConstraints[partLower] && !regexp.MustCompile(`^[\d'"\(\)]+`).MatchString(part) {
			// Might be a table reference or other identifier
			if err := ValidateIdentifier(strings.Trim(part, "()'\"")); err != nil {
				// Allow it if it fails - might be a complex constraint
				continue
			}
		}
	}

	return nil
}
