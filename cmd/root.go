package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "jsonnet-kit",
	Short: "Jsonnet kit is a toolkit that provides a number of different jsonnet-related tools",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(testCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
