package scim

// Attributes represents a list of attributes given to the callback method to create or replace a resource based
// on given attributes.
type Attributes map[string]interface{}

// Resource represents a resource returned by a callback method.
type Resource struct {
	ID         string
	Attributes Attributes
}

func (r Resource) response(resourceType resourceType, location string) Attributes {
	response := r.Attributes
	response["id"] = r.ID
	schemas := []string{resourceType.Schema}
	for _, schema := range resourceType.SchemaExtensions {
		schemas = append(schemas, schema.Schema)
	}
	response["schemas"] = schemas
	response["meta"] = meta{
		ResourceType: resourceType.Name,
		Location:     location,
	}

	return response
}

// ResourceHandler represents a set off callback method that connect the SCIM server with a provider of a certain resource.
type ResourceHandler interface {
	Create(attributes Attributes) (Resource, PostError)
	Get(id string) (Resource, GetError)
	GetAll() ([]Resource, GetError)
	Replace(id string, attributes Attributes) (Resource, PutError)
	Delete(id string) DeleteError
}
