package cmd

import (
	"github.com/marcbran/jsonnet-kit/internal/manifest"
	"github.com/spf13/cobra"
)

var manifestCmd = &cobra.Command{
	Use:   "manifest [flags] directory",
	Short: "Jsonnet manifest is a simple tool to manifests text files",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		dirname := "."
		if len(args) > 0 {
			dirname = args[0]
		}
		jpath, err := cmd.Flags().GetStringArray("jpath")
		if err != nil {
			return err
		}
		err = manifest.RunDir(dirname, jpath)
		if err != nil {
			return err
		}
		return nil
	},
}
