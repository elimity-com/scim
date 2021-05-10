package filter

import (
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
)

// PathValidator represents a path validator.
type PathValidator struct {
	path       filter.Path
	schema     schema.Schema
	extensions []schema.Schema
}

// NewPathValidator constructs a new path validator.
func NewPathValidator(pathFilter string, s schema.Schema, exts ...schema.Schema) (PathValidator, error) {
	f, err := filter.ParsePath([]byte(pathFilter))
	if err != nil {
		return PathValidator{}, err
	}
	return PathValidator{
		path:       f,
		schema:     s,
		extensions: exts,
	}, nil
}

func (v PathValidator) Path() filter.Path {
	return v.path
}

// Validate checks whether the path is a valid path within the given reference schemas.
func (v PathValidator) Validate() error {
	err := v.validatePath(v.schema)
	if err == nil {
		return nil
	}
	for _, e := range v.extensions {
		if err := v.validatePath(e); err == nil {
			return nil
		}
	}
	return err
}

// validatePath tries to validate the path against the given schema.
func (v PathValidator) validatePath(ref schema.Schema) error {
	// e.g. members
	//      ^______
	attr, err := validateAttributePath(ref, v.path.AttributePath)
	if err != nil {
		return err
	}

	// e.g. members[value eq "0"]
	//             ^_____________
	if v.path.ValueExpression != nil {
		if err := validateExpression(
			schema.Schema{
				ID:         ref.ID,
				Attributes: attr.SubAttributes(),
			},
			v.path.ValueExpression,
		); err != nil {
			return err
		}
	}

	// e.g. members[value eq "0"].displayName
	//                            ^__________
	if subAttrName := v.path.SubAttributeName(); subAttrName != "" {
		if err := validateSubAttribute(attr, subAttrName); err != nil {
			return err
		}
	}
	return nil
}
