package schema

import (
	"fmt"
	"github.com/elimity-com/scim/optional"
	"strings"
)

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

func ComplexCoreAttribute(params ComplexParams) CoreAttribute {
	checkAttributeName(params.Name)

	names := map[string]int{}
	var sa []CoreAttribute
	for i, a := range params.SubAttributes {
		if a.name == "" {
			panic(fmt.Errorf("empty name for sub-attribute %d", i))
		}

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
		typ:           attributeTypeComplex,
		uniqueness:    params.Uniqueness.u,
	}
}

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
