package filter

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser"
)

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, f := range []string{
			"userName",
			"urn:ietf:params:scim:schemas:core:2.0:User:userName",
			"name.givenName",
			"emails",
			"emails.value",
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber",
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager.displayName",
		} {
			path := newPath(f)
			if !ValidatePath(path, schema.CoreUserSchema(), schema.ExtensionEnterpriseUser()) {
				t.Errorf("path should be valid: %s", f)
			}

			exp := newExpression(fmt.Sprintf("%s pr", f))
			if !ValidateExpressionPath(exp, schema.CoreUserSchema(), schema.ExtensionEnterpriseUser()) {
				t.Errorf("filter should be valid: %s pr", f)
			}
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, f := range []string{
			"invalid",
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber",
		} {
			path := newPath(f)
			if ValidatePath(path, schema.CoreUserSchema()) {
				t.Errorf("path should not be valid: %s", f)
			}

			exp := newExpression(fmt.Sprintf("%s pr", f))
			if ValidateExpressionPath(exp, schema.CoreUserSchema()) {
				t.Errorf("filter should not be valid: %s pr", f)
			}
		}

		for _, f := range []string{
			"urn:ietf:params:scim:schemas:core:2.0:User:employeeNumber",
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:userName",
		} {
			path := newPath(f)
			if ValidatePath(path, schema.CoreUserSchema(), schema.ExtensionEnterpriseUser()) {
				t.Errorf("path should not be valid: %s", f)
			}

			exp := newExpression(fmt.Sprintf("%s pr", f))
			if ValidateExpressionPath(exp, schema.CoreUserSchema(), schema.ExtensionEnterpriseUser()) {
				t.Errorf("filter should not be valid: %s pr", f)
			}
		}
	})
}

func TestValidateExpressionPath(t *testing.T) {
	// source example filters:
	// https://tools.ietf.org/html/rfc7644#section-3.4.2.2

	// (enterprise) users
	userSchema := schema.CoreUserSchema()
	userSchema.Attributes = append(userSchema.Attributes, schema.CommonAttributes()...)
	for _, f := range []string{
		"userName Eq \"john\"",
		"Username eq \"john\"",

		"userName eq \"bjensen\"",
		"name.familyName co \"O'Malley\"",
		"userName sw \"J\"",
		"urn:ietf:params:scim:schemas:core:2.0:User:userName sw \"J\"",
		"title pr",
		"meta.lastModified gt \"2011-05-13T04:42:34Z\"",
		"meta.lastModified ge \"2011-05-13T04:42:34Z\"",
		"meta.lastModified lt \"2011-05-13T04:42:34Z\"",
		"meta.lastModified le \"2011-05-13T04:42:34Z\"",
		"title pr and userType eq \"Employee\"",
		"title pr or userType eq \"Intern\"",
		"schemas eq \"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User\"",
		"userType eq \"Employee\" and (emails co \"example.com\" or emails.value co \"example.org\")",
		"userType ne \"Employee\" and not (emails co \"example.com\" or emails.value co \"example.org\")",
		"userType eq \"Employee\" and (emails.type eq \"work\")",
		"userType eq \"Employee\" and emails[type eq \"work\" and value co \"@example.com\"]",
		"emails[type eq \"work\" and value co \"@example.com\"] or ims[type eq \"xmpp\" and value co \"@foo.com\"]",
	} {
		if !ValidateExpressionPath(newExpression(f), userSchema, schema.ExtensionEnterpriseUser()) {
			t.Errorf("path should be valid: %s", f)
		}
	}
}

func TestValidatePath(t *testing.T) {
	// source example filters:
	// https://tools.ietf.org/html/rfc7644#section-3.5.2

	// users
	for _, f := range []string{
		"name.familyName",
		"addresses[type eq \"work\"]",
	} {
		if !ValidatePath(newPath(f), schema.CoreUserSchema()) {
			t.Errorf("path should be valid: %s", f)
		}
	}

	// groups
	for _, f := range []string{
		"members",
		"members[value eq \"2819c223-7f76-453a-919d-413861904646\"]",
		"members[value eq \"2819c223-7f76-453a-919d-413861904646\"].display",
	} {
		if !ValidatePath(newPath(f), schema.CoreGroupSchema()) {
			t.Errorf("path should be valid: %s", f)
		}
	}
}

func newExpression(f string) filter.Expression {
	parser := filter.NewParser(strings.NewReader(f))
	exp, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	return exp
}

func newPath(f string) filter.Path {
	parser := filter.NewParser(strings.NewReader(f))
	path, err := parser.ParsePath()
	if err != nil {
		log.Fatal(err)
	}

	return path
}
