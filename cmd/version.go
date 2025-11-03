package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	gitSHA  string // value injected in compilation-time
	version string // value injected in compilation-time
)

func versionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of konfluxctl",
		Long:  "Print the version number of konfluxctl",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("konfluxctl %s (%s)\n", version, gitSHA)
			return nil
		},
	}
	return versionCmd
}
