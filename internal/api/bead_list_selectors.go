package api

import (
	"fmt"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gastownhall/gascity/internal/beadmeta"
	"github.com/gastownhall/gascity/internal/beads"
)

// beadListSelectors keeps HTTP-only selectors out of the Store contract.
// Stores narrow by equality/parent identity; residual checks run before API
// pagination because parent presence and metadata existence have no ListQuery
// spelling. Filtered reads bypass the Count/LIMIT fast path for exact totals.
type beadListSelectors struct {
	metadata map[string]string
	hasKeys  []string
	parent   string
}

func parseBeadListSelectors(input *BeadListInput) (beadListSelectors, error) {
	selectors := beadListSelectors{hasKeys: input.HasMetadata, parent: input.Parent}
	for _, pair := range input.Metadata {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || !beadmeta.ValidKey(key) {
			return selectors, huma.Error400BadRequest(fmt.Sprintf("invalid metadata filter %q: want key=value with a valid metadata key", pair))
		}
		if previous, exists := selectors.metadata[key]; exists && previous != value {
			return selectors, huma.Error400BadRequest(fmt.Sprintf("conflicting metadata filters for %q", key))
		}
		if selectors.metadata == nil {
			selectors.metadata = make(map[string]string, len(input.Metadata))
		}
		selectors.metadata[key] = value
	}
	for _, key := range selectors.hasKeys {
		if !beadmeta.ValidKey(key) {
			return selectors, huma.Error400BadRequest(fmt.Sprintf("invalid has_metadata key %q", key))
		}
	}
	return selectors, nil
}

func (s beadListSelectors) active() bool {
	return s.parent != "" || len(s.metadata) > 0 || len(s.hasKeys) > 0
}

func (s beadListSelectors) apply(query *beads.ListQuery) {
	query.Metadata = s.metadata
	if s.parent != "none" && s.parent != "any" {
		query.ParentID = s.parent
	}
}

func (s beadListSelectors) matches(b beads.Bead) bool {
	switch s.parent {
	case "":
	case "none":
		if b.ParentID != "" {
			return false
		}
	case "any":
		if b.ParentID == "" {
			return false
		}
	default:
		if b.ParentID != s.parent {
			return false
		}
	}
	for key, value := range s.metadata {
		if actual, exists := b.Metadata[key]; !exists || actual != value {
			return false
		}
	}
	for _, key := range s.hasKeys {
		if _, exists := b.Metadata[key]; !exists {
			return false
		}
	}
	return true
}
