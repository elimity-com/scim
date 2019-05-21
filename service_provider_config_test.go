package scim

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestNewServiceProviderConfigFromString(t *testing.T) {
	cases := []struct {
		s   string
		err string
	}{
		{
			s: `{
				"documentationUri": "http://example.com/help/scim.html",
				"patch": {
					"supported": true
				},
				"bulk": {
					"supported": true,
					"maxOperations": 1000,
					"maxPayloadSize": 1048576
				},
				"filter": {
					"supported": true,
					"maxResults": 200
				},
				"changePassword": {
					"supported": true
				},
				"sort": {
					"supported": true
				},
				"etag": {
					"supported": true
				},
				"authenticationSchemes": [
					{
						"name": "OAuth Bearer Token",
						"description": "Authentication scheme using the OAuth Bearer Token Standard",
						"specUri": "http://www.rfc-editor.org/info/rfc6750",
						"documentationUri": "http://example.com/help/oauth.html",
						"type": "oauthbearertoken"
					}
				]
			}`,
			err: "",
		},
		{
			s: `{
				"documentationUri": "http://example.com/help/scim.html",
				"patch": {
					"supported": true
				},
				"bulk": {
					"supported": true,
					"maxOperations": 1.0
				}
			}`,
			err: scimErrorInvalidValue.detail,
		},
		{
			s: `{
				"documentationUri": "http://example.com/help/scim.html",
				"patch": {
					"supported": true
				},
				"bulk": {
					"supported": true,
					"maxOperations": "one"
				}
			}`,
			err: scimErrorInvalidValue.detail,
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("invalid schema %d", idx), func(t *testing.T) {
			if _, err := NewServiceProviderConfig([]byte(test.s)); err == nil || err.Error() != test.err {
				if err != nil || test.err != "" {
					t.Errorf("expected: %s / got: %v", test.err, err)
				}
			}
		})
	}
}

func TestNewServiceProviderConfigFromFile(t *testing.T) {
	rawConfig, err := ioutil.ReadFile("testdata/simple_service_provider_config.json")
	if err != nil {
		t.Error(err)
	}
	_, err = NewServiceProviderConfig(rawConfig)
	if err != nil {
		t.Error(err)
	}

	_, err = NewServiceProviderConfig([]byte(""))
	if err == nil {
		t.Error("expected: no such file or directory")
	}
}

func TestServiceProviderConfigUnmarshalJSON(t *testing.T) {
	var config serviceProviderConfig
	if err := config.UnmarshalJSON([]byte(``)); err == nil {
		t.Error("error expected")
	}
}
