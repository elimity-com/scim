package scim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

// NewSchemaFromFile reads the file from given filepath and returns a validated schema if no errors take place.
func NewSchemaFromFile(filepath string) (Schema, error) {
	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return Schema{}, err
	}

	return NewSchemaFromBytes(raw)
}

// NewSchemaFromString returns a validated schema if no errors take place.
func NewSchemaFromString(s string) (Schema, error) {
	return NewSchemaFromBytes([]byte(s))
}

// NewSchemaFromBytes returns a validated schema if no errors take place.
func NewSchemaFromBytes(raw []byte) (Schema, error) {
	_, scimErr := metaSchema.validate(raw, read)
	if scimErr != scimErrorNil {
		return Schema{}, fmt.Errorf(scimErr.detail)
	}

	var schema schema
	err := json.Unmarshal(raw, &schema)
	if err != nil {
		log.Fatalf("failed parsing schema: %v", err)
	}

	return Schema{schema}, nil
}

// Schema specifies the defined attribute(s) and their characteristics (mutability, returnability, etc).
type Schema struct {
	schema schema
}

// schema specifies the defined attribute(s) and their characteristics (mutability, returnability, etc). For every
// schema URI used in a resource object, there is a corresponding "Schema" resource.
//
// RFC: RFC7643 - https://tools.ietf.org/html/rfc7643#section-7
type schema struct {
	// ID is the unique URI of the schema. REQUIRED.
	ID string
	// Name is the schema's human-readable name. OPTIONAL.
	Name string
	// Description is the schema's human-readable description.  OPTIONAL.
	Description *string
	// Attributes is a collection of a complex type that defines service provider attributes and their qualities.
	Attributes attributes
}

func (s schema) MarshalJSON() ([]byte, error) {
	schema := map[string]interface{}{
		"id":         s.ID,
		"name":       s.Name,
		"attributes": s.Attributes,
		"meta": meta{
			ResourceType: "Schema",
			Location:     "/v2/Schemas/" + s.ID,
		},
	}

	if s.Description != nil {
		schema["description"] = s.Description
	}

	return json.Marshal(schema)
}

// validate validates given bytes based on the schema and validation mode.
func (s schema) validate(raw []byte, mode validationMode) (CoreAttributes, scimError) {
	var m interface{}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()

	err := d.Decode(&m)
	if err != nil {
		return CoreAttributes{}, scimErrorInvalidSyntax
	}
	return s.Attributes.validate(m, mode)
}

// attribute is a complex type that defines service provider attributes and their qualities via the following set of
// sub-attributes.
//
// RFC: https://tools.ietf.org/html/rfc7643#section-7
type attribute struct {
	// Name is the attribute's name.
	Name string `json:"name"`
	// Type is the attribute's data type. Valid values are "string", "boolean", "decimal", "integer", "dateTime",
	// "reference", and "complex".  When an attribute is of type "complex", there SHOULD be a corresponding schema
	// attribute "subAttributes" defined, listing the sub-attributes of the attribute.
	Type attributeType `json:"type"`
	// SubAttributes defines a set of sub-attributes when an attribute is of type "complex". "subAttributes" has the
	// same schema sub-attributes as "attributes".
	SubAttributes attributes `json:"subAttributes,omitempty"`
	// MultiValued is a boolean value indicating the attribute's plurality.
	MultiValued bool `json:"multiValued"`
	// Description is the attribute's human-readable description. When applicable, service providers MUST specify the
	// description.
	Description string `json:"description,omitempty"`
	// Required is a boolean value that specifies whether or not the attribute is required.
	Required bool `json:"required,omitempty"`
	// CanonicalValues is a collection of suggested canonical values that MAY be used (e.g., "work" and "home").
	// OPTIONAL.
	CanonicalValues []string `json:"canonicalValues,omitempty"`
	// CaseExact is a boolean value that specifies whether or not a string attribute is case sensitive.
	CaseExact bool `json:"caseExact,omitempty"`
	// Mutability is a single keyword indicating the circumstances under which the value of the attribute can be
	// (re)defined.
	Mutability attributeMutability `json:"mutability,omitempty"`
	// Returned is a single keyword that indicates when an attribute and associated values are returned in response to
	// a GET request or in response to a PUT, POST, or PATCH request.
	Returned attributeReturned `json:"returned,omitempty"`
	// Uniqueness is a single keyword value that specifies how the service provider enforces uniqueness of attribute
	// values.
	Uniqueness attributeUniqueness `json:"uniqueness,omitempty"`
	// ReferenceTypes is a multi-valued array of JSON strings that indicate the SCIM resource types that may be
	// referenced.
	ReferenceTypes []string `json:"referenceTypes,omitempty"`
}

