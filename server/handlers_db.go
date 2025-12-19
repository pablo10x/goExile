package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const backupDir = "backups"

func EnsureBackupDir() {
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		os.MkdirAll(backupDir, 0755)
	}
}

// --- Overview ---

func GetDatabaseOverviewHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	stats := make(map[string]interface{})

	var sizeBytes int64
	dbConn.Get(&sizeBytes, "SELECT pg_database_size(current_database())")
	stats["size_bytes"] = sizeBytes

	var version string
	dbConn.Get(&version, "SELECT version()")
	stats["version"] = version

	var connections int
	dbConn.Get(&connections, "SELECT count(*) FROM pg_stat_activity")
	stats["connections"] = connections

	var startTime time.Time
	dbConn.Get(&startTime, "SELECT pg_postmaster_start_time()")
	stats["uptime_seconds"] = time.Since(startTime).Seconds()

	writeJSON(w, http.StatusOK, stats)
}

// --- Table Browser ---

func GetTableDataHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	tableName := vars["table"]
	schema := r.URL.Query().Get("schema")
	if schema == "" {
		schema = "public"
	}

	// Security: Validate table name
	// We validate against the list of tables in that schema
	tables, err := ListTablesBySchema(dbConn, schema)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to list tables")
		return
	}
	allowed := false
	for _, t := range tables {
		if t == tableName {
			allowed = true
			break
		}
	}
	if !allowed {
		writeError(w, r, http.StatusForbidden, "access denied to table")
		return
	}

	// Pagination
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if i, err := strconv.Atoi(l); err == nil && i > 0 && i <= 1000 {
			limit = i
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if i, err := strconv.Atoi(o); err == nil && i >= 0 {
			offset = i
		}
	}

	// Fetch Data
	// Use fully qualified name: "schema"."table"
	rows, err := dbConn.Queryx(fmt.Sprintf(`SELECT * FROM %q.%q LIMIT %d OFFSET %d`, schema, tableName, limit, offset))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	results := []map[string]interface{}{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			continue
		}
		// Handle []byte for text columns (common in sqlx/drivers)
		for k, v := range row {
			if b, ok := v.([]byte); ok {
				row[k] = string(b)
			}
		}
		results = append(results, row)
	}

	// Fetch Total Count
	var total int
	dbConn.Get(&total, fmt.Sprintf(`SELECT count(*) FROM %q.%q`, schema, tableName))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":   results,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// --- Internal Backups ---

type BackupFile struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateInternalBackupHandler(w http.ResponseWriter, r *http.Request) {
	EnsureBackupDir()
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		writeError(w, r, http.StatusBadRequest, "DB_DSN not configured")
		return
	}

	filename := fmt.Sprintf("backup-%d.sql", time.Now().Unix())
	filepath := filepath.Join(backupDir, filename)

	cmd := exec.Command("pg_dump", dbDSN, "-f", filepath)
	if output, err := cmd.CombinedOutput(); err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("pg_dump failed: %v, output: %s", err, string(output)))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "backup created", "filename": filename})
}

func ListInternalBackupsHandler(w http.ResponseWriter, r *http.Request) {
	EnsureBackupDir()
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	backups := []BackupFile{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			info, _ := e.Info()
			backups = append(backups, BackupFile{
				Name:      e.Name(),
				Size:      info.Size(),
				CreatedAt: info.ModTime(),
			})
		}
	}
	writeJSON(w, http.StatusOK, backups)
}

func DownloadInternalBackupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	if err := ValidateFilename(filename); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid filename: "+err.Error())
		return
	}

	path := filepath.Join(backupDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		writeError(w, r, http.StatusNotFound, "backup not found")
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	http.ServeFile(w, r, path)
}

func DeleteInternalBackupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	if err := ValidateFilename(filename); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid filename: "+err.Error())
		return
	}

	path := filepath.Join(backupDir, filename)
	if err := os.Remove(path); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func RestoreInternalBackupHandler(w http.ResponseWriter, r *http.Request) {
	EnsureBackupDir()
	var req struct {
		Filename string `json:"filename"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	if err := ValidateFilename(req.Filename); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid filename: "+err.Error())
		return
	}

	path := filepath.Join(backupDir, req.Filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		writeError(w, r, http.StatusNotFound, "backup not found")
		return
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		writeError(w, r, http.StatusBadRequest, "DB_DSN not configured")
		return
	}

	// Use psql to restore (pg_restore is for custom format, we used default/plain sql from pg_dump if no flags)
	// pg_dump default is plain text.
	cmd := exec.Command("psql", dbDSN, "-f", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("restore failed: %v, output: %s", err, string(output)))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "restored successfully"})
}

// --- Postgres Config ---

func GetPostgresConfigHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	rows, err := dbConn.Queryx("SHOW ALL")
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	configs := []map[string]string{}
	for rows.Next() {
		var name, setting, description string
		if err := rows.Scan(&name, &setting, &description); err != nil {
			continue
		}
		configs = append(configs, map[string]string{
			"name":        name,
			"setting":     setting,
			"description": description,
		})
	}
	writeJSON(w, http.StatusOK, configs)
}

func UpdatePostgresConfigHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var req struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	if !isValidColumnName(req.Name) { // Reuse existing validator for parameter names
		writeError(w, r, http.StatusBadRequest, "invalid parameter name")
		return
	}

	// ALTER SYSTEM SET parameter = 'value'
	// Note: Some values might need different quoting, but 'value' works for most strings/numbers in Postgres.
	// For integers, '10' is fine.
	// Sanitize value to prevent SQL injection or syntax errors via single quotes
	safeValue := strings.ReplaceAll(req.Value, "'", "''")
	query := fmt.Sprintf("ALTER SYSTEM SET %s = '%s'", req.Name, safeValue)

	if _, err := dbConn.Exec(query); err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to update config: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "config updated"})
}

func RestartPostgresHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	// Attempt reload first
	var res bool
	if err := dbConn.Get(&res, "SELECT pg_reload_conf()"); err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to reload config: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "configuration reloaded"})
}

func ListSchemasHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	schemas, err := ListSchemas(dbConn)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, schemas)
}

func CreateSchemaHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	var req struct {
		Name  string `json:"name"`
		Owner string `json:"owner"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid json")
		return
	}
	if err := CreateSchema(dbConn, req.Name, req.Owner); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "schema created"})
}

func DeleteSchemaHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	name := vars["name"]
	if err := DropSchema(dbConn, name); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "schema deleted"})
}

func ListRolesHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	roles, err := ListRoles(dbConn)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, roles)
}

func CreateRoleHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	var req struct {
		Name     string   `json:"name"`
		Password string   `json:"password"`
		Options  []string `json:"options"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid json")
		return
	}
	if err := CreateRole(dbConn, req.Name, req.Password, req.Options); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "role created"})
}

func DeleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	name := vars["name"]
	if err := DeleteRole(dbConn, name); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "role deleted"})
}

func ListTablesBySchemaHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	rawSchema := r.URL.Query().Get("schema")
	schema := strings.TrimSpace(rawSchema)
	if schema == "" {
		// STRICT MODE: Do not default to public. Force frontend to specify.
		writeError(w, r, http.StatusBadRequest, "schema parameter required")
		return
	}

	tables, err := ListTablesBySchema(dbConn, schema)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Debug header to help frontend verify what happened
	w.Header().Set("X-Resolved-Schema", schema)

	writeJSON(w, http.StatusOK, tables)
}

func GetAllTablesHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	tables, err := ListAllTables(dbConn)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tables)
}

func ListColumnsHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	schema := r.URL.Query().Get("schema")
	table := r.URL.Query().Get("table")
	if schema == "" || table == "" {
		writeError(w, r, http.StatusBadRequest, "schema and table required")
		return
	}

	cols, err := ListColumns(dbConn, schema, table)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cols)
}

func ExecuteSQLHandler(w http.ResponseWriter, r *http.Request) {
	// Use readOnlyConn if available, otherwise fall back to dbConn (which should be prevented if not safe)
	conn := readOnlyConn
	if conn == nil {
		conn = dbConn
	}

	if conn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	var req struct {
		Query string `json:"query"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	// Validate query is read-only to prevent destructive operations
	// We do this EVEN if we have a separate connection as a defense in depth,
	// unless we are sure the separate connection permissions are perfect.
	if err := IsSafeReadOnlyQuery(req.Query); err != nil {
		writeError(w, r, http.StatusForbidden, "query not allowed: "+err.Error())
		return
	}

	results, err := ExecuteSQL(conn, req.Query)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, SanitizeDBError(err))
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func DebugListAllTablesHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	query := "SELECT table_schema, table_name FROM information_schema.tables WHERE table_schema NOT IN ('information_schema', 'pg_catalog')"
	rows, err := dbConn.Queryx(query)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	results := []map[string]string{}
	for rows.Next() {
		var schema, name string
		if err := rows.Scan(&schema, &name); err != nil {
			continue
		}
		results = append(results, map[string]string{"schema": schema, "table": name})
	}
	writeJSON(w, http.StatusOK, results)
}

