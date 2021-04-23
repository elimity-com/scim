package scim

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/internal/filter"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

// unmarshal unifies the unmarshal of the requests.
func unmarshal(data []byte, v interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	return d.Decode(v)
}

// ResourceType specifies the metadata about a resource type.
type ResourceType struct {
	// ID is the resource type's server unique id. This is often the same value as the "name" attribute.
	ID optional.String
	// Name is the resource type name. This name is referenced by the "meta.resourceType" attribute in all resources.
	Name string
	// Description is the resource type's human-readable description.
	Description optional.String
	// Endpoint is the resource type's HTTP-addressable endpoint relative to the Base URL of the service provider,
	// e.g., "/Users".
	Endpoint string
	// Schema is the resource type's primary/base schema.
	Schema schema.Schema
	// SchemaExtensions is a list of the resource type's schema extensions.
	SchemaExtensions []SchemaExtension

	// Handler is the set of callback method that connect the SCIM server with a provider of the resource type.
	Handler ResourceHandler
}

func (t ResourceType) getRaw() map[string]interface{} {
	return map[string]interface{}{
		"schemas":          []string{"urn:ietf:params:scim:schemas:core:2.0:ResourceType"},
		"id":               t.ID.Value(),
		"name":             t.Name,
		"description":      t.Description.Value(),
		"endpoint":         t.Endpoint,
		"schema":           t.Schema.ID,
		"schemaExtensions": t.getRawSchemaExtensions(),
	}
}

func (t ResourceType) getRawSchemaExtensions() []map[string]interface{} {
	schemas := make([]map[string]interface{}, 0)
	for _, e := range t.SchemaExtensions {
		schemas = append(schemas, map[string]interface{}{
			"schema":   e.Schema.ID,
			"required": e.Required,
		})
	}
	return schemas
}

func (t ResourceType) getSchemaExtensions() []schema.Schema {
	var extensions []schema.Schema
	for _, e := range t.SchemaExtensions {
		extensions = append(extensions, e.Schema)
	}
	return extensions
}

func (t ResourceType) schemaWithCommon() schema.Schema {
	s := t.Schema

	externalID := schema.SimpleCoreAttribute(
		schema.SimpleStringParams(schema.StringParams{
			CaseExact:  true,
			Mutability: schema.AttributeMutabilityReadWrite(),
			Name:       schema.CommonAttributeExternalID,
			Uniqueness: schema.AttributeUniquenessNone(),
		}),
	)

	s.Attributes = append(s.Attributes, externalID)

	return s
}

func (t ResourceType) validate(raw []byte, method string) (ResourceAttributes, *errors.ScimError) {
	var m map[string]interface{}
	if err := unmarshal(raw, &m); err != nil {
		return ResourceAttributes{}, &errors.ScimErrorInvalidSyntax
	}

	attributes, scimErr := t.schemaWithCommon().Validate(m)
	if scimErr != nil {
		return ResourceAttributes{}, scimErr
	}

	for _, extension := range t.SchemaExtensions {
		extensionField := m[extension.Schema.ID]
		if extensionField == nil {
			if extension.Required {
				return ResourceAttributes{}, &errors.ScimErrorInvalidValue
			}
			continue
		}

		extensionAttributes, scimErr := extension.Schema.Validate(extensionField)
		if scimErr != nil {
			return ResourceAttributes{}, scimErr
		}

		attributes[extension.Schema.ID] = extensionAttributes
	}

	return attributes, nil
}

