package errors

// GetError represents an error that is returned by a GET HTTP request.
type GetError int

const (
	// GetErrorNil indicates that no error occurred during handling a GET HTTP request.
	GetErrorNil GetError = iota
	// GetErrorResourceNotFound returns an error with status code 404 and a human readable message containing the identifier
	// of the resource that was requested but not found.
	GetErrorResourceNotFound
)

// PatchError represents an error that is returned by a PATCH HTTP request.
type PatchError int

const (
	// PatchErrorNil indicates that no error occurred during handling a PUT HTTP request.
	PatchErrorNil PatchError = iota
	// PatchErrorUniqueness shall be returned when one or more of the attribute values are already in use or are reserved.
	PatchErrorUniqueness
	// PatchErrorMutability shall be returned when the attempted modification is not compatible with the target
	// attribute's mutability or current state.
	PatchErrorMutability
	// PatchErrorResourceNotFound returns an error with status code 404 and a human readable message containing the
	// identifier of the resource that was requested to be replaced but not found.
	PatchErrorResourceNotFound
	// PatchErrorNotImplemented allows consumers to create a patch handler that simply returns an unsupported error.
	PatchErrorNotImplemented
)

// PostError represents an error that is returned by a POST HTTP request.
type PostError int

const (
	// PostErrorNil indicates that no error occurred during handling a POST HTTP request.
	PostErrorNil PostError = iota
	// PostErrorUniqueness shall be returned when one or more of the attribute values are already in use or are reserved.
	PostErrorUniqueness
)

// PutError represents an error that is returned by a PUT HTTP request.
type PutError int

const (
	// PutErrorNil indicates that no error occurred during handling a PUT HTTP request.
	PutErrorNil PutError = iota
	// PutErrorUniqueness shall be returned when one or more of the attribute values are already in use or are reserved.
	PutErrorUniqueness
	// PutErrorMutability shall be returned when the attempted modification is not compatible with the target
	// attribute's mutability or current state.
	PutErrorMutability
	// PutErrorResourceNotFound returns an error with status code 404 and a human readable message containing the
	// identifier of the resource that was requested to be replaced but not found.
	PutErrorResourceNotFound
)

// DeleteError represents an error that is returned by a DELETE HTTP request.
type DeleteError int

const (
	// DeleteErrorNil indicates that no error occurred during handling a DELETE HTTP request.
	DeleteErrorNil DeleteError = iota
	// DeleteErrorResourceNotFound returns an error with status code 404 and a human readable message containing the
	// identifier of the resource that was requested to be deleted but not found.
	DeleteErrorResourceNotFound
)

// ValidationError represents an error that is returned during a resource validation.
type ValidationError int

const (
	// ValidationErrorNil indicates that no error occurred during a resource validation.
	ValidationErrorNil ValidationError = iota
	// ValidationErrorInvalidSyntax indicates that the request body message structure was invalid or did not conform to
	// the request schema.
	ValidationErrorInvalidSyntax
	// ValidationErrorInvalidValue indicates that a required value was missing or the value specified was not
	// compatible with the operation, attribute type or resource schema.
	ValidationErrorInvalidValue
)
