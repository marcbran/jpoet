package cmd

import (
	"fmt"
	"github.com/marcbran/devsonnet/cmd/pkg"
	"github.com/marcbran/devsonnet/cmd/repo"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "devsonnet",
	Short: "Jsonnet kit is a toolkit that provides a number of different jsonnet-related tools",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(pkg.Cmd)
	rootCmd.AddCommand(repo.Cmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
