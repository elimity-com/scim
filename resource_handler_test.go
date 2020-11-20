package scim

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
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

func (h testResourceHandler) Create(r *http.Request, attributes ResourceAttributes) (Resource, error) {
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
		ExternalID: h.externalID(attributes),
		Attributes: attributes,
		Meta: Meta{
			Created:      &now,
			LastModified: &now,
			Version:      fmt.Sprintf("v%s", id),
		},
	}, nil
}

func (h testResourceHandler) Get(r *http.Request, id string) (Resource, error) {
	// check if resource exists
	data, ok := h.data[id]
	if !ok {
		return Resource{}, errors.ScimErrorResourceNotFound(id)
	}

	created, _ := time.ParseInLocation(time.RFC3339, fmt.Sprintf("%v", data.meta["created"]), time.UTC)
	lastModified, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", data.meta["lastModified"]))

	// return resource with given identifier
	return Resource{
		ID:         id,
		ExternalID: h.externalID(data.resourceAttributes),
		Attributes: data.resourceAttributes,
		Meta: Meta{
			Created:      &created,
			LastModified: &lastModified,
			Version:      fmt.Sprintf("%v", data.meta["version"]),
		},
	}, nil
}

func (h testResourceHandler) GetAll(r *http.Request, params ListRequestParams) (Page, error) {
	resources := make([]Resource, 0)
	i := 1

	for k, v := range h.data {
		if i > (params.StartIndex + params.Count - 1) {
			break
		}

		if i >= params.StartIndex {
			resources = append(resources, Resource{
				ID:         k,
				ExternalID: h.externalID(v.resourceAttributes),
				Attributes: v.resourceAttributes,
			})
		}
		i++
	}

	return Page{
		TotalResults: len(h.data),
		Resources:    resources,
	}, nil
}

func (h testResourceHandler) Replace(r *http.Request, id string, attributes ResourceAttributes) (Resource, error) {
	// check if resource exists
	_, ok := h.data[id]
	if !ok {
		return Resource{}, errors.ScimErrorResourceNotFound(id)
	}

	// replace (all) attributes
	h.data[id] = testData{
		resourceAttributes: attributes,
	}

	// return resource with replaced attributes
	return Resource{
		ID:         id,
		ExternalID: h.externalID(attributes),
		Attributes: attributes,
	}, nil
}

func (h testResourceHandler) Delete(r *http.Request, id string) error {
	// check if resource exists
	_, ok := h.data[id]
	if !ok {
		return errors.ScimErrorResourceNotFound(id)
	}

	// delete resource
	delete(h.data, id)

	return nil
}

func (h testResourceHandler) Patch(r *http.Request, id string, req PatchRequest) (Resource, error) {
	if isRequestNoChange(r) {
		return Resource{}, nil
	}
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
		ExternalID: h.externalID(h.data[id].resourceAttributes),
		Attributes: h.data[id].resourceAttributes,
		Meta: Meta{
			Created:      &created,
			LastModified: &now,
			Version:      fmt.Sprintf("%s.patch", h.data[id].meta["version"]),
		},
	}, nil
}

func (h testResourceHandler) externalID(attributes ResourceAttributes) optional.String {
	if eID, ok := attributes["externalId"]; ok {
		externalID, ok := eID.(string)
		if !ok {
			return optional.String{}
		}
		return optional.NewString(externalID)
	}

	return optional.String{}
}
