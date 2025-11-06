package metadata

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eguzki/konfluxctl/internal/utils"
	"github.com/samber/lo"
)

type Path struct {
	releasePlanAdmission *string
	releasePlan          *string
	release              *string
	application          *string
	sourceRevision       *string
	sourceURL            *string
	snapshot             *string
	componentName        *string
}

func (p Path) String() string {
	return fmt.Sprintf(`ReleasePlanAddmision: %s
Application: %s,
ReleasePlan: %s,
Release: %s,
Snapshot: %s,
Component: %s,
Source URL: %s,
Source Revision: %s`,
		lo.FromPtrOr(p.releasePlanAdmission, "<nil>"),
		lo.FromPtrOr(p.application, "<nil>"),
		lo.FromPtrOr(p.releasePlan, "<nil>"),
		lo.FromPtrOr(p.release, "<nil>"),
		lo.FromPtrOr(p.snapshot, "<nil>"),
		lo.FromPtrOr(p.componentName, "<nil>"),
		lo.FromPtrOr(p.sourceURL, "<nil>"),
		lo.FromPtrOr(p.sourceRevision, "<nil>"),
	)
}

func (p Path) IsComplete() bool {
	return p.releasePlanAdmission != nil &&
		p.releasePlan != nil &&
		p.release != nil &&
		p.application != nil &&
		p.sourceRevision != nil &&
		p.sourceURL != nil &&
		p.snapshot != nil &&
		p.componentName != nil
}

func (p *Path) Clone() Path {
	return Path{
		releasePlanAdmission: p.releasePlanAdmission,
		releasePlan:          p.releasePlan,
		release:              p.release,
		application:          p.application,
		sourceRevision:       p.sourceRevision,
		sourceURL:            p.sourceURL,
		snapshot:             p.snapshot,
		componentName:        p.componentName,
	}
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
			fmt.Println("new path")
			completePaths = append(completePaths, current.Path)
		}
	}
	return completePaths, nil
}
