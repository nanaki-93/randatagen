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
	"github.com/nanaki-93/randatagen/internal/migrate"
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

	getDbProvider, ok := migrate.ProviderFactories[migrateData.DbType]
	if !ok {
		return fmt.Errorf("dbGenerator type %s is not supported", dbType)
	}

	migService := migrate.NewMigrationService(getDbProvider(migrateData))

	source, target, err := migService.MigrationProvider.Open()
	if err != nil {
		return fmt.Errorf("error opening migration provider: %w", err)
	}
	defer func() {
		if cerr := migService.MigrationProvider.Close(source, target); cerr != nil && err == nil {
			err = fmt.Errorf("close error: %w", cerr)
		}
	}()

	tables, err := migService.MigrationProvider.GetTablesToMigrate(source)
	if err != nil {
		return fmt.Errorf("error getting tables from source: %w", err)
	}
	for _, table := range tables {
		err := migService.MigrationProvider.MigrateTable(table)
		if err != nil {
			return fmt.Errorf("error migrating table %s: %w", table, err)
		}
	}
	slog.Info("Migration completed successfully")
	return err
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

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().BoolP("dir", "d", false, "if true, the input and the output will be current directory (for input the file pattern is "+MigrateFilePattern+")")
}
