package kubearchive

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
)

// GetReleasesIterator returns an iterator for lazily fetching releases from the specified namespace.
// The iterator fetches data in pages on demand, using the continue token for pagination.
// It starts with an empty buffer and fetches the first page on the first call to Next().
func (k *KubeArchiveHTTPClient) GetReleasesIterator(ctx context.Context, ns string) ReleaseIterator {
	return &releaseIterator{
		client:    k,
		ctx:       ctx,
		namespace: ns,
		buffer:    []konfluxapi.Release{},
		index:     -1, // Start at -1 so first Next() moves to 0
		done:      false,
	}
}

// getReleases fetches a page of releases from the archive with optional pagination.
// If continueToken is provided, it fetches the next page starting from that token.
// Returns a ReleaseList containing items and a continue token for the next page (if any).
func (k *KubeArchiveHTTPClient) getReleases(ctx context.Context, ns, continueToken string) (konfluxapi.ReleaseList, error) {
	u := &url.URL{
		Scheme: k.scheme,
		Host:   k.kubeArchiveHostname,
		Path:   fmt.Sprintf("/apis/appstudio.redhat.com/v1alpha1/namespaces/%s/releases", ns),
	}

	// Add continue token if provided
	if continueToken != "" {
		query := u.Query()
		query.Set("continue", continueToken)
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return konfluxapi.ReleaseList{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := k.httpClient.Do(req)
	slog.Debug("kubearchive client", "getReleases", u.String(), "error", err)
	if err != nil {
		return konfluxapi.ReleaseList{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return konfluxapi.ReleaseList{}, handleJsonErrResp(resp)
	}

	var decodeInto konfluxapi.ReleaseList

	if err := json.NewDecoder(resp.Body).Decode(&decodeInto); err != nil {
		return konfluxapi.ReleaseList{}, Err{
			code: resp.StatusCode,
			err:  fmt.Sprintf("decoding error - %s", err.Error()),
		}
	}

	return decodeInto, nil
}
