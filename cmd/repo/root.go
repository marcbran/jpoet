package repo

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "repo",
	Short: "Subcommands for managing target repositories",
	Long:  ``,
}

func init() {
	Cmd.AddCommand(indexCmd)
}
