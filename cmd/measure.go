/*
Copyright © 2024 vantrong291 <vantrong291@gmail.com>.
*/
package cmd

import (
	"benchmark-tool/measure"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

const maxAPI = 2

var (
	fileInputPath string
	outputPath    string
	apiAddresses  []string
)

// measureCmd represents the measure command
var measureCmd = &cobra.Command{
	Use:   "measure",
	Short: "Measure benchmark for a http API",
	Long: `Measure benchmark for a http API.
	
Copyright © 2024 vantrong291 <vantrong291@gmail.com>.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println(fileInputPath, outputPath, apiAddresses)
		if len(apiAddresses) > maxAPI {
			fmt.Println(chalk.Red, fmt.Sprintf("Error: The maximum number of apis allowed is %d", maxAPI))
			os.Exit(1)
		}

		fileInput, err := os.Stat(fileInputPath)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println(chalk.Red, "Error : Input file does not existed")
			os.Exit(1)
		}
		if fileInput.IsDir() {
			fmt.Println(chalk.Red, "Error : Input path is a folder, not a file")
			os.Exit(1)
		}

		measurement := measure.NewMeasurement(apiAddresses, fileInputPath, outputPath)
		err = measurement.Run()
		if err != nil {
			fmt.Println(chalk.Red, fmt.Sprintf("Error : %s", err))
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(measureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	var fileInputCmd = "file-input"
	measureCmd.PersistentFlags().StringVarP(&fileInputPath, fileInputCmd, "i", "", "Input file path, contain testcase and concurrent config")
	measureCmd.MarkPersistentFlagRequired(fileInputCmd)
	// viper.BindPFlag(fileInputCmd, measureCmd.PersistentFlags().Lookup(fileInputCmd))

	var apiAddressCmd = "api"
	measureCmd.PersistentFlags().StringSliceVarP(&apiAddresses, apiAddressCmd, "a", []string{}, "API addresses for benchmark measurement")
	measureCmd.MarkPersistentFlagRequired(apiAddressCmd)

	// viper.BindPFlag(apiAddressCmd, measureCmd.PersistentFlags().Lookup(apiAddressCmd))

	var outputCmd = "output"
	measureCmd.PersistentFlags().StringVarP(&outputPath, outputCmd, "o", "./output", "Output folder path")
	// viper.BindPFlag(outputCmd, measureCmd.PersistentFlags().Lookup(outputCmd))
}
