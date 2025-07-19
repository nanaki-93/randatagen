package migrate

import (
	"database/sql"
	"fmt"
	"github.com/nanaki-93/randatagen/internal/config"
	"github.com/nanaki-93/randatagen/internal/db"
	"github.com/nanaki-93/randatagen/internal/generate"
	"github.com/nanaki-93/randatagen/internal/model"
	"log/slog"
	"os"
	"path/filepath"
)

type PostgresDbProvider struct {
	migrationData model.MigrationData
}

func NewPostgresDbProvider(migrationData model.MigrationData) DbProvider {
	return &PostgresDbProvider{
		migrationData: migrationData,
	}
}

func (p *PostgresDbProvider) Open() (*sql.DB, *sql.DB, error) {
	sourceConn, err := db.GetConn(p.migrationData.DbType, p.migrationData.Source)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting source connection: %w", err)
	}
	targetConn, err := db.GetConn(p.migrationData.DbType, p.migrationData.Target)
	if err != nil {
		sourceConn.Close()
		return nil, nil, fmt.Errorf("error getting target connection: %w", err)
	}
	return sourceConn, targetConn, nil

}
func (p *PostgresDbProvider) migrateTableStructure(source, target *sql.DB, table string) error {

	createTablesQuery := fmt.Sprintf(config.Queries[config.DynamicCreateQuery], table)

	rows, err := source.Query(createTablesQuery)
	if err != nil {
		return fmt.Errorf("error getting create table query for %s: %w", table, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close error: %w", cerr)
		}
	}()

	var createTableQuery string
	if rows.Next() {
		if err := rows.Scan(&createTableQuery); err != nil {
			return fmt.Errorf("error scanning table name: %w", err)
		}
	}
	_, err = target.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("error executing create table query for %s: %w", table, err)
	}

	return nil
}

func (p *PostgresDbProvider) migrateTablePkStructure(source, target *sql.DB, table string) error {

	createTablesQuery := fmt.Sprintf(`SELECT
    'ALTER TABLE ' || tc.table_name || ' ADD CONSTRAINT ' || tc.constraint_name ||
    ' PRIMARY KEY (' || string_agg(kcu.column_name, ', ') || ');' AS ddl
FROM information_schema.table_constraints tc
         JOIN information_schema.key_column_usage kcu
              ON tc.constraint_name = kcu.constraint_name
WHERE tc.table_name = '%s'
  AND tc.constraint_type = 'PRIMARY KEY'
GROUP BY tc.table_name, tc.constraint_name;
`, table)

	rows, err := source.Query(createTablesQuery)
	if err != nil {
		return fmt.Errorf("error getting create table query for %s: %w", table, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close error: %w", cerr)
		}
	}()

	var createPkQuery string
	if rows.Next() {
		if err := rows.Scan(&createPkQuery); err != nil {
			return fmt.Errorf("error scanning table name: %w", err)
		}
	}
	pqCreatePkQuery := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = '%s'
          AND constraint_type = 'PRIMARY KEY'
          AND constraint_name = '%s_pk'
    ) THEN
        EXECUTE '%s';
    END IF;
END
$$;
`, table, table, createPkQuery)
	_, err = target.Exec(pqCreatePkQuery)
	if err != nil {
		return fmt.Errorf("error executing create table query for %s: %w", table, err)
	}

	return nil
}

func (p *PostgresDbProvider) migrateTableIndexesStructure(source, target *sql.DB, table string) error {

	indexesDefinitionQuery := fmt.Sprintf(`SELECT indexdef FROM pg_indexes WHERE tablename = '%s' and schemaname='%s';`, table, p.migrationData.Source.DbSchema)

	rows, err := source.Query(indexesDefinitionQuery)
	if err != nil {
		return fmt.Errorf("error getting create table query for %s: %w", table, err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close error: %w", cerr)
		}
	}()

	var indexesCreateQueryList []string
	for rows.Next() {
		var indexCreateQuery string
		if err := rows.Scan(&indexCreateQuery); err != nil {
			return fmt.Errorf("error scanning table name: %w", err)
		}
		indexesCreateQueryList = append(indexesCreateQueryList, indexCreateQuery)
	}
	indexTemplate := `
DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1 FROM pg_indexes
            WHERE tablename = '%tablename%' and schemaname='%schemaname%' AND indexname = '%indexname%'
        ) THEN
            EXECUTE 'CREATE UNIQUE INDEX %indexname% ON %tablename% USING btree (%tablename%_uid);';
        END IF;
    END
$$;
`

	_, err = target.Exec(indexTemplate)
	if err != nil {
		return fmt.Errorf("error executing create table query for %s: %w", table, err)
	}

	return nil
}
func (p *PostgresDbProvider) GetTablesToMigrate(sourceConn *sql.DB) ([]string, error) {
	getTablesQuery := "SELECT table_name FROM information_schema.tables WHERE table_schema = $1"

	_, err := sourceConn.Exec(getTablesQuery, p.migrationData.Source.DbSchema)

	rows, err := sourceConn.Query(getTablesQuery)
	if err != nil {
		return nil, fmt.Errorf("error getting tables from source: %w", err)
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close error: %w", cerr)
		}
	}()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	return tables, err
}
func (p *PostgresDbProvider) MigrateTable(source, target *sql.DB, table string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}
	dataTmpFile := filepath.Join(currentDir, table+".tmp")

	_, err = source.Exec("COPY " + generate.WithDoubleQuote(table) + " TO '" + dataTmpFile + "' WITH CSV HEADER")
	if err != nil {
		return fmt.Errorf("error copying data from source table %s: %w", table, err)
	}
	_, err = target.Exec("COPY " + generate.WithDoubleQuote(table) + " FROM '" + dataTmpFile + "' WITH CSV HEADER")
	if err != nil {
		return fmt.Errorf("error copying data to target table %s: %w", table, err)
	}

	err = os.Remove(dataTmpFile)
	if err != nil {
		return fmt.Errorf("error removing temporary file %s: %w", dataTmpFile, err)
	}
	slog.Info("Migrated table", "table", table)
	return nil
}
func (p *PostgresDbProvider) Close(source, target *sql.DB) error {
	err := source.Close()
	if err != nil {
		return fmt.Errorf("error closing source connection: %w", err)
	}
	err = target.Close()
	if err != nil {
		return fmt.Errorf("error closing target connection: %w", err)
	}
	return err
}
