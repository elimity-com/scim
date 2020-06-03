package scim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

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

// unmarshal unifies the unmarshal of the requests.
func unmarshal(data []byte, v interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	return d.Decode(v)
}

func (t ResourceType) validate(raw []byte) (ResourceAttributes, *errors.ScimError) {
	var m map[string]interface{}
	if err := unmarshal(raw, &m); err != nil {
		return ResourceAttributes{}, &errors.ScimErrorInvalidSyntax
	}

	attributes, scimErr := t.Schema.Validate(m)
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

// validatePatch parse and validate PATCH request
func (t ResourceType) validatePatch(r *http.Request) (PatchRequest, *errors.ScimError) {
	var req PatchRequest

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return req, &errors.ScimErrorInvalidSyntax
	}

	if err := unmarshal(data, &req); err != nil {
		return req, &errors.ScimErrorInvalidSyntax
	}

	// Error causes are currently unused but could be logged or perhaps used to build a more detailed error message.
	errorCauses := make([]string, 0)

	// The body of an HTTP PATCH request MUST contain the attribute "Operations",
	// whose value is an array of one or more PATCH operations.
	if len(req.Operations) < 1 {
		return req, &errors.ScimErrorInvalidValue
	}

	for i := range req.Operations {
		req.Operations[i].Op = strings.ToLower(req.Operations[i].Op)
		errorCauses = append(errorCauses, t.validateOperation(req.Operations[i])...)
	}

	// Denotes all of the errors that have occurred parsing the request.
	if len(errorCauses) > 0 {
		return req, &errors.ScimErrorInvalidSyntax
	}

	return req, nil
}

func (t ResourceType) validateOperation(op PatchOperation) []string {
	errorCauses := make([]string, 0)

	// Ensure the operation is a valid one. "add", "replace", or "remove".
	if !contains(validOps, op.Op) {
		errorCauses = append(
			errorCauses,
			fmt.Sprintf(
				"invalid operation type provided, got %s, expected %s",
				op.Op,
				strings.Join(validOps, ", "),
			),
		)
	}

	// "add" and "replace" operations must have a value
	if (op.Op == PatchOperationAdd || op.Op == PatchOperationReplace) && op.Value == nil {
		errorCauses = append(
			errorCauses,
			"an add or replace patch operation must contain a value",
		)
	}

	// "remove" operations require a path.
	// The "replace" and "add" operations can have implicit paths, which is part of the value.
	if op.Op == PatchOperationRemove && op.Path == "" {
		errorCauses = append(errorCauses, "path is required on a remove operation")
	}

	if err := t.validateOperationValue(op); err != nil {
		return append(errorCauses, fmt.Sprintf("%s operation has an invalid value", op.Op))
	}

	return errorCauses
}

func (t ResourceType) validateOperationValue(op PatchOperation) *errors.ScimError {
	// Not attempting to validate value or path if it is a filter based path.
	// Perhaps we could at least validate the ComparePath
	if op.GetPathFilter() != nil {
		return nil
	}

	mapValue, ok := op.Value.(map[string]interface{})
	if !ok {
		mapValue = map[string]interface{}{op.Path: op.Value}
	}

	// Check if it's a patch on a extension.
	if op.Path != "" {
		if i := strings.LastIndex(op.Path, ":"); i != -1 {
			id := op.Path[:i]
			for _, ext := range t.SchemaExtensions {
				if strings.EqualFold(id, ext.Schema.ID) {
					return ext.Schema.ValidatePatchOperation(op.Op, mapValue, true)
				}
			}
		}
	}

	return t.Schema.ValidatePatchOperationValue(op.Op, mapValue)
}
