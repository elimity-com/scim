package scim

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSchemaValidation(t *testing.T) {
	// validate raw meta schema with meta schema
	if err := metaSchema.validate([]byte(rawMetaSchema)); err != nil {
		t.Errorf(err.Error())
	}

	s := `{
  		"id": "urn:ietf:params:scim:schemas:core:2.0:User",
  		"name": "User",
  		"description": "User Account",
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
    		},
    		{
      			"name": "name",
      			"type": "complex",
      			"multiValued": false,
      			"required": false,
      			"subAttributes": [
        			{
          				"name": "familyName",
          				"type": "string",
          				"multiValued": false,
          				"required": false,
          				"caseExact": false,
          				"mutability": "readWrite",
          				"returned": "default",
          				"uniqueness": "none"
        			},
        			{
          				"name": "givenName",
          				"type": "string",
          				"multiValued": false,
          				"required": false,
          				"caseExact": false,
          				"mutability": "readWrite",
          				"returned": "default",
          				"uniqueness": "none"
        			}
      			],
      			"mutability": "readWrite",
      			"returned": "default",
      			"uniqueness": "none"
    		},
    		{
				"name": "displayName",
      			"type": "string",
      			"multiValued": false,
      			"required": false,
      			"caseExact": false,
      			"mutability": "readWrite",
      			"returned": "default",
      			"uniqueness": "none"
    		},
    		{
      			"name": "emails",
      			"type": "complex",
      			"multiValued": true,
      			"required": false,
      			"subAttributes": [
      				{
      					"name": "value",
      					"type": "string",
      					"multiValued": false,
      					"required": false,
      					"caseExact": false,
      					"mutability": "readWrite",
      					"returned": "default",
      					"uniqueness": "none"
      				},
      				{
      					"name": "display",
      					"type": "string",
      					"multiValued": false,
      					"required": false,
      					"caseExact": false,
      					"mutability": "readWrite",
      					"returned": "default",
      					"uniqueness": "none"
      				},
      				{
      					"name": "type",
      					"type": "string",
      					"multiValued": false,
      					"required": false,
      					"caseExact": false,
      					"canonicalValues": [
      						"work",
      						"home",
      						"other"
      					],
      					"mutability": "readWrite",
      					"returned": "default",
      					"uniqueness": "none"
      				},
      				{
      					"name": "primary",
      					"type": "boolean",
      					"multiValued": false,
      					"required": false,
      					"mutability": "readWrite",
      					"returned": "default"
      				}
      			],
      			"mutability": "readWrite",
      			"returned": "default",
      			"uniqueness": "none"
    		}
  		],
  		"meta": {
    		"resourceType": "Schema",
    		"location": "/v2/Schemas/urn:ietf:params:scim:schemas:core:2.0:User"
  		}
	}`

	// validate simple user schema with meta schema
	if err := metaSchema.validate([]byte(s)); err != nil {
		t.Errorf(err.Error())
	}

	var schema schema
	if err := json.Unmarshal([]byte(s), &schema); err != nil {
		t.Errorf(err.Error())
	}

	// validate user with simple user schema
	if err := schema.validate([]byte(`{
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
		t.Errorf(err.Error())
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
			if err := schema.validate([]byte(test.s)); err == nil || err.Error() != test.err {
				t.Errorf("expected: %s / got: %v", test.err, err)
			}
		})
	}
}

func TestInvalidJSON(t *testing.T) {
	cases := []struct {
		s   string
		err string
	}{
		{
			s:   ``,
			err: "unexpected end of JSON input",
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("json %d", idx), func(t *testing.T) {
			if err := metaSchema.validate([]byte(test.s)); err == nil || err.Error() != test.err {
				t.Errorf("expected: %s / got: %v", test.err, err)
			}
		})
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
