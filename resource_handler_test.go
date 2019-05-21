package scim

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/elimity-com/scim/errors"
)

func newTestResourceHandler() testResourceHandler {
	data := make(map[string]ResourceAttributes)
	data["0001"] = ResourceAttributes{
		"userName": "test",
	}

	return testResourceHandler{
		data: data,
	}
}

type testResourceHandler struct {
	data map[string]ResourceAttributes
}

func (h testResourceHandler) Create(attributes ResourceAttributes) (Resource, errors.PostError) {
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%04d", rand.Intn(9999))
	h.data[id] = attributes
	return Resource{
		ID:         id,
		Attributes: attributes,
	}, errors.PostErrorNil
}

func (h testResourceHandler) Get(id string) (Resource, errors.GetError) {
	data, ok := h.data[id]
	if !ok {
		return Resource{}, errors.GetErrorResourceNotFound
	}
	return Resource{
		ID:         id,
		Attributes: data,
	}, errors.GetErrorNil
}

func (h testResourceHandler) GetAll() ([]Resource, errors.GetAllError) {
	all := make([]Resource, 0)
	for k, v := range h.data {
		all = append(all, Resource{
			ID:         k,
			Attributes: v,
		})
	}
	return all, errors.GetAllErrorNil
}

func (h testResourceHandler) Replace(id string, attributes ResourceAttributes) (Resource, errors.PutError) {
	_, ok := h.data[id]
	if !ok {
		return Resource{}, errors.PutErrorResourceNotFound
	}
	h.data[id] = attributes
	return Resource{
		ID:         id,
		Attributes: attributes,
	}, errors.PutErrorNil
}

func (h testResourceHandler) Delete(id string) errors.DeleteError {
	_, ok := h.data[id]
	if !ok {
		return errors.DeleteErrorResourceNotFound
	}
	delete(h.data, id)
	return errors.DeleteErrorNil
}
