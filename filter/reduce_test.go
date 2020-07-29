package filter

import (
	"log"
	"strings"
	"testing"

	filter "github.com/di-wu/scim-filter-parser"
	"github.com/elimity-com/scim/schema"
)

func newFilter(f string) Filter {
	parser := filter.NewParser(strings.NewReader(f))
	exp, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	userSchema := schema.CoreUserSchema()
	userSchema.Attributes = append(userSchema.Attributes, schema.CommonAttributes()...)

	return Filter{
		Expression: exp,
		schema:     userSchema,
	}
}

func testResources() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"schemas": []string{
				"urn:ietf:params:scim:schemas:core:2.0:User",
			},
			"userName": "di-wu",
			"userType": "admin",
			"name": map[string]interface{}{
				"familyName": "di",
				"givenName":  "wu",
			},
			"emails": []map[string]interface{}{
				{
					"value": "quint@elimity.com",
					"type":  "work",
				},
			},
			"meta": map[string]interface{}{
				"lastModified": "2020-07-26T20:02:34Z",
			},
		},
		{
			"schemas": []interface{}{
				"urn:ietf:params:scim:schemas:core:2.0:User",
			},
			"userName": "interface",
			"emails": []interface{}{
				map[string]interface{}{
					"value": "noreply@elimity.com",
					"type":  "work",
				},
			},
		},
		{
			"userName": "admin",
			"userType": "admin",
			"name": map[string]interface{}{
				"familyName": "ad",
				"givenName":  "min",
			},
		},
		{"userName": "guest"},
		{
			"userName": "unknown",
			"name": map[string]interface{}{
				"familyName": "un",
				"givenName":  "known",
			},
		},
		{"userName": "another"},
	}
}

func TestReduce(t *testing.T) {
	for _, test := range []struct {
		name   string
		len    int
		filter string
	}{
		{name: "eq", len: 1, filter: "userName eq \"di-wu\""},
		{name: "ne", len: 5, filter: "userName ne \"di-wu\""},
		{name: "co", len: 3, filter: "userName co \"u\""},
		{name: "co", len: 2, filter: "name.familyName co \"d\""},
		{name: "sw", len: 2, filter: "userName sw \"a\""},
		{name: "sw", len: 2, filter: "urn:ietf:params:scim:schemas:core:2.0:User:userName sw \"a\""},
		{name: "ew", len: 2, filter: "userName ew \"n\""},
		{name: "pr", len: 6, filter: "userName pr"},
		{name: "gt", len: 2, filter: "userName gt \"guest\""},
		{name: "ge", len: 3, filter: "userName ge \"guest\""},
		{name: "lt", len: 3, filter: "userName lt \"guest\""},
		{name: "le", len: 4, filter: "userName le \"guest\""},
		{name: "value", len: 2, filter: "emails[type eq \"work\"]"},
		{name: "and", len: 1, filter: "name.familyName eq \"ad\" and userType eq \"admin\""},
		{name: "or", len: 2, filter: "name.familyName eq \"ad\" or userType eq \"admin\""},
		{name: "not", len: 5, filter: "not userName eq \"di-wu\""},
		{name: "meta", len: 1, filter: "meta.lastModified gt \"2011-05-13T04:42:34Z\""},
		{name: "schemas", len: 2, filter: "schemas eq \"urn:ietf:params:scim:schemas:core:2.0:User\""},
	} {
		t.Run(test.name, func(t *testing.T) {
			resources, err := newFilter(test.filter).Reduce(testResources())
			if err != nil {
				t.Errorf("no error expected, got %s", err)
			}
			if len(resources) != test.len {
				t.Errorf("expected %d resources, got %d\n%s", test.len, len(resources), resources)
			}
		})
	}
}

func TestReduceErrors(t *testing.T) {
	for _, test := range []struct {
		err    string
		filter string
	}{
		{
			err:    "invalid attribute name \"none\"",
			filter: "none eq \"di-wu\"",
		},
		{
			err:    "attribute \"username\" has no sub attribute \"none\"",
			filter: "userName.none eq \"di-wu\"",
		},
		{
			err:    "invalid sub attribute name \"none\"",
			filter: "name.none eq \"di-wu\"",
		},
		{
			err:    "invalid attribute type \"binary\"",
			filter: "x509Certificates.value gt \"0\"",
		},
		{
			err:    "invalid attribute type \"boolean\"",
			filter: "active gt \"true\"",
		},
	} {
		resources, err := newFilter(test.filter).Reduce(testResources())
		if err == nil {
			t.Errorf("expected error, got none")
			continue
		}
		if !strings.HasSuffix(err.Error(), test.err) {
			t.Errorf("expected %s, got %s", test.err, err.Error())
		}
		if resources != nil {
			t.Errorf("expected nil, got %d resources", len(resources))
		}
	}
}
