package pkg

import (
	"github.com/marcbran/jpoet/internal/pkg"
	"github.com/spf13/cobra"
	"path/filepath"
)

var buildCmd = &cobra.Command{
	Use:   "build [flags] directory",
	Short: "Builds Jsonnet packages",
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
		err = pkg.Build(cmd.Context(), pkgDir, buildDir)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	buildCmd.Flags().StringP("build", "b", "build", "The path to the build directory, relative to the package directory")
}
