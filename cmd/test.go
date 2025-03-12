package cmd

import (
	"github.com/marcbran/jsonnet-kit/internal/test"
	"github.com/spf13/cobra"
	"os"
)

var testCmd = &cobra.Command{
	Use:   "test [flags] directory",
	Short: "Jsonnet test is a simple tool to run tests for Jsonnet files",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		run, err := test.RunDir(args[0])
		if err != nil {
			return err
		}
		if run.PassedCount < run.TotalCount {
			os.Exit(1)
		}
		os.Exit(0)
		return nil
	},
}
