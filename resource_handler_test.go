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

type testData struct {
	resourceAttributes ResourceAttributes
	meta               map[string]string
}

// simple in-memory resource database
type testResourceHandler struct {
	data map[string]testData
}

func (h testResourceHandler) Create(r *http.Request, attributes ResourceAttributes) (Resource, errors.PostError) {
	// create unique identifier
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%04d", rand.Intn(9999))

	// store resource
	h.data[id] = testData{
		resourceAttributes: attributes,
	}

	now := time.Now()

	// return stored resource
	return Resource{
		ID:         id,
		Attributes: attributes,
		Meta: Meta{
			Created: &now,
			LastModified: &now,
			Version: fmt.Sprintf("v%s", id),
		},
	}, errors.PostErrorNil
}

func (h testResourceHandler) Get(r *http.Request, id string) (Resource, errors.GetError) {
	// check if resource exists
	data, ok := h.data[id]
	if !ok {
		return Resource{}, errors.GetErrorResourceNotFound
	}

	created, _ := time.ParseInLocation(time.RFC3339, fmt.Sprintf("%v", data.meta["created"]), time.UTC)
	lastModified, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", data.meta["lastModified"]))

	// return resource with given identifier
	return Resource{
		ID:         id,
		Attributes: data.resourceAttributes,
		Meta: Meta{
			Created:      &created,
			LastModified: &lastModified,
			Version:      fmt.Sprintf("%v", data.meta["version"]),
		},
	}, errors.GetErrorNil
}

func (h testResourceHandler) GetAll(r *http.Request, params ListRequestParams) (Page, errors.GetError) {
	resources := make([]Resource, 0)
	i := 1

	for k, v := range h.data {
		if i > (params.StartIndex + params.Count - 1) {
			break
		}

		if i >= params.StartIndex {
			resources = append(resources, Resource{
				ID:         k,
				Attributes: v.resourceAttributes,
			})
		}
		i++
	}

	return Page{
		TotalResults: len(h.data),
		Resources:    resources,
	}, errors.GetErrorNil
}

func (h testResourceHandler) Replace(
	r *http.Request, id string, attributes ResourceAttributes) (Resource, errors.PutError) {
	// check if resource exists
	_, ok := h.data[id]
	if !ok {
		return Resource{}, errors.PutErrorResourceNotFound
	}

	// replace (all) attributes
	h.data[id] = testData{
		resourceAttributes: attributes,
	}

	// return resource with replaced attributes
	return Resource{
		ID:         id,
		Attributes: attributes,
	}, errors.PutErrorNil
}

func (h testResourceHandler) Delete(r *http.Request, id string) errors.DeleteError {
	// check if resource exists
	_, ok := h.data[id]
	if !ok {
		return errors.DeleteErrorResourceNotFound
	}

	// delete resource
	delete(h.data, id)

	return errors.DeleteErrorNil
}

func (h testResourceHandler) Patch(r *http.Request, id string, req PatchRequest) (Resource, errors.PatchError) {
	for _, op := range req.Operations {
		switch op.Op {
		case PatchOperationAdd:
			if op.Path != "" {
				h.data[id].resourceAttributes[op.Path] = op.Value
			} else {
				valueMap := op.Value.(map[string]interface{})
				for k, v := range valueMap {
					if arr, ok := h.data[id].resourceAttributes[k].([]interface{}); ok {
						arr = append(arr, v)
						h.data[id].resourceAttributes[k] = arr
					} else {
						h.data[id].resourceAttributes[k] = v
					}
				}
			}
		case PatchOperationReplace:
			if op.Path != "" {
				h.data[id].resourceAttributes[op.Path] = op.Value
			} else {
				valueMap := op.Value.(map[string]interface{})
				for k, v := range valueMap {
					h.data[id].resourceAttributes[k] = v
				}
			}
		case PatchOperationRemove:
			h.data[id].resourceAttributes[op.Path] = nil
		}
	}

	created, _ := time.ParseInLocation(time.RFC3339, fmt.Sprintf("%v", h.data[id].meta["created"]), time.UTC)
	now := time.Now()

	// return resource with replaced attributes
	return Resource{
		ID:         id,
		Attributes: h.data[id].resourceAttributes,
		Meta: Meta{
			Created:      &created,
			LastModified: &now,
			Version:      fmt.Sprintf("%s.patch", h.data[id].meta["version"]),
		},
	}, errors.PatchErrorNil
}
