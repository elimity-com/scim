package schema

import (
	"fmt"
	"regexp"
)

func checkAttributeName(name string) {
	match, err := regexp.MatchString(`^[A-Za-z][\w-]*$`, name)
	if err != nil {
		panic(err)
	}

	if !match {
		panic(fmt.Sprintf("invalid attribute name %q", name))
	}
}

type AttributeMutability struct {
	m attributeMutability
}

var (
	AttributeMutabilityImmutable = AttributeMutability{m: attributeMutabilityImmutable}
	AttributeMutabilityReadOnly  = AttributeMutability{m: attributeMutabilityReadOnly}
	AttributeMutabilityReadWrite = AttributeMutability{m: attributeMutabilityReadWrite}
	AttributeMutabilityWriteOnly = AttributeMutability{m: attributeMutabilityWriteOnly}
)

type attributeMutability int

const (
	attributeMutabilityReadWrite attributeMutability = iota
	attributeMutabilityImmutable
	attributeMutabilityReadOnly
	attributeMutabilityWriteOnly
)

type AttributeReferenceType string

const (
	AttributeReferenceTypeExternal AttributeReferenceType = "external"
	AttributeReferenceTypeURI      AttributeReferenceType = "uri"
)

type AttributeReturned struct {
	r attributeReturned
}

var (
	AttributeReturnedAlways  = AttributeReturned{r: attributeReturnedAlways}
	AttributeReturnedDefault = AttributeReturned{r: attributeReturnedDefault}
	AttributeReturnedNever   = AttributeReturned{r: attributeReturnedNever}
	AttributeReturnedRequest = AttributeReturned{r: attributeReturnedRequest}
)

type attributeReturned int

const (
	attributeReturnedDefault attributeReturned = iota
	attributeReturnedAlways
	attributeReturnedNever
	attributeReturnedRequest
)

type AttributeType struct {
	t attributeType
}

var (
	AttributeTypeBinary   = AttributeType{t: attributeTypeBinary}
	AttributeTypeBoolean  = AttributeType{t: attributeTypeBoolean}
	AttributeTypeDateTime = AttributeType{t: attributeTypeDateTime}
	AttributeTypeDecimal  = AttributeType{t: attributeTypeDecimal}
	AttributeTypeInteger  = AttributeType{t: attributeTypeInteger}
)

type attributeType int

const (
	attributeTypeBinary attributeType = iota
	attributeTypeBoolean
	attributeTypeComplex
	attributeTypeDateTime
	attributeTypeDecimal
	attributeTypeInteger
	attributeTypeReference
	attributeTypeString
)

type AttributeUniqueness struct {
	u attributeUniqueness
}

var (
	AttributeUniquenessGlobal = AttributeUniqueness{u: attributeUniquenessGlobal}
	AttributeUniquenessNone   = AttributeUniqueness{u: attributeUniquenessNone}
	AttributeUniquenessServer = AttributeUniqueness{u: attributeUniquenessServer}
)

type attributeUniqueness int

const (
	attributeUniquenessNone attributeUniqueness = iota
	attributeUniquenessGlobal
	attributeUniquenessServer
)
