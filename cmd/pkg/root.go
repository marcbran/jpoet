package pkg

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "pkg",
	Short: "Subcommands for building packages and managing them in the target repository",
	Long:  ``,
}

func init() {
	Cmd.AddCommand(buildCmd)
	Cmd.AddCommand(pushCmd)
	Cmd.AddCommand(removeCmd)
}
