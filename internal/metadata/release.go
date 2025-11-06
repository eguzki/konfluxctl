package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	applicationapi "github.com/konflux-ci/application-api/api/v1alpha1"
	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
	"github.com/samber/lo"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eguzki/konfluxctl/internal/utils"
)

type ReleaseAdvisory struct {
	InternalURL string `json:"internal_url"`
	URL         string `json:"url"`
}

type ReleaseArtifacts struct {
	Advisory ReleaseAdvisory `json:"advisory"`
}

type ReleaseElement konfluxapi.Release

func (r *ReleaseElement) String() string {
	return fmt.Sprintf("%s: %s", "Release", r.Name)
}

func (r *ReleaseElement) Visit(path *Path) {
	path.Release = &r.Name
	path.Advisory = ptr.To("<unknown>")

	if r.Status.Artifacts != nil {
		var artifacts ReleaseArtifacts
		if err := json.Unmarshal(r.Status.Artifacts.Raw, &artifacts); err == nil {
			path.Advisory = &artifacts.Advisory.URL
		}
	}
}

func (r *ReleaseElement) Children(ctx context.Context, k8sClient client.Client, imageURL *utils.ImageURL) ([]Element, error) {
	snapshot := &applicationapi.Snapshot{}
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: r.Namespace,
		Name:      r.Spec.Snapshot,
	}, snapshot)
	if err != nil {
		return nil, err
	}

	component, ok := lo.Find(snapshot.Spec.Components, func(comp applicationapi.SnapshotComponent) bool {
		containerImageURL, err := utils.ParseImageURL(comp.ContainerImage)
		if err != nil {
			return false
		}
		return containerImageURL.Digest() == imageURL.Digest()
	})

	if !ok {
		return nil, nil
	}

	return []Element{&SnapshotElement{
		rawSnapshot: snapshot,
		component:   &component,
	}}, nil
}
