package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// NewRootCmd constructs the root Cobra command for go-cli-tool-template.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "go-cli-tool-template",
		Short: "A template CLI tool written in Go",
		Long: `go-cli-tool-template is a starter template for building Go CLI tools
with Cobra, Taskfile, golangci-lint, and GitHub Actions releases.`,
		Example: `  # Print a greeting
  go-cli-tool-template hello

  # Print a personalized greeting
  go-cli-tool-template hello --name Alice

  # Print the CLI version
  go-cli-tool-template version`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(newHelloCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

// Execute runs the root CLI command and exits with a non-zero code on error.
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
