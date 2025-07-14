package migrate

import (
	"database/sql"
	"fmt"
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
