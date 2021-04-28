package patch

import (
	"encoding/json"
	"fmt"
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
	if p, err := filter.ParsePath([]byte(operation.Path)); err == nil {
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

func (v OperationValidator) Validate() error {
	switch v.op {
	case OperationAdd:
		return v.ValidateAdd()
	default:
		return fmt.Errorf("invalid operation op: %s", v.op)
	}
}
