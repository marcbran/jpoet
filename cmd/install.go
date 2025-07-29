package cmd

import (
	"github.com/marcbran/jpoet/internal/pkg"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [flags] directory",
	Short: "Jsonnet test is a simple tool to install tests for Jsonnet files",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		pkgDir := "."
		if len(args) > 0 {
			pkgDir = args[0]
		}
		err := pkg.Install(cmd.Context(), pkgDir)
		if err != nil {
			return err
		}
		return nil
	},
}
