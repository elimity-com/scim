package filter

import (
	"fmt"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"log"
)

// validateAttributePath checks whether the given attribute path is a valid path within the given reference schema.
func validateAttributePath(ref schema.Schema, attrPath filter.AttributePath) (schema.CoreAttribute, error) {
	if uri := attrPath.URI(); uri != "" && uri != ref.ID {
		return schema.CoreAttribute{}, fmt.Errorf("the uri does not match the schema id: %s", uri)
	}

	attr, ok := ref.Attributes.ContainsAttribute(attrPath.AttributeName)
	if !ok {
		return schema.CoreAttribute{}, fmt.Errorf(
			"the reference schema does not have an attribute with the name: %s",
			attrPath.AttributeName,
		)
	}
	// e.g. name.givenName
	//           ^________
	if subAttrName := attrPath.SubAttributeName(); subAttrName != "" {
		if err := validateSubAttribute(attr, subAttrName); err != nil {
			return schema.CoreAttribute{}, err
		}
	}
	return attr, nil
}

// validateExpression checks whether the given expression is a valid expression within the given reference schema.
func validateExpression(ref schema.Schema, e filter.Expression) error {
	switch e := e.(type) {
	case *filter.ValuePath:
		attr, err := validateAttributePath(ref, e.AttributePath)
		if err != nil {
			return nil
		}
		if err := validateExpression(
			schema.Schema{
				ID:         ref.ID,
				Attributes: attr.SubAttributes(),
			},
			e.ValueFilter,
		); err != nil {
			return err
		}
	case *filter.AttributeExpression:
		if _, err := validateAttributePath(ref, e.AttributePath); err != nil {
			return err
		}
	case *filter.LogicalExpression:
		if err := validateExpression(ref, e.Left); err != nil {
			return err
		}
		if err := validateExpression(ref, e.Right); err != nil {
			return err
		}
	case *filter.NotExpression:
		if err := validateExpression(ref, e.Expression); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown expression type: %v", e)
	}
	return nil
}

// validateSubAttribute checks whether the given attribute name is a attribute within the given reference attribute.
func validateSubAttribute(attr schema.CoreAttribute, subAttrName string) error {
	if !attr.HasSubAttributes() {
		return fmt.Errorf("the attribute has no sub-attributes")
	}

	if _, ok := attr.SubAttributes().ContainsAttribute(subAttrName); !ok {
		return fmt.Errorf("the attribute has no sub-attributes named: %s", subAttrName)
	}
	return nil
}

// PathValidator represents a path validator.
type PathValidator struct {
	path       filter.Path
	schema     schema.Schema
	extensions []schema.Schema
}

// NewPathValidator constructs a new path validator.
func NewPathValidator(p filter.Path, s schema.Schema, exts ...schema.Schema) PathValidator {
	return PathValidator{
		path:       p,
		schema:     s,
		extensions: exts,
	}
}

// Validate checks whether the path is a valid path within the given reference schemas.
func (v PathValidator) Validate() error {
	err := v.validatePath(v.schema)
	if err == nil {
		return nil
	}
	for _, e := range v.extensions {
		if err := v.validatePath(e); err == nil {
			return nil
		}
	}
	return err
}

// validatePath tries to validate the path against the given schema.
func (v PathValidator) validatePath(ref schema.Schema) error {
	// e.g. members
	//      ^______
	attr, err := validateAttributePath(ref, v.path.AttributePath)
	if err != nil {
		return err
	}

	// e.g. members[value eq "0"]
	//			   ^_____________
	if v.path.ValueExpression != nil {
		if err := validateExpression(
			schema.Schema{
				ID:         ref.ID,
				Attributes: attr.SubAttributes(),
			},
			v.path.ValueExpression,
		); err != nil {
			return err
		}
	}

	// e.g. members[value eq "0"].displayName
	//			                  ^__________
	if subAttrName := v.path.SubAttributeName(); subAttrName != "" {
		if err := validateSubAttribute(attr, subAttrName); err != nil {
			return err
		}
	}
	return nil
}

// Validator represents a filter validator.
type Validator struct {
	filter     filter.Expression
	schema     schema.Schema
	extensions []schema.Schema
}

// NewValidator constructs a new path validator.
func NewValidator(e filter.Expression, s schema.Schema, exts ...schema.Schema) Validator {
	return Validator{
		filter:     e,
		schema:     s,
		extensions: exts,
	}
}

