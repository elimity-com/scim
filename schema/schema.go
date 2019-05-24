package schema

import (
	"fmt"
	"github.com/elimity-com/scim/optional"
	"strings"
)

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

type Schema struct {
	id          string
	name        string
	description optional.String
	attributes  []CoreAttribute
}
