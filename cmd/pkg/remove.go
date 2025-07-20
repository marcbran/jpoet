package pkg

import (
	"github.com/marcbran/devsonnet/internal/repo"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [flags] directory",
	Short: "Removes Jsonnet packages from the target repository",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		pkgDir := "."
		if len(args) > 0 {
			pkgDir = args[0]
		}
		authMethod, err := repo.NewAuthMethodFromEnv()
		if err != nil {
			return err
		}
		err = repo.Remove(cmd.Context(), pkgDir, authMethod)
		if err != nil {
			return err
		}
		return nil
	},
}
