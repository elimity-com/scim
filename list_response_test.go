package scim

import (
	"testing"
	"time"

	"github.com/elimity-com/scim/schema"
)

func TestRawResourcesDoesNotMutateHandlerAttributes(t *testing.T) {
	t.Parallel()

	t.Run("outer map not mutated", func(t *testing.T) {
		t.Parallel()
		attrs := ResourceAttributes{"displayName": "Test User"}
		original := make(ResourceAttributes, len(attrs))
		for k, v := range attrs {
			original[k] = v
		}
		p := Page{Resources: []Resource{{ID: "42", Attributes: attrs}}}

		_ = p.rawResources()

		for k := range attrs {
			if _, ok := original[k]; !ok {
				t.Errorf("rawResources() injected %q into handler-owned attributes map", k)
			}
		}
	})

	t.Run("nested meta map not mutated", func(t *testing.T) {
		t.Parallel()
		// Handler provides a meta map (e.g. with resourceType pre-filled).
		existingMeta := map[string]interface{}{"resourceType": "User"}
		originalMeta := make(map[string]interface{}, len(existingMeta))
		for k, v := range existingMeta {
			originalMeta[k] = v
		}
		attrs := ResourceAttributes{
			"displayName":              "Test User",
			schema.CommonAttributeMeta: existingMeta,
		}
		now := time.Now()
		p := Page{
			Resources: []Resource{{
				ID:         "42",
				Attributes: attrs,
				Meta:       Meta{Created: &now, LastModified: &now, Version: "W/\"1\""},
			}},
		}

		_ = p.rawResources()

		for k := range existingMeta {
			if _, ok := originalMeta[k]; !ok {
				t.Errorf("rawResources() injected %q into handler-owned meta map", k)
			}
		}
	})
}
