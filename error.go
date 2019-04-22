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

func scimErrorResourceNotFound(id string) scimError {
	return scimError{
		Detail: fmt.Sprintf("Resource %s not found.", id),
		Status: http.StatusNotFound,
	}
}

var scimErrorNil scimError

var (
	scimErrorUniqueness = scimError{
		ScimType: scimTypeUniqueness,
		Detail:   "One or more of the attribute values are already in use or are reserved.",
		Status:   http.StatusConflict,
	}
	scimErrorMutability = scimError{
		ScimType: scimTypeMutability,
		Detail:   "The attempted modification is not compatible with the target attribute's mutability or current state.",
		Status:   http.StatusBadRequest,
	}
	scimErrorInvalidSyntax = scimError{
		ScimType: scimTypeInvalidSyntax,
		Detail:   "The request body message structure was invalid or did not conform to the request schema.",
		Status:   http.StatusBadRequest,
	}
	scimErrorInvalidValue = scimError{
		ScimType: scimTypeInvalidValue,
		Detail:   "A required value was missing, or the value specified was not compatible with the operation or attribute type, or resource schema.",
		Status:   http.StatusBadRequest,
	}
	scimErrorInternalServer = scimError{
		Status: http.StatusInternalServerError,
	}
)

// RFC: https://tools.ietf.org/html/rfc7644#section-3.12
type scimError struct {
	// scimType is a SCIM detail error keyword. OPTIONAL.
	ScimType scimType
	// Detail is a detailed human-readable message. OPTIONAL.
	Detail string
	// status is the HTTP status code expressed as a JSON string. REQUIRED.
	Status int
}

func (e scimError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Schemas  []string
		ScimType scimType `json:",omitempty"`
		Detail   string   `json:",omitempty"`
		Status   string
	}{
		Schemas:  []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		ScimType: e.ScimType,
		Detail:   e.Detail,
		Status:   strconv.Itoa(e.Status),
	})
}

func (e *scimError) UnmarshalJSON(data []byte) error {
	var tmpScimError struct {
		ScimType scimType
		Detail   string
		Status   string
	}

	err := json.Unmarshal(data, &tmpScimError)
	if err != nil {
		return err
	}

	status, err := strconv.Atoi(tmpScimError.Status)
	if err != nil {
		return err
	}

	*e = scimError{
		ScimType: tmpScimError.ScimType,
		Detail:   tmpScimError.Detail,
		Status:   status,
	}

	return nil
}

// GetError represents an error that is returned by a GET HTTP request.
type GetError struct {
	err scimError
}

// GetErrorNil indicates that no error occurred during handling a GET HTTP request.
var GetErrorNil GetError

var (
	// GetErrorInvalidValue shall be returned when a required field is missing or a value is not compatible with the
	// attribute type.
	GetErrorInvalidValue = GetError{err: scimErrorInvalidValue}
)

// NewResourceNotFoundGetError returns an error with status code 404 and a human readable message containing the identifier
// of the resource that was requested but not found.
func NewResourceNotFoundGetError(id string) GetError {
	return GetError{scimErrorResourceNotFound(id)}
}

// PostError represents an error that is returned by a POST HTTP request.
type PostError struct {
	err scimError
}

// PostErrorNil indicates that no error occurred during handling a POST HTTP request.
var PostErrorNil PostError

var (
	// PostErrorUniqueness shall be returned when one or more of the attribute values are already in use or are reserved.
	PostErrorUniqueness = PostError{err: scimErrorUniqueness}
	// PostErrorInvalidSyntax shall be returned when the request body message structure was invalid or did not conform
	// to the request schema.
	PostErrorInvalidSyntax = PostError{err: scimErrorInvalidSyntax}
	// PostErrorInvalidValue shall be returned when a required field is missing or a value is not compatible with the
	// attribute type.
	PostErrorInvalidValue = PostError{err: scimErrorInvalidValue}
)

// PutError represents an error that is returned by a PUT HTTP request.
type PutError struct {
	err scimError
}

// PutErrorNil indicates that no error occurred during handling a PUT HTTP request.
var PutErrorNil PutError

var (
	// PutErrorUniqueness shall be returned when one or more of the attribute values are already in use or are reserved.
	PutErrorUniqueness = PutError{err: scimErrorUniqueness}
	// PutErrorMutability shall be returned when the attempted modification is not compatible with the target
	// attribute's mutability or current state.
	PutErrorMutability = PutError{err: scimErrorMutability}
	// PutErrorInvalidSyntax shall be returned when the request body message structure was invalid or did not conform
	// to the request schema.
	PutErrorInvalidSyntax = PutError{err: scimErrorInvalidSyntax}
	// PutErrorInvalidValue shall be returned when a required field is missing or a value is not compatible with the
	// attribute type.
	PutErrorInvalidValue = PutError{err: scimErrorInvalidValue}
)

// NewResourceNotFoundPutError returns an error with status code 404 and a human readable message containing the identifier
// of the resource that was requested to be replaced but not found.
func NewResourceNotFoundPutError(id string) PutError {
	return PutError{scimErrorResourceNotFound(id)}
}

// DeleteError represents an error that is returned by a DELETE HTTP request.
type DeleteError struct {
	err scimError
}

// DeleteErrorNil indicates that no error occurred during handling a DELETE HTTP request.
var DeleteErrorNil DeleteError

// NewResourceNotFoundDeleteError returns an error with status code 404 and a human readable message containing the identifier
// of the resource that was requested to be deleted but not found.
func NewResourceNotFoundDeleteError(id string) DeleteError {
	return DeleteError{scimErrorResourceNotFound(id)}
}