func (t ResourceType) validateOperation(op PatchOperation) *errors.ScimError {
	// Ensure the operation is a valid one. "add", "replace", or "remove".
	if !contains(validOps, op.Op) {
		return &errors.ScimErrorInvalidFilter
	}

	// "add" and "replace" operations must have a value.
	if (op.Op == PatchOperationAdd || op.Op == PatchOperationReplace) && op.Value == nil {
		return &errors.ScimErrorInvalidFilter
	}

	var (
		validator filter.PathValidator
		err       error
	)
	if op.Op == PatchOperationRemove {
		// "remove" operations require a path.
		validator, err = filter.NewPathValidator(op.Path, t.schemaWithCommon(), t.getSchemaExtensions()...)
		if err != nil {
			return &errors.ScimErrorInvalidPath
		}
	} else {
		// "replace" and "add" can have implicit paths, which is part of the value.
		validator, err = filter.NewPathValidator(op.Path, t.schemaWithCommon(), t.getSchemaExtensions()...)
	}

	if err == nil {
		// Checks whether the path is a valid path within the given reference schemas.
		if err := validator.Validate(); err != nil {
			return &errors.ScimErrorInvalidPath
		}
		// Validates the operation value itself.
		if err := t.validateOperationValue(op); err != nil {
			return err
		}
	}
	return nil
}

func (t ResourceType) validateOperationValue(op PatchOperation) *errors.ScimError {
	validator, err := filter.NewPathValidator(op.Path, t.schemaWithCommon(), t.getSchemaExtensions()...)
	if err != nil {
		scimErr := errors.CheckScimError(err, http.MethodPatch)
		return &scimErr
	}
	var (
		path             = validator.Path()
		attributeName    = path.AttributePath.AttributeName
		subAttributeName = path.AttributePath.SubAttributeName()
	)
	// The sub attribute can come from: `emails.value` or `emails[type eq "work"].value`.
	// The path filter `x.y[].z` is impossible
	if subAttributeName == "" {
		subAttributeName = path.SubAttributeName()
	}

	if attributeName != "" {
		if err := validator.Validate(); err != nil {
			return &errors.ScimErrorInvalidPath
		}
	}

	var mapValue map[string]interface{}
	switch v := op.Value.(type) {
	case map[string]interface{}:
		mapValue = v

	default:
		if subAttributeName == "" {
			mapValue = map[string]interface{}{attributeName: v}
			break
		}
		mapValue = map[string]interface{}{attributeName: map[string]interface{}{
			subAttributeName: v,
		}}
	}

	// Check if it's a patch on an extension.
	if attributeName != "" {
		if id := path.AttributePath.URI(); id != "" {
			for _, ext := range t.SchemaExtensions {
				if strings.EqualFold(id, ext.Schema.ID) {
					return ext.Schema.ValidatePatchOperation(op.Op, mapValue, true)
				}
			}
		}
	}

	return t.schemaWithCommon().ValidatePatchOperationValue(op.Op, mapValue)
}

// validatePatch parse and validate PATCH request.
func (t ResourceType) validatePatch(r *http.Request) (PatchRequest, *errors.ScimError) {
	var req PatchRequest

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return req, &errors.ScimErrorInvalidSyntax
	}

	if err := unmarshal(data, &req); err != nil {
		return req, &errors.ScimErrorInvalidSyntax
	}

	// The body of an HTTP PATCH request MUST contain the attribute "Operations",
	// whose value is an array of one or more PATCH operations.
	if len(req.Operations) < 1 {
		return req, &errors.ScimErrorInvalidValue
	}

	// Evaluation continues until all operations are successfully applied or until an error condition is encountered.
	for i := range req.Operations {
		req.Operations[i].Op = strings.ToLower(req.Operations[i].Op)
		if err := t.validateOperation(req.Operations[i]); err != nil {
			return req, err
		}
	}

	return req, nil
}

// SchemaExtension is one of the resource type's schema extensions.
type SchemaExtension struct {
	// Schema is the URI of an extended schema, e.g., "urn:edu:2.0:Staff".
	Schema schema.Schema
	// Required is a boolean value that specifies whether or not the schema extension is required for the resource
	// type. If true, a resource of this type MUST include this schema extension and also include any attributes
	// declared as required in this schema extension. If false, a resource of this type MAY omit this schema
	// extension.
	Required bool
}
