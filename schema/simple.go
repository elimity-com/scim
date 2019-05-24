package schema

import "github.com/elimity-com/scim/optional"

type SimpleParams struct {
	canonicalValues []string
	caseExact       bool
	description     optional.String
	multiValued     bool
	mutability      attributeMutability
	name            string
	referenceTypes  []AttributeReferenceType
	required        bool
	returned        attributeReturned
	typ             attributeType
	uniqueness      attributeUniqueness
}

func SimpleReferenceParams(params ReferenceParams) SimpleParams {
	checkAttributeName(params.Name)

	return SimpleParams{
		description:    params.Description,
		multiValued:    params.MultiValued,
		mutability:     params.Mutability.m,
		name:           params.Name,
		referenceTypes: params.ReferenceTypes,
		required:       params.Required,
		returned:       params.Returned.r,
		typ:            attributeTypeReference,
		uniqueness:     params.Uniqueness.u,
	}
}

type ReferenceParams struct {
	Description    optional.String
	MultiValued    bool
	Mutability     AttributeMutability
	Name           string
	ReferenceTypes []AttributeReferenceType
	Required       bool
	Returned       AttributeReturned
	Uniqueness     AttributeUniqueness
}

func SimpleStringParams(params StringParams) SimpleParams {
	checkAttributeName(params.Name)

	return SimpleParams{
		canonicalValues: params.CanonicalValues,
		caseExact:       params.CaseExact,
		description:     params.Description,
		multiValued:     params.MultiValued,
		mutability:      params.Mutability.m,
		name:            params.Name,
		required:        params.Required,
		returned:        params.Returned.r,
		typ:             attributeTypeString,
		uniqueness:      params.Uniqueness.u,
	}
}

type StringParams struct {
	CanonicalValues []string
	CaseExact       bool
	Description     optional.String
	MultiValued     bool
	Mutability      AttributeMutability
	Name            string
	Required        bool
	Returned        AttributeReturned
	Uniqueness      AttributeUniqueness
}

func SimpleDefaulParams(params DefaultParams) SimpleParams {
	checkAttributeName(params.Name)

	return SimpleParams{
		description: params.Description,
		multiValued: params.MultiValued,
		mutability:  params.Mutability.m,
		name:        params.Name,
		required:    params.Required,
		returned:    params.Returned.r,
		typ:         params.Type.t,
		uniqueness:  params.Uniqueness.u,
	}
}

type DefaultParams struct {
	Description optional.String
	MultiValued bool
	Mutability  AttributeMutability
	Name        string
	Required    bool
	Returned    AttributeReturned
	Type        AttributeType
	Uniqueness  AttributeUniqueness
}
