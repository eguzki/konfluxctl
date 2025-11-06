package utils

import (
	// required to parse OCI images with 256 digest
	_ "crypto/sha256"
	"fmt"

	"github.com/distribution/reference"
)

type ImageURL struct {
	hostname     string
	familiarName string
	repository   string
	digest       string
}

func (i ImageURL) Hostname() string {
	return i.hostname
}

func (i ImageURL) Repository() string {
	return i.repository
}

func (i ImageURL) FamiliarName() string {
	return i.familiarName
}

func (i ImageURL) Digest() string {
	return i.digest
}

func ParseImageURL(imageURL string) (*ImageURL, error) {
	// 1. Parse the reference string
	ref, err := reference.ParseAnyReference(imageURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing image reference: %w", err)
	}

	// 2. Extract Hostname and Path (Repository)
	named, ok := ref.(reference.Named)
	if !ok {
		return nil, fmt.Errorf("image reference is not a named reference: %s", ref.String())
	}

	// 3. Extract Digest
	canonical, ok := ref.(reference.Canonical)
	if !ok {
		return nil, fmt.Errorf("reference does not contain a digest: %s", ref.String())
	}

	return &ImageURL{
		hostname:     reference.Domain(named),
		familiarName: reference.FamiliarName(named),
		repository:   reference.Path(named),
		digest:       canonical.Digest().String(),
	}, nil
}
