package scim

import (
	"strings"

	scim "github.com/di-wu/scim-filter-parser"
)

const (
	add     = "add"
	remove  = "remove"
	replace = "replace"
)

var (
	validOps = []string{add, remove, replace}
)

type (
	// PatchOperation represents a single PATCH operation.
	PatchOperation struct {
		Op    string      `json:"op"`
		Path  string      `json:"path"`
		Value interface{} `json:"value,omitempty"`
	}

	// PatchRequest represents a resource PATCH request.
	PatchRequest struct {
		Schemas    []string         `json:"schemas"`
		Operations []PatchOperation `json:"Operations"`
	}
)

// HasPathFilter whether or not the path in the operation is in filter syntax or not.
// ex:
// "emails[type eq \"work\" and value ew \"example.com\"]" -> true
// "emails"                                                -> false
func (p PatchOperation) HasPathFilter() bool {
	return p.GetPathFilter() != nil
}

// GetPathFilter parses patch operation path to determine if it is a filter. If it is, scim.Expression will be returned
// if it is not, it will be nil.
func (p PatchOperation) GetPathFilter() *scim.AttributeExpression {
	parser := scim.NewParser(strings.NewReader(p.Path))
	filter, err := parser.Parse()

	// We can assume the path provided is not a filter
	if err != nil {
		return nil
	}

	// Only attribute expressions are supported with PATCH paths
	// PATH = attrPath / valuePath [subAttr]
	if attrFilter, ok := filter.(scim.AttributeExpression); ok {
		return &attrFilter
	}

	return nil
}
