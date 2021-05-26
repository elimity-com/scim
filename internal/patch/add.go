package patch

import (
	"fmt"
	f "github.com/elimity-com/scim/internal/filter"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
)

// validateAdd validates the add operation contained within the validator based on on Section 3.5.2.1 in RFC 7644.
// More info: https://datatracker.ietf.org/doc/html/rfc7644#section-3.5.2.1
func (v OperationValidator) validateAdd() (interface{}, error) {
	// The operation must contain a "value" member whose content specifies the value to be added.
	if v.value == nil {
		return nil, fmt.Errorf("an add operation must contain a value member")
	}

	// If "path" is omitted, the target location is assumed to be the resource itself.
	if v.path == nil {
		return v.validateAddEmptyPath()
	}

	refAttr, err := v.getRefAttribute(v.path.AttributePath)
	if err != nil {
		return nil, err
	}
	if v.path.ValueExpression != nil {
		if err := f.NewFilterValidator(v.path.ValueExpression, schema.Schema{
			Attributes: refAttr.SubAttributes(),
		}).Validate(); err != nil {
			return nil, err
		}
	}
	if subAttrName := v.path.SubAttributeName(); subAttrName != "" {
		refSubAttr, err := v.getRefSubAttribute(refAttr, subAttrName)
		if err != nil {
			return nil, err
		}
		refAttr = refSubAttr
	}

	if !refAttr.MultiValued() {
		attr, scimErr := refAttr.ValidateSingular(v.value)
		if scimErr != nil {
			return nil, scimErr
		}
		return attr, nil
	}

	if list, ok := v.value.([]interface{}); ok {
		var attrs []interface{}
		for _, value := range list {
			attr, scimErr := refAttr.ValidateSingular(value)
			if scimErr != nil {
				return nil, scimErr
			}
			attrs = append(attrs, attr)
		}
		return attrs, nil
	}

	attr, scimErr := refAttr.ValidateSingular(v.value)
	if scimErr != nil {
		return nil, scimErr
	}
	return []interface{}{attr}, nil
}

// validateAddEmptyPath validates paths that don't have a "path" value. In this case the target location is assumed to
// be the resource itself. The "value" parameter contains a set of attributes to be added to the resource.
func (v OperationValidator) validateAddEmptyPath() (interface{}, error) {
	attributes, ok := v.value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("the given value should be a complex attribute if path is empty")
	}

	rootValue := map[string]interface{}{}
	for p, value := range attributes {
		path, err := filter.ParsePath([]byte(p))
		if err != nil {
			return nil, fmt.Errorf("invalid attribute path: %s", p)
		}
		validator := OperationValidator{
			op:      v.op,
			path:    &path,
			value:   value,
			schema:  v.schema,
			schemas: v.schemas,
		}
		v, err := validator.validateAdd()
		if err != nil {
			return nil, err
		}
		rootValue[p] = v
	}
	return rootValue, nil
}
