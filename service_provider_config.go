package scim

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

// NewServiceProviderConfigFromFile reads the file from given filepath and returns a validated service provider config
// if no errors take place.
func NewServiceProviderConfigFromFile(filepath string) (ServiceProviderConfig, error) {
	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return ServiceProviderConfig{}, err
	}

	return NewServiceProviderConfigFromBytes(raw)
}

// NewServiceProviderConfigFromString returns a validated service provider config if no errors take place.
func NewServiceProviderConfigFromString(s string) (ServiceProviderConfig, error) {
	return NewServiceProviderConfigFromBytes([]byte(s))
}

// NewServiceProviderConfigFromBytes returns a validated service provider config if no errors take place.
func NewServiceProviderConfigFromBytes(raw []byte) (ServiceProviderConfig, error) {
	_, scimErr := serviceProviderConfigSchema.validate(raw, read)
	if scimErr != scimErrorNil {
		return ServiceProviderConfig{}, fmt.Errorf(scimErr.Detail)
	}

	var serviceProviderConfig serviceProviderConfig
	err := json.Unmarshal(raw, &serviceProviderConfig)
	if err != nil {
		log.Fatalf("failed parsing service provider config: %v", err)
	}

	return ServiceProviderConfig{serviceProviderConfig}, nil
}

// ServiceProviderConfig enables a service provider to discover SCIM specification features in a standardized form as
// well as provide additional implementation details to clients.
type ServiceProviderConfig struct {
	config serviceProviderConfig
}

// serviceProviderConfig enables a service provider to discover SCIM specification features in a standardized form as
// well as provide additional implementation details to clients.
//
// RFC: https://tools.ietf.org/html/rfc7643#section-5
type serviceProviderConfig struct {
	// DocumentationURI is an HTTP-addressable URL pointing to the service provider's human-consumable help
	// documentation. OPTIONAL.
	DocumentationURI *string `json:",omitempty"`
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
	AuthenticationSchemes []authenticationScheme
}

// RFC: https://tools.ietf.org/html/rfc7643#section-8.5
func (config serviceProviderConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"schemas":          []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"},
		"documentationUri": config.DocumentationURI,
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

func (config *serviceProviderConfig) UnmarshalJSON(data []byte) error {
	var tmpConfig struct {
		DocumentationURI *string
		Patch            struct {
			Supported bool
		}
		Bulk struct {
			Supported      bool
			MaxOperations  int
			MaxPayloadSize int
		}
		Filter struct {
			Supported bool
		}
		ChangePassword struct {
			Supported bool
		}
		Sort struct {
			Supported bool
		}
		ETag struct {
			Supported bool
		}
		AuthenticationSchemes []authenticationScheme
	}

	err := json.Unmarshal(data, &tmpConfig)
	if err != nil {
		return err
	}

	*config = serviceProviderConfig{
		DocumentationURI:        tmpConfig.DocumentationURI,
		PatchSupported:          tmpConfig.Patch.Supported,
		BulkSupported:           tmpConfig.Bulk.Supported,
		MaxBulkOperations:       tmpConfig.Bulk.MaxOperations,
		MaxBulkPayloadSize:      tmpConfig.Bulk.MaxPayloadSize,
		FilterSupported:         tmpConfig.Filter.Supported,
		ChangePasswordSupported: tmpConfig.ChangePassword.Supported,
		SortSupported:           tmpConfig.Sort.Supported,
		ETagSupported:           tmpConfig.ETag.Supported,
		AuthenticationSchemes:   tmpConfig.AuthenticationSchemes,
	}

	return nil
}

// authenticationScheme specifies a supported authentication scheme property.
type authenticationScheme struct {
	// Type is the authentication scheme. This specification defines the values "oauth", "oauth2", "oauthbearertoken",
	// "httpbasic", and "httpdigest".
	Type authenticationType
	// Name is the common authentication scheme name, e.g., HTTP Basic.
	Name string
	// Description of the authentication scheme.
	Description string
	// SpecURI is an HTTP-addressable URL pointing to the authentication scheme's specification. OPTIONAL.
	SpecURI *string `json:",omitempty"`
	// DocumentationURI is an HTTP-addressable URL pointing to the authentication scheme's usage documentation. OPTIONAL.
	DocumentationURI *string `json:",omitempty"`
	// Primary is a boolean value indicating the 'primary' or preferred authentication scheme.
	Primary *bool `json:",omitempty"`
}

type authenticationType string

// TODO: authentication types
// const (
// authenticationTypeOauth            authenticationType = "oauth"
// authenticationTypeOauth2           authenticationType = "oauth2"
// authenticationTypeOauthBearerToken authenticationType = "oauthbearertoken"
// authenticationTypeHTTPBasic        authenticationType = "httpbasic"
// authenticationTypeHTTPDigest       authenticationType = "httpdigest"
// )

var serviceProviderConfigSchema schema

func init() {
	if err := json.Unmarshal([]byte(rawServiceProviderConfigSchema), &serviceProviderConfigSchema); err != nil {
		panic(err)
	}
}
