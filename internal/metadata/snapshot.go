package metadata

import (
	"context"
	"fmt"

	applicationapi "github.com/konflux-ci/application-api/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eguzki/konfluxctl/internal/utils"
)

type SnapshotElement struct {
	rawSnapshot *applicationapi.Snapshot
	component   *applicationapi.SnapshotComponent
}

func (s *SnapshotElement) String() string {
	return fmt.Sprintf("%s: %s", "Snapshot", s.rawSnapshot.Name)
}

func (s *SnapshotElement) Visit(path *Path) {
	path.Snapshot = &s.rawSnapshot.Name
	path.ComponentName = &s.component.Name
	path.SourceRevision = &s.component.Source.GitSource.Revision
	path.SourceURL = &s.component.Source.GitSource.URL
}

func (s *SnapshotElement) Children(ctx context.Context, k8sClient client.Client, imageURL *utils.ImageURL) ([]Element, error) {
	application := &applicationapi.Application{}
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: s.rawSnapshot.Namespace,
		Name:      s.rawSnapshot.Spec.Application,
	}, application)
	if err != nil {
		return nil, err
	}

	tmp := ApplicationElement(*application)

	return []Element{&tmp}, nil
}
