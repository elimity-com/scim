package schema

import (
	"fmt"
	"regexp"
	"strings"
	"time"

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

func (a CoreAttribute) validate(attribute interface{}) bool {
	if attribute == nil {
		return !a.required
	}

	if a.multiValued {
		arr, ok := attribute.([]interface{})
		if !ok {
			return false
		}

		if a.required && len(arr) == 0 {
			return false
		}

		for _, ele := range arr {
			if !a.validateSingular(ele) {
				return false
			}
		}
		return true
	}

	return a.validateSingular(attribute)
}

func (a CoreAttribute) validateSingular(attribute interface{}) bool {
	switch a.typ {
	case attributeDataTypeBinary:
		bin, ok := attribute.(string)
		if !ok {
			return false
		}

		match, err := regexp.MatchString(`^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`, bin)
		if err != nil {
			panic(err)
		}

		return match
	case attributeDataTypeBoolean:
		if _, ok := attribute.(bool); !ok {
			return false
		}
		return true
	case attributeDataTypeComplex:
		complex, ok := attribute.(map[string]interface{})
		if !ok {
			return false
		}

		for _, sub := range a.subAttributes {
			var hit interface{}
			var found bool
			for k, v := range complex {
				if strings.EqualFold(sub.name, k) {
					if found {
						return false
					}
					found = true
					hit = v
				}
			}

			if !sub.validate(hit) {
				return false
			}
		}
		return true
	case attributeDataTypeDateTime:
		date, ok := attribute.(string)
		if !ok {
			return false
		}
		_, err := time.Parse(time.RFC3339Nano, date)
		if err != nil {
			return false
		}
		return true
	case attributeDataTypeDecimal:
		_, ok := attribute.(float64)
		if !ok {
			return false
		}
		return true
	case attributeDataTypeInteger:
		_, ok := attribute.(int)
		if !ok {
			return false
		}
		return true
	case attributeDataTypeString, attributeDataTypeReference:
		if _, ok := attribute.(string); !ok {
			return false
		}
		return true
	default:
		return false
	}
}
