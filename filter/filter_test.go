package filter

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/di-wu/scim-filter-parser"
	"github.com/elimity-com/scim/schema"
)

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
