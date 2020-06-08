package schema

import (
	"encoding/json"
	"fmt"

	"github.com/elimity-com/scim/optional"
)

// ParseJSONSchema converts raw json data into a SCIM Schema.
// RFC: https://tools.ietf.org/html/rfc7643#section-7
func ParseJSONSchema(raw []byte) (Schema, error) {
	var jsonSchema map[string]interface{}
	if err := json.Unmarshal(raw, &jsonSchema); err != nil {
		return Schema{}, err
	}

	var schema Schema
	var jsonAttributes []interface{}
	for k, v := range jsonSchema {
		switch k {
		case "id":
			id, ok := v.(string)
			if !ok {
				return Schema{}, fmt.Errorf("id is not a string")
			}
			schema.ID = id
		case "name":
			name, ok := v.(string)
			if !ok {
				return Schema{}, fmt.Errorf("name is not a string")
			}
			schema.Name = optional.NewString(name)
		case "description":
			desc, ok := v.(string)
			if !ok {
				return Schema{}, fmt.Errorf("name is not a string")
			}
			schema.Description = optional.NewString(desc)
		case "attributes":
			attrs, ok := v.([]interface{})
			if !ok {
				return Schema{}, fmt.Errorf("attributes is not an array")
			}
			jsonAttributes = attrs
		}
	}

	if schema.ID == "" {
		return Schema{}, fmt.Errorf("id is empty")
	}

	schemaAttributes, err := parseAttributes(jsonAttributes)
	if err != nil {
		return Schema{}, err
	}
	schema.Attributes = schemaAttributes

	return schema, nil
}

func parseAttributes(attributes []interface{}) ([]CoreAttribute, error) {
	var schemaAttributes []CoreAttribute
	for _, a := range attributes {
		jsonAttribute, ok := a.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("attribute is not an object")
		}

		var attribute CoreAttribute
		for k, v := range jsonAttribute {
			switch k {
			case "name":
				name, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("name is not a string")
				}
				attribute.name = name
			case "type":
				typ, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("type is not a string")
				}
				switch typ {
				case "string":
					attribute.typ = attributeDataTypeString
				case "boolean":
					attribute.typ = attributeDataTypeBoolean
				case "decimal":
					attribute.typ = attributeDataTypeDecimal
				case "integer":
					attribute.typ = attributeDataTypeInteger
				case "dateTime":
					attribute.typ = attributeDataTypeDateTime
				case "binary":
					attribute.typ = attributeDataTypeBinary
				case "reference":
					attribute.typ = attributeDataTypeReference
				case "complex":
					attribute.typ = attributeDataTypeComplex
				default:
					return nil, fmt.Errorf("invalid attribute type: %s", typ)
				}
			case "subAttributes":
				jsonSubAttribute, ok := v.([]interface{})
				if !ok {
					return nil, fmt.Errorf("sub attribute is not an object")
				}
				subAttributes, err := parseAttributes(jsonSubAttribute)
				if err != nil {
					return nil, err
				}
				attribute.subAttributes = subAttributes
			case "multiValued":
				t, ok := v.(bool)
				if !ok {
					return nil, fmt.Errorf("multi valued is not a boolean")
				}
				attribute.multiValued = t
			case "description":
				desc, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("description is not a string")
				}
				attribute.description = optional.NewString(desc)
			case "required":
				t, ok := v.(bool)
				if !ok {
					return nil, fmt.Errorf("required is not a boolean")
				}
				attribute.required = t
			case "canonicalValues":
				cv, ok := v.([]interface{})
				if !ok {
					return nil, fmt.Errorf("canonical values is not an array of strings")
				}
				var values []string
				for _, s := range cv {
					vs, ok := s.(string)
					if !ok {
						return nil, fmt.Errorf("canonical value is not a string")
					}
					values = append(values, vs)
				}
				attribute.canonicalValues = values
			case "caseExact":
				t, ok := v.(bool)
				if !ok {
					return nil, fmt.Errorf("case exact is not a boolean")
				}
				attribute.caseExact = t
			case "mutability":
				mut, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("mutability is not a string")
				}
				switch mut {
				case "readOnly":
					attribute.mutability = attributeMutabilityReadOnly
				case "readWrite":
					attribute.mutability = attributeMutabilityReadWrite
				case "immutable":
					attribute.mutability = attributeMutabilityImmutable
				case "writeOnly":
					attribute.mutability = attributeMutabilityWriteOnly
				default:
					return nil, fmt.Errorf("invalid mutability type: %s", mut)
				}
			case "returned":
				ret, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("returned is not a string")
				}
				switch ret {
				case "always":
					attribute.returned = attributeReturnedAlways
				case "never":
					attribute.returned = attributeReturnedNever
				case "default":
					attribute.returned = attributeReturnedDefault
				case "request":
					attribute.returned = attributeReturnedRequest
				default:
					return nil, fmt.Errorf("invalid returned type: %s", ret)
				}
			case "uniqueness":
				uni, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("uniqueness is not a string")
				}
				switch uni {
				case "none":
					attribute.uniqueness = attributeUniquenessNone
				case "server":
					attribute.uniqueness = attributeUniquenessServer
				case "global":
					attribute.uniqueness = attributeUniquenessGlobal
				default:
					return nil, fmt.Errorf("invalid uniqueness type: %s", uni)
				}
			case "referenceTypes":
				rt, ok := v.([]interface{})
				if !ok {
					return nil, fmt.Errorf("reference types is not an array of strings")
				}
				var values []AttributeReferenceType
				for _, s := range rt {
					vs, ok := s.(string)
					if !ok {
						return nil, fmt.Errorf("reference type is not a string")
					}
					switch vs {
					case "external":
						values = append(values, AttributeReferenceTypeExternal)
					case "uri":
						values = append(values, AttributeReferenceTypeURI)
					default:
						values = append(values, AttributeReferenceType(vs))
					}
				}
				attribute.referenceTypes = values
			}
		}

		if attribute.typ == attributeDataTypeComplex && len(attribute.subAttributes) == 0 {
			return nil, fmt.Errorf("complex attributes should have sub attributes")
		}

		schemaAttributes = append(schemaAttributes, attribute)
	}
	return schemaAttributes, nil
}