// PassesFilter checks whether given resources passes the filter.
func (v Validator) PassesFilter(resource map[string]interface{}) bool {
	switch e := v.filter.(type) {
	case *filter.ValuePath:
		ref, attr, ok := v.referenceContains(e.AttributePath)
		if !ok {
			// Could not find an attribute that matches the attribute path.
			return false
		}
		if !attr.MultiValued() {
			// Value path filters can only be applied to multi-valued attributes.
			return false
		}

		value, ok := resource[attr.Name()]
		if !ok {
			// Also try with the id as prefix.
			value, ok = resource[fmt.Sprintf("%s:%s", ref.ID, attr.Name())]
			if !ok {
				// The give resource does not have the wanted attribute.
				return false
			}
		}
		valueFilter := NewValidator(
			e.ValueFilter,
			schema.Schema{
				ID:         ref.ID,
				Attributes: attr.SubAttributes(),
			},
		)
		switch value := value.(type) {
		case []interface{}:
			for _, a := range value {
				attr, ok := a.(map[string]interface{})
				if !ok {
					return false
				}
				if valueFilter.PassesFilter(attr) {
					return true
				}
			}
		}
		return false
	case *filter.AttributeExpression:
		ref, attr, ok := v.referenceContains(e.AttributePath)
		if !ok {
			// Could not find an attribute that matches the attribute path.
			return false
		}

		value, ok := resource[attr.Name()]
		if !ok {
			// Also try with the id as prefix.
			value, ok = resource[fmt.Sprintf("%s:%s", ref.ID, attr.Name())]
			if !ok {
				// The give resource does not have the wanted attribute.
				return false
			}
		}

		var (
			// cmpAttr will be the attribute to validate the filter against.
			cmpAttr = attr

			subAttr     schema.CoreAttribute
			subAttrName = e.AttributePath.SubAttributeName()
		)

		if subAttrName != "" {
			if !attr.HasSubAttributes() {
				// The attribute has no sub-attributes.
				return false
			}
			subAttr, ok = attr.SubAttributes().ContainsAttribute(subAttrName)
			if !ok {
				return false
			}

			attr, ok := value.(map[string]interface{})
			if !ok {
				return false
			}
			value, ok = attr[subAttr.Name()]
			if !ok {
				return false
			}

			cmpAttr = subAttr
		}

		// If the attribute has a non-empty or non-null value or if it contains a non-empty node for complex attributes, there is a match.
		if e.Operator == filter.PR {
			// We already found a value.
			return true
		}

		cmp, err := createCompareFunction(e, cmpAttr)
		if err != nil {
			// TODO replace booleans w/ errors.
			log.Println(err)
			return false
		}

		if !attr.MultiValued() {
			return cmp(value)
		}

		switch value := value.(type) {
		case []interface{}:
			for _, v := range value {
				if cmp(v) {
					return true
				}
			}
		}
	case *filter.LogicalExpression:
		switch e.Operator {
		case filter.AND:
			if !NewValidator(
				e.Left,
				v.schema,
				v.extensions...,
			).PassesFilter(resource) {
				return false
			}
			return NewValidator(
				e.Right,
				v.schema,
				v.extensions...,
			).PassesFilter(resource)
		case filter.OR:
			if NewValidator(
				e.Left,
				v.schema,
				v.extensions...,
			).PassesFilter(resource) {
				return true
			}
			return NewValidator(
				e.Right,
				v.schema,
				v.extensions...,
			).PassesFilter(resource)
		}
	case *filter.NotExpression:
		if !NewValidator(
			e.Expression,
			v.schema,
			v.extensions...,
		).PassesFilter(resource) {
			return true
		}
	}
	return false
}

// Validate checks whether the expression is a valid path within the given reference schemas.
func (v Validator) Validate() error {
	err := validateExpression(v.schema, v.filter)
	if err == nil {
		return nil
	}
	for _, e := range v.extensions {
		if err := validateExpression(e, v.filter); err == nil {
			return nil
		}
	}
	return err
}

// referenceContains returns the schema and attribute to which the attribute path applies.
func (v Validator) referenceContains(attrPath filter.AttributePath) (schema.Schema, schema.CoreAttribute, bool) {
	for _, s := range append([]schema.Schema{v.schema}, v.extensions...) {
		if uri := attrPath.URI(); uri != "" && s.ID != uri {
			continue
		}
		if attr, ok := s.Attributes.ContainsAttribute(attrPath.AttributeName); ok {
			return s, attr, true
		}
	}
	return schema.Schema{}, schema.CoreAttribute{}, false
}
