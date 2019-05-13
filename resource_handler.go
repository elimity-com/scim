package scim

import (
	"fmt"
	"net/url"
)

// Attributes represents a list of attributes given to the callback method to create or replace a resource based
// on given attributes.
type Attributes map[string]interface{}

// Resource represents a resource returned by a callback method.
type Resource struct {
	ID         string
	Attributes Attributes
}

func (r Resource) response(resourceType resourceType) Attributes {
	response := r.Attributes
	response["id"] = r.ID
	schemas := []string{resourceType.Schema}
	for _, schema := range resourceType.SchemaExtensions {
		schemas = append(schemas, schema.Schema)
	}
	response["schemas"] = schemas
	response["meta"] = meta{
		ResourceType: resourceType.Name,
		Location:     fmt.Sprintf("%s/%s", resourceType.Endpoint[1:], url.PathEscape(r.ID)),
	}

	return response
}

// ResourceHandler represents a set off callback method that connect the SCIM server with a provider of a certain resource.
type ResourceHandler interface {
	// Create stores given attributes. Returns a resource with the attributes that are stored and a (new) unique identifier.
	Create(attributes Attributes) (Resource, PostError)
	// Get returns the resource corresponding with the given identifier.
	Get(id string) (Resource, GetError)
	// GetAll returns all the resources.
	GetAll() ([]Resource, GetError)
	// Replace replaces ALL existing attributes of the resource with given identifier. Given attributes that are empty
	// are to be deleted. Returns a resource with the attributes that are stored.
	Replace(id string, attributes Attributes) (Resource, PutError)
	// Delete removes the resource with corresponding id.
	Delete(id string) DeleteError
}
