/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/nanaki-93/randatagen/internal/db"
	"github.com/nanaki-93/randatagen/internal/generate"
	"github.com/nanaki-93/randatagen/internal/model"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"path/filepath"
)

const MigrateFilePattern = "migrate*.json"

// migrateCmd represents the gen command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate a list of sql table data",
	Long:  `migrate a list of sql table data`,
	Run:   execMigrateCmd,
}

func execMigrateCmd(cmd *cobra.Command, args []string) {
	inputFile := args[0]

	isDir, _ := cmd.Flags().GetBool("dir")
	if isDir {
		err := MigrateAll()
		if err != nil {
			slog.Error("Error migrating all schemas", "error", err)
			os.Exit(1)
		}
	} else {
		migrateData, err := toMigrateData(inputFile)
		if err != nil {
			slog.Error("Error getting migrate data from input file", "error", err)
			os.Exit(1)
		}
		err = MigrateSchema(migrateData)
		if err != nil {
			slog.Error("Error migrating schema", "error", err)
			os.Exit(1)
		}
		return
	}

}

func MigrateSchema(migrateData model.MigrationData) error {
	sourceConn, targetConn, err := getDbConnections(migrateData)
	if err != nil {
		return fmt.Errorf("error getting database connections: %w", err)
	}
	defer func() {
		if cerr := sourceConn.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close error: %w", cerr)
		}
	}()
	defer func() {
		if cerr := targetConn.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close error: %w", cerr)
		}
	}()

	tables, err := GetTablesToMigrate(migrateData, sourceConn)
	if err != nil {
		return fmt.Errorf("error getting tables from source: %w", err)
	}
	for _, table := range tables {
		err := migrateTable(sourceConn, targetConn, table)
		if err != nil {
			return fmt.Errorf("error migrating table %s: %w", table, err)
		}
	}
	slog.Info("Migration completed successfully")
	return err
}

func GetTablesToMigrate(migrateData model.MigrationData, sourceConn *sql.DB) ([]string, error) {
	getTablesQuery := "SELECT table_name FROM information_schema.tables WHERE table_schema = $1"
	if migrateData.DbType == "oracle" {
		getTablesQuery = "SELECT table_name FROM all_tables WHERE owner = $1"
	}
	_, err := sourceConn.Exec(getTablesQuery, migrateData.Source.DbSchema)

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

func getDbConnections(data model.MigrationData) (*sql.DB, *sql.DB, error) {
	sourceConn, err := db.GetConn(data.DbType, data.Source)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting source connection: %w", err)
	}
	targetConn, err := db.GetConn(data.DbType, data.Target)
	if err != nil {
		sourceConn.Close()
		return nil, nil, fmt.Errorf("error getting target connection: %w", err)
	}
	return sourceConn, targetConn, nil
}

func MigrateAll() error {
	inputFiles, err := filepath.Glob(MigrateFilePattern)
	if err != nil {
		return fmt.Errorf("error finding files with pattern %s: %v", MigrateFilePattern, err)
	}

	for _, inputFile := range inputFiles {
		fmt.Printf("Processing file: %s\n", inputFile)
		migData, err := toMigrateData(inputFile)
		if err != nil {
			return fmt.Errorf("error getting migrate data: %w", err)
		}
		err = MigrateSchema(migData)
		if err != nil {
			return fmt.Errorf("error migrating file %s: %w", inputFile, err)
		}
	}
	return nil
}

func toMigrateData(inputFile string) (model.MigrationData, error) {
	fmt.Printf("input file: %s,\n", inputFile)
	err := checkFileExists(inputFile)
	if err != nil {
		return model.MigrationData{}, fmt.Errorf("error checking file path: %v", err)
	}
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return model.MigrationData{}, fmt.Errorf("error reading file: %v", err)

	}
	var migrateData model.MigrationData
	err = json.Unmarshal(data, &migrateData)
	if err != nil {
		return model.MigrationData{}, fmt.Errorf("error unmarshalling json: %v", err)
	}

	return migrateData, nil
}

// todo abstact for different db types
func migrateTable(source *sql.DB, target *sql.DB, table string) error {
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

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().BoolP("dir", "d", false, "if true, the input and the output will be current directory (for input the file pattern is "+MigrateFilePattern+")")
}
