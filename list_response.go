package scim

import (
	"encoding/json"
	"time"

	"github.com/elimity-com/scim/schema"
)

// Page represents a paginated resource query response.
type Page struct {
	// TotalResults is the total number of results returned by the list or query operation.
	TotalResults int
	// Resources is a multi-valued list of complex objects containing the requested resources.
	Resources []Resource
}

// rawResources returns resources as raw interface values for root queries (GET /).
//
// Unlike the resource-type-specific resources() method, rawResources does NOT have access to a ResourceType and
// therefore cannot inject:
//   - "schemas": requires knowing the resource type's schema URIs.
//   - "meta.resourceType": requires the resource type name.
//   - "meta.location": requires the resource type endpoint.
//
// The caller (RootQueryHandler) is responsible for including these fields in the Resource's Attributes map.
//
// rawResources DOES inject the following fields from the Resource struct into each resource's attributes:
//   - "id": from Resource.ID (always required per RFC 7643 Section 3.1).
//   - "externalId": from Resource.ExternalID, when present.
//   - "meta.created": from Resource.Meta.Created, when non-nil.
//   - "meta.lastModified": from Resource.Meta.LastModified, when non-nil.
//   - "meta.version": from Resource.Meta.Version, when non-empty.
//
// These meta fields are merged into any existing "meta" map in Attributes. If the caller already provides a "meta"
// map (e.g. with "resourceType"), the injected fields are added alongside it without overwriting existing keys.
func (p Page) rawResources() []interface{} {
	if len(p.Resources) == 0 {
		if p.Resources != nil {
			return []interface{}{}
		}
		return nil
	}

	var resources []interface{}
	for _, v := range p.Resources {
		attrs := v.Attributes
		if attrs == nil {
			attrs = ResourceAttributes{}
		}

		attrs[schema.CommonAttributeID] = v.ID
		if v.ExternalID.Present() {
			attrs[schema.CommonAttributeExternalID] = v.ExternalID.Value()
		}

		// Merge Meta fields into the existing "meta" map if present, or create a new one.
		var metaMap map[string]interface{}
		if existing, ok := attrs[schema.CommonAttributeMeta]; ok {
			if m, ok := existing.(map[string]interface{}); ok {
				metaMap = m
			}
		}
		hasMeta := false
		if v.Meta.Created != nil {
			if metaMap == nil {
				metaMap = map[string]interface{}{}
			}
			metaMap["created"] = v.Meta.Created.Format(time.RFC3339)
			hasMeta = true
		}
		if v.Meta.LastModified != nil {
			if metaMap == nil {
				metaMap = map[string]interface{}{}
			}
			metaMap["lastModified"] = v.Meta.LastModified.Format(time.RFC3339)
			hasMeta = true
		}
		if len(v.Meta.Version) != 0 {
			if metaMap == nil {
				metaMap = map[string]interface{}{}
			}
			metaMap["version"] = v.Meta.Version
			hasMeta = true
		}
		if hasMeta {
			attrs[schema.CommonAttributeMeta] = metaMap
		}

		resources = append(resources, attrs)
	}
	return resources
}

func (p Page) resources(resourceType ResourceType, baseURL string) []interface{} {
	// If the page.Resources is nil, then it will also be represented as a `null` in the response.
	// Otherwise is it is an empty slice then it will result in an empty array `[]`.
	if len(p.Resources) == 0 {
		if p.Resources != nil {
			return []interface{}{}
		}
		return nil
	}

	var resources []interface{}
	for _, v := range p.Resources {
		location := resourceLocation(resourceType, v.ID, baseURL)
		resources = append(
			resources,
			v.response(resourceType, location),
		)
	}
	return resources
}

// listResponse identifies a query response.
type listResponse struct {
	// TotalResults is the total number of results returned by the list or query operation.
	// The value may be larger than the number of resources returned, such as when returning
	// a single page of results where multiple pages are available.
	// REQUIRED
	TotalResults int

	// MaxResults is the number of resources returned in a list response page.
	// REQUIRED when partial results are returned due to pagination.
	ItemsPerPage int

	// StartIndex is a 1-based index of the first result in the current set of the list results.
	// REQUIRED when partial results are returned due to pagination.
	StartIndex int

	// Resources is a multi-valued list of complex objects containing the requested resources.
	// This may be a subset of the full set of resources if pagination is requested.
	// REQUIRED if TotalResults is non-zero.
	Resources []interface{}
}

func (l listResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"schemas":      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		"totalResults": l.TotalResults,
		"itemsPerPage": l.ItemsPerPage,
		"startIndex":   l.StartIndex,
		"Resources":    l.Resources,
	})
}
