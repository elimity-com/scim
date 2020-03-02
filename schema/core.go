package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	datetime "github.com/di-wu/xsd-datetime"
	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
)

// SimpleCoreAttribute creates a non-complex attribute based on given parameters.
func SimpleCoreAttribute(params SimpleParams) CoreAttribute {
	checkAttributeName(params.name)

	return CoreAttribute{
		canonicalValues: params.canonicalValues,
		caseExact:       params.caseExact,
		description:     params.description,
		multiValued:     params.multiValued,
		mutability:      params.mutability,
		name:            params.name,
		referenceTypes:  params.referenceTypes,
		required:        params.required,
		returned:        params.returned,
		typ:             params.typ,
		uniqueness:      params.uniqueness,
	}
}

// ComplexCoreAttribute creates a complex attribute based on given parameters.
func ComplexCoreAttribute(params ComplexParams) CoreAttribute {
	checkAttributeName(params.Name)

	names := map[string]int{}
	var sa []CoreAttribute
	for i, a := range params.SubAttributes {
		name := strings.ToLower(a.name)
		if j, ok := names[name]; ok {
			panic(fmt.Errorf("duplicate name %q for sub-attributes %d and %d", name, i, j))
		}
		names[name] = i

		sa = append(sa, CoreAttribute{
			canonicalValues: a.canonicalValues,
			caseExact:       a.caseExact,
			description:     a.description,
			multiValued:     a.multiValued,
			mutability:      a.mutability,
			name:            a.name,
			referenceTypes:  a.referenceTypes,
			required:        a.required,
			returned:        a.returned,
			typ:             a.typ,
			uniqueness:      a.uniqueness,
		})
	}

	return CoreAttribute{
		description:   params.Description,
		multiValued:   params.MultiValued,
		mutability:    params.Mutability.m,
		name:          params.Name,
		required:      params.Required,
		returned:      params.Returned.r,
		subAttributes: sa,
		typ:           attributeDataTypeComplex,
		uniqueness:    params.Uniqueness.u,
	}
}

// CoreAttribute represents those attributes that sit at the top level of the JSON object together with the common
// attributes (such as the resource "id").
type CoreAttribute struct {
	canonicalValues []string
	caseExact       bool
	description     optional.String
	multiValued     bool
	mutability      attributeMutability
	name            string
	referenceTypes  []AttributeReferenceType
	required        bool
	returned        attributeReturned
	subAttributes   []CoreAttribute
	typ             attributeType
	uniqueness      attributeUniqueness
}

func (a CoreAttribute) validate(attribute interface{}) (interface{}, errors.ValidationError) {
	// return false if the attribute is not present but required.
	if attribute == nil {
		if !a.required {
			return nil, errors.ValidationErrorNil
		}
		return nil, errors.ValidationErrorInvalidValue
	}

	if a.multiValued {
		// return false if the multivalued attribute is not a slice.
		arr, ok := attribute.([]interface{})
		if !ok {
			return nil, errors.ValidationErrorInvalidSyntax
		}

		// return false if the multivalued attribute is empty.
		if a.required && len(arr) == 0 {
			return nil, errors.ValidationErrorInvalidValue
		}

		attributes := make([]interface{}, 0)
		for _, ele := range arr {
			attr, scimErr := a.validateSingular(ele)
			if scimErr != errors.ValidationErrorNil {
				return nil, scimErr
			}
			attributes = append(attributes, attr)
		}
		return attributes, errors.ValidationErrorNil
	}

	return a.validateSingular(attribute)
}

func (a CoreAttribute) validateSingular(attribute interface{}) (interface{}, errors.ValidationError) {
	switch a.typ {
	case attributeDataTypeBinary:
		bin, ok := attribute.(string)
		if !ok {
			return nil, errors.ValidationErrorInvalidValue
		}

		match, err := regexp.MatchString(`^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`, bin)
		if err != nil {
			panic(err)
		}

		if !match {
			return nil, errors.ValidationErrorInvalidValue
		}

		return bin, errors.ValidationErrorNil
	case attributeDataTypeBoolean:
		b, ok := attribute.(bool)
		if !ok {
			return nil, errors.ValidationErrorInvalidValue
		}
		return b, errors.ValidationErrorNil
	case attributeDataTypeComplex:
		complex, ok := attribute.(map[string]interface{})
		if !ok {
			return nil, errors.ValidationErrorInvalidValue
		}

		attributes := make(map[string]interface{})
		for _, sub := range a.subAttributes {
			var hit interface{}
			var found bool
			for k, v := range complex {
				if strings.EqualFold(sub.name, k) {
					if found {
						return nil, errors.ValidationErrorInvalidSyntax
					}
					found = true
					hit = v
				}
			}

			attr, scimErr := sub.validate(hit)
			if scimErr != errors.ValidationErrorNil {
				return nil, scimErr
			}
			attributes[sub.name] = attr
		}
		return attributes, errors.ValidationErrorNil
	case attributeDataTypeDateTime:
		date, ok := attribute.(string)
		if !ok {
			return nil, errors.ValidationErrorInvalidValue
		}
		_, err := datetime.Parse(date)
		if err != nil {
			return nil, errors.ValidationErrorInvalidValue
		}
		return date, errors.ValidationErrorNil
	case attributeDataTypeDecimal:
		switch n := attribute.(type) {
		case json.Number:
			f, err := n.Float64()
			if err != nil {
				return nil, errors.ValidationErrorInvalidValue
			}
			return f, errors.ValidationErrorNil
		case float64:
			return n, errors.ValidationErrorNil
		default:
			return nil, errors.ValidationErrorInvalidValue
		}
	case attributeDataTypeInteger:
		switch n := attribute.(type) {
		case json.Number:
			i, err := n.Int64()
			if err != nil {
				return nil, errors.ValidationErrorInvalidValue
			}
			return i, errors.ValidationErrorNil
		case int, int8, int16, int32, int64:
			return n, errors.ValidationErrorNil
		default:
			return nil, errors.ValidationErrorInvalidValue
		}
	case attributeDataTypeString, attributeDataTypeReference:
		s, ok := attribute.(string)
		if !ok {
			return nil, errors.ValidationErrorInvalidValue
		}
		return s, errors.ValidationErrorNil
	default:
		return nil, errors.ValidationErrorInvalidSyntax
	}
}

func (a *CoreAttribute) getRawAttributes() map[string]interface{} {
	attributes := map[string]interface{}{
		"description": a.description.Value(),
		"multiValued": a.multiValued,
		"mutability":  a.mutability,
		"name":        a.name,
		"required":    a.required,
		"returned":    a.returned,
		"type":        a.typ,
	}

	if a.canonicalValues != nil {
		attributes["canonicalValues"] = a.canonicalValues
	}

	if a.referenceTypes != nil {
		attributes["referenceTypes"] = a.referenceTypes
	}

	rawSubAttributes := make([]map[string]interface{}, len(a.subAttributes))
	for i, subAttr := range a.subAttributes {
		rawSubAttributes[i] = subAttr.getRawAttributes()
	}
	if a.subAttributes != nil && len(a.subAttributes) != 0 {
		attributes["subAttributes"] = rawSubAttributes
	}

	if a.typ != attributeDataTypeComplex && a.typ != attributeDataTypeBoolean {
		attributes["caseExact"] = a.caseExact
		attributes["uniqueness"] = a.uniqueness
	}

	return attributes
}
