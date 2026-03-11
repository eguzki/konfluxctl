package kubearchive

import (
	"context"

	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
)

type releaseIterator struct {
	client        *KubeArchiveHTTPClient
	ctx           context.Context
	namespace     string
	continueToken string
	buffer        []konfluxapi.Release
	index         int
	done          bool
	err           error
}

func (r *releaseIterator) Next() bool {
	// If we have items in buffer, advance to next
	if r.index < len(r.buffer)-1 {
		r.index++
		return true
	}

	// If we're done (no continue token from last fetch), return false
	if r.done {
		return false
	}

	// Fetch next page
	list, err := r.client.getReleases(r.ctx, r.namespace, r.continueToken)
	if err != nil {
		r.err = err
		r.done = true
		return false
	}

	// Update buffer and token
	r.buffer = list.Items
	r.index = 0
	r.continueToken = list.Continue

	// If no continue token, we're done after this page
	if r.continueToken == "" {
		r.done = true
	}

	// Return true if we have items
	return len(r.buffer) > 0
}

func (r *releaseIterator) Value() konfluxapi.Release {
	return r.buffer[r.index]
}

func (r *releaseIterator) Err() error {
	return r.err
}
