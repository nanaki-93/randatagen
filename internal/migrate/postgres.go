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
	"strings"
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

	createTablesQuery := fmt.Sprintf(config.DynamicQueries.Postgres.Table, table)

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

	createTablesQuery := fmt.Sprintf(config.DynamicQueries.Postgres.ExtractPrimaryKey, table)

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

	pqCreatePkQuery := fmt.Sprintf(config.DynamicQueries.Postgres.CreatePrimaryKey, table, table, createPkQuery)
	_, err = target.Exec(pqCreatePkQuery)
	if err != nil {
		return fmt.Errorf("error executing create table query for %s: %w", table, err)
	}

	return nil
}

func (p *PostgresDbProvider) migrateTableIndexesStructure(source, target *sql.DB, table string) error {

	indexesDefinitionQuery := fmt.Sprintf(config.DynamicQueries.Postgres.ExtractIndex, table, p.migrationData.Source.DbSchema)

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
	var indexNameList []string
	for rows.Next() {
		var indexCreateQuery string
		if err := rows.Scan(&indexCreateQuery); err != nil {
			return fmt.Errorf("error scanning table name: %w", err)
		}
		indexesCreateQueryList = append(indexesCreateQueryList, indexCreateQuery)
		indexNameList = append(indexNameList, generate.WithDoubleQuote(strings.Split(indexCreateQuery, " ")[2]))
		fmt.Println("Index Name:", generate.WithDoubleQuote(strings.Split(indexCreateQuery, " ")[2]))
	}

	for i, extractedIndex := range indexesCreateQueryList {
		createIndexTemplate := config.DynamicQueries.Postgres.CreateIndex
		createIndexTemplate = strings.ReplaceAll(createIndexTemplate, "%tablename%", table)
		createIndexTemplate = strings.ReplaceAll(createIndexTemplate, "%schemaname%", p.migrationData.Source.DbSchema)
		createIndexTemplate = strings.ReplaceAll(createIndexTemplate, "%indexname%", indexNameList[i])
		createIndexTemplate = strings.ReplaceAll(createIndexTemplate, "%indexCreateQuery%", extractedIndex)

		_, err = target.Exec(createIndexTemplate)
		if err != nil {
			return fmt.Errorf("error executing create table query for %s: %w", table, err)
		}
	}

	return nil
}
func (p *PostgresDbProvider) GetTablesToMigrate(sourceConn *sql.DB) ([]string, error) {
	query := config.DynamicQueries.Postgres.GetTableNames
	fmt.Println(query)
	rows, err := sourceConn.Query(query, p.migrationData.Source.DbSchema)
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

	_, err = source.Exec(config.DynamicQueries.Postgres.CopyTo, generate.WithDoubleQuote(table), dataTmpFile)
	if err != nil {
		return fmt.Errorf("error copying data from source table %s: %w", table, err)
	}
	_, err = target.Exec(config.DynamicQueries.Postgres.CopyFrom, generate.WithDoubleQuote(table), dataTmpFile)
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
