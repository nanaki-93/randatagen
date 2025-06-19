/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/nanaki-93/randatagen/internal/model"
	"github.com/nanaki-93/randatagen/internal/service"
	"github.com/nanaki-93/randatagen/internal/template"
	"github.com/spf13/cobra"
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
		fmt.Println(err)
		os.Exit(1)
	}

	if !isDir {
		migrateData := getMigrateFromInputFile(args[0])
		MigrationFromFile(migrateData)
		return
	}
	MigrateAll()

}

func MigrationFromFile(migrateData model.MigrateData) {

	sourceConn := service.GetPostgresConn(migrateData.Source)
	defer sourceConn.Close()
	targetConn := service.GetPostgresConn(migrateData.Target)
	defer targetConn.Close()

	rows, err := sourceConn.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = $1", migrateData.Source.DbSchema)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			panic(err)
		}
		tables = append(tables, tableName)
	}
	fmt.Println("Tables to migrate: ", tables)
	for _, table := range tables {
		migrateTable(sourceConn, targetConn, table)
	}
}
func MigrateAll() {
	inputFiles, err := filepath.Glob(MigrateFilePattern)
	if err != nil {
		fmt.Printf("[!] error reading files: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Input files: %v\n", inputFiles)
	for _, inputFile := range inputFiles {
		migData := getMigrateFromInputFile(inputFile)
		MigrationFromFile(migData)
	}
}

func getMigrateFromInputFile(inputFile string) model.MigrateData {
	fmt.Printf("input file: %s,\n", inputFile)
	inputFilePath, err := checkFilePath(inputFile)
	if err != nil {
		fmt.Printf("[!] %s\n", err)
		os.Exit(1)
	}
	migrateData, err := getMigrateData(inputFilePath)
	if err != nil {
		fmt.Printf("[!] %s\n", err)
		os.Exit(1)
	}
	return migrateData
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

func migrateTable(source *sql.DB, target *sql.DB, table string) {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Error getting current directory: %v", err))
	}
	tmpFile := filepath.Join(currentDir, table+".tmp")

	_, err = source.Exec("COPY " + template.WithDoubleQuote(table) + " TO '" + tmpFile + "' WITH CSV HEADER")
	if err != nil {
		panic(err)
	}
	_, err = target.Exec("COPY " + template.WithDoubleQuote(table) + " FROM '" + tmpFile + "' WITH CSV HEADER")
	if err != nil {
		panic(err)
	}

	err = os.Remove(tmpFile)
	if err != nil {
		panic(err)
	}

}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().BoolP("dir", "d", false, "if true, the input and the output will be current directory (for input the file pattern is "+MigrateFilePattern+")")
}
