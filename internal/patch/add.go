package patch

import (
	"fmt"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"strings"
)

func (v OperationValidator) ValidateAdd() error {
	// The operation must contain a "value" member whose content specifies the value to be added.
	if v.value == nil {
		return fmt.Errorf("an add operation must contain a value member")
	}

	// If "path" is omitted, the target location is assumed to be the resource itself.
	if v.path == nil {
		return v.validateAddEmptyPath()
	}

	if v.path.ValueExpression != nil || v.path.SubAttribute != nil {
		return fmt.Errorf("an add operation does not support value expressions")
	}

	var (
		refSchema = v.schema
		attrPath  = v.path.AttributePath
		attrName  = attrPath.AttributeName
	)
	if uri := attrPath.URI(); uri != "" {
		// It can only be an extension if it has a uri prefix.
		var ok bool
		if refSchema, ok = v.schemas[uri]; !ok {
			return fmt.Errorf("invalid uri prefix: %s", uri)
		}
	}

	var refAttr *schema.CoreAttribute
	for _, attr := range refSchema.Attributes {
		if strings.EqualFold(attr.Name(), attrName) {
			refAttr = &attr
			break
		}
	}
	if refAttr == nil {
		return fmt.Errorf("could not find attribute %s", v.path)
	}
	if subAttrName := attrPath.SubAttributeName(); subAttrName != "" {
		if !refAttr.HasSubAttributes() {
			return fmt.Errorf("the referred attribute has no sub-attributes: %s", v.path)
		}

		var refSubAttr *schema.CoreAttribute
		for _, attr := range refAttr.SubAttributes() {
			if strings.EqualFold(attr.Name(), subAttrName) {
				refSubAttr = &attr
				break
			}
		}
		if refSubAttr == nil {
			return fmt.Errorf("could not find attribute %s", v.path)
		}
		refAttr = refSubAttr
	}

	if refAttr.MultiValued() {
		if list, ok := v.value.([]interface{}); !ok {
			attr, err := refAttr.ValidateSingular(v.value)
			if err != nil {
				return err
			}
			v.value = []interface{}{attr}
		} else {
			var attrs []interface{}
			for _, value := range list {
				attr, err := refAttr.ValidateSingular(value)
				if err != nil {
					return err
				}
				attrs = append(attrs, attr)
			}
			v.value = attrs
		}
	} else {
		attr, err := refAttr.ValidateSingular(v.value)
		if err != nil {
			return err
		}
		v.value = attr
	}
	return nil
}

func (v OperationValidator) validateAddEmptyPath() error {
	attributes, ok := v.value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("the given value should be a complex attribute if path is empty")
	}

	for p, value := range attributes {
		path, err := filter.ParseAttrPath([]byte(p))
		if err != nil || path.SubAttribute != nil {
			return fmt.Errorf("invalid attribute path: %s", p)
		}
		refSchema := v.schema
		if uri := path.URI(); uri != "" {
			// It can only be an extension if it has a uri prefix.
			var ok bool
			if refSchema, ok = v.schemas[uri]; !ok {
				return fmt.Errorf("invalid uri prefix: %s", uri)
			}
		}

		var refAttr *schema.CoreAttribute
		for _, attr := range refSchema.Attributes {
			if strings.EqualFold(attr.Name(), path.AttributeName) {
				refAttr = &attr
				break
			}
		}
		if refAttr == nil {
			return fmt.Errorf("could not find attribute %s", path)
		}

		if refAttr.MultiValued() {
			if list, ok := value.([]interface{}); !ok {
				attr, err := refAttr.ValidateSingular(value)
				if err != nil {
					return err
				}
				v.value = []interface{}{attr}
			} else {
				var attrs []interface{}
				for _, value := range list {
					attr, err := refAttr.ValidateSingular(value)
					if err != nil {
						return err
					}
					attrs = append(attrs, attr)
				}
				v.value = attrs
			}
		} else {
			attr, err := refAttr.ValidateSingular(value)
			if err != nil {
				return err
			}
			v.value = attr
		}
	}
	return nil
}
