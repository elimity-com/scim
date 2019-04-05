package scim

// ListResponse is a standard set of query parameters that can be used to filter, sort, and
// paginate to return zero or more resources in a query response.
// INFO: RFC7644 - 3.4.2. Query Resources
type ListResponse struct {
	// Schemas is used to identify responses.
	// URI: "urn:ietf:params:scim:api:messages:2.0:ListResponse"
	Schemas []string

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
	Resources interface{}
}
