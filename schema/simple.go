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
	checkAttributeName(params.Name)

	return SimpleParams{
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
		typ:             attributeDataTypeString,
		uniqueness:      params.Uniqueness.u,
	}
}

// StringParams are the parameters used to create a simple attribute with a data type of "string".
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
