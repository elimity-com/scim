package schema

import (
	"encoding/json"
	"strings"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
)

// Schema is a collection of attribute definitions that describe the contents of an entire or partial resource.
type Schema struct {
	Attributes  []CoreAttribute
	Description optional.String
	ID          string
	Name        optional.String
}

// Validate validates given resource based on the schema.
func (s Schema) Validate(resource interface{}) (map[string]interface{}, errors.ValidationError) {
	core, ok := resource.(map[string]interface{})
	if !ok {
		return nil, errors.ValidationErrorInvalidSyntax
	}

	attributes := make(map[string]interface{})
	for _, attribute := range s.Attributes {
		var hit interface{}
		var found bool
		for k, v := range core {
			if strings.EqualFold(attribute.name, k) {
				if found {
					return nil, errors.ValidationErrorInvalidSyntax
				}
				found = true
				hit = v
			}
		}

		attr, scimErr := attribute.validate(hit)
		if scimErr != errors.ValidationErrorNil {
			return nil, scimErr
		}
		attributes[attribute.name] = attr
	}
	return attributes, errors.ValidationErrorNil
}

// ValidatePatchOperationValue validates an individual operation and its related value
func (s Schema) ValidatePatchOperationValue(operation string, operationValue map[string]interface{}) errors.ValidationError {
	for k, v := range operationValue {
		var attr *CoreAttribute
		scimErr := errors.ValidationErrorNil

		for _, attribute := range s.Attributes {
			if strings.EqualFold(attribute.name, k) {
				attr = &attribute
				break
			}
		}

		// Attribute does not exist in the schema, thus it is an invalid request.
		// Immutable attrs can only be added and Readonly attrs cannot be patched
		if attr == nil || cannotBePatched(operation, *attr) {
			return errors.ValidationErrorInvalidValue
		}

		// "remove" operations simply have to exist
		if operation != "remove" {
			_, scimErr = attr.validate(v)
		}

		if scimErr != errors.ValidationErrorNil {
			return scimErr
		}
	}

	return errors.ValidationErrorNil
}

func cannotBePatched(op string, attr CoreAttribute) bool {
	return isImmutable(op, attr) || isReadOnly(attr)
}

func isImmutable(op string, attr CoreAttribute) bool {
	return attr.mutability == attributeMutabilityImmutable && (op == "replace" || op == "remove")
}

func isReadOnly(attr CoreAttribute) bool {
	return attr.mutability == attributeMutabilityReadOnly
}

// MarshalJSON converts the schema struct to its corresponding json representation.
func (s Schema) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":          s.ID,
		"name":        s.Name,
		"description": s.Description.Value(),
		"attributes":  s.getRawAttributes(),
	})
}

func (s Schema) getRawAttributes() []map[string]interface{} {
	attributes := make([]map[string]interface{}, 0)
	for _, a := range s.Attributes {
		attributes = append(attributes, map[string]interface{}{
			"canonicalValues": a.canonicalValues,
			"caseExact":       a.caseExact,
			"description":     a.description.Value(),
			"multiValued":     a.multiValued,
			"mutability":      a.mutability,
			"name":            a.name,
			"referenceTypes":  a.referenceTypes,
			"required":        a.required,
			"returned":        a.returned,
			"subAttributes":   a.subAttributes,
			"type":            a.typ,
			"uniqueness":      a.uniqueness,
		})
	}
	return attributes
}
