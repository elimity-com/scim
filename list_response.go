package scim

import (
	"encoding/json"
)

// NewListResponse creates a new list response with provided paramters.
func NewListResponse(resources []Resource, startIndex, totalResults, itemsPerPage int) ListResponse {
	return ListResponse{
		TotalResults: totalResults,
		ItemsPerPage: itemsPerPage,
		StartIndex:   startIndex,
		Resources:    resources,
	}
}

func (l *ListResponse) toInternalListResponse(resourceType ResourceType) listResponse {
	convertedResources := make([]interface{}, len(l.Resources))

	for i, res := range l.Resources {
		convertedResources[i] = res.response(resourceType)
	}

	return newListResponse(convertedResources)
}

func newListResponse(resources []interface{}) listResponse {
	return listResponse{
		TotalResults: len(resources),
		ItemsPerPage: len(resources),
		StartIndex:   1,
		Resources:    resources,
	}
}

type (
	// ListResponse identifies a paginated query response.
	ListResponse struct {
		// TotalResults is the total number of results returned by the list or query operation.
		// The value may be larger than the number of resources returned, such as when returning
		// a single page of results where multiple pages are available.
		// REQUIRED
		TotalResults int

		// ItemsPerPage is the number of resources returned in a list response page.
		// REQUIRED when partial results are returned due to pagination.
		ItemsPerPage int

		// StartIndex is a 1-based index of the first result in the current set of the list results.
		// REQUIRED when partial results are returned due to pagination.
		StartIndex int

		// Resources is a multi-valued list of complex objects containing the requested resources.
		// This may be a subset of the full set of resources if pagination is requested.
		// REQUIRED if TotalResults is non-zero.
		Resources []Resource
	}

	// listResponse identifies a query response.
	listResponse struct {
		// TotalResults is the total number of results returned by the list or query operation.
		// The value may be larger than the number of resources returned, such as when returning
		// a single page of results where multiple pages are available.
		// REQUIRED
		TotalResults int

		// ItemsPerPage is the number of resources returned in a list response page.
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
)

func (l listResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"schemas":      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		"totalResults": l.TotalResults,
		"itemsPerPage": l.ItemsPerPage,
		"startIndex":   l.StartIndex,
		"Resources":    l.Resources,
	})
}
