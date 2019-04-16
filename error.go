package scim

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type scimType string

const (
	// One or more of the attribute values are already in use or are reserved.
	scimTypeUniqueness = "uniqueness"
	// The attempted modification is not compatible with the target attribute's mutability or current state (e.g.,
	// modification of an "immutable" attribute with an existing value).
	scimTypeMutability = "mutability"
	// The request body message structure was invalid or did not conform to the request schema.
	scimTypeInvalidSyntax = "invalidSyntax"
	// A required value was missing, or the value specified was not compatible with the operation or attribute type,
	// or resource schema.
	scimTypeInvalidValue = "invalidValue"
	// The specified SCIM protocol version is not supported.
)

var uniqueness = scimError{
	scimType: scimTypeUniqueness,
	detail:   "One or more of the attribute values are already in use or are reserved.",
	status:   http.StatusConflict,
}

var mutability = scimError{
	scimType: scimTypeMutability,
	detail:   "The attempted modification is not compatible with the target attribute's mutability or current state.",
	status:   http.StatusBadRequest,
}

var invalidSyntax = scimError{
	scimType: scimTypeInvalidSyntax,
	detail:   "The request body message structure was invalid or did not conform to the request schema.",
	status:   http.StatusBadRequest,
}

var invalidValue = scimError{
	scimType: scimTypeInvalidValue,
	detail:   "A required value was missing, or the value specified was not compatible with the operation or attribute type, or resource schema.",
	status:   http.StatusBadRequest,
}

func resourceNotFound(id string) scimError {
	return scimError{
		detail: fmt.Sprintf("Resource %s not found.", id),
		status: http.StatusNotFound,
	}
}

// RFC: https://tools.ietf.org/html/rfc7644#section-3.12
type scimError struct {
	// scimType is a SCIM detail error keyword. OPTIONAL.
	scimType scimType
	// detail is a detailed human-readable message. OPTIONAL.
	detail string
	// status is the HTTP status code expressed as a JSON string. REQUIRED.
	status int
}

func (e scimError) Error() string {
	raw, _ := e.MarshalJSON()
	return string(raw)
}

func (e scimError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"schemas":  []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		"scimType": e.scimType,
		"detail":   e.detail,
		"status":   e.status,
	})
}

var GetErrorNil GetError

type GetError struct {
	err scimError
}

// A required value was missing, or the value specified was not compatible with the operation or attribute type or resource.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.4.2
func NewInvalidValueGetError() GetError {
	return GetError{invalidValue}
}

// NewResourceNotFoundGetError returns an error with status code 404 and a human readable message containing the identifier
// of the resource that was requested but not found.
func NewResourceNotFoundGetError(id string) GetError {
	return GetError{resourceNotFound(id)}
}

var PostErrorNil PostError

type PostError struct {
	err scimError
}

// If the service provider determines that the creation of the requested resource conflicts with existing resources
// (e.g., a "User" resource with a duplicate "userName"), the service provider MUST return HTTP status code 409
// (Conflict) with a "scimType" error code of "uniqueness".
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.3
func NewUniquenessPostError() PostError {
	return PostError{uniqueness}
}

// Request is unparsable, syntactically incorrect, or violates schema.
func NewInvalidSyntaxPostError() PostError {
	return PostError{invalidSyntax}
}

// A required value was missing, or the value specified was not compatible with the operation or attribute type or resource.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.3
//		https://tools.ietf.org/html/rfc7644#section-3.4.3
func NewInvalidValuePostError() PostError {
	return PostError{invalidValue}
}

var PutErrorNil PutError

type PutError struct {
	err scimError
}

// If the service provider determines that the creation of the requested resource conflicts with existing resources
// (e.g., a "User" resource with a duplicate "userName"), the service provider MUST return HTTP status code 409
// (Conflict) with a "scimType" error code of "uniqueness".
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.5.1
func NewUniquenessPutError() PutError {
	return PutError{uniqueness}
}

// If the attribute is immutable and one or more values are already set for the attribute, the input value(s) MUST match,
// or HTTP status code 400 SHOULD be returned with a "scimType" error code of "mutability". If the service provider has
// no existing values, the new value(s) SHALL be applied.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.5.1
func NewMutabilityPutError() PutError {
	return PutError{mutability}
}

// Request is unparsable, syntactically incorrect, or violates schema.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.4.3
func NewInvalidSyntaxPutError() PutError {
	return PutError{invalidSyntax}
}

// A required value was missing, or the value specified was not compatible with the operation or attribute type or resource.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.5.1
func NewInvalidValuePutError() PutError {
	return PutError{invalidValue}
}

// NewResourceNotFoundPutError returns an error with status code 404 and a human readable message containing the identifier
// of the resource that was requested to be replaced but not found.
func NewResourceNotFoundPutError(id string) PutError {
	return PutError{resourceNotFound(id)}
}

var DeleteErrorNil DeleteError

type DeleteError struct {
	err scimError
}

// NewResourceNotFoundDeleteError returns an error with status code 404 and a human readable message containing the identifier
// of the resource that was requested to be deleted but not found.
func NewResourceNotFoundDeleteError(id string) DeleteError {
	return DeleteError{resourceNotFound(id)}
}
