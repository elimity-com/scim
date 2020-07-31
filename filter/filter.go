package filter

import (
	"fmt"
	"net/http"
	"strings"

	filter "github.com/di-wu/scim-filter-parser"
	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schema"
)

// GetPathFilter returns a parsed filter path.
func GetPathFilter(path string) (filter.Path, error) {
	if path != "" {
		parser := filter.NewParser(strings.NewReader(path))
		return parser.ParsePath()
	}
	return filter.Path{}, nil
}

// ValidateExpressionPath returns whether the path in the expression can be found in one of the given schemas.
func ValidateExpressionPath(exp filter.Expression, s schema.Schema, extensions ...schema.Schema) bool {
	if validateExpressionPath(exp, s) {
		return true
	}

	for _, e := range extensions {
		if validateExpressionPath(exp, e) {
			return true
		}
	}
	return false
}

// ValidatePath returns whether the path can be found in one of the given schemas.
func ValidatePath(path filter.Path, s schema.Schema, extensions ...schema.Schema) bool {
	// get correct reference schema matching path
	var ok bool
	var refSchema schema.Schema
	if ok = validateAttributeNames(path.URIPrefix, path.AttributeName, path.SubAttribute, s); ok {
		refSchema = s
	} else if path.URIPrefix != "" { // extensions must have a path URI prefix
		for _, e := range extensions {
			if validateAttributeNames(path.URIPrefix, path.AttributeName, path.SubAttribute, e) {
				ok = true
				refSchema = e
				break
			}
		}
	}

	// no reference schema found, not in main schemas as extensions.
	if !ok {
		return false
	}

	// path has no value filter expression
	if path.ValueExpression == nil {
		return true
	}

	attr, _ := s.Attributes.ContainsAttribute(path.AttributeName)
	if !attr.HasSubAttributes() {
		return false
	}
	return validateExpressionPath(path.ValueExpression, schema.Schema{
		Attributes: attr.SubAttributes(),
		ID:         refSchema.ID,
	})
}

func invalidFilterError(msg string) *errors.ScimError {
	return &errors.ScimError{
		ScimType: errors.ScimTypeInvalidFilter,
		Detail:   msg,
		Status:   http.StatusBadRequest,
	}
}

func unknownExpressionTypeError(exp filter.Expression) *errors.ScimError {
	return invalidFilterError(fmt.Sprintf("unknown expression type: %s", exp))
}

func unknownOperatorError(token filter.Token, exp filter.Expression) *errors.ScimError {
	return invalidFilterError(fmt.Sprintf("unknown operator in expression: %s %s", token, exp))
}

func validateAttributeNames(uriPrefix, attrName, subAttrName string, s schema.Schema) bool {
	if uriPrefix != "" && uriPrefix != s.ID {
		return false
	}

	attr, ok := s.Attributes.ContainsAttribute(attrName)
	if !ok {
		return false
	}

	if subAttrName != "" {
		if !attr.HasSubAttributes() {
			return false
		}

		if _, ok := attr.SubAttributes().ContainsAttribute(subAttrName); !ok {
			return false
		}
		return true
	}
	return true
}

func validateExpressionPath(exp filter.Expression, s schema.Schema) bool {
	switch exp := exp.(type) {
	case filter.AttributeExpression:
		attrPath := exp.AttributePath
		return validateAttributeNames(attrPath.URIPrefix, attrPath.AttributeName, attrPath.SubAttribute, s)
	case filter.ValuePath:
		if !validateAttributeNames("", exp.AttributeName, "", s) {
			return false
		}

		attr, _ := s.Attributes.ContainsAttribute(exp.AttributeName)
		if !attr.HasSubAttributes() {
			return false
		}

		return validateExpressionPath(exp.ValueExpression, schema.Schema{
			Attributes: attr.SubAttributes(),
			ID:         s.ID,
		})
	case filter.BinaryExpression:
		return validateExpressionPath(exp.X, s) && validateExpressionPath(exp.Y, s)
	case filter.UnaryExpression:
		return validateExpressionPath(exp.X, s)
	case nil:
		return true
	default:
		return false
	}
}

// Filter represents a parsed SCIM Filter.
type Filter struct {
	filter.Expression
	schema     schema.Schema
	extensions []schema.Schema
}

// NewFilter get the filter from the request, parses it and adds reference schemas to validate against.
func NewFilter(r *http.Request, s schema.Schema, extensions ...schema.Schema) (Filter, error) {
	rawFilter := strings.TrimSpace(r.URL.Query().Get("filter"))
	if rawFilter != "" {
		parser := filter.NewParser(strings.NewReader(rawFilter))
		exp, err := parser.Parse()
		if err != nil {
			return Filter{}, err
		}
		
		return Filter{
			Expression: exp,
			schema:     s,
			extensions: extensions,
		}, nil
	}
	return Filter{}, nil
}
