package scim

import (
	"encoding/json"

	"github.com/elimity-com/scim/optional"
)

// ServiceProviderConfig enables a service provider to discover SCIM specification features in a standardized form as
// well as provide additional implementation details to clients.
type ServiceProviderConfig struct {
	// DocumentationURI is an HTTP-addressable URL pointing to the service provider's human-consumable help
	// documentation.
	DocumentationURI optional.String
	// AuthenticationSchemes is a multi-valued complex type that specifies supported authentication scheme properties.
	AuthenticationSchemes []AuthenticationScheme
	// ItemsPerPage denotes the maximum and default count on a list request. It defaults to 100.
	ItemsPerPage int
	// SupportsFiltering whether your app supports filtering.
	SupportsFiltering bool
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
	// SpecURI is an HTTP-addressable URL pointing to the authentication scheme's specification.
	SpecURI optional.String
	// DocumentationURI is an HTTP-addressable URL pointing to the authentication scheme's usage documentation.
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
			"supported": false,
		},
		"bulk": map[string]interface{}{
			"supported":      false,
			"maxOperations":  1000,
			"maxPayloadSize": 1048576,
		},
		"filter": map[string]interface{}{
			"supported":  config.SupportsFiltering,
			"maxResults": config.ItemsPerPage,
		},
		"changePassword": map[string]bool{
			"supported": false,
		},
		"sort": map[string]bool{
			"supported": false,
		},
		"etag": map[string]bool{
			"supported": false,
		},
		"authenticationSchemes": getRawAuthSchemes(config.AuthenticationSchemes),
	})
}

// GetItemsPerPage retrieves the configured default count. It falls back to 100 when not configured.
func (config ServiceProviderConfig) GetItemsPerPage() int {
	if config.ItemsPerPage < 1 {
		return fallbackCount
	}

	return config.ItemsPerPage
}

func getRawAuthSchemes(arr []AuthenticationScheme) []map[string]interface{} {
	rawAuthScheme := make([]map[string]interface{}, len(arr))

	for i, auth := range arr {
		rawAuthScheme[i] = auth.Value()
	}

	return rawAuthScheme
}

// Value builds a map based on the values in the AuthenticationScheme.
func (auth AuthenticationScheme) Value() map[string]interface{} {
	return map[string]interface{}{
		"description":      auth.Description,
		"documentationUri": auth.DocumentationURI.Value(),
		"name":             auth.Name,
		"primary":          auth.Primary,
		"specUri":          auth.SpecURI.Value(),
		"type":             auth.Type,
	}
}
