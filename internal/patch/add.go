package patch

import (
	"fmt"
	f "github.com/elimity-com/scim/internal/filter"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"strings"
)

// getRefAttribute returns the corresponding attribute based on the given attribute path.
//
// e.g.
//  - `userName` would return the userNameAttribute.
//	- `name.givenName` would return the `givenName` attribute.
//  - `ext:employeeNumber` would return the `employeeNumber` from the extension.
func (v OperationValidator) getRefAttribute(attrPath filter.AttributePath) (*schema.CoreAttribute, error) {
	// Get the corresponding schema, this can be the main schema or an extension.
	var refSchema = v.schema
	if uri := attrPath.URI(); uri != "" {
		// It can also be an extension if it has a uri prefix.
		var ok bool
		if refSchema, ok = v.schemas[uri]; !ok {
			return nil, fmt.Errorf("invalid uri prefix: %s", uri)
		}
	}

	// Get the correct attribute corresponding to the given attribute path.
	var (
		refAttr  *schema.CoreAttribute
		attrName = attrPath.AttributeName
	)
	for _, attr := range refSchema.Attributes {
		if strings.EqualFold(attr.Name(), attrName) {
			refAttr = &attr
			break
		}
	}
	if refAttr == nil {
		return nil, fmt.Errorf("could not find attribute %s", v.path)
	}
	if subAttrName := attrPath.SubAttributeName(); subAttrName != "" {
		refSubAttr, err := v.getRefSubAttribute(refAttr, subAttrName)
		if err != nil {
			return nil, err
		}
		refAttr = refSubAttr
	}
	return refAttr, nil
}

// getRefSubAttribute returns the sub-attribute of the reference attribute that matches the given subAttrName, if none
// are found it will return an error.
func (v OperationValidator) getRefSubAttribute(refAttr *schema.CoreAttribute, subAttrName string) (*schema.CoreAttribute, error) {
	if !refAttr.HasSubAttributes() {
		return nil, fmt.Errorf("the referred attribute has no sub-attributes: %s", v.path)
	}
	var refSubAttr *schema.CoreAttribute
	for _, attr := range refAttr.SubAttributes() {
		if strings.EqualFold(attr.Name(), subAttrName) {
			refSubAttr = &attr
			break
		}
	}
	if refSubAttr == nil {
		return nil, fmt.Errorf("could not find attribute %s", v.path)
	}
	return refSubAttr, nil
}

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

func (v OperationValidator) validateAddEmptyPath() (interface{}, error) {
	attributes, ok := v.value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("the given value should be a complex attribute if path is empty")
	}

	rootValue := make(map[string]interface{})
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
