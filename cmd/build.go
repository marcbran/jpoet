package cmd

import (
	"github.com/marcbran/jsonnet-kit/internal/pkg"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [flags] directory",
	Short: "Jsonnet build is a simple tool to build Jsonnet modules",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := cmd.Flags().GetString("input")
		if err != nil {
			return err
		}
		output, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}
		err = pkg.Build(cmd.Context(), input, output)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	buildCmd.Flags().StringP("input", "i", ".", "The path to the main input directory")
	buildCmd.Flags().StringP("output", "o", "out", "The path to the packaged output directory")
}
