package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/samber/lo"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eguzki/konfluxctl/internal/utils"
)

type Path struct {
	ReleasePlanAdmission *string  `json:"releasePlanAdmission"`
	ReleasePlan          *string  `json:"releasePlan"`
	Release              *string  `json:"release"`
	Application          *string  `json:"application"`
	SourceRevision       *string  `json:"sourceRevision"`
	SourceURL            *string  `json:"sourceURL"`
	Snapshot             *string  `json:"snapshot"`
	ComponentName        *string  `json:"componentName"`
	ImageTags            []string `json:"imageTags"`
}

func (p Path) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil

}

func (p Path) ToYAML() (string, error) {
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	yamlBytes, err := yaml.JSONToYAML(jsonBytes) // use `omitempty`'s from the json Marshal
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

func (p Path) String() string {
	return fmt.Sprintf(`ReleasePlanAddmision: %s
Application: %s,
ReleasePlan: %s,
Release: %s,
Snapshot: %s,
Component: %s,
Source URL: %s,
Source Revision: %s,
Image Tags: %s`,
		lo.FromPtrOr(p.ReleasePlanAdmission, "<nil>"),
		lo.FromPtrOr(p.Application, "<nil>"),
		lo.FromPtrOr(p.ReleasePlan, "<nil>"),
		lo.FromPtrOr(p.Release, "<nil>"),
		lo.FromPtrOr(p.Snapshot, "<nil>"),
		lo.FromPtrOr(p.ComponentName, "<nil>"),
		lo.FromPtrOr(p.SourceURL, "<nil>"),
		lo.FromPtrOr(p.SourceRevision, "<nil>"),
		strings.Join(p.ImageTags, ","),
	)
}

func (p Path) IsComplete() bool {
	return p.ReleasePlanAdmission != nil &&
		p.ReleasePlan != nil &&
		p.Release != nil &&
		p.Application != nil &&
		p.SourceRevision != nil &&
		p.SourceURL != nil &&
		p.Snapshot != nil &&
		p.ComponentName != nil &&
		len(p.ImageTags) != 0
}

func (p *Path) Clone() Path {
	// shallow copy
	return *p
}

type Node struct {
	Element Element
	Path    Path
}

type Element interface {
	Visit(path *Path)
	Children(ctx context.Context, k8sClient client.Client, imageURL *utils.ImageURL) ([]Element, error)
	String() string
}

func DepthFirstSearch(ctx context.Context, k8sClient client.Client, imageURL *utils.ImageURL, elements []Element) ([]Path, error) {
	completePaths := []Path{}
	queue := []Node{}

	for _, element := range elements {
		queue = append(queue, Node{Element: element, Path: Path{}})
	}

	for len(queue) > 0 {
		current := queue[0]
		//fmt.Printf("node: %s\n", current.Element.String())
		queue = queue[1:]
		current.Element.Visit(&current.Path)
		children, err := current.Element.Children(ctx, k8sClient, imageURL)
		if err != nil {
			return nil, err
		}

		for _, child := range children {
			// prepend
			queue = append([]Node{Node{Element: child, Path: current.Path.Clone()}}, queue...)
		}

		if current.Path.IsComplete() {
			completePaths = append(completePaths, current.Path)
		}
	}
	return completePaths, nil
}
