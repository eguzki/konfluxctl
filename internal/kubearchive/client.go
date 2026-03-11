package kubearchive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	applicationapi "github.com/konflux-ci/application-api/api/v1alpha1"
	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
	"k8s.io/client-go/rest"
)

type Client interface {
	GetReleases(ctx context.Context, ns string) (konfluxapi.ReleaseList, error)
	GetSnapshot(ctx context.Context, ns, name string) (applicationapi.Snapshot, error)
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

func (k *KubeArchiveHTTPClient) GetReleases(ctx context.Context, ns string) (konfluxapi.ReleaseList, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   k.kubeArchiveHostname,
		Path:   fmt.Sprintf("/apis/appstudio.redhat.com/v1alpha1/namespaces/%s/releases", ns),
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return konfluxapi.ReleaseList{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := k.httpClient.Do(req)
	slog.Debug("kubearchive client", "GetReleases", u.String(), "error", err)
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

func (k *KubeArchiveHTTPClient) GetSnapshot(ctx context.Context, ns, name string) (applicationapi.Snapshot, error) {
	u := &url.URL{
		Scheme: "https",
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
