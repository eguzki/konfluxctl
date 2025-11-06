package metadata

import (
	"context"
	"fmt"

	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eguzki/konfluxctl/internal/utils"
)

type ReleasePlanElement konfluxapi.ReleasePlan

func (r *ReleasePlanElement) String() string {
	return fmt.Sprintf("%s: %s", "ReleasePlan", r.Name)
}

func (r *ReleasePlanElement) Visit(path *Path) {
	path.ReleasePlan = &r.Name
}

func (r *ReleasePlanElement) Children(ctx context.Context, k8sClient client.Client, imageURL *utils.ImageURL) ([]Element, error) {
	releaseList := &konfluxapi.ReleaseList{}
	err := k8sClient.List(ctx, releaseList, client.InNamespace(r.Namespace))
	if err != nil {
		return nil, err
	}

	planReleaseList := lo.Filter(releaseList.Items, func(release konfluxapi.Release, _ int) bool {
		return release.Spec.ReleasePlan == r.Name &&
			meta.IsStatusConditionTrue(release.Status.Conditions, "Released")
	})

	return lo.Map(planReleaseList, func(e konfluxapi.Release, _ int) Element {
		tmp := ReleaseElement(e)
		return &tmp
	}), nil
}
