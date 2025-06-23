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
	"github.com/spf13/cobra"

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
		isDir, err := cmd.Flags().GetBool("dir")
		PrintErrAndExit(err)

		if !isDir {
			if len(args) != 1 {
				return fmt.Errorf("you have to use exactly 1 arg, the input file")
			}
		}
		return nil
	}
}

func execGenCmd(cmd *cobra.Command, args []string) {

	isDir, err := cmd.Flags().GetBool("dir")
	PrintErrAndExit(err)

	isToFile, err := cmd.Flags().GetBool("toFile")
	PrintErrAndExit(err)

	ranDataService := generate.NewRandataService(isToFile)

	if !isDir {
		fromSingleFile(args[0], ranDataService)
		return
	}
	FromCurrentDir(ranDataService)
}

func FromCurrentDir(w generate.RanDataWriter) {
	inputFiles, err := filepath.Glob(GenFilePattern)
	PrintErrAndExit(err)

	fmt.Printf("Input files: %v\n", inputFiles)

	for _, inputFile := range inputFiles {
		ranData, err := getRanData(inputFile)
		PrintErrAndExit(err)
		genSingleInsert(ranData, w)
	}
}

func fromSingleFile(inputFile string, w generate.RanDataWriter) {
	dataGen := getDataGenFromArgs(inputFile)

	genSingleInsert(dataGen, w)
}

func getDataGenFromArgs(inputFile string) model.RanData {
	fmt.Printf("input file: %s,\n", inputFile)
	inputFilePath, err := checkFilePath(inputFile)
	PrintErrAndExit(err)

	dataGen, err := getRanData(inputFilePath)
	PrintErrAndExit(err)

	return dataGen
}

func genSingleInsert(dataGen model.RanData, w generate.RanDataWriter) {
	w.Open(dataGen)
	defer w.Close()

	if !slices.Contains(validDb, strings.ToLower(dataGen.Target.DbType)) {
		fmt.Printf("[!] dbTemplate type %s is not supported\n", dataGen.Target.DbType)
		os.Exit(1)
	}

	var dataGenerator generate.DataGenerator
	if dataGen.Target.DbType == "postgres" {
		dataGenerator = generate.NewPostgresTemplate()
	} else if dataGen.Target.DbType == "oracle" {
		dataGenerator = generate.NewOracleTemplate()
	} else {
		fmt.Printf("[!] dbTemplate type %s is not supported\n", dataGen.Target.DbType)
		os.Exit(1)
	}
	dbTemplate := generate.NewService(dataGenerator)
	fmt.Println("[+] Generating insert statements for " + dataGen.Target.DbTable)

	insertSqlSlice := dbTemplate.GetSqlTemplate(dataGen)

	for _, insertSql := range insertSqlSlice {
		_, err := w.Write([]byte(insertSql))
		PrintErrAndExit(err)
	}
	fmt.Println("Successfully inserted into database!")
}

func getRanData(inputFile string) (model.RanData, error) {
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

func checkFilePath(inputFilePath string) (string, error) {
	exists := fileExists(inputFilePath)
	if !exists {
		return "", fmt.Errorf("[!] %s\n", inputFilePath+" does not exist")
	}
	return inputFilePath, nil
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
