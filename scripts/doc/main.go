package main

import (
	"bytes"
	"fmt"
	"github.com/marcbran/jpoet/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"log"
	"os"
	"strings"
)

func main() {
	var buf bytes.Buffer

	err := generateMarkdown(cmd.Cmd, &buf)
	if err != nil {
		log.Fatalf("Error generating documentation: %v", err)
	}

	f, err := os.Create("README.md")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer f.Close()

	_, err = f.Write(buf.Bytes())
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	fmt.Println("Documentation written to README.md")
}

func generateMarkdown(cmd *cobra.Command, buf *bytes.Buffer) error {
	if err := doc.GenMarkdownCustom(cmd, buf, func(s string) string {
		return fmt.Sprintf("#%s", strings.TrimSuffix(strings.ReplaceAll(s, "_", "-"), ".md"))
	}); err != nil {
		return err
	}
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := generateMarkdown(c, buf); err != nil {
			return err
		}
	}
	return nil
}
