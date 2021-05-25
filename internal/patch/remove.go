package patch

import (
	"github.com/elimity-com/scim/errors"
	f "github.com/elimity-com/scim/internal/filter"
	"github.com/elimity-com/scim/schema"
	"net/http"
)

// multiValuedFilterAttributes returns the attributes of the given attribute on which can be filtered. In the case of a
// complex attribute, the sub-attributes get returned. Otherwise if the given attribute is not complex, a "value" sub-
// attribute gets created to filter against.
func multiValuedFilterAttributes(attr *schema.CoreAttribute) schema.Attributes {
	switch attr.AttributeType() {
	case "decimal":
		return schema.Attributes{
			schema.SimpleCoreAttribute(schema.SimpleNumberParams(schema.NumberParams{
				Name: "value",
				Type: schema.AttributeTypeDecimal(),
			})),
		}
	case "integer":
		return schema.Attributes{schema.SimpleCoreAttribute(
			schema.SimpleNumberParams(schema.NumberParams{
				Name: "value",
				Type: schema.AttributeTypeInteger(),
			})),
		}
	case "binary":
		return schema.Attributes{schema.SimpleCoreAttribute(
			schema.SimpleBinaryParams(schema.BinaryParams{Name: "value"}),
		)}
	case "boolean":
		return schema.Attributes{schema.SimpleCoreAttribute(
			schema.SimpleBooleanParams(schema.BooleanParams{Name: "value"})),
		}
	case "complex":
		return attr.SubAttributes()
	case "dateTime":
		return schema.Attributes{schema.SimpleCoreAttribute(
			schema.SimpleDateTimeParams(schema.DateTimeParams{Name: "value"})),
		}
	case "reference":
		return schema.Attributes{schema.SimpleCoreAttribute(
			schema.SimpleReferenceParams(schema.ReferenceParams{Name: "value"})),
		}
	default:
		return schema.Attributes{schema.SimpleCoreAttribute(
			schema.SimpleStringParams(schema.StringParams{Name: "value"})),
		}
	}
}

func (v OperationValidator) validateRemove() error {
	// If "path" is unspecified, the operation fails with HTTP status code 400 and a "scimType" error code of "noTarget".
	if v.path == nil {
		return &errors.ScimError{
			ScimType: errors.ScimTypeNoTarget,
			Status:   http.StatusBadRequest,
		}
	}

	refAttr, err := v.getRefAttribute(v.path.AttributePath)
	if err != nil {
		return err
	}
	if v.path.ValueExpression != nil {
		if err := f.NewFilterValidator(v.path.ValueExpression, schema.Schema{
			Attributes: multiValuedFilterAttributes(refAttr),
		}).Validate(); err != nil {
			return err
		}
	}
	if subAttrName := v.path.SubAttributeName(); subAttrName != "" {
		if _, err := v.getRefSubAttribute(refAttr, subAttrName); err != nil {
			return err
		}
	}
	return nil
}
