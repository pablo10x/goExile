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
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		writeError(w, r, http.StatusBadRequest, "invalid filename")
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
	if filename == "" || strings.Contains(filename, "..") {
		writeError(w, r, http.StatusBadRequest, "invalid filename")
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

	if req.Filename == "" || strings.Contains(req.Filename, "..") {
		writeError(w, r, http.StatusBadRequest, "invalid filename")
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
        Name string `json:"name"`
    }
    if err := decodeJSON(r, &req); err != nil {
        writeError(w, r, http.StatusBadRequest, "invalid json")
        return
    }
    if err := CreateSchema(dbConn, req.Name); err != nil {
        writeError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, map[string]string{"message": "schema created"})
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
    schema := strings.TrimSpace(r.URL.Query().Get("schema"))
    if schema == "" {
        schema = "public"
    }
    
    // Explicitly check if schema exists to avoid confusion? 
    // No, standard behavior is empty list if schema doesn't exist or has no tables.
    // However, if the user sees public tables when querying 'app', it implies schema variable IS 'public'.
    
    tables, err := ListTablesBySchema(dbConn, schema)
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
    if dbConn == nil {
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
    
    // Basic safety check (very minimal, relying on admin auth)
    if req.Query == "" {
        writeError(w, r, http.StatusBadRequest, "empty query")
        return
    }

    results, err := ExecuteSQL(dbConn, req.Query)
    if err != nil {
        writeError(w, r, http.StatusBadRequest, err.Error())
        return
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
	id := vars["id"]
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

	if len(data) == 0 {
		writeError(w, r, http.StatusBadRequest, "no data to update")
		return
	}

	// Build query
	setParts := []string{}
	args := []interface{}{}
	i := 1
	for k, v := range data {
		if !isValidColumnName(k) {
			writeError(w, r, http.StatusBadRequest, "invalid column name: "+k)
			return
		}
		// Skip 'id' in update
		if k == "id" {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%q = $%d", k, i))
		args = append(args, v)
		i++
	}

	if len(setParts) == 0 {
		writeJSON(w, http.StatusOK, map[string]string{"message": "no changes"})
		return
	}

	args = append(args, id) // Add ID as last arg
	query := fmt.Sprintf("UPDATE %q.%q SET %s WHERE id = $%d", schema, tableName, strings.Join(setParts, ", "), i)

	// Rebind for driver (pgx uses $n, but sqlx Rebind helps if we switched drivers, though we manually built $n here for Postgres)
	// Actually, Rebind is better if we build with `?`.
	// Let's rebuild with `?` and use Rebind.
	setParts = []string{}
	args = []interface{}{}
	for k, v := range data {
		if k == "id" { continue }
		setParts = append(setParts, fmt.Sprintf("%q = ?", k))
		args = append(args, v)
	}
	args = append(args, id)
	query = fmt.Sprintf("UPDATE %q.%q SET %s WHERE id = ?", schema, tableName, strings.Join(setParts, ", "))
	query = dbConn.Rebind(query)

	if _, err := dbConn.Exec(query, args...); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
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
		if k == "id" { continue } // Let DB handle ID generation (SERIAL) usually
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
	id := vars["id"]
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

	query := fmt.Sprintf("DELETE FROM %q.%q WHERE id = ?", schema, tableName)
	query = dbConn.Rebind(query)

	if _, err := dbConn.Exec(query, id); err != nil {
		writeError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "row deleted"})
}
