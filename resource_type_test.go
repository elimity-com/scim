package scim

import (
	"fmt"
	"testing"
)

func TestNewResourceTypeFromString(t *testing.T) {
	cases := []struct {
		s   string
		err string
	}{
		{
			s: `{
				"name": "User",
				"endpoint": "/Users",
				"schema": "urn:ietf:params:scim:schemas:core:2.0:User"
			}`,
			err: "",
		},
		{
			s: `{
				"id": "User",
				"endpoint": "/Users",
				"schema": "urn:ietf:params:scim:schemas:core:2.0:User"
			}`,
			err: scimErrorInvalidValue.detail,
		},
		{
			s: `{
				"id": "User",
				"name": "User",
				"endpoint": "/Users",
				"description": "User Account",
				"schema": "urn:ietf:params:scim:schemas:core:2.0:User",
				"schemaExtensions": [
					{
						"schema": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
					}
				]
			}`,
			err: scimErrorInvalidValue.detail,
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("invalid schema %d", idx), func(t *testing.T) {
			if _, err := NewResourceTypeFromString(test.s, nil); err == nil || err.Error() != test.err {
				if err != nil || test.err != "" {
					t.Errorf("expected: %s / got: %v", test.err, err)
				}
			}
		})
	}
}

func TestNewResourceTypeFromFile(t *testing.T) {
	_, err := NewResourceTypeFromFile("testdata/simple_user_resource_type.json", nil)
	if err != nil {
		t.Error(err)
	}

	_, err = NewResourceTypeFromFile("", nil)
	if err == nil {
		t.Error("expected: no such file or directory")
	}
}

func TestResourceTypeValidation(t *testing.T) {
	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}

	cases := []struct {
		s   string
		err scimError
	}{
		{
			s:   `true`,
			err: scimErrorInvalidSyntax,
		},
		{
			s: `{
				"id": "test"
			}`,
			err: scimErrorInvalidValue,
		},
		{
			s: `{
				"userName": "other"
			}`,
			err: scimErrorNil,
		},
		{
			s: `{
				"userName": "other",
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
					"organization": "elimity"
				}
			}`,
			err: scimErrorNil,
		},
		{
			s: `{
				"userName": "other",
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": false
			}`,
			err: scimErrorInvalidSyntax,
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("invalid resource %d", idx), func(t *testing.T) {
			if _, err := server.resourceTypes["User"].validate(server.schemas, []byte(test.s), validationConfig{mode: write}); err != test.err {
				t.Errorf("expected: %v / got: %v", test.err, err)
			}
		})
	}

	server.resourceTypes["User"].SchemaExtensions[0].Required = true

	cases = []struct {
		s   string
		err scimError
	}{
		{
			s: `{
				"userName": "other"
			}`,
			err: scimErrorInvalidValue,
		},
		{
			s: `{
				"userName": "other",
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": null
			}`,
			err: scimErrorInvalidValue,
		},
		{
			s: `{
				"userName": "other",
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
					"organization": "elimity"
				}
			}`,
			err: scimErrorNil,
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("invalid resource %d", idx), func(t *testing.T) {
			if _, err := server.resourceTypes["User"].validate(server.schemas, []byte(test.s), validationConfig{mode: write}); err != test.err {
				t.Errorf("expected: %v / got: %v", test.err, err)
			}
		})
	}
}
