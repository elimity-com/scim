package scim

import (
	"net/http"
	"time"

	"github.com/elimity-com/scim/filter"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

// ListRequestParams request parameters sent to the API via a "GetAll" route.
type ListRequestParams struct {
	// Count specifies the desired maximum number of query results per page. A negative value SHALL be interpreted as "0".
	// A value of "0" indicates that no resource results are to be returned except for "totalResults".
	Count int

	// Filter is the raw filter expression string. For resource-type-specific queries, the filter
	// is also parsed and available via FilterValidator. For root queries (RootQueryHandler),
	// only this raw string is provided since filter validation requires a known schema.
	Filter string

	// FilterValidator represents the parsed and tokenized filter query parameter.
	// It is an optional parameter and thus will be nil when the parameter is not present.
	// For root queries (RootQueryHandler), this is always nil since filter validation requires
	// a known schema.
	FilterValidator *filter.Validator

	// StartIndex The 1-based index of the first query result. A value less than 1 SHALL be interpreted as 1.
	StartIndex int
}

// Meta represents the metadata of a resource.
type Meta struct {
	// Created is the time that the resource was added to the service provider.
	Created *time.Time
	// LastModified is the most recent time that the details of this resource were updated at the service provider.
	LastModified *time.Time
	// Version is the version / entity-tag of the resource
	Version string
}

// Resource represents an entity returned by a callback method.
type Resource struct {
	// ID is the unique identifier created by the callback method "Create".
	ID string
	// ExternalID is an identifier for the resource as defined by the provisioning client.
	ExternalID optional.String
	// Attributes is a list of attributes defining the resource.
	Attributes ResourceAttributes
	// Meta contains dates and the version of the resource.
	Meta Meta
}

func (r Resource) response(resourceType ResourceType, location string) ResourceAttributes {
	response := r.Attributes
	if response == nil {
		response = ResourceAttributes{}
	}

	response[schema.CommonAttributeID] = r.ID
	if r.ExternalID.Present() {
		response[schema.CommonAttributeExternalID] = r.ExternalID.Value()
	}
	schemas := []string{resourceType.Schema.ID}
	for _, s := range resourceType.SchemaExtensions {
		schemas = append(schemas, s.Schema.ID)
	}

	response["schemas"] = schemas

	m := meta{
		ResourceType: resourceType.Name,
		Location:     location,
	}

	if r.Meta.Created != nil {
		m.Created = r.Meta.Created.Format(time.RFC3339)
	}

	if r.Meta.LastModified != nil {
		m.LastModified = r.Meta.LastModified.Format(time.RFC3339)
	}

	if len(r.Meta.Version) != 0 {
		m.Version = r.Meta.Version
	}

	response[schema.CommonAttributeMeta] = m

	return response
}

// ResourceAttributes represents a list of attributes given to the callback method to create or replace
// a resource based on the given attributes.
type ResourceAttributes map[string]interface{}

// ResourceHandler represents a set of callback method that connect the SCIM server with a provider of a certain resource.
type ResourceHandler interface {
	// Create stores given attributes. Returns a resource with the attributes that are stored and a (new) unique identifier.
	Create(r *http.Request, attributes ResourceAttributes) (Resource, error)
	// Get returns the resource corresponding with the given identifier.
	Get(r *http.Request, id string) (Resource, error)
	// GetAll returns a paginated list of resources.
	// An empty list of resources will be represented as `null` in the JSON response if `nil` is assigned to the
	// Page.Resources. Otherwise, is an empty slice is assigned, an empty list will be represented as `[]`.
	GetAll(r *http.Request, params ListRequestParams) (Page, error)
	// Replace replaces ALL existing attributes of the resource with given identifier. Given attributes that are empty
	// are to be deleted. Returns a resource with the attributes that are stored.
	Replace(r *http.Request, id string, attributes ResourceAttributes) (Resource, error)
	// Delete removes the resource with corresponding ID.
	Delete(r *http.Request, id string) error
	// Patch update one or more attributes of a SCIM resource using a sequence of
	// operations to "add", "remove", or "replace" values.
	// If you return no Resource.Attributes, a 204 No Content status code will be returned.
	// This case is only valid in the following scenarios:
	// 1. the Add/Replace operation should return No Content only when the value already exists AND is the same.
	// 2. the Remove operation should return No Content when the value to be remove is already absent.
	// More information in Section 3.5.2 of RFC 7644: https://tools.ietf.org/html/rfc7644#section-3.5.2
	Patch(r *http.Request, id string, operations []PatchOperation) (Resource, error)
}

// ResourceTypeFilter associates a resource type with a validated filter.
type ResourceTypeFilter struct {
	// ResourceType is the resource type whose schema the filter validated against.
	ResourceType ResourceType
	// Validator is the filter validator for this resource type's schema.
	Validator filter.Validator
}

// ValidateFilterForResourceTypes validates a raw filter expression against each of the given resource types' schemas.
// It returns a ResourceTypeFilter for each resource type whose schema the filter is valid for.
// This is useful for RootQueryHandler implementations to determine which resource types a filter applies to.
// An empty result means the filter is not valid for any of the given resource types.
// A parse error in the filter expression results in an empty result.
func ValidateFilterForResourceTypes(rawFilter string, resourceTypes []ResourceType) []ResourceTypeFilter {
	var results []ResourceTypeFilter
	for _, rt := range resourceTypes {
		s := rt.Schema
		attrs := make([]schema.CoreAttribute, len(s.Attributes), len(s.Attributes)+len(schema.CommonAttributes()))
		copy(attrs, s.Attributes)
		s.Attributes = append(attrs, schema.CommonAttributes()...)
		v, err := filter.NewValidator(rawFilter, s, rt.getSchemaExtensions()...)
		if err != nil {
			return nil
		}
		if err := v.Validate(); err != nil {
			continue
		}
		results = append(results, ResourceTypeFilter{
			ResourceType: rt,
			Validator:    v,
		})
	}
	return results
}

// RootQueryHandler represents an optional callback that handles queries against the server root endpoint (GET /).
// Per RFC 7644 Section 3.4.2.1, a query against the server root indicates that all resources within the server
// shall be included, subject to filtering.
//
// The server does not validate or parse the filter for root queries because there is no single target schema.
// ListRequestParams.FilterValidator will always be nil for root queries. The raw filter string can be
// obtained from the request via r.URL.Query().Get("filter"). The handler is responsible for interpreting
// the filter (e.g. meta.resourceType eq "User") as appropriate for its backing store.
type RootQueryHandler interface {
	// GetAll returns a paginated list of resources across all resource types.
	GetAll(r *http.Request, params ListRequestParams) (Page, error)
}
