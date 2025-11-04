package image

import (
	"context"
	_ "crypto/sha256"
	"fmt"
	"log"
	"os"

	"github.com/distribution/reference"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
	_, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	configuration, err := config.GetConfig()
	if err != nil {
		return err
	}

	_, err = client.New(configuration, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}

	//topologyKey := client.ObjectKey{Name: "topology", Namespace: topologyNS}
	//topologyConfigMap := &corev1.ConfigMap{}
	//err = k8sClient.Get(ctx, topologyKey, topologyConfigMap)
	//logf.Log.V(1).Info("Reading topology ConfigMap", "object", topologyKey, "error", err)
	//if err != nil {
	//	return err
	//}

	// 1. Parse the reference string
	ref, err := reference.ParseAnyReference(imageURL)
	if err != nil {
		log.Fatalf("Error parsing image reference: %v", err)
	}

	// 2. Extract Hostname and Path (Repository)
	named, ok := ref.(reference.Named)
	if !ok {
		log.Fatalf("Reference is not a Named reference: %s", ref.String())
	}

	// SplitHostname returns the domain and the remainder (the path/repository)
	hostname := reference.Domain(named)
	path := reference.Path(named)
	familiarName := reference.FamiliarName(named)

	// 3. Extract Digest
	canonical, ok := ref.(reference.Canonical)
	if !ok {
		log.Fatalf("Reference does not contain a digest: %s", ref.String())
	}
	digest := canonical.Digest().String()

	// 4. Output the results
	fmt.Println("--- Parsed OCI Image Reference ---")
	fmt.Printf("Original: %s\n", imageURL)
	fmt.Printf("Hostname: %s\n", hostname)
	fmt.Printf("FamiliarName: %s\n", familiarName)
	fmt.Printf("Path (Repository): %s\n", path)
	fmt.Printf("Digest: %s\n", digest)

	return nil
}
