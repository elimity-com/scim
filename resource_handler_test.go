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

func (h testResourceHandler) Create(attributes CoreAttributes) (Resource, PostError) {
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%04d", rand.Intn(9999))
	h.data[id] = attributes
	return Resource{
		ID:             id,
		CoreAttributes: attributes,
	}, PostErrorNil
}

func (h testResourceHandler) Get(id string) (Resource, GetError) {
	data, ok := h.data[id]
	if !ok {
		return Resource{}, NewResourceNotFoundGetError(id)
	}
	return Resource{
		ID:             id,
		CoreAttributes: data,
	}, GetErrorNil
}

func (h testResourceHandler) GetAll() ([]Resource, GetError) {
	all := make([]Resource, 0)
	for k, v := range h.data {
		all = append(all, Resource{
			ID:             k,
			CoreAttributes: v,
		})
	}
	return all, GetErrorNil
}

func (h testResourceHandler) Replace(id string, attributes CoreAttributes) (Resource, PutError) {
	_, ok := h.data[id]
	if !ok {
		return Resource{}, NewResourceNotFoundPutError(id)
	}
	h.data[id] = attributes
	return Resource{
		ID:             id,
		CoreAttributes: attributes,
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
