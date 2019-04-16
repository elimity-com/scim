package scim

import (
	"fmt"
	"testing"
)

func TestNewServer(t *testing.T) {
	config, _ := NewServiceProviderConfigFromFile("testdata/simple_service_provider_config.json")
	cases := []struct {
		s   []Schema
		t   []ResourceType
		err string
	}{
		{
			s: []Schema{
				{
					schema: schema{
						ID: "id",
					},
				},
				{
					schema: schema{
						ID: "id",
					},
				},
			},
			t: []ResourceType{
				{
					resourceType: resourceType{
						Name:   "name",
						Schema: "id",
					},
				},
			},
			err: "duplicate schema with id: id",
		},
		{
			s: []Schema{},
			t: []ResourceType{
				{
					resourceType: resourceType{
						Name:   "name",
						Schema: "id",
					},
				},
			},
			err: "schemas does not contain a schema with id: id, referenced by resource type: name",
		},
		{
			s: []Schema{
				{
					schema: schema{
						ID: "id",
					},
				},
			},
			t: []ResourceType{
				{
					resourceType: resourceType{
						Name:   "name",
						Schema: "id",
					},
				},
				{
					resourceType: resourceType{
						Name:   "name",
						Schema: "id",
					},
				},
			},
			err: "duplicate resource type with name: name",
		},
		{
			s: []Schema{
				{
					schema: schema{
						ID: "id",
					},
				},
			},
			t: []ResourceType{
				{
					resourceType: resourceType{
						Name:   "name",
						Schema: "id",
						SchemaExtensions: []schemaExtension{
							{
								Schema: "other",
							},
						},
					},
				},
			},
			err: "schemas does not contain a schema with id: other, referenced by resource type extension with index: 0",
		},
		{
			s: []Schema{
				{
					schema: schema{
						ID: "id",
					},
				},
			},
			t: []ResourceType{
				{
					resourceType: resourceType{
						Name:     "name",
						Endpoint: "/",
						Schema:   "id",
					},
				},
				{
					resourceType: resourceType{
						Name:     "other",
						Endpoint: "/",
						Schema:   "id",
					},
				},
			},
			err: "duplicate endpoints in resource types: /",
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("invalid schema %d", idx), func(t *testing.T) {
			if _, err := NewServer(config, test.s, test.t); err == nil || err.Error() != test.err {
				if err != nil || test.err != "" {
					t.Errorf("expected: %s / got: %v", test.err, err)
				}
			}
		})
	}
}
