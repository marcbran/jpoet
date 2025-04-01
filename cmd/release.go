package cmd

import (
	"github.com/marcbran/jsonnet-kit/internal/release"
	"github.com/spf13/cobra"
	"path"
	"path/filepath"
)

var releaseCmd = &cobra.Command{
	Use:   "release [flags] directory",
	Short: "Jsonnet release is a simple tool to release Jsonnet modules",
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
		repo, err := cmd.Flags().GetString("repo")
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
		authMethod, err := release.NewAuthMethodFromEnv()
		if err != nil {
			return err
		}
		err = release.Run(cmd.Context(), source, repo, branch, p, authMethod)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	releaseCmd.Flags().StringP("repo", "r", "", "The git repository targeted for release")
	releaseCmd.Flags().StringP("branch", "b", "", "The module's branch name")
	releaseCmd.Flags().StringP("path", "p", "", "The folder name of the module")
}