func (a attribute) validate(i interface{}, mode validationMode) (CoreAttributes, scimError) {
	// validate required
	if i == nil {
		if a.Required {
			return CoreAttributes{}, scimErrorInvalidValue
		}
		return CoreAttributes{}, scimErrorNil
	}

	if a.MultiValued {
		arr, ok := i.([]interface{})
		if !ok {
			return CoreAttributes{}, scimErrorInvalidSyntax
		}

		// empty array = omitted/nil
		if len(arr) == 0 && a.Required {
			return CoreAttributes{}, scimErrorInvalidValue
		}

		coreAttributes := make([]CoreAttributes, 0)
		for _, sub := range arr {
			attributes, err := a.validateSingular(sub, mode)
			if err != scimErrorNil {
				return CoreAttributes{}, err
			}
			coreAttributes = append(coreAttributes, attributes)
		}

		if mode != read {
			return CoreAttributes{a.Name: coreAttributes}, scimErrorNil
		}
		return CoreAttributes{}, scimErrorNil
	}

	return a.validateSingular(i, mode)
}

func (a attribute) validateSingular(i interface{}, mode validationMode) (CoreAttributes, scimError) {
	if mode == replace {
		switch a.Mutability {
		case attributeMutabilityImmutable:
			return CoreAttributes{}, scimErrorMutability
		case attributeMutabilityReadOnly:
			return CoreAttributes{}, scimErrorNil
		}
	}

	switch a.Type {
	case attributeTypeBoolean:
		_, ok := i.(bool)
		if !ok {
			return CoreAttributes{}, scimErrorInvalidValue
		}
	case attributeTypeComplex:
		if _, err := a.SubAttributes.validate(i, mode); err != scimErrorNil {
			return CoreAttributes{}, err
		}
	case attributeTypeString, attributeTypeReference:
		_, ok := i.(string)
		if !ok {
			return CoreAttributes{}, scimErrorInvalidValue
		}
	case attributeTypeInteger:
		n, ok := i.(json.Number)
		if !ok {
			return CoreAttributes{}, scimErrorInvalidValue
		}
		if strings.Contains(n.String(), ".") || strings.Contains(n.String(), "e") {
			return CoreAttributes{}, scimErrorInvalidValue
		}
	default:
		log.Fatalf("attribute type not implemented: %s", a.Type)
		return CoreAttributes{}, scimErrorNil
	}

	if mode != read && (a.Returned == attributeReturnedAlways || a.Returned == attributeReturnedDefault) {
		return CoreAttributes{a.Name: i}, scimErrorNil
	}
	return CoreAttributes{}, scimErrorNil
}

type attributes []attribute

func (as attributes) validate(i interface{}, mode validationMode) (CoreAttributes, scimError) {
	coreAttributes := make(CoreAttributes)

	c, ok := i.(map[string]interface{})
	if !ok {
		return CoreAttributes{}, scimErrorInvalidSyntax
	}

	for _, attribute := range as {
		// validate duplicate
		var hit interface{}
		var found bool
		for k, v := range c {
			if strings.EqualFold(attribute.Name, k) {
				if found {
					return CoreAttributes{}, scimErrorUniqueness
				}
				found = true
				hit = v
			}
		}

		attribute, err := attribute.validate(hit, mode)
		if err != scimErrorNil {
			return CoreAttributes{}, err
		}

		if mode != read {
			for k, v := range attribute {
				coreAttributes[k] = v
			}
		}
	}
	return coreAttributes, scimErrorNil
}

type attributeType string

// TODO: binary, dateTime and decimal
const (
	// attributeTypeBinary    attributeType = "binary"
	attributeTypeBoolean attributeType = "boolean"
	attributeTypeComplex attributeType = "complex"
	// attributeTypeDateTime  attributeType = "dateTime"
	// attributeTypeDecimal   attributeType = "decimal"
	attributeTypeInteger   attributeType = "integer"
	attributeTypeReference attributeType = "reference"
	attributeTypeString    attributeType = "string"
)

type attributeMutability string

// TODO: readWrite and writeOnly
const (
	attributeMutabilityImmutable attributeMutability = "immutable"
	attributeMutabilityReadOnly  attributeMutability = "readOnly"
	// attributeMutabilityReadWrite attributeMutability = "readWrite"
	// attributeMutabilityWriteOnly attributeMutability = "writeOnly"
)

type attributeReturned string

// TODO: never and request
const (
	attributeReturnedAlways  attributeReturned = "always"
	attributeReturnedDefault attributeReturned = "default"
	// attributeReturnedNever   attributeReturned = "never"
	// attributeReturnedRequest attributeReturned = "request"
)

type attributeUniqueness string

// TODO global, none and server
// const (
// attributeUniquenessGlobal attributeUniqueness = "global"
// attributeUniquenessNone   attributeUniqueness = "none"
// attributeUniquenessServer attributeUniqueness = "server"
// )

type validationMode int

const (
	// read will validate required and the type, but does not return core attributes.
	read validationMode = iota
	// write will validate required, returnability and the type.
	write
	// replace will validate required, mutability, returnability and type.
	replace
)

var metaSchema schema

func init() {
	if err := json.Unmarshal([]byte(rawSchemaSchema), &metaSchema); err != nil {
		panic(err)
	}
}
