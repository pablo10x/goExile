package database

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// SeedFunctions creates default useful PostgreSQL functions for the Exile system.
func SeedFunctions(db *sqlx.DB) error {
	// Only run on PostgreSQL
	if db.DriverName() != "pgx" {
		return nil
	}

	functions := []string{
		// 1. Get Pretty Database Size
		`CREATE OR REPLACE FUNCTION exile_get_db_size()
		RETURNS text AS $$
		BEGIN
			RETURN pg_size_pretty(pg_database_size(current_database()));
		END;
		$$ LANGUAGE plpgsql VOLATILE;`,

		// 2. Cleanup Old Logs (System Logs)
		`CREATE OR REPLACE FUNCTION exile_cleanup_logs(days_to_keep integer)
		RETURNS integer AS $$
		DECLARE
			deleted_count integer;
		BEGIN
			DELETE FROM system_logs 
			WHERE timestamp < extract(epoch from (now() - (days_to_keep || ' days')::interval));
			GET DIAGNOSTICS deleted_count = ROW_COUNT;
			RETURN deleted_count;
		END;
		$$ LANGUAGE plpgsql VOLATILE;`,

		// 3. Dynamic Row Count
		`CREATE OR REPLACE FUNCTION exile_count_rows(schema_name text, table_name text)
		RETURNS integer AS $$
		DECLARE
			row_count integer;
		BEGIN
			EXECUTE format('SELECT count(*) FROM %I.%I', schema_name, table_name) INTO row_count;
			RETURN row_count;
		END;
		$$ LANGUAGE plpgsql STABLE;`,

		// 4. System Health Check (Example)
		`CREATE OR REPLACE FUNCTION exile_system_health()
		RETURNS text AS $$
		DECLARE
			node_count integer;
			error_count integer;
		BEGIN
			SELECT count(*) INTO node_count FROM nodes WHERE status = 'Online';
			SELECT count(*) INTO error_count FROM system_logs 
			WHERE level = 'ERROR' AND timestamp > extract(epoch from (now() - interval '1 hour'));
			
			IF error_count > 50 THEN
				RETURN 'CRITICAL: High Error Rate (' || error_count || '/hr)';
			ELSIF node_count = 0 THEN
				RETURN 'WARNING: No Active Nodes';
			ELSE
				RETURN 'OPTIMAL: ' || node_count || ' nodes active';
			END IF;
		END;
		$$ LANGUAGE plpgsql STABLE;`,
	}

	for _, sql := range functions {
		if _, err := db.Exec(sql); err != nil {
			log.Printf("Warning: failed to seed function: %v", err)
			return fmt.Errorf("failed to create function: %w", err)
		}
	}

	return nil
}
