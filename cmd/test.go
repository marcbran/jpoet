package cmd

import (
	"encoding/json"
	"github.com/marcbran/devsonnet/internal/terminal"
	"github.com/marcbran/devsonnet/internal/test"
	"github.com/spf13/cobra"
	"os"
)

var testCmd = &cobra.Command{
	Use:   "test [flags] directory",
	Short: "Jsonnet test is a simple tool to run tests for Jsonnet files",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		dirname := "."
		if len(args) > 0 {
			dirname = args[0]
		}
		j, err := cmd.Flags().GetBool("json")
		if err != nil {
			return err
		}
		run, err := test.RunDir(dirname)
		if err != nil {
			return err
		}
		if j {
			b, err := json.Marshal(run)
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(b)
			if err != nil {
				return err
			}
		} else {
			terminal.Space()
			for _, result := range run.Results {
				if !result.Equal {
					terminal.Failf("      Name: %s", result.Name)
					terminal.Failf("    Actual: %s", result.Actual)
					terminal.Failf("  Expected: %s", result.Expected)
					terminal.Space()
				}
			}
			if run.PassedCount < run.TotalCount {
				terminal.Failf("Passed: %d/%d", run.PassedCount, run.TotalCount)
			} else {
				terminal.Successf("Passed: %d/%d", run.PassedCount, run.TotalCount)
			}
			terminal.Space()
		}

		if run.PassedCount < run.TotalCount {
			os.Exit(1)
		}
		os.Exit(0)
		return nil
	},
}

func init() {
	testCmd.Flags().BoolP("json", "j", false, "Outputs the test results in JSON")
}
