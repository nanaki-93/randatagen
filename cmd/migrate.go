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
	"io"
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

	isDir, err := cmd.Flags().GetBool("dir")
	if err != nil {
		slog.Error("Error getting dir flag", "error", err)
		os.Exit(1)
	}

	if !isDir {
		migrateData, err := getMigrateFromInputFile(args[0])
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
	err = MigrateAll()
	if err != nil {
		slog.Error("Error migrating all schemas", "error", err)
		os.Exit(1)
	}

}

func MigrateSchema(migrateData model.MigrateData) error {
	sourceConn, targetConn := getDbConnections(migrateData)
	defer CloseWithErrorCheck(sourceConn, "sourceConn")
	defer CloseWithErrorCheck(targetConn, "targetConn")

	tables, err := GetTablesFromSource(migrateData, sourceConn)
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
	return nil
}

func GetTablesFromSource(migrateData model.MigrateData, sourceConn *sql.DB) ([]string, error) {
	rows, err := sourceConn.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = $1", migrateData.Source.DbSchema)
	if err != nil {
		return nil, fmt.Errorf("error getting tables from source: %w", err)
	}
	defer CloseWithErrorCheck(rows, "rows")

	tables := make([]string, 0)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %w", err)
		}
		tables = append(tables, tableName)
	}
	slog.Info("Tables to migrate: ", tables)
	return tables, nil
}

func CloseWithErrorCheck(c io.Closer, name string) {
	if err := c.Close(); err != nil {
		slog.Error(fmt.Sprintf("Error closing %s: %v", name, err))
		os.Exit(1)
	}
}

func getDbConnections(data model.MigrateData) (*sql.DB, *sql.DB) {
	//todo return the right db connections based on the data
	sourceConn := db.GetConn(data.Source)
	targetConn := db.GetConn(data.Target)
	return sourceConn, targetConn
}

func MigrateAll() error {
	inputFiles, err := filepath.Glob(MigrateFilePattern)
	if err != nil {
		fmt.Printf("[!] error reading files: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Input files: %v\n", inputFiles)
	for _, inputFile := range inputFiles {
		migData, err := getMigrateFromInputFile(inputFile)
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

func getMigrateFromInputFile(inputFile string) (model.MigrateData, error) {
	fmt.Printf("input file: %s,\n", inputFile)
	inputFilePath, err := checkFilePath(inputFile)
	if err != nil {
		return model.MigrateData{}, fmt.Errorf("error checking file path: %v", err)
	}
	migrateData, err := getMigrateData(inputFilePath)
	if err != nil {
		return model.MigrateData{}, fmt.Errorf("error getting migrate data: %v", err)
	}
	return migrateData, nil
}

func getMigrateData(inputFile string) (model.MigrateData, error) {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return model.MigrateData{}, fmt.Errorf("error reading file: %v", err)

	}
	var dataGen model.MigrateData
	err = json.Unmarshal(data, &dataGen)
	if err != nil {
		return model.MigrateData{}, fmt.Errorf("error unmarshalling json: %v", err)
	}

	return dataGen, nil
}

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
