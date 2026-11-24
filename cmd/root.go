package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// NewRootCmd constructs the root Cobra command for proto2md.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "proto2md",
		Short: "Tool for converting Protobuf file definitions into Markdown documentation",
		Long:  "Tool for converting Protobuf file definitions into Markdown documentation",
		Example: `  # Print a greeting
  proto2md hello

  # Print a personalized greeting
  proto2md hello --name Alice

  # Print the CLI version
  proto2md version`,
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
