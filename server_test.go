package scim_test

import (
	"fmt"
	"github.com/elimity-com/scim/logging"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
)

func checkBodyNotEmpty(r *http.Request) error {
	// Check whether the request body is empty.
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("passed body is empty")
	}
	return nil
}

// externalID extracts the external identifier of the given attributes.
func externalID(attributes scim.ResourceAttributes) optional.String {
	if eID, ok := attributes["externalId"]; ok {
		if externalID, ok := eID.(string); ok {
			return optional.NewString(externalID)
		}
	}
	return optional.String{}
}

// Some things that are not checked:
// - Whether a reference to another entity really exists.
//   e.g. if a member gets added, does this entity exist?

func newTestServer() scim.Server {
	return scim.NewServer(scim.ServiceProviderConfig{},
		[]scim.ResourceType{
			{
				ID:          optional.NewString("User"),
				Name:        "User",
				Endpoint:    "/Users",
				Description: optional.NewString("User Account"),
				Schema:      schema.CoreUserSchema(),
				Handler: &testResourceHandler{
					data: map[string]testData{
						"0001": {attributes: map[string]interface{}{}},
					},
					schema: schema.CoreUserSchema(),
				},
			},
			{
				ID:          optional.NewString("Group"),
				Name:        "Group",
				Endpoint:    "/Groups",
				Description: optional.NewString("Group"),
				Schema:      schema.CoreGroupSchema(),
				Handler: &testResourceHandler{
					data: map[string]testData{
						"0001": {attributes: map[string]interface{}{}},
					},
					schema: schema.CoreGroupSchema(),
				},
			},
		},
		logging.NullLogger{},
	)
}

// testData represents a resource entity.
type testData struct {
	attributes scim.ResourceAttributes
	meta       scim.Meta
}

// testResourceHandler is a simple in-memory resource 'database'.
// This is for test/example purposes only! Do NOT use in production.
type testResourceHandler struct {
	// nextID is the id of the next resource that gets created.
	nextID int
	// data is a map[id]resource in which the resources are stored.
	data map[string]testData
	// schema is the reference schema of the resource handler.
	schema schema.Schema
}

func (h *testResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}

	var (
		id   = h.createID()
		now  = time.Now()
		meta = scim.Meta{
			Created:      &now,
			LastModified: &now,
			Version:      fmt.Sprintf("v%d", now.Unix()),
		}
	)
	h.data[id] = testData{
		attributes: attributes,
		meta:       meta,
	}
	return scim.Resource{
		ID:         id,
		ExternalID: externalID(attributes),
		Attributes: attributes,
		Meta:       meta,
	}, nil
}

func (h *testResourceHandler) Delete(r *http.Request, id string) error {
	if _, ok := h.data[id]; !ok {
		return errors.ScimErrorResourceNotFound(id)
	}
	delete(h.data, id)
	return nil
}

func (h testResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	resource, ok := h.data[id]
	if !ok {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}
	return scim.Resource{
		ID:         id,
		ExternalID: externalID(resource.attributes),
		Attributes: resource.attributes,
		Meta:       resource.meta,
	}, nil
}

func (h testResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	if params.Count == 0 {
		return scim.Page{
			TotalResults: len(h.data),
		}, nil
	}
	var (
		resources []scim.Resource
		index     int
	)
	for k, v := range h.data {
		index++ // 1-indexed
		if index < params.StartIndex {
			continue
		}
		if len(resources) == params.Count {
			break
		}

		if err := params.FilterValidator.PassesFilter(v.attributes); err != nil {
			continue
		}

		resources = append(resources, scim.Resource{
			ID:         k,
			ExternalID: externalID(v.attributes),
			Attributes: v.attributes,
			Meta:       v.meta,
		})
	}
	return scim.Page{
		TotalResults: len(h.data),
		Resources:    resources,
	}, nil
}

