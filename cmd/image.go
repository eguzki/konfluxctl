package cmd

import (
	"github.com/eguzki/konfluxctl/cmd/image"
	"github.com/spf13/cobra"
)

func imageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Docker/OCI image related utility",
		Long:  "Docker/OCI image related utility",
	}

	cmd.AddCommand(image.MetadataCommand())
	return cmd
}
