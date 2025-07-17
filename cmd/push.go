package cmd

import (
	"github.com/marcbran/jsonnet-kit/internal/repo"
	"github.com/spf13/cobra"
	"path"
	"path/filepath"
)

var pushCmd = &cobra.Command{
	Use:   "push [flags] directory",
	Short: "Jsonnet push is a simple tool to push Jsonnet modules",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		source := "."
		if len(args) > 0 {
			source = args[0]
		}
		err := cmd.MarkFlagRequired("repo")
		if err != nil {
			return err
		}
		r, err := cmd.Flags().GetString("repo")
		if err != nil {
			return err
		}
		branch, err := cmd.Flags().GetString("branch")
		if err != nil {
			return err
		}
		p, err := cmd.Flags().GetString("path")
		if err != nil {
			return err
		}
		if branch == "" && p != "" {
			branch = p
		} else if branch != "" && p == "" {
			p = branch
		} else if branch == "" && p == "" {
			abs, err := filepath.Abs(source)
			if err != nil {
				return err
			}
			branch = path.Base(abs)
			p = path.Base(abs)
		}
		authMethod, err := repo.NewAuthMethodFromEnv()
		if err != nil {
			return err
		}
		err = repo.Push(cmd.Context(), source, r, branch, p, authMethod)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	pushCmd.Flags().StringP("repo", "r", "", "The git repository targeted for the push")
	pushCmd.Flags().StringP("branch", "b", "", "The module's branch name")
	pushCmd.Flags().StringP("path", "p", "", "The folder name of the module")
}
