package cmd

import (
	"fmt"
	"github.com/marcbran/jpoet/cmd/pkg"
	"github.com/marcbran/jpoet/cmd/repo"
	"github.com/spf13/cobra"
	"os"
)

var Cmd = &cobra.Command{
	Use:   "jpoet",
	Short: "Jpoet provides a set of tools that makes it easier to write and reuse Jsonnet code.",
	Long: `Jsonnet is a powerful and flexible configuration language that extends JSON with advanced programming features.
It supports conditionals, loops, functions, and object-oriented constructs, enabling more concise and reusable configuration code.

In addition to these language features, Jsonnet provides additional tooling.
For example to import external Jsonnet files and to write output to multiple files.
Overall, this promotes modular design and facilitates reuse across different projects, whether authored internally or sourced externally.

However, the standard Jsonnet toolchain does have limitations.
Extending the standard library with custom native functions, for instance, is non-trivial.
It requires developers to create a dedicated Go binary, which must then replace the default Jsonnet CLI to execute configurations.

While the inclusion of additional native functions introduces potential security and side-effect risks, the absence of commonly needed features (such as regular expression support) makes this functionality highly desirable.
A plugin mechanism that allows selective inclusion of safe, useful functions, without having to write new binaries, would significantly improve developer experience.

This is where Jpoet comes into play.
Jpoet introduces a plugin management system built on [go-plugin](https://github.com/hashicorp/go-plugin), the same robust framework used in projects like Terraform and Vault.
With Jpoet, developers can install Jsonnet plugins locally and evaluate configurations via the Jpoet binary.

For detailed usage instructions, refer to the documentation of the respective commands.`,

	DisableAutoGenTag: true,
}

func init() {
	Cmd.AddCommand(testCmd)
	Cmd.AddCommand(installCmd)
	Cmd.AddCommand(evalCmd)
	Cmd.AddCommand(pkg.Cmd)
	Cmd.AddCommand(repo.Cmd)
}

func Execute() {
	if err := Cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
