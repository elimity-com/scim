package scim

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

var metaSchema schema

func init() {
	_ = json.Unmarshal([]byte(rawMetaSchema), &metaSchema)
}

// Schema specifies the defined attribute(s) and their characteristics (mutability, returnability, etc). For every
// schema URI used in a resource object, there is a corresponding "Schema" resource.
// INFO: RFC7643 - 7.  Schema Definition
type schema struct {
	// ID is the unique URI of the schema. REQUIRED.
	ID string
	// Name is the schema's human-readable name. OPTIONAL.
	Name string
	// Description is the schema's human-readable description.  OPTIONAL.
	Description string
	// Attributes is a collection of a complex type that defines service provider attributes and their qualities.
	Attributes []*attribute
}

// Attribute is a complex type that defines service provider attributes and their qualities via the following set of
// sub-attributes.
// INFO: RFC7643 - 7.  Schema Definition
type attribute struct {
	// Name is the attribute's name.
	Name string
	// Type is the attribute's data type. Valid values are "string", "boolean", "decimal", "integer", "dateTime",
	// "reference", and "complex".  When an attribute is of type "complex", there SHOULD be a corresponding schema
	// attribute "subAttributes" defined, listing the sub-attributes of the attribute.
	Type attributeType
	// SubAttributes defines a set of sub-attributes when an attribute is of type "complex". "subAttributes" has the
	// same schema sub-attributes as "attributes".
	SubAttributes []*attribute
	// MultiValued is a boolean value indicating the attribute's plurality.
	MultiValued bool
	// Description is the attribute's human-readable description. When applicable, service providers MUST specify the
	// description.
	Description string
	// Required is a boolean value that specifies whether or not the attribute is required.
	Required bool
	// CanonicalValues is a collection of suggested canonical values that MAY be used (e.g., "work" and "home").
	// OPTIONAL.
	CanonicalValues []string
	// CaseExact is a boolean value that specifies whether or not a string attribute is case sensitive.
	CaseExact bool
	// Mutability is a single keyword indicating the circumstances under which the value of the attribute can be
	// (re)defined.
	Mutability attributeMutability
	// Returned is a single keyword that indicates when an attribute and associated values are returned in response to
	// a GET request or in response to a PUT, POST, or PATCH request.
	Returned attributeReturned
	// Uniqueness is a single keyword value that specifies how the service provider enforces uniqueness of attribute
	// values.
	Uniqueness attributeUniqueness
	// ReferenceTypes is a multi-valued array of JSON strings that indicate the SCIM resource types that may be
	// referenced.
	ReferenceTypes []string
}

type attributeType string

const (
	attributeTypeBinary    attributeType = "binary"
	attributeTypeBoolean                 = "boolean"
	attributeTypeComplex                 = "complex"
	attributeTypeDateTime                = "dateTime"
	attributeTypeDecimal                 = "decimal"
	attributeTypeInteger                 = "integer"
	attributeTypeReference               = "reference"
	attributeTypeString                  = "string"
)

type attributeMutability string

const (
	attributeMutabilityImmutable attributeMutability = "immutable"
	attributeMutabilityReadOnly                      = "readOnly"
	attributeMutabilityReadWrite                     = "readWrite"
	attributeMutabilityWriteOnly                     = "writeOnly"
)

type attributeReturned string

const (
	attributeReturnedAlways  attributeReturned = "always"
	attributeReturnedDefault                   = "default"
	attributeReturnedNever                     = "never"
	attributeReturnedRequest                   = "request"
)

type attributeUniqueness string

const (
	attributeUniquenessGlobal attributeUniqueness = "global"
	attributeUniquenessNone                       = "none"
	attributeUniquenessServer                     = "server"
)

// validate reads all bytes from given stream then unmarshals it into a map[string]interface.
// all keys in the resulting map will all be converted to lower case before validation.
func (s *schema) validate(stream io.Reader) error {
	raw, err := ioutil.ReadAll(stream)
	if err != nil {
		return err
	}

	var m map[string]interface{}
	err = json.Unmarshal(raw, &m)
	if err != nil {
		return err
	}

	i, err := lower(m)
	if err != nil {
		return err
	}

	m, _ = i.(map[string]interface{})
	return validateAttributes(s.Attributes, m)
}

// lower converts all the keys in given interface to lower case.
// returns an error on duplicate keys (e.g. "id" and "Id" are duplicates).
func lower(v interface{}) (_ interface{}, err error) {
	switch v := v.(type) {
	case []interface{}:
		for i := range v {
			v[i], err = lower(v[i])
			if err != nil {
				return nil, err
			}
		}
		return v, nil
	case map[string]interface{}:
		m := make(map[string]interface{}, len(v))
		for k, v := range v {
			// if key already exists
			if _, ok := m[strings.ToLower(k)]; ok {
				return nil, fmt.Errorf("duplicate key: %s", strings.ToLower(k))
			}
			m[strings.ToLower(k)], err = lower(v)
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	default:
		return v, nil
	}
}

// validateAttributes checks whether all required fields are present in given map and validates the type.
// ignores fields of given attributes such as: description, required, caseExact, mutability, returned and uniqueness.
func validateAttributes(ref []*attribute, m map[string]interface{}) error {
	for _, attr := range ref {
		_name, ok := m[strings.ToLower(attr.Name)]
		if !ok {
			if attr.Required {
				return fmt.Errorf("could not find required value %s in %v", strings.ToLower(attr.Name), m)
			}
			// attribute not found and not required
			continue
		}

		if attr.MultiValued {
			arr, ok := _name.([]interface{})
			if !ok {
				return fmt.Errorf("could not convert %v to type %s", _name, attr.Type)
			}

			switch attr.Type {
			case attributeTypeComplex:
				for _, sub := range arr {
					m, ok := sub.(map[string]interface{})
					if !ok {
						return fmt.Errorf("element of slice was not a complex value: %v", sub)
					}
					if err := validateAttributes(attr.SubAttributes, m); err != nil {
						return err
					}
				}
			case attributeTypeString:
				for _, sub := range arr {
					_, ok := sub.(string)
					if !ok {
						return fmt.Errorf("could not convert %v to type %s", sub, attr.Type)
					}
				}
			default:
				return fmt.Errorf("not implemented/invalid type: %v", attr.Type)
			}
			continue
		}

		switch attr.Type {
		case attributeTypeBoolean:
			_, ok := _name.(bool)
			if !ok {
				return fmt.Errorf("could not convert %v to type %s", _name, attr.Type)
			}
		case attributeTypeComplex:
			c, ok := _name.(map[string]interface{})
			if !ok {
				return fmt.Errorf("could not convert %v to type %s", _name, attr.Type)
			}
			if err := validateAttributes(attr.SubAttributes, c); err != nil {
				return err
			}
		case attributeTypeString:
			name, ok := _name.(string)
			if !ok {
				return fmt.Errorf("could not convert %v to type %s", _name, attr.Type)
			}

			if attr.CanonicalValues != nil {
				var contains bool
				for _, v := range attr.CanonicalValues {
					if v == name {
						contains = true
					}
				}
				if !contains {
					return fmt.Errorf("%v not in canonical values %v", name, attr.CanonicalValues)

				}
			}
		default:
			return fmt.Errorf("not implemented/invalid type: %v", attr.Type)
		}
	}
	return nil
}
