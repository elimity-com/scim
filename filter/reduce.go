package filter

import (
	"fmt"

	filter "github.com/di-wu/scim-filter-parser"
	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schema"
)

// IsValid checks whether given resources passed the filter.
func (f Filter) IsValid(resource map[string]interface{}) (bool, *errors.ScimError) {
	switch exp := f.Expression.(type) {
	case filter.AttributeExpression:
		return f.isValidAttributeExpression(resource)
	case filter.ValuePath:
		return f.isValidValuePath(resource)
	case filter.UnaryExpression:
		ok, err := Filter{
			Expression: exp.X,
			schema:     f.schema,
			extensions: f.extensions,
		}.IsValid(resource)
		if err != nil {
			return false, err
		}
		if !ok {
			return true, nil
		}
		return false, nil
	case filter.BinaryExpression:
		switch exp.CompareOperator {
		case filter.AND:
			ok, err := Filter{
				Expression: exp.X,
				schema:     f.schema,
				extensions: f.extensions,
			}.IsValid(resource)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
			ok, err = Filter{
				Expression: exp.Y,
				schema:     f.schema,
				extensions: f.extensions,
			}.IsValid(resource)
			if err != nil {
				return false, err
			}
			return ok, nil
		case filter.OR:
			ok1, err := Filter{
				Expression: exp.X,
				schema:     f.schema,
				extensions: f.extensions,
			}.IsValid(resource)
			if err != nil {
				return false, err
			}
			ok2, err := Filter{
				Expression: exp.Y,
				schema:     f.schema,
				extensions: f.extensions,
			}.IsValid(resource)
			if err != nil {
				return false, err
			}
			return ok1 || ok2, nil
		default:
			return false, unknownOperatorError(exp.CompareOperator, exp)
		}
	case nil:
		return true, nil // failsafe if filter does not contain an expression
	default:
		return false, unknownExpressionTypeError(exp)
	}
}

// Reduce removes all resources that do not pass the filter.
func (f Filter) Reduce(resources []map[string]interface{}) ([]map[string]interface{}, error) {
	var idx int
	for i := 0; i < len(resources); i++ {
		r := resources[i]
		ok, err := f.IsValid(r)
		if err != nil {
			return nil, err
		}
		if ok {
			resources[i] = r
			idx++
		}
	}
	return resources[:idx], nil
}

func (f Filter) containsAttribute(uriPrefix, attrName string) (schema.Schema, schema.CoreAttribute, bool) {
	attr, ok := f.schema.Attributes.ContainsAttribute(attrName)
	if ok {
		return f.schema, attr, true
	}

	if uriPrefix != "" {
		for _, e := range f.extensions {
			if uriPrefix != e.ID {
				continue
			}

			attr, ok := e.Attributes.ContainsAttribute(attrName)
			if ok {
				return e, attr, true
			}
		}
	}

	return schema.Schema{}, schema.CoreAttribute{}, false
}

