// Package kubearchive provides a client for interacting with the KubeArchive API.
// KubeArchive is a system that archives Kubernetes resources for historical queries.
// See https://kubearchive.github.io/kubearchive/main/reference/api.html
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

// Client provides methods to query archived Konflux resources from KubeArchive.
type Client interface {
	// GetReleasesIterator returns an iterator for lazily fetching releases from the archive.
	// The iterator handles pagination automatically, fetching pages on demand.
	GetReleasesIterator(ctx context.Context, ns string) ReleaseIterator
	// GetSnapshot retrieves a specific snapshot by name from the archive.
	GetSnapshot(ctx context.Context, ns, name string) (applicationapi.Snapshot, error)
}

// ReleaseIterator provides lazy pagination over releases.
// Use Next() to advance through items and Value() to access the current release.
// Always check Err() after iteration completes to detect any errors.
type ReleaseIterator interface {
	// Next advances to the next release. Returns false when no more items or error occurred.
	Next() bool
	// Value returns the current release. Only valid after Next() returns true.
	Value() konfluxapi.Release
	// Err returns any error that occurred during iteration.
	Err() error
}

// KubeArchiveHTTPClient implements Client using HTTP requests to the KubeArchive API.
// Results are memoized to avoid redundant API calls.
type KubeArchiveHTTPClient struct {
	httpClient          *http.Client
	kubeArchiveHostname string
	scheme              string // defaults to "https", can be overridden for testing

	// Memoization caches for successful results only
	releasesCache  map[string]konfluxapi.ReleaseList
	snapshotsCache map[string]applicationapi.Snapshot
}

func kubeArchiveHostnameFromClusterHostname(clusterAPIHost string) string {
	clusterAPIURL, _ := url.Parse(clusterAPIHost)
	suffix := strings.TrimPrefix(clusterAPIURL.Hostname(), "api.")
	// https://konflux.pages.redhat.com/docs/users/faq/kubearchive.html
	return fmt.Sprintf("kubearchive-api-server-product-kubearchive.apps.%s", suffix)
}

// ClientFor creates a new KubeArchive client using the provided Kubernetes REST config.
// It constructs an authenticated HTTP client using the credentials from the config
// and derives the KubeArchive hostname from the cluster API host.
// The client memoizes successful results to avoid redundant API calls.
// See https://kubearchive.github.io/kubearchive/main/reference/api.html
func ClientFor(config *rest.Config) (Client, error) {
	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}
	return &KubeArchiveHTTPClient{
		httpClient:          httpClient,
		kubeArchiveHostname: kubeArchiveHostnameFromClusterHostname(config.Host),
		scheme:              "https",
		releasesCache:       make(map[string]konfluxapi.ReleaseList),
		snapshotsCache:      make(map[string]applicationapi.Snapshot),
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
