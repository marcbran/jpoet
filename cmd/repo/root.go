package repo

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "repo",
	Short: "Subcommands for managing target repositories",
	Long:  ``,

	DisableAutoGenTag: true,
}

func init() {
	Cmd.AddCommand(indexCmd)
}
