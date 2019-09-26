package scim

import (
	"encoding/json"

	"github.com/elimity-com/scim/optional"
)

// ServiceProviderConfig enables a service provider to discover SCIM specification features in a standardized form as
// well as provide additional implementation details to clients.
type ServiceProviderConfig struct {
	// DocumentationURI is an HTTP-addressable URL pointing to the service provider's human-consumable help
	// documentation. OPTIONAL.
	DocumentationURI optional.String
	// PatchSupported is a boolean value specifying whether or not PATCH is supported.
	PatchSupported bool
	// BulkSupported is a boolean value specifying whether or not bulk is supported.
	BulkSupported bool
	// MaxBulkOperations is an integer value specifying the maximum number of bulk operations.
	MaxBulkOperations int
	// MaxBulkPayloadSize is an integer value specifying the maximum bulk payload size in bytes.
	MaxBulkPayloadSize int
	// FilterSupported is a boolean value specifying whether or not FILTER is supported.
	FilterSupported bool
	// MaxFilterResults is an integer value specifying the maximum number of resources returned in a filter response.
	MaxFilterResults int
	// ChangePasswordSupported is a boolean value specifying whether or not changing a password is supported.
	ChangePasswordSupported bool
	// SortSupported is a boolean value specifying whether or not sorting is supported.
	SortSupported bool
	// ETagSupported is a boolean value specifying whether or not ETag is supported.
	ETagSupported bool
	// AuthenticationSchemes is a multi-valued complex type that specifies supported authentication scheme properties.
	AuthenticationSchemes []AuthenticationScheme
}

// AuthenticationScheme specifies a supported authentication scheme property.
type AuthenticationScheme struct {
	// Type is the authentication scheme. This specification defines the values "oauth", "oauth2", "oauthbearertoken",
	// "httpbasic", and "httpdigest".
	Type AuthenticationType
	// Name is the common authentication scheme name, e.g., HTTP Basic.
	Name string
	// Description of the authentication scheme.
	Description string
	// SpecURI is an HTTP-addressable URL pointing to the authentication scheme's specification. OPTIONAL.
	SpecURI optional.String
	// DocumentationURI is an HTTP-addressable URL pointing to the authentication scheme's usage documentation. OPTIONAL.
	DocumentationURI optional.String
	// Primary is a boolean value indicating the 'primary' or preferred authentication scheme.
	Primary bool
}

// AuthenticationType is a single keyword indicating the authentication type of the authentication scheme.
type AuthenticationType string

const (
	// AuthenticationTypeOauth indicates that the authentication type is OAuth.
	AuthenticationTypeOauth AuthenticationType = "oauth"
	// AuthenticationTypeOauth2 indicates that the authentication type is OAuth2.
	AuthenticationTypeOauth2 AuthenticationType = "oauth2"
	// AuthenticationTypeOauthBearerToken indicates that the authentication type is OAuth2 Bearer Token.
	AuthenticationTypeOauthBearerToken AuthenticationType = "oauthbearertoken"
	// AuthenticationTypeHTTPBasic indicated that the authentication type is Basic Access Authentication.
	AuthenticationTypeHTTPBasic AuthenticationType = "httpbasic"
	// AuthenticationTypeHTTPDigest indicated that the authentication type is Digest Access Authentication.
	AuthenticationTypeHTTPDigest AuthenticationType = "httpdigest"
)

// MarshalJSON converts the service provider config struct to its corresponding json representation.
func (config ServiceProviderConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"schemas":          []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"},
		"documentationUri": config.DocumentationURI.Value(),
		"patch": map[string]bool{
			"supported": config.PatchSupported,
		},
		"bulk": map[string]interface{}{
			"supported":      config.BulkSupported,
			"maxOperations":  config.MaxBulkOperations,
			"maxPayloadSize": config.MaxBulkPayloadSize,
		},
		"filter": map[string]interface{}{
			"supported":  config.FilterSupported,
			"maxResults": config.MaxFilterResults,
		},
		"changePassword": map[string]bool{
			"supported": config.ChangePasswordSupported,
		},
		"sort": map[string]bool{
			"supported": config.SortSupported,
		},
		"etag": map[string]bool{
			"supported": config.ETagSupported,
		},
		"authenticationSchemes": config.AuthenticationSchemes,
	})
}
