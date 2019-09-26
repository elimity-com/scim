package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	datetime "github.com/di-wu/xsd-datetime"
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

func (a CoreAttribute) validate(attribute interface{}) (interface{}, bool) {
	// return false if the attribute is not present but required.
	if attribute == nil {
		return nil, !a.required
	}

	if a.multiValued {
		// return false if the multivalued attribute is not a slice.
		arr, ok := attribute.([]interface{})
		if !ok {
			return nil, false
		}

		// return false if the multivalued attribute is empty.
		if a.required && len(arr) == 0 {
			return nil, false
		}

		attributes := make([]interface{}, 0)
		for _, ele := range arr {
			attr, ok := a.validateSingular(ele)
			if !ok {
				return nil, false
			}
			attributes = append(attributes, attr)
		}
		return attributes, true
	}

	return a.validateSingular(attribute)
}

func (a CoreAttribute) validateSingular(attribute interface{}) (interface{}, bool) {
	switch a.typ {
	case attributeDataTypeBinary:
		bin, ok := attribute.(string)
		if !ok {
			return nil, false
		}

		match, err := regexp.MatchString(`^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`, bin)
		if err != nil {
			panic(err)
		}

		return bin, match
	case attributeDataTypeBoolean:
		b, ok := attribute.(bool)
		if !ok {
			return nil, false
		}
		return b, true
	case attributeDataTypeComplex:
		complex, ok := attribute.(map[string]interface{})
		if !ok {
			return nil, false
		}

		attributes := make(map[string]interface{})
		for _, sub := range a.subAttributes {
			var hit interface{}
			var found bool
			for k, v := range complex {
				if strings.EqualFold(sub.name, k) {
					if found {
						return nil, false
					}
					found = true
					hit = v
				}
			}

			attr, ok := sub.validate(hit)
			if !ok {
				return nil, false
			}
			attributes[sub.name] = attr
		}
		return attributes, true
	case attributeDataTypeDateTime:
		date, ok := attribute.(string)
		if !ok {
			return nil, false
		}
		_, err := datetime.Parse(date)
		if err != nil {
			return nil, false
		}
		return date, true
	case attributeDataTypeDecimal:
		number, ok := attribute.(json.Number)
		if !ok {
			return nil, false
		}
		f, err := strconv.ParseFloat(string(number), 64)
		if err != nil {
			return nil, false
		}
		return f, true
	case attributeDataTypeInteger:
		number, ok := attribute.(json.Number)
		if !ok {
			return nil, false
		}
		i, err := strconv.ParseInt(string(number), 10, 64)
		if err != nil {
			return nil, false
		}
		return i, true
	case attributeDataTypeString, attributeDataTypeReference:
		s, ok := attribute.(string)
		if !ok {
			return nil, false
		}
		return s, true
	default:
		return nil, false
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
