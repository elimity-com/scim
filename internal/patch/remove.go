package patch

import (
	"net/http"

	"github.com/elimity-com/scim/errors"
	f "github.com/elimity-com/scim/internal/filter"
	"github.com/elimity-com/scim/schema"
)

// validateRemove validates the remove operation contained within the validator based on on Section 3.5.2.2 in RFC 7644.
// More info: https://datatracker.ietf.org/doc/html/rfc7644#section-3.5.2.2
func (v OperationValidator) validateRemove() (interface{}, error) {
	// If "path" is unspecified, the operation fails with HTTP status code 400 and a "scimType" error code of "noTarget".
	if v.Path == nil {
		return nil, &errors.ScimError{
			ScimType: errors.ScimTypeNoTarget,
			Status:   http.StatusBadRequest,
		}
	}

	refAttr, err := v.getRefAttribute(v.Path.AttributePath)
	if err != nil {
		return nil, err
	}
	if v.Path.ValueExpression != nil {
		if err := f.NewFilterValidator(v.Path.ValueExpression, schema.Schema{
			Attributes: f.MultiValuedFilterAttributes(*refAttr),
		}).Validate(); err != nil {
			return nil, err
		}
	}
	if subAttrName := v.Path.SubAttributeName(); subAttrName != "" {
		refSubAttr, err := v.getRefSubAttribute(refAttr, subAttrName)
		if err != nil {
			return nil, err
		}
		refAttr = refSubAttr
	}
	if v.value == nil {
		return nil, nil
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
