package schema

import (
	"fmt"
	"strings"

	"github.com/elimity-com/scim/optional"
)

// NewSchema creates a schema with given identifier, name, description and attributes.
func NewSchema(id, name string, desc optional.String, attributes []CoreAttribute) Schema {
	checkAttributeName(name)

	names := map[string]int{}
	for i, a := range attributes {
		name := strings.ToLower(a.name)
		if j, ok := names[name]; ok {
			panic(fmt.Errorf("duplicate name %q for sub-attributes %d and %d", name, i, j))
		}
		names[name] = i
	}

	return Schema{
		id:          id,
		name:        name,
		description: desc,
		attributes:  attributes,
	}
}

// Schema is a collection of attribute definitions that describe the contents of an entire or partial resource.
type Schema struct {
	id          string
	name        string
	description optional.String
	attributes  []CoreAttribute
}

func (s Schema) validate(resource interface{}) bool {
	core, ok := resource.(map[string]interface{})
	if !ok {
		return false
	}

	for _, attribute := range s.attributes {
		var hit interface{}
		var found bool
		for k, v := range core {
			if strings.EqualFold(attribute.name, k) {
				if found {
					return false
				}
				found = true
				hit = v
			}
		}

		if !attribute.validate(hit) {
			return false
		}
	}
	return true
}