func (f Filter) isValidAttributeExpression(resource map[string]interface{}) (bool, *errors.ScimError) {
	path := f.Expression.(filter.AttributeExpression).AttributePath

	var refSchema schema.Schema
	var refAttr, refSubAttr schema.CoreAttribute
	var ok bool
	refSchema, refAttr, ok = f.containsAttribute(path.URIPrefix, path.AttributeName)
	if !ok {
		return false, invalidFilterError(fmt.Sprintf("invalid attribute name %q", path.AttributeName))
	}

	if path.SubAttribute != "" {
		if !refAttr.HasSubAttributes() {
			return false, invalidFilterError(
				fmt.Sprintf(
					"attribute %q has no sub attribute %q",
					path.AttributeName, path.SubAttribute,
				),
			)
		}
		refSubAttr, ok = refAttr.SubAttributes().ContainsAttribute(path.SubAttribute)
		if !ok {
			return false, invalidFilterError(fmt.Sprintf("invalid sub attribute name %q", path.SubAttribute))
		}
	}

	attr := refAttr
	if path.SubAttribute != "" {
		attr = refSubAttr
	}
	comp, err := f.reduceAttributeExpression(attr)
	if err != nil {
		return false, invalidFilterError(err.Error())
	}

	value, ok := resource[refAttr.Name()]
	if !ok {
		value, ok = resource[fmt.Sprintf("%s:%s", refSchema.ID, refAttr.Name())]
		if !ok {
			return false, nil
		}
	}
	if path.SubAttribute != "" {
		subAttr, ok := value.(map[string]interface{})
		if !ok {
			return false, nil
		}
		value, ok = subAttr[refSubAttr.Name()]
		if !ok {
			return false, nil
		}
	}

	if attr.MultiValued() {
		var values []string
		switch value := value.(type) {
		case []interface{}:
			for _, v := range value {
				str, ok := v.(string)
				if !ok {
					return false, nil
				}
				values = append(values, str)
			}
		case []string:
			values = value
		default:
			return false, nil
		}

		for _, v := range values {
			if comp(v) {
				return true, nil
			}
		}
		return false, nil
	}
	return comp(value), nil
}

func (f Filter) isValidValuePath(resource map[string]interface{}) (bool, *errors.ScimError) {
	exp := f.Expression.(filter.ValuePath)

	refSchema, refAttr, ok := f.containsAttribute(exp.URIPrefix, exp.AttributeName)
	if !ok {
		return false, invalidFilterError(fmt.Sprintf("invalid attribute name %q", exp.AttributeName))
	}

	if !refAttr.MultiValued() {
		return false, invalidFilterError(fmt.Sprintf("attribute %q is not multi valued", exp.AttributeName))
	}

	iAttrs, ok := resource[refAttr.Name()]
	if !ok {
		iAttrs, ok = resource[fmt.Sprintf("%s:%s", refSchema.ID, refAttr.Name())]
		if !ok {
			return false, nil
		}
	}

	var attrs []map[string]interface{}
	switch iAttrs := iAttrs.(type) {
	case []interface{}:
		for _, a := range iAttrs {
			attr, ok := a.(map[string]interface{})
			if !ok {
				return false, nil
			}
			attrs = append(attrs, attr)
		}
	case []map[string]interface{}:
		attrs = iAttrs
	default:
		return false, nil
	}

	valueFilter := Filter{
		Expression: exp.ValueExpression,
		schema: schema.Schema{
			Attributes: refAttr.SubAttributes(),
		},
	}

	for _, attr := range attrs {
		ok, err := valueFilter.IsValid(attr)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (f Filter) reduceAttributeExpression(attr schema.CoreAttribute) (func(interface{}) bool, error) {
	switch exp := f.Expression.(filter.AttributeExpression); exp.CompareOperator {
	case filter.EQ:
		return eq(exp.CompareValue, attr), nil
	case filter.NE:
		return func(i interface{}) bool {
			return !eq(exp.CompareValue, attr)(i)
		}, nil
	case filter.CO:
		return co(exp.CompareValue, attr), nil
	case filter.SW:
		return sw(exp.CompareValue, attr), nil
	case filter.EW:
		return ew(exp.CompareValue, attr), nil
	case filter.PR:
		return func(value interface{}) bool {
			return value != nil
		}, nil
	case filter.GT:
		return c(exp.CompareValue, attr, func(i int) bool {
			return i > 0
		})
	case filter.GE:
		return c(exp.CompareValue, attr, func(i int) bool {
			return i >= 0
		})
	case filter.LT:
		return c(exp.CompareValue, attr, func(i int) bool {
			return i < 0
		})
	case filter.LE:
		return c(exp.CompareValue, attr, func(i int) bool {
			return i <= 0
		})
	default:
		return func(value interface{}) bool {
			return false
		}, unknownOperatorError(exp.CompareOperator, exp)
	}
}
