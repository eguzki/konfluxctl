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

// getReleases fetches a page of releases with optional continue token for pagination
func (k *KubeArchiveHTTPClient) getReleases(ctx context.Context, ns, continueToken string) (konfluxapi.ReleaseList, error) {
	u := &url.URL{
		Scheme: "https",
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
