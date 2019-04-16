package scim

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

func (e scimError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		schemas  []string
		scimType scimType
		detail   string
		status   string
	}{
		schemas:  []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		scimType: e.scimType,
		detail:   e.detail,
		status:   strconv.Itoa(e.status),
	})
}

// GetErrorNil indicates that no error occurred during handling a GET HTTP request.
var GetErrorNil GetError

// GetError represents an error that is returned by a GET HTTP request.
type GetError struct {
	err scimError
}

var (
	// InvalidValueGetError shall be returned when a required field is missing or a value is not compatible with the
	// attribute type.
	InvalidValueGetError = GetError{invalidValue}
)

// NewResourceNotFoundGetError returns an error with status code 404 and a human readable message containing the identifier
// of the resource that was requested but not found.
func NewResourceNotFoundGetError(id string) GetError {
	return GetError{resourceNotFound(id)}
}

// PostErrorNil indicates that no error occurred during handling a POST HTTP request.
var PostErrorNil PostError

// PostError represents an error that is returned by a POST HTTP request.
type PostError struct {
	err scimError
}

var (
	// UniquenessPostError shall be returned when one or more of the attribute values are already in use or are reserved.
	UniquenessPostError = PostError{uniqueness}
	// InvalidSyntaxPostError shall be returned when the request body message structure was invalid or did not conform
	// to the request schema.
	InvalidSyntaxPostError = PostError{invalidSyntax}
	// InvalidValuePostError shall be returned when a required field is missing or a value is not compatible with the
	// attribute type.
	InvalidValuePostError = PostError{invalidValue}
)

// PutErrorNil indicates that no error occurred during handling a PUT HTTP request.
var PutErrorNil PutError

// PutError represents an error that is returned by a PUT HTTP request.
type PutError struct {
	err scimError
}

var (
	// UniquenessPutError shall be returned when one or more of the attribute values are already in use or are reserved.
	UniquenessPutError = PutError{uniqueness}
	// MutabilityPutError shall be returned when the attempted modification is not compatible with the target
	// attribute's mutability or current state.
	MutabilityPutError = PutError{mutability}
	// InvalidSyntaxPutError shall be returned when the request body message structure was invalid or did not conform
	// to the request schema.
	InvalidSyntaxPutError = PutError{invalidSyntax}
	// InvalidValuePutError shall be returned when a required field is missing or a value is not compatible with the
	// attribute type.
	InvalidValuePutError = PutError{invalidValue}
)

// NewResourceNotFoundPutError returns an error with status code 404 and a human readable message containing the identifier
// of the resource that was requested to be replaced but not found.
func NewResourceNotFoundPutError(id string) PutError {
	return PutError{resourceNotFound(id)}
}

// DeleteErrorNil indicates that no error occurred during handling a DELETE HTTP request.
var DeleteErrorNil DeleteError

// DeleteError represents an error that is returned by a DELETE HTTP request.
type DeleteError struct {
	err scimError
}

// NewResourceNotFoundDeleteError returns an error with status code 404 and a human readable message containing the identifier
// of the resource that was requested to be deleted but not found.
func NewResourceNotFoundDeleteError(id string) DeleteError {
	return DeleteError{resourceNotFound(id)}
}
