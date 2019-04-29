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
			err: scimErrorInvalidValue.detail,
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
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("invalid schema %d", idx), func(t *testing.T) {
			_, err := NewSchemaFromString(test.s)
			if err != nil && err.Error() != test.err {
				t.Errorf("expected: %s / got: %v", test.err, err)
			} else if err == nil && test.err != "" {
				t.Errorf("no error expected")
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
		t.Error("error expected")
	}
}

func TestSchemaValidation(t *testing.T) {
	// validate raw meta schema with meta schema
	if _, err := metaSchema.validate([]byte(rawSchemaSchema), read); err != scimErrorNil {
		t.Error(err)
	}

	raw, err := ioutil.ReadFile("testdata/simple_user_schema.json")
	if err != nil {
		t.Error(err)
	}

	// validate simple user schema with meta schema
	if _, err := metaSchema.validate(raw, read); err != scimErrorNil {
		t.Error(err)
	}

	schema, err := NewSchemaFromBytes(raw)
	if err != nil {
		t.Error(err)
	}

	// validate user with simple user schema
	if _, err := schema.schema.validate([]byte(`{
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
	}`), read); err != scimErrorNil {
		t.Error(err)
	}

	// invalid user
	cases := []struct {
		s   string
		err scimError
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
			err: scimErrorInvalidSyntax,
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
			err: scimErrorInvalidValue,
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("invalid user %d", idx), func(t *testing.T) {
			if _, err := schema.schema.validate([]byte(test.s), read); err != test.err {
				t.Errorf("expected: %v / got: %v", test.err, err)
			}
		})
	}
}

func TestInvalidJSON(t *testing.T) {
	if _, err := metaSchema.validate([]byte(``), read); err != scimErrorInvalidSyntax {
		t.Errorf("invalid error: %v", err)
	}
}

func TestDuplicate(t *testing.T) {
	cases := []struct {
		s   string
		err scimError
	}{
		{
			s: `{
				"id": "test",
				"ID": "test"
			}`,
			err: scimErrorUniqueness,
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
			err: scimErrorUniqueness,
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("duplicate %d", idx), func(t *testing.T) {
			if _, err := metaSchema.validate([]byte(test.s), read); err != test.err {
				t.Errorf("expected: %v / got: %v", test.err, err)
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
			if _, err := metaSchema.validate([]byte(test.s), read); err != scimErrorNil {
				t.Errorf("no error expected / got: %v", err)
			}
		})
	}
}

func TestRequired(t *testing.T) {
	cases := []struct {
		s   string
		err scimError
	}{
		{
			s:   `{}`,
			err: scimErrorInvalidValue,
		},
		{
			s: `{
				"id": "test",
				"name": "test"
			}`,
			err: scimErrorInvalidValue,
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
			err: scimErrorInvalidValue,
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("required %d", idx), func(t *testing.T) {
			if _, err := metaSchema.validate([]byte(test.s), read); err != test.err {
				t.Errorf("expected: %v / got: %v", test.err, err)
			}
		})
	}
}

func TestConverting(t *testing.T) {
	cases := []struct {
		s   string
		err scimError
	}{
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": "test"
			}`,
			err: scimErrorInvalidSyntax,
		},
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": [
					"test"
				]
			}`,
			err: scimErrorInvalidSyntax,
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
			err: scimErrorInvalidSyntax,
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
			err: scimErrorInvalidValue,
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
			err: scimErrorInvalidValue,
		},
		{
			s: `{
				"id": {},
				"name": "test",
				"attributes": [
				]
			}`,
			err: scimErrorInvalidValue,
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("converting %d", idx), func(t *testing.T) {
			if _, err := metaSchema.validate([]byte(test.s), read); err != test.err {
				t.Errorf("expected: %v / got: %v", test.err, err)
			}
		})
	}
}

func TestNil(t *testing.T) {
	cases := []struct {
		s   string
		err scimError
	}{
		{
			s: `{
				"id": "test",
				"name": "test",
				"attributes": [
				]
			}`,
			err: scimErrorInvalidValue,
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("canonical %d", idx), func(t *testing.T) {
			if _, err := metaSchema.validate([]byte(test.s), read); err != test.err {
				t.Errorf("expected: %v / got: %v", test.err, err)
			}
		})
	}
}

func TestValidationModeRead(t *testing.T) {
	raw, err := ioutil.ReadFile("testdata/simple_user_schema.json")
	if err != nil {
		t.Error(err)
	}

	attributes, scimErr := metaSchema.validate(raw, read)
	if scimErr != scimErrorNil {
		t.Error(scimErr)
	}

	if len(attributes) != 0 {
		t.Errorf("no attributes exprected")
	}
}

func TestValidationModeWrite(t *testing.T) {
	raw, err := ioutil.ReadFile("testdata/simple_user_schema.json")
	if err != nil {
		t.Error(err)
	}

	attributes, scimErr := metaSchema.validate(raw, write)
	if scimErr != scimErrorNil {
		t.Error(scimErr)
	}

	if len(attributes) == 0 {
		t.Errorf("no attributes exprected")
	}
}

func TestValidationModeReplace(t *testing.T) {
	raw, err := ioutil.ReadFile("testdata/simple_user_schema.json")
	if err != nil {
		t.Error(err)
	}

	attributes, scimErr := metaSchema.validate(raw, replace)
	if scimErr != scimErrorNil {
		t.Error(scimErr)
	}

	if len(attributes) == 0 {
		t.Errorf("no attributes exprected")
	}
}
