package kubearchive

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	applicationapi "github.com/konflux-ci/application-api/api/v1alpha1"
	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
	"k8s.io/client-go/rest"
)

type Client interface {
	GetReleasesIterator(ctx context.Context, ns string) ReleaseIterator
	GetSnapshot(ctx context.Context, ns, name string) (applicationapi.Snapshot, error)
}

// ReleaseIterator provides lazy pagination over releases
type ReleaseIterator interface {
	// Next advances to the next release. Returns false when no more items or error occurred.
	Next() bool
	// Value returns the current release. Only valid after Next() returns true.
	Value() konfluxapi.Release
	// Err returns any error that occurred during iteration.
	Err() error
}

type KubeArchiveHTTPClient struct {
	httpClient          *http.Client
	kubeArchiveHostname string
}

func kubeArchiveHostnameFromClusterHostname(clusterAPIHost string) string {
	clusterAPIURL, _ := url.Parse(clusterAPIHost)
	suffix := strings.TrimPrefix(clusterAPIURL.Hostname(), "api.")
	// https://konflux.pages.redhat.com/docs/users/faq/kubearchive.html
	return fmt.Sprintf("kubearchive-api-server-product-kubearchive.apps.%s", suffix)
}

// https://kubearchive.github.io/kubearchive/main/reference/api.html
func ClientFor(config *rest.Config) (Client, error) {
	// Create authenticated HTTP client with K8s credentials
	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}
	return &KubeArchiveHTTPClient{
		httpClient:          httpClient,
		kubeArchiveHostname: kubeArchiveHostnameFromClusterHostname(config.Host),
	}, nil
}

func handleJsonErrResp(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	return Err{
		code: resp.StatusCode,
		err:  string(body),
	}
}
