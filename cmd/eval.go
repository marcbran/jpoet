package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcbran/jpoet/pkg/jpoet"
	"github.com/spf13/cobra"
)

var evalCmd = &cobra.Command{
	Use:   "eval [flags] input",
	Short: "Jsonnext eval is a simple tool to eval Jsonnet files",
	Long:  ``,

	DisableAutoGenTag: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		directory, err := cmd.Flags().GetString("directory")
		if err != nil {
			return err
		}
		code, err := cmd.Flags().GetBool("code")
		if err != nil {
			return err
		}
		str, err := cmd.Flags().GetBool("string")
		if err != nil {
			return err
		}

		arg := ""
		if len(args) > 0 {
			arg = args[0]
		}
		if !code {
			switch arg {
			case "":
				arg = "main.jsonnet"
			case "-":
				arg, err = bufio.NewReader(os.Stdin).ReadString('\n')
				if err != nil {
					return fmt.Errorf("error reading input: %v", err)
				}
				code = true
			}
		}
		var input jpoet.Input
		if code {
			input = jpoet.SnippetInput{Filename: "main.jsonnet", Snippet: arg}
		} else {
			input = jpoet.FileInput{Filename: filepath.Join(directory, arg)}
		}

		plugins, err := jpoet.NewPluginsDir(filepath.Join(directory, ".jpoet", "plugins"))
		if err != nil {
			return err
		}

		err = jpoet.NewEval().
			PluginSet(plugins...).
			Input(input).
			Serialize(!str).
			Eval()
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	evalCmd.Flags().StringP("directory", "d", ".", "Context directory for the evaluation")
	evalCmd.Flags().BoolP("code", "c", false, "Treat provided input as code")
	evalCmd.Flags().BoolP("string", "s", false, "Output raw string instead of Json serialization but fails if evaluated output is not a string")
}
