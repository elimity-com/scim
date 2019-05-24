package schema

import (
	"fmt"
	"regexp"
)

func checkAttributeName(name string) {
	// starts w/ a A-Za-z followed by a A-Za-z0-9, a hyphen or an underscore
	match, err := regexp.MatchString(`^[A-Za-z][\w-]*$`, name)
	if err != nil {
		panic(err)
	}

	if !match {
		panic(fmt.Sprintf("invalid attribute name %q", name))
	}
}

// AttributeMutability is a single keyword indicating the circumstances under which the value of the attribute can be
// (re)defined.
type AttributeMutability struct {
	m attributeMutability
}

var (
	// AttributeMutabilityImmutable indicates that the attribute MAY be defined at resource creation (e.g., POST) or at
	// record replacement via a request (e.g., a PUT). The attribute SHALL NOT be updated.
	AttributeMutabilityImmutable = AttributeMutability{m: attributeMutabilityImmutable}
	// AttributeMutabilityReadOnly indicates that the attribute SHALL NOT be modified.
	AttributeMutabilityReadOnly = AttributeMutability{m: attributeMutabilityReadOnly}
	// AttributeMutabilityReadWrite indicates that the attribute MAY be updated and read at any time.
	// This is the default value.
	AttributeMutabilityReadWrite = AttributeMutability{m: attributeMutabilityReadWrite}
	// AttributeMutabilityWriteOnly indicates that the attribute MAY be updated at any time. Attribute values SHALL NOT
	// be returned (e.g., because the value is a stored hash).
	// Note: An attribute with a mutability of "writeOnly" usually also has a returned setting of "never".
	AttributeMutabilityWriteOnly = AttributeMutability{m: attributeMutabilityWriteOnly}
)

type attributeMutability int

const (
	attributeMutabilityReadWrite attributeMutability = iota
	attributeMutabilityImmutable
	attributeMutabilityReadOnly
	attributeMutabilityWriteOnly
)

// AttributeReferenceType is a single keyword indicating the reference type of the SCIM resource that may be referenced.
// This attribute is only applicable for attributes that are of type "reference".
type AttributeReferenceType string

const (
	// AttributeReferenceTypeExternal indicates that the resource is an external resource.
	AttributeReferenceTypeExternal AttributeReferenceType = "external"
	// AttributeReferenceTypeURI indicates that the reference is to a service endpoint or an identifier.
	AttributeReferenceTypeURI AttributeReferenceType = "uri"
)

// AttributeReturned is a single keyword indicating the circumstances under which an attribute and associated values are
// returned in response to a GET request or in response to a PUT, POST, or PATCH request.
type AttributeReturned struct {
	r attributeReturned
}

var (
	// AttributeReturnedAlways indicates that the attribute is always returned.
	AttributeReturnedAlways = AttributeReturned{r: attributeReturnedAlways}
	// AttributeReturnedDefault indicates that the attribute is returned by default in all SCIM operation responses
	// where attribute values are returned.
	AttributeReturnedDefault = AttributeReturned{r: attributeReturnedDefault}
	// AttributeReturnedNever indicates that the attribute is never returned.
	AttributeReturnedNever = AttributeReturned{r: attributeReturnedNever}
	// AttributeReturnedRequest indicates that the attribute is returned in response to any PUT, POST, or PATCH
	// operations if the attribute was specified by the client (for example, the attribute was modified).
	AttributeReturnedRequest = AttributeReturned{r: attributeReturnedRequest}
)

type attributeReturned int

const (
	attributeReturnedDefault attributeReturned = iota
	attributeReturnedAlways
	attributeReturnedNever
	attributeReturnedRequest
)

// AttributeDataType is a single keyword indicating the derived data type from JSON.
type AttributeDataType struct {
	t attributeType
}

var (
	// AttributeTypeDecimal indicates that the data type is a real number with at least one digit to the left and right of the period.
	// This is the default value.
	AttributeTypeDecimal = AttributeDataType{t: attributeDataTypeDecimal}
	// AttributeTypeInteger indicates that the data type is a whole number with no fractional digits or decimal.
	AttributeTypeInteger = AttributeDataType{t: attributeDataTypeInteger}
)

type attributeType int

const (
	attributeDataTypeDecimal attributeType = iota
	attributeDataTypeInteger

	attributeDataTypeBinary
	attributeDataTypeBoolean
	attributeDataTypeComplex
	attributeDataTypeDateTime
	attributeDataTypeReference
	attributeDataTypeString
)

// AttributeUniqueness is a single keyword value that specifies how the service provider enforces uniqueness of attribute values.
type AttributeUniqueness struct {
	u attributeUniqueness
}

var (
	// AttributeUniquenessGlobal indicates that the value SHOULD be globally unique (e.g., an email address, a GUID, or
	// other value). No two resources on any server SHOULD possess the same value.
	AttributeUniquenessGlobal = AttributeUniqueness{u: attributeUniquenessGlobal}
	// AttributeUniquenessNone indicates that the values are not intended to be unique in any way.
	// This is the default value.
	AttributeUniquenessNone = AttributeUniqueness{u: attributeUniquenessNone}
	// AttributeUniquenessServer indicates that the value SHOULD be unique within the context of the current SCIM
	// endpoint (or tenancy).  No two resources on the same server SHOULD possess the same value.
	AttributeUniquenessServer = AttributeUniqueness{u: attributeUniquenessServer}
)

type attributeUniqueness int

const (
	attributeUniquenessNone attributeUniqueness = iota
	attributeUniquenessGlobal
	attributeUniquenessServer
)
