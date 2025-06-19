/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"

	"fmt"
	_ "github.com/lib/pq"
	"github.com/nanaki-93/randatagen/internal/model"
	"github.com/nanaki-93/randatagen/internal/service"
	"github.com/nanaki-93/randatagen/internal/template"
	"github.com/spf13/cobra"

	"log"
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
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	isToFile, err := cmd.Flags().GetBool("toFile")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ranDataService := service.NewRandataService(isToFile)

	if !isDir {
		fromSingleFile(args[0], ranDataService)
		return
	}
	allRandataFromCurrentDir(ranDataService)
}

func allRandataFromCurrentDir(w service.RanDataWriter) {
	inputFiles, err := filepath.Glob(GenFilePattern)
	if err != nil {
		fmt.Printf("[!] error reading files: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Input files: %v\n", inputFiles)
	for _, inputFile := range inputFiles {
		dataGen, err := getDataGen(inputFile)
		if err != nil {
			fmt.Printf("[!] %s\n", err)
			os.Exit(1)
		}
		genInsertSqlFile(dataGen, w)
	}
}

func fromSingleFile(inputFile string, w service.RanDataWriter) {
	dataGen := getDataGenFromArgs(inputFile)

	genInsertSqlFile(dataGen, w)
}

func getDataGenFromArgs(inputFile string) model.GenerateData {
	fmt.Printf("input file: %s,\n", inputFile)
	inputFilePath, err := checkFilePath(inputFile)
	if err != nil {
		fmt.Printf("[!] %s\n", err)
		os.Exit(1)
	}
	dataGen, err := getDataGen(inputFilePath)
	if err != nil {
		fmt.Printf("[!] %s\n", err)
		os.Exit(1)
	}
	return dataGen
}

func genInsertSqlFile(dataGen model.GenerateData, w service.RanDataWriter) {
	w.Open(dataGen)
	defer w.Close()

	if !slices.Contains(validDb, strings.ToLower(dataGen.Target.DbType)) {
		fmt.Printf("[!] dbTemplate type %s is not supported\n", dataGen.Target.DbType)
		os.Exit(1)
	}

	var dataGenerator template.DataGenerator
	if dataGen.Target.DbType == "postgres" {
		dataGenerator = template.NewPostgresTemplate()
	} else if dataGen.Target.DbType == "oracle" {
		dataGenerator = template.NewOracleTemplate()
	} else {
		fmt.Printf("[!] dbTemplate type %s is not supported\n", dataGen.Target.DbType)
		os.Exit(1)
	}
	dbTemplate := template.NewService(dataGenerator)
	fmt.Println("[+] Generating insert statements for " + dataGen.Target.DbTable)

	insertSqlSlice := dbTemplate.GetSqlTemplate(dataGen)

	for _, insertSql := range insertSqlSlice {
		_, err := w.Write([]byte(insertSql))
		if err != nil {
			log.Fatalf("[!] %s\n", err)
		}
	}
	fmt.Println("Successfully inserted into database!")
}

func getDataGen(inputFile string) (model.GenerateData, error) {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return model.GenerateData{}, fmt.Errorf("error reading file: %v", err)

	}
	var dataGen model.GenerateData
	err = json.Unmarshal(data, &dataGen)
	if err != nil {
		return model.GenerateData{}, fmt.Errorf("error unmarshalling json: %v", err)
	}

	outputFilePath := strings.Replace(inputFile, ".json", "", 1) + "-output-" + strconv.Itoa(int(time.Now().UnixMilli())) + ".sql"
	dataGen.OutputFilePath = outputFilePath
	return dataGen, nil
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