func (h *testResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}

	if _, ok := h.data[id]; !ok {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}

	var changed bool // Whether or not changes where made
	for _, op := range operations {
		// Target is the root node.
		if op.Path == nil {
			for k, v := range op.Value.(map[string]interface{}) {
				if v == nil {
					continue
				}

				path, _ := filter.ParseAttrPath([]byte(k))
				if subAttrName := path.SubAttributeName(); subAttrName != "" {
					if old, ok := h.data[id].attributes[path.AttributeName]; ok {
						m := old.(map[string]interface{})
						if sub, ok := m[subAttrName]; ok {
							if sub == v {
								continue
							}
						}
						changed = true
						m[subAttrName] = v
						h.data[id].attributes[path.AttributeName] = m
						continue
					}
					changed = true
					h.data[id].attributes[path.AttributeName] = map[string]interface{}{
						subAttrName: v,
					}
					continue
				}
				old, ok := h.data[id].attributes[k]
				if !ok {
					changed = true
					h.data[id].attributes[k] = v
					continue
				}
				switch v := v.(type) {
				case []interface{}:
					changed = true
					h.data[id].attributes[k] = append(old.([]interface{}), v...)
				case map[string]interface{}:
					m := old.(map[string]interface{})
					var changed_ bool
					for attr, value := range v {
						if value == nil {
							continue
						}

						if v, ok := m[attr]; ok {
							if v == nil || v == value {
								continue
							}
						}
						changed = true
						changed_ = true
						m[attr] = value
					}
					if changed_ {
						h.data[id].attributes[k] = m
					}
				default:
					if old == v {
						continue
					}
					changed = true
					h.data[id].attributes[k] = v // replace
				}
			}
			continue
		}

		var (
			attrName    = op.Path.AttributePath.AttributeName
			subAttrName = op.Path.AttributePath.SubAttributeName()
			valueExpr   = op.Path.ValueExpression
		)

		// Attribute does not exist yet.
		old, ok := h.data[id].attributes[attrName]
		if !ok {
			switch {
			case subAttrName != "":
				changed = true
				h.data[id].attributes[attrName] = map[string]interface{}{
					subAttrName: op.Value,
				}
			case valueExpr != nil:
				// Do nothing since there is nothing to match the filter?
			default:
				changed = true
				h.data[id].attributes[attrName] = op.Value
			}
			continue
		}

		switch op.Op {
		case "add":
			switch v := op.Value.(type) {
			case []interface{}:
				changed = true
				h.data[id].attributes[attrName] = append(old.([]interface{}), v...)
			default:
				if subAttrName != "" {
					m := old.(map[string]interface{})
					if value, ok := old.(map[string]interface{})[subAttrName]; ok {
						if v == value {
							continue
						}
					}
					changed = true
					m[subAttrName] = v
					h.data[id].attributes[attrName] = m
					continue
				}
				switch v := v.(type) {
				case map[string]interface{}:
					m := old.(map[string]interface{})
					var changed_ bool
					for attr, value := range v {
						if value == nil {
							continue
						}

						if v, ok := m[attr]; ok {
							if v == nil || v == value {
								continue
							}
						}
						changed = true
						changed_ = true
						m[attr] = value
					}
					if changed_ {
						h.data[id].attributes[attrName] = m
					}
				default:
					if old == v {
						continue
					}
					changed = true
					h.data[id].attributes[attrName] = v // replace
				}
			}
		}
	}

	if !changed {
		// StatusNoContent
		return scim.Resource{}, nil
	}

	resource := h.data[id]
	return scim.Resource{
		ID:         id,
		ExternalID: externalID(resource.attributes),
		Attributes: resource.attributes,
		Meta:       resource.meta,
	}, nil
}

func (h *testResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	if err := checkBodyNotEmpty(r); err != nil {
		return scim.Resource{}, err
	}

	resource, ok := h.data[id]
	if !ok {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}
	var (
		now  = time.Now()
		meta = scim.Meta{
			Created:      resource.meta.Created,
			LastModified: &now,
			Version:      fmt.Sprintf("v%d", now.Unix()),
		}
	)
	h.data[id] = testData{
		attributes: attributes,
		meta:       meta,
	}
	return scim.Resource{
		ID:         id,
		ExternalID: externalID(attributes),
		Attributes: attributes,
		Meta:       meta,
	}, nil
}

// createID returns a unique identifier for a resource.
func (h *testResourceHandler) createID() string {
	id := fmt.Sprintf("%04d", h.nextID)
	h.nextID++
	return id
}