// --- Table Editor ---

func isValidColumnName(col string) bool {
	// Simple validation to prevent SQL injection in column names
	for _, r := range col {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

func UpdateTableRowHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	tableName := vars["table"]
	pkValue := vars["id"] // This is actually the primary key value, not necessarily "id"
	schema := r.URL.Query().Get("schema")
	pkColumn := r.URL.Query().Get("pk") // Primary key column name
	if schema == "" {
		schema = "public"
	}
	if pkColumn == "" {
		pkColumn = "id" // Default to "id" for backward compatibility
	}

	// Security: Validate table name
	tables, err := ListTablesBySchema(dbConn, schema)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to list tables")
		return
	}
	allowed := false
	for _, t := range tables {
		if t == tableName {
			allowed = true
			break
		}
	}
	if !allowed {
		writeError(w, r, http.StatusForbidden, "access denied to table")
		return
	}

	// Security: Validate primary key column name
	if !isValidColumnName(pkColumn) {
		writeError(w, r, http.StatusBadRequest, "invalid primary key column name")
		return
	}

	// Verify the pk column exists in this table
	cols, err := ListColumns(dbConn, schema, tableName)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get columns")
		return
	}
	pkColumnExists := false
	for _, col := range cols {
		if col.Name == pkColumn {
			pkColumnExists = true
			break
		}
	}
	if !pkColumnExists {
		writeError(w, r, http.StatusBadRequest, fmt.Sprintf("primary key column '%s' does not exist in table", pkColumn))
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	if len(data) == 0 {
		writeError(w, r, http.StatusBadRequest, "no data to update")
		return
	}

	// Build query using ? placeholders and Rebind
	setParts := []string{}
	args := []interface{}{}
	for k, v := range data {
		if !isValidColumnName(k) {
			writeError(w, r, http.StatusBadRequest, "invalid column name: "+k)
			return
		}
		// Skip the primary key column in SET clause
		if k == pkColumn {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%q = ?", k))
		args = append(args, v)
	}

	if len(setParts) == 0 {
		writeJSON(w, http.StatusOK, map[string]string{"message": "no changes"})
		return
	}

	args = append(args, pkValue) // Add PK value as last arg
	query := fmt.Sprintf("UPDATE %q.%q SET %s WHERE %q = ?", schema, tableName, strings.Join(setParts, ", "), pkColumn)
	query = dbConn.Rebind(query)

	if _, err := dbConn.Exec(query, args...); err != nil {
		writeError(w, r, http.StatusInternalServerError, SanitizeDBError(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "row updated"})
}

func InsertTableRowHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	tableName := vars["table"]
	schema := r.URL.Query().Get("schema")
	if schema == "" {
		schema = "public"
	}

	// Security: Validate table name
	tables, err := ListTablesBySchema(dbConn, schema)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to list tables")
		return
	}
	allowed := false
	for _, t := range tables {
		if t == tableName {
			allowed = true
			break
		}
	}
	if !allowed {
		writeError(w, r, http.StatusForbidden, "access denied to table")
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	cols := []string{}
	placeholders := []string{}
	args := []interface{}{}

	for k, v := range data {
		if !isValidColumnName(k) {
			writeError(w, r, http.StatusBadRequest, "invalid column name: "+k)
			return
		}
		if k == "id" {
			continue
		} // Let DB handle ID generation (SERIAL) usually
		cols = append(cols, fmt.Sprintf("%q", k))
		placeholders = append(placeholders, "?")
		args = append(args, v)
	}

	query := fmt.Sprintf("INSERT INTO %q.%q (%s) VALUES (%s)", schema, tableName, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	query = dbConn.Rebind(query)

	if _, err := dbConn.Exec(query, args...); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "row inserted"})
}

func DeleteTableRowHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	tableName := vars["table"]
	pkValue := vars["id"] // This is actually the primary key value, not necessarily "id"
	schema := r.URL.Query().Get("schema")
	pkColumn := r.URL.Query().Get("pk") // Primary key column name
	if schema == "" {
		schema = "public"
	}
	if pkColumn == "" {
		pkColumn = "id" // Default to "id" for backward compatibility
	}

	// Security: Validate table name
	tables, err := ListTablesBySchema(dbConn, schema)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to list tables")
		return
	}
	allowed := false
	for _, t := range tables {
		if t == tableName {
			allowed = true
			break
		}
	}
	if !allowed {
		writeError(w, r, http.StatusForbidden, "access denied to table")
		return
	}

	// Security: Validate primary key column name
	if !isValidColumnName(pkColumn) {
		writeError(w, r, http.StatusBadRequest, "invalid primary key column name")
		return
	}

	// Verify the pk column exists in this table
	cols, err := ListColumns(dbConn, schema, tableName)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get columns")
		return
	}
	pkColumnExists := false
	for _, col := range cols {
		if col.Name == pkColumn {
			pkColumnExists = true
			break
		}
	}
	if !pkColumnExists {
		writeError(w, r, http.StatusBadRequest, fmt.Sprintf("primary key column '%s' does not exist in table", pkColumn))
		return
	}

	query := fmt.Sprintf("DELETE FROM %q.%q WHERE %q = ?", schema, tableName, pkColumn)
	query = dbConn.Rebind(query)

	if _, err := dbConn.Exec(query, pkValue); err != nil {
		writeError(w, r, http.StatusInternalServerError, SanitizeDBError(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "row deleted"})
}

// --- DDL Operations ---

func CreateTableHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	var req struct {
		Schema  string   `json:"schema"`
		Name    string   `json:"name"`
		Columns []string `json:"columns"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Schema == "" {
		req.Schema = "public"
	}

	// Validate schema and table names
	if err := ValidateSchemaName(req.Schema); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateTableName(req.Name); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Validate each column definition
	for _, col := range req.Columns {
		if err := ValidateColumnDefinition(col); err != nil {
			writeError(w, r, http.StatusBadRequest, "invalid column definition: "+err.Error())
			return
		}
	}

	if err := CreateTable(dbConn, req.Schema, req.Name, req.Columns); err != nil {
		writeError(w, r, http.StatusInternalServerError, SanitizeDBError(err))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "table created"})
}

func DropTableHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	table := vars["table"]
	schema := r.URL.Query().Get("schema")
	if schema == "" {
		schema = "public"
	}

	if err := DropTable(dbConn, schema, table); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "table dropped"})
}

func AlterTableHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	table := vars["table"]

	var req struct {
		Schema  string `json:"schema"`
		Action  string `json:"action"` // add_column, drop_column, rename_column, alter_column_type, alter_column_nullable, alter_column_default
		Column  string `json:"column"`
		Type    string `json:"type,omitempty"`     // for add_column, alter_column_type
		NewName string `json:"new_name,omitempty"` // for rename_column
		NotNull bool   `json:"not_null,omitempty"` // for alter_column_nullable
		Default string `json:"default,omitempty"`  // for alter_column_default
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Schema == "" {
		req.Schema = "public"
	}

	// Validate schema and table names
	if err := ValidateSchemaName(req.Schema); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateTableName(table); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Validate column name
	if req.Column != "" {
		if err := ValidateColumnName(req.Column); err != nil {
			writeError(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}

	var err error
	switch req.Action {
	case "add_column":
		// Validate type for add_column
		if err := ValidateSQLType(req.Type); err != nil {
			writeError(w, r, http.StatusBadRequest, "invalid column type: "+err.Error())
			return
		}
		err = AddColumn(dbConn, req.Schema, table, req.Column, req.Type)
	case "drop_column":
		err = DropColumn(dbConn, req.Schema, table, req.Column)
	case "rename_column":
		// Validate new column name
		if err := ValidateColumnName(req.NewName); err != nil {
			writeError(w, r, http.StatusBadRequest, "invalid new column name: "+err.Error())
			return
		}
		err = RenameColumn(dbConn, req.Schema, table, req.Column, req.NewName)
	case "alter_column_type":
		// Validate type
		if err := ValidateSQLType(req.Type); err != nil {
			writeError(w, r, http.StatusBadRequest, "invalid column type: "+err.Error())
			return
		}
		err = AlterColumnType(dbConn, req.Schema, table, req.Column, req.Type)
	case "alter_column_nullable":
		err = AlterColumnNullable(dbConn, req.Schema, table, req.Column, req.NotNull)
	case "alter_column_default":
		// Validate default value
		if err := ValidateDefaultValue(req.Default); err != nil {
			writeError(w, r, http.StatusBadRequest, "invalid default value: "+err.Error())
			return
		}
		err = AlterColumnDefault(dbConn, req.Schema, table, req.Column, req.Default)
	default:
		writeError(w, r, http.StatusBadRequest, "invalid action")
		return
	}

	if err != nil {
		writeError(w, r, http.StatusInternalServerError, SanitizeDBError(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "table altered successfully"})
}

// PostgreSQL Function represents a database function
type PGFunction struct {
	OID           int64  `json:"oid" db:"oid"`
	Schema        string `json:"schema" db:"schema"`
	Name          string `json:"name" db:"name"`
	ResultType    string `json:"result_type" db:"result_type"`
	ArgumentTypes string `json:"argument_types" db:"argument_types"`
	Type          string `json:"type" db:"type"`             // func, proc, agg, window
	Volatility    string `json:"volatility" db:"volatility"` // immutable, stable, volatile
	Language      string `json:"language" db:"language"`
	Source        string `json:"source" db:"source"`
	Owner         string `json:"owner" db:"owner"`
	Description   string `json:"description" db:"description"`
}

// ListFunctionsHandler returns all functions in the database
func ListFunctionsHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	schema := r.URL.Query().Get("schema")
	if schema == "" {
		schema = "public"
	}

	// Compatible query for PostgreSQL 11+ (prokind) and older versions (proisagg, proiswindow)
	query := `
		SELECT
			p.oid::bigint as oid,
			n.nspname as schema,
			p.proname as name,
			COALESCE(pg_catalog.pg_get_function_result(p.oid), 'void') as result_type,
			COALESCE(pg_catalog.pg_get_function_identity_arguments(p.oid), '') as argument_types,
			CASE
				WHEN p.prokind = 'p' THEN 'procedure'
				WHEN p.prokind = 'a' THEN 'aggregate'
				WHEN p.prokind = 'w' THEN 'window'
				ELSE 'function'
			END as type,
			CASE p.provolatile
				WHEN 'i' THEN 'immutable'
				WHEN 's' THEN 'stable'
				ELSE 'volatile'
			END as volatility,
			l.lanname as language,
			'' as source,
			pg_catalog.pg_get_userbyid(p.proowner) as owner,
			COALESCE(obj_description(p.oid, 'pg_proc'), '') as description
		FROM pg_catalog.pg_proc p
		LEFT JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
		LEFT JOIN pg_catalog.pg_language l ON l.oid = p.prolang
		WHERE n.nspname = $1
		ORDER BY p.proname
	`

	rows, err := dbConn.Queryx(query, schema)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	functions := []PGFunction{}
	for rows.Next() {
		var fn PGFunction
		if err := rows.StructScan(&fn); err != nil {
			// Log error but continue with other rows
			continue
		}

		// Try to get function definition separately (only works for regular functions)
		if fn.Type == "function" {
			var source string
			sourceQuery := "SELECT pg_catalog.pg_get_functiondef($1::oid)"
			if err := dbConn.Get(&source, sourceQuery, fn.OID); err == nil {
				fn.Source = source
			}
		}

		functions = append(functions, fn)
	}

	writeJSON(w, http.StatusOK, functions)
}

// GetFunctionHandler returns details of a specific function
func GetFunctionHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	oid := r.URL.Query().Get("oid")
	if oid == "" {
		writeError(w, r, http.StatusBadRequest, "oid is required")
		return
	}

	query := `
		SELECT
			p.oid::bigint as oid,
			n.nspname as schema,
			p.proname as name,
			COALESCE(pg_catalog.pg_get_function_result(p.oid), 'void') as result_type,
			COALESCE(pg_catalog.pg_get_function_identity_arguments(p.oid), '') as argument_types,
			CASE
				WHEN p.prokind = 'p' THEN 'procedure'
				WHEN p.prokind = 'a' THEN 'aggregate'
				WHEN p.prokind = 'w' THEN 'window'
				ELSE 'function'
			END as type,
			CASE p.provolatile
				WHEN 'i' THEN 'immutable'
				WHEN 's' THEN 'stable'
				ELSE 'volatile'
			END as volatility,
			l.lanname as language,
			'' as source,
			pg_catalog.pg_get_userbyid(p.proowner) as owner,
			COALESCE(obj_description(p.oid, 'pg_proc'), '') as description
		FROM pg_catalog.pg_proc p
		LEFT JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
		LEFT JOIN pg_catalog.pg_language l ON l.oid = p.prolang
		WHERE p.oid = $1
	`

	var fn PGFunction
	if err := dbConn.Get(&fn, query, oid); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Try to get function definition separately (only works for regular functions)
	if fn.Type == "function" {
		var source string
		sourceQuery := "SELECT pg_catalog.pg_get_functiondef($1::oid)"
		if err := dbConn.Get(&source, sourceQuery, fn.OID); err == nil {
			fn.Source = source
		}
	}

	writeJSON(w, http.StatusOK, fn)
}

// CreateFunctionHandler creates a new function
func CreateFunctionHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var req struct {
		Schema          string `json:"schema"`
		Name            string `json:"name"`
		Arguments       string `json:"arguments"`
		Returns         string `json:"returns"`
		Language        string `json:"language"`
		Body            string `json:"body"`
		Volatility      string `json:"volatility"`
		IsStrict        bool   `json:"is_strict"`
		SecurityDefiner bool   `json:"security_definer"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.Name == "" || req.Returns == "" || req.Body == "" {
		writeError(w, r, http.StatusBadRequest, "name, returns, and body are required")
		return
	}

	if req.Schema == "" {
		req.Schema = "public"
	}
	if req.Language == "" {
		req.Language = "plpgsql"
	}
	if req.Volatility == "" {
		req.Volatility = "VOLATILE"
	}

	// Security: Validate all inputs
	if err := ValidateSchemaName(req.Schema); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionName(req.Name); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionLanguage(req.Language); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateVolatility(req.Volatility); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionArguments(req.Arguments); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionReturnType(req.Returns); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionBody(req.Body); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Security: Disallow SECURITY DEFINER to prevent privilege escalation
	if req.SecurityDefiner {
		writeError(w, r, http.StatusForbidden, "SECURITY DEFINER functions are not allowed for security reasons")
		return
	}

	// Build the CREATE FUNCTION statement using quoted identifiers
	sql := fmt.Sprintf(
		`CREATE OR REPLACE FUNCTION %q.%q(%s) RETURNS %s AS $$ %s $$ LANGUAGE %s %s`,
		req.Schema, req.Name, req.Arguments, req.Returns, req.Body, req.Language, strings.ToUpper(req.Volatility),
	)

	if req.IsStrict {
		sql += " STRICT"
	}

	if _, err := dbConn.Exec(sql); err != nil {
		writeError(w, r, http.StatusInternalServerError, SanitizeDBError(err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "function created successfully"})
}

// UpdateFunctionHandler updates an existing function
func UpdateFunctionHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var req struct {
		OID             int64  `json:"oid"`
		Schema          string `json:"schema"`
		Name            string `json:"name"`
		Arguments       string `json:"arguments"`
		Returns         string `json:"returns"`
		Language        string `json:"language"`
		Body            string `json:"body"`
		Volatility      string `json:"volatility"`
		IsStrict        bool   `json:"is_strict"`
		SecurityDefiner bool   `json:"security_definer"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.Name == "" || req.Returns == "" || req.Body == "" {
		writeError(w, r, http.StatusBadRequest, "name, returns, and body are required")
		return
	}

	if req.Schema == "" {
		req.Schema = "public"
	}
	if req.Language == "" {
		req.Language = "plpgsql"
	}
	if req.Volatility == "" {
		req.Volatility = "VOLATILE"
	}

	// Security: Validate all inputs
	if err := ValidateSchemaName(req.Schema); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionName(req.Name); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionLanguage(req.Language); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateVolatility(req.Volatility); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionArguments(req.Arguments); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionReturnType(req.Returns); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionBody(req.Body); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Security: Disallow SECURITY DEFINER to prevent privilege escalation
	if req.SecurityDefiner {
		writeError(w, r, http.StatusForbidden, "SECURITY DEFINER functions are not allowed for security reasons")
		return
	}

	// Use CREATE OR REPLACE to update with quoted identifiers
	sql := fmt.Sprintf(
		`CREATE OR REPLACE FUNCTION %q.%q(%s) RETURNS %s AS $$ %s $$ LANGUAGE %s %s`,
		req.Schema, req.Name, req.Arguments, req.Returns, req.Body, req.Language, strings.ToUpper(req.Volatility),
	)

	if req.IsStrict {
		sql += " STRICT"
	}

	if _, err := dbConn.Exec(sql); err != nil {
		writeError(w, r, http.StatusInternalServerError, SanitizeDBError(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "function updated successfully"})
}

// DeleteFunctionHandler drops a function
func DeleteFunctionHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var req struct {
		Schema    string `json:"schema"`
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
		Cascade   bool   `json:"cascade"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, r, http.StatusBadRequest, "function name is required")
		return
	}

	if req.Schema == "" {
		req.Schema = "public"
	}

	// Security: Validate all inputs
	if err := ValidateSchemaName(req.Schema); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionName(req.Name); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if req.Arguments != "" {
		if err := ValidateFunctionArguments(req.Arguments); err != nil {
			writeError(w, r, http.StatusBadRequest, err.Error())
			return
		}
	}

	sql := fmt.Sprintf("DROP FUNCTION %q.%q(%s)", req.Schema, req.Name, req.Arguments)
	if req.Cascade {
		sql += " CASCADE"
	}

	if _, err := dbConn.Exec(sql); err != nil {
		writeError(w, r, http.StatusInternalServerError, SanitizeDBError(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "function dropped successfully"})
}

