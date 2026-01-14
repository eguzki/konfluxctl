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

type Repository struct {
	Url  string   `json:"url"`
	Tags []string `json:"tags"`
}

type ReleasePlanAdmissionDataComponent struct {
	Name         string       `json:"name"`
	Repositories []Repository `json:"repositories"`
}

type ReleasePlanAdmissionDataMapping struct {
	Components []ReleasePlanAdmissionDataComponent `json:"components"`
}

type ReleasePlanAdmissionData struct {
	Mappping ReleasePlanAdmissionDataMapping `json:"mapping"`
}

type ReleasePlanAdmissionElement struct {
	rawRPA konfluxapi.ReleasePlanAdmission
	tags   []string
}

func (r *ReleasePlanAdmissionElement) String() string {
	return fmt.Sprintf("%s: %s", "ReleasePlanAdmission", r.rawRPA.Name)
}

func (r *ReleasePlanAdmissionElement) Visit(path *Path) {
	path.ReleasePlanAdmission = &r.rawRPA.Name
	path.ImageTags = r.tags
}

func (r *ReleasePlanAdmissionElement) Children(ctx context.Context, k8sClient client.Client, imageURL *utils.ImageURL) ([]Element, error) {
	children := []*konfluxapi.ReleasePlan{}
	for _, matchedReleasePlan := range r.rawRPA.Status.ReleasePlans {
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

func ReleasePlanAdmissionList(ctx context.Context, k8sClient client.Client, imageName string) ([]Element, error) {
	rpaList := &konfluxapi.ReleasePlanAdmissionList{}
	err := k8sClient.List(ctx, rpaList, client.InNamespace("rhtap-releng-tenant"))
	if err != nil {
		return nil, err
	}

	return lo.FilterMap(rpaList.Items, func(rpa konfluxapi.ReleasePlanAdmission, index int) (Element, bool) {
		var data ReleasePlanAdmissionData
		if err := json.Unmarshal(rpa.Spec.Data.Raw, &data); err != nil {
			return nil, false
		}

		repositories := lo.FlatMap(data.Mappping.Components, func(comp ReleasePlanAdmissionDataComponent, _ int) []Repository {
			return comp.Repositories
		})

		repository, ok := lo.Find(repositories, func(repo Repository) bool {
			return repo.Url == imageName
		})

		if !ok {
			return nil, false
		}

		return &ReleasePlanAdmissionElement{
			rawRPA: rpa,
			tags:   repository.Tags,
		}, true
	}), nil
}
