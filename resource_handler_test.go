package scim

import (
	"fmt"
	"math/rand"
	"time"
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

func (h testResourceHandler) Create(attributes ResourceAttributes) (Resource, PostError) {
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%04d", rand.Intn(9999))
	h.data[id] = attributes
	return Resource{
		ID:         id,
		Attributes: attributes,
	}, PostErrorNil
}

func (h testResourceHandler) Get(id string) (Resource, GetError) {
	data, ok := h.data[id]
	if !ok {
		return Resource{}, NewResourceNotFoundGetError(id)
	}
	return Resource{
		ID:         id,
		Attributes: data,
	}, GetErrorNil
}

func (h testResourceHandler) GetAll() ([]Resource, GetError) {
	all := make([]Resource, 0)
	for k, v := range h.data {
		all = append(all, Resource{
			ID:         k,
			Attributes: v,
		})
	}
	return all, GetErrorNil
}

func (h testResourceHandler) Replace(id string, attributes ResourceAttributes) (Resource, PutError) {
	_, ok := h.data[id]
	if !ok {
		return Resource{}, NewResourceNotFoundPutError(id)
	}
	h.data[id] = attributes
	return Resource{
		ID:         id,
		Attributes: attributes,
	}, PutErrorNil
}

func (h testResourceHandler) Delete(id string) DeleteError {
	_, ok := h.data[id]
	if !ok {
		return NewResourceNotFoundDeleteError(id)
	}
	delete(h.data, id)
	return DeleteErrorNil
}
