package scim

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/elimity-com/scim/schema"

	scim "github.com/di-wu/scim-filter-parser"
)

// ListRequestParams request parameters sent to the API via a "GetAll" route.
type ListRequestParams struct {
	// Count specifies the desired maximum number of query results per page. A negative value SHALL be interpreted as "0".
	// A value of "0" indicates that no resource results are to be returned except for "totalResults".
	Count int

	// Filter represents the parsed and tokenized filter query parameter.
	// It is an optional parameter and thus will be nil when the parameter is not present.
	Filter scim.Expression

	// StartIndex The 1-based index of the first query result. A value less than 1 SHALL be interpreted as 1.
	StartIndex int
}

// ResourceAttributes represents a list of attributes given to the callback method to create or replace
// a resource based on the given attributes.
type ResourceAttributes map[string]interface{}

// Meta represents the metadata of a resource
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
	ExternalID string
	// Attributes is a list of attributes defining the resource.
	Attributes ResourceAttributes
	// Meta contains dates and the version of the resource.
	Meta Meta
}

func (r Resource) response(resourceType ResourceType) ResourceAttributes {
	response := r.Attributes
	response[schema.CommonAttributeID] = r.ID
	if r.ExternalID != "" {
		response[schema.CommonAttributeExternalID] = r.ExternalID
	}
	schemas := []string{resourceType.Schema.ID}
	for _, s := range resourceType.SchemaExtensions {
		schemas = append(schemas, s.Schema.ID)
	}

	response["schemas"] = schemas

	m := meta{
		ResourceType: resourceType.Name,
		Location:     fmt.Sprintf("%s/%s", resourceType.Endpoint[1:], url.PathEscape(r.ID)),
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

// ResourceHandler represents a set of callback method that connect the SCIM server with a provider of a certain resource.
type ResourceHandler interface {
	// Create stores given attributes. Returns a resource with the attributes that are stored and a (new) unique identifier.
	Create(r *http.Request, attributes ResourceAttributes) (Resource, error)
	// Get returns the resource corresponding with the given identifier.
	Get(r *http.Request, id string) (Resource, error)
	// GetAll returns a paginated list of resources.
	GetAll(r *http.Request, params ListRequestParams) (Page, error)
	// Replace replaces ALL existing attributes of the resource with given identifier. Given attributes that are empty
	// are to be deleted. Returns a resource with the attributes that are stored.
	Replace(r *http.Request, id string, attributes ResourceAttributes) (Resource, error)
	// Delete removes the resource with corresponding ID.
	Delete(r *http.Request, id string) error
	// Patch update one or more attributes of a SCIM resource using a sequence of
	// operations to "add", "remove", or "replace" values.
	Patch(r *http.Request, id string, request PatchRequest) (Resource, error)
}
