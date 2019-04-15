package scim

type CoreAttributes map[string]interface{}

type Resource struct {
	ID             string
	CoreAttributes CoreAttributes
}

type ResourceHandler interface {
	Create(attributes CoreAttributes) (Resource, error)
	Get(id string) (Resource, error)
	GetAll() ([]Resource, error)
	Replace(id string, attributes CoreAttributes) (Resource, error)
	Delete(id string) error
}
