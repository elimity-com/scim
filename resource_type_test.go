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
			err: "cannot find required value name",
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
			err: "cannot find required value required",
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
