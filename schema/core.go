package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
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
		number, ok := attribute.(json.Number)
		if !ok {
			return nil, errors.ValidationErrorInvalidValue
		}
		f, err := strconv.ParseFloat(string(number), 64)
		if err != nil {
			return nil, errors.ValidationErrorInvalidValue
		}
		return f, errors.ValidationErrorNil
	case attributeDataTypeInteger:
		if reflect.TypeOf(attribute).Kind() != reflect.Int {
			return nil, errors.ValidationErrorInvalidValue
		}
		return attribute.(int), errors.ValidationErrorNil
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

// MarshalJSON converts the attribute struct to its corresponding json representation.
func (a CoreAttribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"canonicalValues": a.canonicalValues,
		"caseExact":       a.caseExact,
		"description":     a.description.Value(),
		"multiValued":     a.multiValued,
		"mutability":      a.mutability,
		"name":            a.name,
		"referenceTypes":  a.referenceTypes,
		"required":        a.required,
		"returned":        a.returned,
		"subAttributes":   a.subAttributes,
		"type":            a.typ,
		"uniqueness":      a.uniqueness,
	})
}
