package scim

import (
	"fmt"
	"math/rand"
	"time"
)

func newTestResourceHandler() testResourceHandler {
	data := make(map[string]CoreAttributes)
	data["0001"] = CoreAttributes{
		"userName": "test",
	}

	return testResourceHandler{
		data: data,
	}
}

type testResourceHandler struct {
	data map[string]CoreAttributes
}

func (h testResourceHandler) Create(attributes CoreAttributes) (Resource, error) {
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%04d", rand.Intn(9999))
	h.data[id] = attributes
	return Resource{
		ID:             id,
		CoreAttributes: attributes,
	}, nil
}

func (h testResourceHandler) Get(id string) (Resource, error) {
	data, ok := h.data[id]
	if !ok {
		return Resource{}, fmt.Errorf("resource not found with id: %s", id)
	}
	return Resource{
		ID:             id,
		CoreAttributes: data,
	}, nil
}

func (h testResourceHandler) GetAll() ([]Resource, error) {
	all := make([]Resource, 0)
	for k, v := range h.data {
		all = append(all, Resource{
			ID:             k,
			CoreAttributes: v,
		})
	}
	return all, nil
}

func (h testResourceHandler) Replace(id string, attributes CoreAttributes) (Resource, error) {
	_, ok := h.data[id]
	if !ok {
		return Resource{}, fmt.Errorf("resource not found with id: %s", id)
	}
	h.data[id] = attributes
	return Resource{
		ID:             id,
		CoreAttributes: attributes,
	}, nil
}

func (h testResourceHandler) Delete(id string) error {
	_, ok := h.data[id]
	if !ok {
		return fmt.Errorf("resource not found with id: %s", id)
	}
	delete(h.data, id)
	return nil
}
