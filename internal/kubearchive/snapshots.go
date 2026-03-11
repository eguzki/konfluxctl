package kubearchive

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	applicationapi "github.com/konflux-ci/application-api/api/v1alpha1"
)

// GetSnapshot retrieves a specific snapshot by name from the specified namespace in the archive.
// Returns an error if the snapshot is not found or if the request fails.
func (k *KubeArchiveHTTPClient) GetSnapshot(ctx context.Context, ns, name string) (applicationapi.Snapshot, error) {
	u := &url.URL{
		Scheme: k.scheme,
		Host:   k.kubeArchiveHostname,
		Path:   fmt.Sprintf("/apis/appstudio.redhat.com/v1alpha1/namespaces/%s/snapshots/%s", ns, name),
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return applicationapi.Snapshot{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := k.httpClient.Do(req)
	slog.Debug("kubearchive client", "GetSnapshot", u.String(), "error", err)
	if err != nil {
		return applicationapi.Snapshot{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return applicationapi.Snapshot{}, handleJsonErrResp(resp)
	}

	var decodeInto applicationapi.Snapshot

	if err := json.NewDecoder(resp.Body).Decode(&decodeInto); err != nil {
		return applicationapi.Snapshot{}, Err{
			code: resp.StatusCode,
			err:  fmt.Sprintf("decoding error - %s", err.Error()),
		}
	}

	return decodeInto, nil
}
