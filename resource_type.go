package scim

import (
	"bytes"
	"encoding/json"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

// ResourceType specifies the metadata about a resource type.
type ResourceType struct {
	// ID is the resource type's server unique id. This is often the same value as the "name" attribute.
	ID optional.String
	// Name is the resource type name. This name is referenced by the "meta.resourceType" attribute in all resources.
	Name string
	// Description is the resource type's human-readable description.
	Description optional.String
	// Endpoint is the resource type's HTTP-addressable endpoint relative to the Base URL of the service provider,
	// e.g., "/Users".
	Endpoint string
	// Schema is the resource type's primary/base schema.
	Schema schema.Schema
	// SchemaExtensions is a list of the resource type's schema extensions.
	SchemaExtensions []SchemaExtension

	// Handler is the set of callback method that connect the SCIM server with a provider of the resource type.
	Handler ResourceHandler
}

// SchemaExtension is one of the resource type's schema extensions.
type SchemaExtension struct {
	// Schema is the URI of an extended schema, e.g., "urn:edu:2.0:Staff".
	Schema schema.Schema
	// Required is a boolean value that specifies whether or not the schema extension is required for the resource
	// type. If true, a resource of this type MUST include this schema extension and also include any attributes
	// declared as required in this schema extension. If false, a resource of this type MAY omit this schema
	// extension.
	Required bool
}

func (t ResourceType) validate(raw []byte) (ResourceAttributes, errors.ValidationError) {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()

	var m map[string]interface{}
	err := d.Decode(&m)
	if err != nil {
		return ResourceAttributes{}, errors.ValidationErrorInvalidSyntax
	}

	attributes, scimErr := t.Schema.Validate(m)
	if scimErr != errors.ValidationErrorNil {
		return ResourceAttributes{}, scimErr
	}

	for _, extension := range t.SchemaExtensions {
		extensionField := m[extension.Schema.ID]
		if extensionField == nil {
			if extension.Required {
				return ResourceAttributes{}, errors.ValidationErrorInvalidValue
			}
			continue
		}

		extensionAttributes, scimErr := extension.Schema.Validate(extensionField)
		if scimErr != errors.ValidationErrorNil {
			return ResourceAttributes{}, scimErr
		}

		attributes[extension.Schema.ID] = extensionAttributes
	}

	return attributes, errors.ValidationErrorNil
}

// MarshalJSON converts the resource type struct to its corresponding json representation.
func (t ResourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":               t.ID.Value(),
		"name":             t.Name,
		"description":      t.Description.Value(),
		"endpoint":         t.Endpoint,
		"schema":           t.Schema.ID,
		"schemaExtensions": t.SchemaExtensions,
	})
}

// MarshalJSON converts the schema extensions struct to its corresponding json representation.
func (t SchemaExtension) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"schema":   t.Schema.ID,
		"required": t.Required,
	})
}
