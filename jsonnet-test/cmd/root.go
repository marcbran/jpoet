package cmd

import (
	"fmt"
	"github.com/marcbran/jsonnet-libs/jsonnet-test/internal"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "jsonnet-test [flags] directory",
	Short: "Jsonnet test is a simple tool to run tests for Jsonnet files",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.TestDir(args[0])
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
