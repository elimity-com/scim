package scim

// CoreAttributes represents a list of attributes given to the callback method to create or replace a resource based
// on given attributes.
type CoreAttributes map[string]interface{}

// Resource represents a resource returned by a callback method.
type Resource struct {
	ID             string
	CoreAttributes CoreAttributes
}

func (r Resource) response(resourceType resourceType, location string) CoreAttributes {
	response := r.CoreAttributes
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
	Create(attributes CoreAttributes) (Resource, PostError)
	Get(id string) (Resource, GetError)
	GetAll() ([]Resource, GetError)
	Replace(id string, attributes CoreAttributes) (Resource, PutError)
	Delete(id string) DeleteError
}
