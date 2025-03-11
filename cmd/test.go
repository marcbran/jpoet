package cmd

import (
	"github.com/marcbran/jsonnet-kit/internal/test"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [flags] directory",
	Short: "Jsonnet test is a simple tool to run tests for Jsonnet files",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return test.RunDir(args[0])
	},
}
