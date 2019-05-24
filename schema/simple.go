package schema

import "github.com/elimity-com/scim/optional"

// SimpleParams are the parameters used to create a simple attribute.
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

// SimpleReferenceParams converts given reference parameters to their corresponding simple parameters.
func SimpleReferenceParams(params ReferenceParams) SimpleParams {
	return SimpleParams{
		caseExact:      true,
		description:    params.Description,
		multiValued:    params.MultiValued,
		mutability:     params.Mutability.m,
		name:           params.Name,
		referenceTypes: params.ReferenceTypes,
		required:       params.Required,
		returned:       params.Returned.r,
		typ:            attributeDataTypeReference,
		uniqueness:     params.Uniqueness.u,
	}
}

// ReferenceParams are the parameters used to create a simple attribute with a data type of "reference".
// A reference is case exact. A reference has a "referenceTypes" attribute that indicates what types of resources may
// be linked.
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

// SimpleStringParams converts given string parameters to their corresponding simple parameters.
func SimpleStringParams(params StringParams) SimpleParams {
	return SimpleParams{
		canonicalValues: params.CanonicalValues,
		caseExact:       params.CaseExact,
		description:     params.Description,
		multiValued:     params.MultiValued,
		mutability:      params.Mutability.m,
		name:            params.Name,
		required:        params.Required,
		returned:        params.Returned.r,
		typ:             attributeDataTypeString,
		uniqueness:      params.Uniqueness.u,
	}
}

// StringParams are the parameters used to create a simple attribute with a data type of "string".
// A string is a sequence of zero or more Unicode characters encoded using UTF-8.
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

// SimpleDefaultParams converts given default parameters to their corresponding simple parameters.
func SimpleDefaultParams(params DefaultParams) SimpleParams {
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

// DefaultParams are the parameters used to create a simple attribute with a data type other than "string" and "reference".
type DefaultParams struct {
	Description optional.String
	MultiValued bool
	Mutability  AttributeMutability
	Name        string
	Required    bool
	Returned    AttributeReturned
	Type        AttributeDataType
	Uniqueness  AttributeUniqueness
}
