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

		// 5. Table Size Analysis
		`CREATE OR REPLACE FUNCTION exile_table_sizes()
		RETURNS TABLE (
			schema_name text,
			table_name text,
			total_size text,
			data_size text,
			index_size text
		) AS $$
		BEGIN
			RETURN QUERY
			SELECT
				ns.nspname::text,
				tbl.relname::text,
				pg_size_pretty(pg_total_relation_size(tbl.oid)),
				pg_size_pretty(pg_relation_size(tbl.oid)),
				pg_size_pretty(pg_total_relation_size(tbl.oid) - pg_relation_size(tbl.oid))
			FROM pg_class tbl
			JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
			WHERE ns.nspname NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
			AND tbl.relkind = 'r'
			ORDER BY pg_total_relation_size(tbl.oid) DESC;
		END;
		$$ LANGUAGE plpgsql STABLE;`,

		// 6. Active Query Inspector
		`CREATE OR REPLACE FUNCTION exile_active_queries()
		RETURNS TABLE (
			pid integer,
			duration interval,
			query text,
			state text
		) AS $$
		BEGIN
			RETURN QUERY
			SELECT
				pg_stat_activity.pid,
				(now() - pg_stat_activity.query_start)::interval,
				pg_stat_activity.query,
				pg_stat_activity.state
			FROM pg_stat_activity
			WHERE state != 'idle' 
			AND pid <> pg_backend_pid()
			ORDER BY query_start ASC;
		END;
		$$ LANGUAGE plpgsql VOLATILE;`,

		// 7. Kill Long Running Queries
		`CREATE OR REPLACE FUNCTION exile_kill_long_queries(max_duration interval)
		RETURNS integer AS $$
		DECLARE
			killed_count integer := 0;
			r record;
		BEGIN
			FOR r IN 
				SELECT pid, query 
				FROM pg_stat_activity 
				WHERE state != 'idle' 
				AND (now() - query_start) > max_duration 
				AND pid <> pg_backend_pid()
			LOOP
				PERFORM pg_terminate_backend(r.pid);
				killed_count := killed_count + 1;
			END LOOP;
			RETURN killed_count;
		END;
		$$ LANGUAGE plpgsql VOLATILE;`,

		// 8. Maintenance Statistics
		`CREATE OR REPLACE FUNCTION exile_maintenance_stats()
		RETURNS TABLE (
			metric text,
			value text
		) AS $$
		BEGIN
			RETURN QUERY VALUES 
				('Total Database Size', pg_size_pretty(pg_database_size(current_database()))),
				('Cache Hit Ratio', (
					SELECT round(sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) * 100, 2)::text || '%'
					FROM pg_statio_user_tables
					WHERE heap_blks_read > 0
				)),
				('Index Usage Rate', (
					SELECT round(sum(idx_scan) / (sum(idx_scan) + sum(seq_scan)) * 100, 2)::text || '%'
					FROM pg_stat_user_tables
					WHERE seq_scan + idx_scan > 0
				)),
				('Dead Tuple Estimate', (
					SELECT sum(n_dead_tup)::text
					FROM pg_stat_user_tables
				));
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
