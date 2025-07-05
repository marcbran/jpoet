package cmd

import (
	"github.com/marcbran/jsonnet-kit/internal/bundle"
	"github.com/spf13/cobra"
)

var bundleCmd = &cobra.Command{
	Use:   "bundle [flags] directory",
	Short: "Jsonnet bundle is a simple tool to bundle Jsonnet modules",
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
		err = bundle.Run(cmd.Context(), input, output)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	bundleCmd.Flags().StringP("input", "i", "", "The path to the main input file")
	bundleCmd.Flags().StringP("output", "o", "", "The path to the bundled output file")
}
