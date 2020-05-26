package scim

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/elimity-com/scim/errors"
)

func ExampleResourceHandler() {
	var r interface{} = testResourceHandler{}
	_, ok := r.(ResourceHandler)
	fmt.Println(ok)
	// Output: true
}

// simple in-memory resource database
type testResourceHandler struct {
	data map[string]ResourceAttributes
}

func (h testResourceHandler) Create(r *http.Request, attributes ResourceAttributes) (Resource, *errors.ScimError) {
	// create unique identifier
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%04d", rand.Intn(9999))

	// store resource
	h.data[id] = attributes

	// return stored resource
	return Resource{
		ID:         id,
		Attributes: attributes,
	}, nil
}

func (h testResourceHandler) Get(r *http.Request, id string) (Resource, *errors.ScimError) {
	// check if resource exists
	data, ok := h.data[id]
	if !ok {
		return Resource{}, errors.ScimErrorResourceNotFound(id)
	}

	// return resource with given identifier
	return Resource{
		ID:         id,
		Attributes: data,
	}, nil
}

func (h testResourceHandler) GetAll(r *http.Request, params ListRequestParams) (Page, *errors.ScimError) {
	resources := make([]Resource, 0)
	i := 1

	for k, v := range h.data {
		if i > (params.StartIndex + params.Count - 1) {
			break
		}

		if i >= params.StartIndex {
			resources = append(resources, Resource{
				ID:         k,
				Attributes: v,
			})
		}
		i++
	}

	return Page{
		TotalResults: len(h.data),
		Resources:    resources,
	}, nil
}

func (h testResourceHandler) Replace(r *http.Request, id string, attributes ResourceAttributes) (Resource, *errors.ScimError) {
	// check if resource exists
	_, ok := h.data[id]
	if !ok {
		return Resource{}, errors.ScimErrorResourceNotFound(id)
	}

	// replace (all) attributes
	h.data[id] = attributes

	// return resource with replaced attributes
	return Resource{
		ID:         id,
		Attributes: attributes,
	}, nil
}

func (h testResourceHandler) Delete(r *http.Request, id string) *errors.ScimError {
	// check if resource exists
	_, ok := h.data[id]
	if !ok {
		return errors.ScimErrorResourceNotFound(id)
	}

	// delete resource
	delete(h.data, id)

	return nil
}

func (h testResourceHandler) Patch(r *http.Request, id string, req PatchRequest) (Resource, *errors.ScimError) {
	for _, op := range req.Operations {
		switch op.Op {
		case PatchOperationAdd:
			if op.Path != "" {
				h.data[id][op.Path] = op.Value
			} else {
				valueMap := op.Value.(map[string]interface{})
				for k, v := range valueMap {
					if arr, ok := h.data[id][k].([]interface{}); ok {
						arr = append(arr, v)
						h.data[id][k] = arr
					} else {
						h.data[id][k] = v
					}
				}
			}
		case PatchOperationReplace:
			if op.Path != "" {
				h.data[id][op.Path] = op.Value
			} else {
				valueMap := op.Value.(map[string]interface{})
				for k, v := range valueMap {
					h.data[id][k] = v
				}
			}
		case PatchOperationRemove:
			h.data[id][op.Path] = nil
		}
	}

	// return resource with replaced attributes
	return Resource{
		ID:         id,
		Attributes: h.data[id],
	}, nil
}
