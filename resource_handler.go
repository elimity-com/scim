package scim

import (
	"fmt"
	"net/url"

	scim "github.com/di-wu/scim-filter-parser"
	"github.com/elimity-com/scim/errors"
)

type (
	// ListRequestParams request parameters sent to the API via the "GetAll" route.
	ListRequestParams struct {
		// Count specifies the desired maximum number of query results per page. A negative value
		// SHALL be interpreted as "0". A value of "0" indicates that no resource
		// results are to be returned except for "totalResults".
		Count int

		// Filter represents the parsed and tokenized filter query parameter. It is an optional param
		// and thus will be nil when the param is not present.
		// https://github.com/di-wu/scim-filter-parser
		Filter scim.Expression

		// StartIndex The 1-based index of the first query result.
		// A value less than 1 SHALL be interpreted as 1.
		StartIndex int
	}

	// ResourceAttributes represents a list of attributes given to the callback method to create or replace
	// a resource based on the given attributes.
	ResourceAttributes map[string]interface{}

	// Resource represents an entity returned by a callback method.
	Resource struct {
		// ID is the unique identifier created by the callback method "Create".
		ID string
		// Attributes is a list of attributes defining the resource.
		Attributes ResourceAttributes
	}
)

func (r Resource) response(resourceType ResourceType) ResourceAttributes {
	response := r.Attributes
	response["id"] = r.ID
	schemas := []string{resourceType.Schema.ID}
	for _, schema := range resourceType.SchemaExtensions {
		schemas = append(schemas, schema.Schema.ID)
	}
	response["schemas"] = schemas
	response["meta"] = meta{
		ResourceType: resourceType.Name,
		Location:     fmt.Sprintf("%s/%s", resourceType.Endpoint[1:], url.PathEscape(r.ID)),
	}

	return response
}

// ResourceHandler represents a set of callback method that connect the SCIM server with a provider of a certain resource.
type ResourceHandler interface {
	// Create stores given attributes. Returns a resource with the attributes that are stored and a (new) unique identifier.
	Create(attributes ResourceAttributes) (Resource, errors.PostError)
	// Get returns the resource corresponding with the given identifier.
	Get(id string) (Resource, errors.GetError)
	// GetAll returns a paginated list of resources.
	GetAll(params ListRequestParams) (Page, errors.GetError)
	// Replace replaces ALL existing attributes of the resource with given identifier. Given attributes that are empty
	// are to be deleted. Returns a resource with the attributes that are stored.
	Replace(id string, attributes ResourceAttributes) (Resource, errors.PutError)
	// Delete removes the resource with corresponding ID.
	Delete(id string) errors.DeleteError
}
