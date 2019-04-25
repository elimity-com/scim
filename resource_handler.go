package scim

// CoreAttributes represents a list of attributes given to the callback method to create or replace a resource based
// on given attributes.
type CoreAttributes map[string]interface{}

// Resource represents a resource returned by a callback method.
type Resource struct {
	ID             string
	CoreAttributes CoreAttributes
}

// ResourceHandler represents a set off callback method that connect the SCIM server with a provider of a certain resource.
type ResourceHandler interface {
	Create(attributes CoreAttributes) (Resource, PostError)
	Get(id string) (Resource, GetError)
	GetAll() ([]Resource, GetError)
	Replace(id string, attributes CoreAttributes) (Resource, PutError)
	Delete(id string) DeleteError
}
