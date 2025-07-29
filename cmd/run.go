package cmd

import (
	"github.com/marcbran/jpoet/internal/pkg"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] filename",
	Short: "Jsonnext run is a simple tool to run Jsonnet files",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}
		err := pkg.Run(filename)
		if err != nil {
			return err
		}
		return nil
	},
}
