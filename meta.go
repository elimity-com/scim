package scim

// meta is a complex attribute containing resource metadata. All "meta" sub-attributes are assigned by the service
// provider (have a "mutability" of "readOnly"), and all of these sub-attributes have a "returned" characteristic of
// "default". This attribute SHALL be ignored when provided by clients.
//
// RFC: https://tools.ietf.org/html/rfc7643#section-3.1
type meta struct {
	// ResourceType is the name of the resource type of the resource.
	ResourceType string
	// Created is the "DateTime" that the resource was added to the service provider.
	Created string `json:",omitempty"`
	// LastModified is the most recent DateTime that the details of this resource were updated at the service provider.
	// If this resource has never been modified since its initial creation, the value MUST be the same as the value of
	// "created".
	LastModified string `json:",omitempty"`
	// Location is the URI of the resource being returned. This value MUST be the same as the "Content-Location" HTTP
	// response header.
	Location string
	// Version is the version of the resource being returned.  This value must be the same as the entity-tag (ETag) HTTP
	// response header.
	Version string `json:",omitempty"`
}