// ExecuteFunctionHandler executes a function and returns results
func ExecuteFunctionHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var req struct {
		Schema    string   `json:"schema"`
		Name      string   `json:"name"`
		Arguments []string `json:"arguments"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, r, http.StatusBadRequest, "function name is required")
		return
	}

	if req.Schema == "" {
		req.Schema = "public"
	}

	// Security: Validate schema and function name
	if err := ValidateSchemaName(req.Schema); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidateFunctionName(req.Name); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Build parameterized query to prevent SQL injection
	// We use positional parameters ($1, $2, etc.) for function arguments
	placeholders := make([]string, len(req.Arguments))
	args := make([]interface{}, len(req.Arguments))
	for i, arg := range req.Arguments {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = arg
	}

	// Use quoted identifiers for schema and function name
	query := fmt.Sprintf("SELECT * FROM %q.%q(%s)", req.Schema, req.Name, strings.Join(placeholders, ", "))

	rows, err := dbConn.Queryx(query, args...)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, SanitizeDBError(err))
		return
	}
	defer rows.Close()

	results := []map[string]interface{}{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			writeError(w, r, http.StatusInternalServerError, SanitizeDBError(err))
			return
		}
		// Convert []byte to string for JSON serialization
		for k, v := range row {
			if b, ok := v.([]byte); ok {
				row[k] = string(b)
			}
		}
		results = append(results, row)
	}

	writeJSON(w, http.StatusOK, results)
}
