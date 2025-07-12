/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/nanaki-93/randatagen/internal/generate"
	"github.com/nanaki-93/randatagen/internal/model"
	"github.com/nanaki-93/randatagen/internal/writer"
	"github.com/spf13/cobra"
	"log/slog"

	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

const GenFilePattern = "generate*.json"

var validDb = []string{"postgres", "oracle"}

// generateCmd represents the gen command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate some random data for sql or nosql db",
	Long:  `generate some random data for sql or nosql db`,
	Run:   execGenCmd,
	Args:  validateArgs(),
}

func validateArgs() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		isDir, _ := cmd.Flags().GetBool("dir")

		if !isDir {
			if len(args) != 1 {
				return fmt.Errorf("you have to use exactly 1 arg, the input file")
			}
		}
		return nil
	}
}

func execGenCmd(cmd *cobra.Command, args []string) {

	isDir, _ := cmd.Flags().GetBool("dir")
	isToFile, _ := cmd.Flags().GetBool("toFile")
	ranDataService := writer.NewRandataService(isToFile)
	var err error

	if isDir {
		err = processDir(ranDataService)
	} else {
		err = processFile(args[0], ranDataService)
	}

	if err != nil {
		slog.Error("Error generating data", "error", err)
		os.Exit(1)
	}
}

func processDir(w writer.RanDataWriter) error {
	inputFiles, err := filepath.Glob(GenFilePattern)
	if err != nil {
		return fmt.Errorf("error finding files: %v", err)
	}

	fmt.Printf("Input files: %v\n", inputFiles)

	for _, inputFile := range inputFiles {
		ranData, err := getRanData(inputFile)
		if err != nil {
			return fmt.Errorf("error getting ran data from file %s: %v", inputFile, err)
		}
		if err = genSingleInsert(ranData, w); err != nil {
			return fmt.Errorf("error generating single insert for file %s: %v", inputFile, err)
		}
	}
	return nil
}

func processFile(inputFile string, w writer.RanDataWriter) error {
	fmt.Printf("input file: %s,\n", inputFile)
	dataGen, err := getRanData(inputFile)
	if err != nil {
		return fmt.Errorf("error getting ran data from file %s: %v", inputFile, err)
	}

	if err := genSingleInsert(dataGen, w); err != nil {
		return fmt.Errorf("error generating single insert: %v", err)
	}
	return nil
}

func genSingleInsert(dataGen model.RanData, w writer.RanDataWriter) error {
	err := w.Open(dataGen)
	if err != nil {
		return fmt.Errorf("error opening writer: %v", err)
	}
	defer func() {
		if cerr := w.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close error: %w", cerr)
		}
	}()
	dbType := strings.ToLower(dataGen.DbType)

	if !slices.Contains(validDb, dbType) {
		return fmt.Errorf("dbGenerator type %s is not supported", dbType)
	}

	getDataProvider, ok := generate.ProviderFactories[dbType]
	if !ok {
		return fmt.Errorf("dbGenerator type %s is not supported", dbType)
	}

	dbGenerator := generate.NewGeneratorService(getDataProvider())
	fmt.Println("[+] Generating Data for " + dataGen.Target.DbTable)

	insertSqlSlice := dbGenerator.GenerateSql(dataGen)

	for _, insertSql := range insertSqlSlice {
		_, err := w.Write([]byte(insertSql))
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
	}
	fmt.Println("Data generated successfully!")
	return err
}

func getRanData(inputFile string) (model.RanData, error) {
	err := checkFileExists(inputFile)
	if err != nil {
		return model.RanData{}, fmt.Errorf("error checking file existence: %v", err)
	}
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return model.RanData{}, fmt.Errorf("error reading file: %v", err)

	}
	var ranData model.RanData
	err = json.Unmarshal(data, &ranData)
	if err != nil {
		return model.RanData{}, fmt.Errorf("error unmarshalling json: %v", err)
	}

	outputFilePath := strings.Replace(inputFile, ".json", "", 1) + "-output-" + strconv.Itoa(int(time.Now().UnixMilli())) + ".sql"
	ranData.OutputFilePath = outputFilePath
	return ranData, nil
}

func checkFileExists(inputFilePath string) error {
	exists := fileExists(inputFilePath)
	if !exists {
		return fmt.Errorf("[!] %s\n", inputFilePath+" does not exist")
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().BoolP("dir", "d", false, "if true, the input and the output will be current directory (for input the file pattern is "+GenFilePattern+")")
	generateCmd.Flags().BoolP("toFile", "f", false, "if true, write the output to a file")

}
