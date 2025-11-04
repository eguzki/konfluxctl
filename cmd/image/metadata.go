package image

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

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

type ReleasePlanAdmissionDataComponent struct {
	Name       string   `json:"name"`
	Repository string   `json:"repository"`
	Tags       []string `json:"tags"`
}

type ReleasePlanAdmissionDataMapping struct {
	Components []ReleasePlanAdmissionDataComponent `json:"components"`
}

type ReleasePlanAdmissionData struct {
	Mappping ReleasePlanAdmissionDataMapping `json:"mapping"`
}

func releasePlanAdmissionList(ctx context.Context, k8sClient client.Client, imageName string) ([]konfluxapi.ReleasePlanAdmission, error) {
	rpaList := &konfluxapi.ReleasePlanAdmissionList{}
	err := k8sClient.List(ctx, rpaList, client.InNamespace("rhtap-releng-tenant"))
	if err != nil {
		return nil, err
	}

	return lo.Filter(rpaList.Items, func(rpa konfluxapi.ReleasePlanAdmission, index int) bool {
		var data ReleasePlanAdmissionData
		if err := json.Unmarshal(rpa.Spec.Data.Raw, &data); err != nil {
			return false
		}

		return lo.ContainsBy(data.Mappping.Components, func(comp ReleasePlanAdmissionDataComponent) bool {
			return comp.Repository == imageName
		})
	}), nil
}

func runMetadata(cmd *cobra.Command, args []string) error {
	scheme := k8sruntime.NewScheme()
	konfluxapi.AddToScheme(scheme)

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

	rpaList, err := releasePlanAdmissionList(ctx, k8sClient, imageRef.FamiliarName())
	if err != nil {
		return err
	}

	fmt.Println("releasePlanAdmissionList =====")
	for _, rpa := range rpaList {
		fmt.Printf("name: %s\n", rpa.Name)
	}
	fmt.Println("releasePlanAdmissionList =====")

	return nil
}
