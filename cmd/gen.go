/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/nanaki-93/randatagen/internal"
	"github.com/nanaki-93/randatagen/internal/model"
	"github.com/spf13/cobra"
	"os"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "generate random database data",
	Long:  `generate random database data`,
	Run:   execGenCmd,
}

func execGenCmd(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile("test.json")
	if err != nil {
		fmt.Errorf("error reading file: %v", err)
	}
	var dataGen model.DataGen
	err = json.Unmarshal(data, &dataGen)
	if err != nil {
		fmt.Errorf("error unmarshalling json: %v", err)
	}
	columns := dataGen.Columns
	insertSql := internal.GetSqlTemplate(columns)

	fmt.Println(insertSql)

}

func init() {
	rootCmd.AddCommand(genCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// genCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// genCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
