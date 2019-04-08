package scim

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestNewSchemaFromString(t *testing.T) {
	cases := []struct {
		s   string
		err string
	}{
		{
			s: `{
				"id": "urn:ietf:params:scim:schemas:core:2.0:User",
				"name": "User",
				"attributes": []
			}`,
			err: "required array is empty",
		},
		{
			s: `{
				"id": "urn:ietf:params:scim:schemas:core:2.0:User",
				"name": "User",
				"attributes": [
					{
						"name": "userName",
			  			"type": "string",
			  			"multiValued": false,
			  			"required": true,
			  			"caseExact": false,
			  			"mutability": "readWrite",
			  			"returned": "default",
			  			"uniqueness": "server"
					}
				]
			}`,
			err: "",
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("invalid schema %d", idx), func(t *testing.T) {
			if _, err := NewSchemaFromString(test.s); err == nil || err.Error() != test.err {
				if err != nil || test.err != "" {
					t.Errorf("expected: %s / got: %v", test.err, err)
				}
			}
		})
	}
}

func TestNewSchemaFromFile(t *testing.T) {
	_, err := NewSchemaFromFile("testdata/simple_user_schema.json")
	if err != nil {
		t.Error(err)
	}

	_, err = NewSchemaFromFile("")
	if err == nil {
		t.Error("expected: no such file or directory")
	}
}

func TestSchemaValidation(t *testing.T) {
	// validate raw meta schema with meta schema
	if err := metaSchema.validate([]byte(rawMetaSchema)); err != nil {
		t.Error(err)
	}

	raw, err := ioutil.ReadFile("testdata/simple_user_schema.json")
	if err != nil {
		t.Error(err)
	}

	// validate simple user schema with meta schema
	if err := metaSchema.validate(raw); err != nil {
		t.Error(err)
	}

	schema, err := NewSchemaFromBytes(raw)
	if err != nil {
		t.Error(err)
	}

	// validate user with simple user schema
	if err := schema.schema.validate([]byte(`{
		"schemas": [
			"schemas"
		],
		"id": "id",
		"userName": "username",
		"externalId": "eid",
		"name": {
			"familyName": "family name",
			"givenName": "given name"
		}
	}`)); err != nil {
		t.Error(err)
	}

	// invalid user
	cases := []struct {
		s   string
		err string
	}{
		{
			s: `{
				"schemas": [
					"schemas"
				],
				"id": "id",
			 	"userName": "username",
			 	"externalId": "eid",
				"name": "name"
			}`,
			err: "cannot convert name to type complex",
		},
		{
			s: `{
				"schemas": [
					"test"
				],
				"id": "test",
			 	"userName": "test",
			 	"externalId": "test",
				"name": {
					"familyName": {}
				}
			}`,
			err: "cannot convert map[] to type string",
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("invalid user %d", idx), func(t *testing.T) {
			if err := schema.schema.validate([]byte(test.s)); err == nil || err.Error() != test.err {
				t.Errorf("expected: %s / got: %v", test.err, err)
			}
		})
	}
}

func TestInvalidJSON(t *testing.T) {
	if err := metaSchema.validate([]byte(``)); err.Error() != "unexpected end of JSON input" {
		t.Errorf("expected: unexpected end of JSON input / got: %v", err)
	}
}

func TestDuplicate(t *testing.T) {
	cases := []struct {
		s   string
		err string
	}{
		{
			s: `{
				"id": "test",
				"ID": "test"
			}`,
			err: "duplicate key: id",
		},
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": [
					{
						"name": "test2",
						"NAME": "test2",
						"type": "string",
						"multiValued": false
					}
				]
			}`,
			err: "duplicate key: name",
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("duplicate %d", idx), func(t *testing.T) {
			if err := metaSchema.validate([]byte(test.s)); err == nil || err.Error() != test.err {
				t.Errorf("expected: %s / got: %v", test.err, err)
			}
		})
	}

	valid := []struct {
		s string
	}{
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": [
					{
						"name": "test2",
						"type": "string",
						"multiValued": false
					}
				],
				"test": "test",
				"TEST": "test"
			}`,
		},
	}

	for idx, test := range valid {
		t.Run(fmt.Sprintf("valid %d", idx), func(t *testing.T) {
			if err := metaSchema.validate([]byte(test.s)); err != nil {
				t.Errorf("no error expected / got: %v", err)
			}
		})
	}
}

func TestRequired(t *testing.T) {
	cases := []struct {
		s   string
		err string
	}{
		{
			s:   `{}`,
			err: "cannot find required value id",
		},
		{
			s: `{
				"id": "test",
				"name": "test"
			}`,
			err: "cannot find required value attributes",
		},
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": [
					{
						"name": "test2"
					}
				]
			}`,
			err: "cannot find required value type",
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("required %d", idx), func(t *testing.T) {
			if err := metaSchema.validate([]byte(test.s)); err == nil || err.Error() != test.err {
				t.Errorf("expected: %s / got: %v", test.err, err)
			}
		})
	}
}

func TestConverting(t *testing.T) {
	cases := []struct {
		s   string
		err string
	}{
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": "test"
			}`,
			err: "cannot convert test to a slice",
		},
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": [
					"test"
				]
			}`,
			err: "cannot convert test to type complex",
		},
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": [
					{
						"name": "test2",
						"type": "string",
						"multiValued": true,
						"canonicalValues": {}
					}
				]
			}`,
			err: "cannot convert map[] to a slice",
		},
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": [
					{
						"name": "test2",
						"type": "string",
						"multiValued": true,
						"canonicalValues": [
							{}
						]
					}
				]
			}`,
			err: "cannot convert map[] to type string",
		},
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": [
					{
						"name": "test2",
						"type": "string",
						"multiValued": "true"
					}
				]
			}`,
			err: "cannot convert true to type boolean",
		},
		{
			s: `{
				"id": {},
				"name": "test",
				"attributes": [
				]
			}`,
			err: "cannot convert map[] to type string",
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("converting %d", idx), func(t *testing.T) {
			if err := metaSchema.validate([]byte(test.s)); err == nil || err.Error() != test.err {
				t.Errorf("expected: %s / got: %v", test.err, err)
			}
		})
	}
}

func TestNil(t *testing.T) {
	cases := []struct {
		s   string
		err string
	}{
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": [
				]
			}`,
			err: "required array is empty",
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("canonical %d", idx), func(t *testing.T) {
			if err := metaSchema.validate([]byte(test.s)); err == nil || err.Error() != test.err {
				t.Errorf("expected: %s / got: %v", test.err, err)
			}
		})
	}
}
