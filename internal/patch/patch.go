package patch

import (
	"encoding/json"
	"fmt"
	f "github.com/elimity-com/scim/internal/filter"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
)

type Op string

const (
	// OperationAdd is used to add a new attribute value to an existing resource.
	OperationAdd Op = "add"
	// OperationRemove removes the value at the target location specified by the required attribute "path".
	OperationRemove Op = "remove"
	// OperationReplace replaces the value at the target location specified by the "path".
	OperationReplace Op = "replace"
)

type OperationValidator struct {
	op    Op
	path  *filter.Path
	value interface{}

	schema  schema.Schema
	schemas map[string]schema.Schema
}

// NewValidator creates an OperationValidator based on the given JSON string and reference schemas.
// Returns an error if patchReq is not valid.
func NewValidator(patchReq string, s schema.Schema, exts ...schema.Schema) (OperationValidator, error) {
	var operation struct {
		Op    string
		Path  string
		Value interface{}
	}
	if err := json.Unmarshal([]byte(patchReq), &operation); err != nil {
		return OperationValidator{}, err
	}

	var path *filter.Path
	if operation.Path != "" {
		validator, err := f.NewPathValidator(operation.Path, s, exts...)
		if err != nil {
			return OperationValidator{}, err
		}
		if err := validator.Validate(); err != nil {
			return OperationValidator{}, err
		}
		p := validator.Path()
		path = &p
	}

	schemas := map[string]schema.Schema{
		s.ID: s,
	}
	for _, e := range exts {
		schemas[e.ID] = e
	}
	return OperationValidator{
		op:    Op(operation.Op),
		path:  path,
		value: operation.Value,

		schema:  s,
		schemas: schemas,
	}, nil
}

// Validate validates the PATCH operation. Unknown attributes in complex values are ignored. The returned interface
// contains a (sanitised) version of given value based on the attribute it targets. Multi-valued attributes will always
// be returned wrapped in a slice, even if it is just one value that was defined within the operation.
func (v OperationValidator) Validate() (interface{}, error) {
	switch v.op {
	case OperationAdd:
		return v.validateAdd()
	default:
		return nil, fmt.Errorf("invalid operation op: %s", v.op)
	}
}
