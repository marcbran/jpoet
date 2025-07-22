package repo

import (
	"github.com/marcbran/devsonnet/internal/repo"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index [flags] directory",
	Short: "Indexes Jsonnet repository and updates index README",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		r := "."
		if len(args) > 0 {
			r = args[0]
		}
		authMethod, err := repo.NewAuthMethodFromEnv()
		if err != nil {
			return err
		}
		err = repo.Index(cmd.Context(), r, authMethod)
		if err != nil {
			return err
		}
		return nil
	},
}
