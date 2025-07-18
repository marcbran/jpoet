package cmd

import (
	"github.com/marcbran/devsonnet/internal/repo"
	"github.com/spf13/cobra"
	"path/filepath"
)

var pushCmd = &cobra.Command{
	Use:   "push [flags] directory",
	Short: "Jsonnet push is a simple tool to push Jsonnet modules",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		pkgDir := "."
		if len(args) > 0 {
			pkgDir = args[0]
		}
		buildDir, err := cmd.Flags().GetString("build")
		if err != nil {
			return err
		}
		buildDir = filepath.Join(pkgDir, buildDir)
		authMethod, err := repo.NewAuthMethodFromEnv()
		if err != nil {
			return err
		}
		err = repo.Push(cmd.Context(), pkgDir, buildDir, authMethod)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	pushCmd.Flags().StringP("build", "b", "build", "The path to the build directory, relative to the package directory")
}
