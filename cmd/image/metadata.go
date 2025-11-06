package image

import (
	"context"
	"fmt"
	"os"

	applicationapi "github.com/konflux-ci/application-api/api/v1alpha1"
	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
	"github.com/spf13/cobra"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/eguzki/konfluxctl/internal/metadata"
	"github.com/eguzki/konfluxctl/internal/utils"
)

//konfluxctl image metadata --image IMAGE_URL

var (
	imageURL            string
	imageMetadataFormat string
)

func MetadataCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "Returns Docker/OCI image related konflux metadata",
		Long:  "Returns Docker/OCI image related konflux metadata",
		RunE:  runMetadata,
	}

	cmd.Flags().StringVar(&imageURL, "image", "", "Docker/OCI image URL (required)")
	cmd.Flags().StringVarP(&imageMetadataFormat, "output-format", "o", "yaml", "Output format: 'yaml' or 'json'.")

	if err := cmd.MarkFlagRequired("image"); err != nil {
		fmt.Println("Error setting 'image' flag as required:", err)
		os.Exit(1)
	}

	return cmd
}

func runMetadata(cmd *cobra.Command, args []string) error {
	scheme := k8sruntime.NewScheme()
	if err := konfluxapi.AddToScheme(scheme); err != nil {
		return err
	}
	if err := applicationapi.AddToScheme(scheme); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	configuration, err := config.GetConfig()
	if err != nil {
		return err
	}

	k8sClient, err := client.New(configuration, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	// 1. Parse the reference string
	imageRef, err := utils.ParseImageURL(imageURL)
	if err != nil {
		return err
	}

	rpaList, err := metadata.ReleasePlanAdmissionList(ctx, k8sClient, imageRef.FamiliarName())
	if err != nil {
		return err
	}

	paths, err := metadata.DepthFirstSearch(ctx, k8sClient, imageRef, rpaList)

	if len(paths) == 0 {
		fmt.Println("🧐 No metadata found")
		return nil

	}

	fmt.Println("Paths =====")
	for _, path := range paths {
		fmt.Println("Path =====")
		fmt.Println(path)
		fmt.Println("")
	}
	fmt.Println("Paths =====")

	return nil
}
