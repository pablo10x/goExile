package database

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"exile/server/utils"
)

// GetAllTablesHandler lists all tables
func GetAllTablesHandler(w http.ResponseWriter, r *http.Request) {
	tables, err := ListAllTables(DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, tables)
}

// DebugListAllTablesHandler returns table counts
func DebugListAllTablesHandler(w http.ResponseWriter, r *http.Request) {
    counts, err := GetTableCounts(DBConn)
    if err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    utils.WriteJSON(w, http.StatusOK, counts)
}

// CreateInternalBackupHandler - Placeholder
func CreateInternalBackupHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "backup_created"})
}

// GetDatabaseOverviewHandler
func GetDatabaseOverviewHandler(w http.ResponseWriter, r *http.Request) {
    stats, err := GetAdvancedDBStats(DBConn)
    if err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    utils.WriteJSON(w, http.StatusOK, stats)
}

// GetTableDataHandler - Generic select
func GetTableDataHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    table := vars["table"]
    // Basic validation
    if err := utils.ValidateTableName(table); err != nil {
        utils.WriteError(w, r, http.StatusBadRequest, "Invalid table name")
        return
    }
    
    rows, err := DBConn.Queryx(fmt.Sprintf("SELECT * FROM %s LIMIT 100", table))
    if err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    defer rows.Close()

    results := []map[string]interface{}{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			continue
		}
        // Handle bytes
		for k, v := range row {
			if b, ok := v.([]byte); ok {
				row[k] = string(b)
			}
		}
		results = append(results, row)
	}
    utils.WriteJSON(w, http.StatusOK, results)
}

// InsertTableRowHandler - Stub
func InsertTableRowHandler(w http.ResponseWriter, r *http.Request) {
     utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// UpdateTableRowHandler - Stub
func UpdateTableRowHandler(w http.ResponseWriter, r *http.Request) {
     utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DeleteTableRowHandler - Generic Delete
func DeleteTableRowHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    table := vars["table"]
    id := vars["id"]
     if err := utils.ValidateTableName(table); err != nil {
        utils.WriteError(w, r, http.StatusBadRequest, "Invalid table name")
        return
    }
    // Assume id column is 'id'
    _, err := DBConn.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = $1", table), id)
    if err != nil {
         utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
         return
    }
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetPostgresConfigHandler - Stub
func GetPostgresConfigHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, []string{})
}

// UpdatePostgresConfigHandler - Stub
func UpdatePostgresConfigHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// RestartPostgresHandler - Stub
func RestartPostgresHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListSchemasHandler
func ListSchemasHandler(w http.ResponseWriter, r *http.Request) {
    schemas, err := ListSchemas(DBConn)
    if err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    utils.WriteJSON(w, http.StatusOK, schemas)
}

// CreateSchemaHandler
func CreateSchemaHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name string `json:"name"`
        Owner string `json:"owner"`
    }
    if err := utils.DecodeJSON(r, &req); err != nil {
        utils.WriteError(w, r, http.StatusBadRequest, err.Error())
        return
    }
    if err := CreateSchema(DBConn, req.Name, req.Owner); err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    utils.WriteJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// DeleteSchemaHandler
func DeleteSchemaHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    if err := DropSchema(DBConn, name); err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListTablesBySchemaHandler
func ListTablesBySchemaHandler(w http.ResponseWriter, r *http.Request) {
    schema := r.URL.Query().Get("schema")
    if schema == "" {
        schema = "public"
    }
    tables, err := ListTablesBySchema(DBConn, schema)
    if err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    utils.WriteJSON(w, http.StatusOK, tables)
}

// CreateTableHandler
func CreateTableHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Schema string `json:"schema"`
        Name string `json:"name"`
        Columns []string `json:"columns"`
    }
     if err := utils.DecodeJSON(r, &req); err != nil {
        utils.WriteError(w, r, http.StatusBadRequest, err.Error())
        return
    }
    if err := CreateTable(DBConn, req.Schema, req.Name, req.Columns); err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    utils.WriteJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// DropTableHandler
func DropTableHandler(w http.ResponseWriter, r *http.Request) {
     vars := mux.Vars(r)
     table := vars["table"] 
     schema := r.URL.Query().Get("schema")
     if schema == "" { schema = "public" }
     
     if err := DropTable(DBConn, schema, table); err != nil {
         utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
         return
     }
     utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "dropped"})
}

// AlterTableHandler - Stub to support "ADD COLUMN" etc
func AlterTableHandler(w http.ResponseWriter, r *http.Request) {
     utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListColumnsHandler
func ListColumnsHandler(w http.ResponseWriter, r *http.Request) {
    schema := r.URL.Query().Get("schema")
    table := r.URL.Query().Get("table")
    if schema == "" { schema = "public" }
    
    cols, err := ListColumns(DBConn, schema, table)
    if err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    utils.WriteJSON(w, http.StatusOK, cols)
}

// ExecuteSQLHandler
func ExecuteSQLHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Query string `json:"query"`
    }
    if err := utils.DecodeJSON(r, &req); err != nil {
        utils.WriteError(w, r, http.StatusBadRequest, err.Error())
        return
    }
    results, err := ExecuteSQL(DBConn, req.Query)
    if err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    utils.WriteJSON(w, http.StatusOK, results)
}

// ListRolesHandler
func ListRolesHandler(w http.ResponseWriter, r *http.Request) {
    roles, err := ListRoles(DBConn)
    if err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
    utils.WriteJSON(w, http.StatusOK, roles)
}

// CreateRoleHandler
func CreateRoleHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name string `json:"name"`
        Password string `json:"password"`
        Options []string `json:"options"`
    }
    if err := utils.DecodeJSON(r, &req); err != nil {
        utils.WriteError(w, r, http.StatusBadRequest, err.Error())
        return
    }
    if err := CreateRole(DBConn, req.Name, req.Password, req.Options); err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
     utils.WriteJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// DeleteRoleHandler
func DeleteRoleHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    if err := DeleteRole(DBConn, name); err != nil {
        utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
        return
    }
     utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListFunctionsHandler - Stub
func ListFunctionsHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, []string{})
}

// GetFunctionHandler - Stub
func GetFunctionHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, map[string]string{})
}

// CreateFunctionHandler - Stub
func CreateFunctionHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// UpdateFunctionHandler - Stub
func UpdateFunctionHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DeleteFunctionHandler - Stub
func DeleteFunctionHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ExecuteFunctionHandler - Stub
func ExecuteFunctionHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListInternalBackupsHandler - Stub
func ListInternalBackupsHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, []string{})
}

// DownloadInternalBackupHandler - Stub
func DownloadInternalBackupHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound)
}

// DeleteInternalBackupHandler - Stub
func DeleteInternalBackupHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// RestoreInternalBackupHandler - Stub
func RestoreInternalBackupHandler(w http.ResponseWriter, r *http.Request) {
    utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}