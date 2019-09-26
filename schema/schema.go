package schema

import (
	"encoding/json"
	"strings"

	"github.com/elimity-com/scim/optional"
)

// Schema is a collection of attribute definitions that describe the contents of an entire or partial resource.
type Schema struct {
	ID          string
	Name        string
	Description optional.String
	Attributes  []CoreAttribute
}

// Validate validates given resource based on the schema.
func (s Schema) Validate(resource interface{}) (map[string]interface{}, bool) {
	core, ok := resource.(map[string]interface{})
	if !ok {
		return nil, false
	}

	attributes := make(map[string]interface{})
	for _, attribute := range s.Attributes {
		var hit interface{}
		var found bool
		for k, v := range core {
			if strings.EqualFold(attribute.name, k) {
				if found {
					return nil, false
				}
				found = true
				hit = v
			}
		}

		attr, ok := attribute.validate(hit)
		if !ok {
			return nil, false
		}
		attributes[attribute.name] = attr
	}
	return attributes, true
}

// MarshalJSON converts the schema struct to its corresponding json representation.
func (s Schema) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":          s.ID,
		"name":        s.Name,
		"description": s.Description.Value(),
		"attributes":  s.Attributes,
	})
}
