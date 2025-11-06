package metadata

import (
	"context"
	"fmt"

	applicationapi "github.com/konflux-ci/application-api/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eguzki/konfluxctl/internal/utils"
)

type ApplicationElement applicationapi.Application

func (a *ApplicationElement) String() string {
	return fmt.Sprintf("%s: %s", "Application", a.Name)
}

func (a *ApplicationElement) Visit(path *Path) {
	path.Application = &a.Name
}

func (s *ApplicationElement) Children(ctx context.Context, k8sClient client.Client, imageURL *utils.ImageURL) ([]Element, error) {
	// leaf node
	return nil, nil
}
