package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eguzki/konfluxctl/internal/utils"
)

type ReleasePlanAdmissionElement konfluxapi.ReleasePlanAdmission

func (r *ReleasePlanAdmissionElement) String() string {
	return fmt.Sprintf("%s: %s", "ReleasePlanAdmission", r.Name)
}

func (r *ReleasePlanAdmissionElement) Visit(path *Path) {
	path.ReleasePlanAdmission = &r.Name
}

func (r *ReleasePlanAdmissionElement) Children(ctx context.Context, k8sClient client.Client, imageURL *utils.ImageURL) ([]Element, error) {
	children := []*konfluxapi.ReleasePlan{}
	for _, matchedReleasePlan := range r.Status.ReleasePlans {
		namespacedName := strings.Split(matchedReleasePlan.Name, "/")

		releasePlan := &konfluxapi.ReleasePlan{}
		err := k8sClient.Get(ctx, client.ObjectKey{
			Namespace: namespacedName[0],
			Name:      namespacedName[1],
		}, releasePlan)
		if err != nil {
			return nil, err
		}
		children = append(children, releasePlan)
	}

	// Filter out those in bad condition
	validReleasePlans := lo.Filter(children, func(c *konfluxapi.ReleasePlan, _ int) bool {
		return meta.IsStatusConditionTrue(c.Status.Conditions, string(konfluxapi.MatchedConditionType))
	})

	return lo.Map(validReleasePlans, func(e *konfluxapi.ReleasePlan, _ int) Element {
		tmp := ReleasePlanElement(*e)
		return &tmp
	}), nil
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

func ReleasePlanAdmissionList(ctx context.Context, k8sClient client.Client, imageName string) ([]Element, error) {
	rpaList := &konfluxapi.ReleasePlanAdmissionList{}
	err := k8sClient.List(ctx, rpaList, client.InNamespace("rhtap-releng-tenant"))
	if err != nil {
		return nil, err
	}

	rpaForImageList := lo.Filter(rpaList.Items, func(rpa konfluxapi.ReleasePlanAdmission, index int) bool {
		var data ReleasePlanAdmissionData
		if err := json.Unmarshal(rpa.Spec.Data.Raw, &data); err != nil {
			return false
		}

		return lo.ContainsBy(data.Mappping.Components, func(comp ReleasePlanAdmissionDataComponent) bool {
			return comp.Repository == imageName
		})
	})

	return lo.Map(rpaForImageList, func(e konfluxapi.ReleasePlanAdmission, _ int) Element {
		tmp := ReleasePlanAdmissionElement(e)
		return &tmp
	}), nil
}
