package filter

import (
	"fmt"
	"net/http"
	"strings"

	filter "github.com/di-wu/scim-filter-parser"
	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schema"
)

func invalidFilterError(msg string) *errors.ScimError {
	return &errors.ScimError{
		ScimType: errors.ScimTypeInvalidFilter,
		Detail:   msg,
		Status:   http.StatusBadRequest,
	}
}

func unknownOperatorError(token filter.Token, exp filter.Expression) *errors.ScimError {
	return invalidFilterError(fmt.Sprintf("unknown operator in expression: %s %s", token, exp))
}

func unknownExpressionTypeError(exp filter.Expression) *errors.ScimError {
	return invalidFilterError(fmt.Sprintf("unknown expression type: %s", exp))
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
		}, nil
	}
	return Filter{}, nil
}

// GetPathFilter returns a parsed filter path.
func GetPathFilter(path string) (filter.Path, error) {
	if path != "" {
		parser := filter.NewParser(strings.NewReader(path))
		return parser.ParsePath()
	}
	return filter.Path{}, nil
}

// ValidatePath returns whether the path can be found in one of the given schemas.
func ValidatePath(path filter.Path, s schema.Schema, extensions ...schema.Schema) bool {
	ok1 := validateAttributeNames(path.URIPrefix, path.AttributeName, path.SubAttribute, s)
	var ok2 bool
	for _, e := range extensions {
		if validateAttributeNames(path.URIPrefix, path.AttributeName, path.SubAttribute, e) {
			ok2 = true
			break
		}
	}

	fmt.Println(path, ok1, ok2)

	if !ok1 && !ok2 {
		return false
	}

	return ValidateExpressionPath(path.ValueExpression, s, extensions...)
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

// Filter represents a parsed SCIM Filter.
type Filter struct {
	filter.Expression
	schema schema.Schema
}

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

func validateExpressionPath(exp filter.Expression, s schema.Schema) bool {
	switch exp := exp.(type) {
	case filter.AttributeExpression:
		attrPath := exp.AttributePath
		return validateAttributeNames(attrPath.URIPrefix, attrPath.AttributeName, attrPath.SubAttribute, s)
	case filter.ValuePath:
		if !validateAttributeNames("", exp.AttributeName, "", s) {
			return false
		}
		return validateExpressionPath(exp.ValueExpression, s)
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
