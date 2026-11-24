package cmd

import (
	"fmt"

	"github.com/oxio/go-cli-tool-template/internal/hello"
	"github.com/spf13/cobra"
)

func newHelloCmd() *cobra.Command {
	var name string

	helloCmd := &cobra.Command{
		Use:   "hello",
		Short: "Print a Hello World greeting",
		Example: `  # Print the default greeting
  go-cli-tool-template hello

  # Print a personalized greeting
  go-cli-tool-template hello --name Alice`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), hello.Greeting(name))
			return err
		},
	}

	helloCmd.Flags().StringVarP(&name, "name", "n", "", "Name to greet (defaults to World)")

	return helloCmd
}
