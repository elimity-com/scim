package scim

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/elimity-com/scim/errors"
)

// simple in-memory resource database
type testResourceHandler struct {
	data map[string]ResourceAttributes
}

func (h testResourceHandler) Create(attributes ResourceAttributes) (Resource, errors.PostError) {
	// create unique identifier
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%04d", rand.Intn(9999))

	// store resource
	h.data[id] = attributes

	// return stored resource
	return Resource{
		ID:         id,
		Attributes: attributes,
	}, errors.PostErrorNil
}

func (h testResourceHandler) Get(id string) (Resource, errors.GetError) {
	// check if resource exists
	data, ok := h.data[id]
	if !ok {
		return Resource{}, errors.GetErrorResourceNotFound
	}

	// return resource with given identifier
	return Resource{
		ID:         id,
		Attributes: data,
	}, errors.GetErrorNil
}

func (h testResourceHandler) GetAll() []Resource {
	// get all existing resources
	all := make([]Resource, 0)
	for k, v := range h.data {
		all = append(all, Resource{
			ID:         k,
			Attributes: v,
		})
	}

	// return all resources
	return all
}

func (h testResourceHandler) Replace(id string, attributes ResourceAttributes) (Resource, errors.PutError) {
	// check if resource exists
	_, ok := h.data[id]
	if !ok {
		return Resource{}, errors.PutErrorResourceNotFound
	}

	// replace (all) attributes
	h.data[id] = attributes

	// return resource with replaced attributes
	return Resource{
		ID:         id,
		Attributes: attributes,
	}, errors.PutErrorNil
}

func (h testResourceHandler) Delete(id string) errors.DeleteError {
	// check if resource exists
	_, ok := h.data[id]
	if !ok {
		return errors.DeleteErrorResourceNotFound
	}

	// delete resource
	delete(h.data, id)

	return errors.DeleteErrorNil
}

func Example_resourceHandler() {}
