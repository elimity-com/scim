package filter

import (
	"net/http"
	"strings"

	filter "github.com/di-wu/scim-filter-parser"
	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schema"
)

func invalidFilterError(msg string) *errors.ScimError {
	return &errors.ScimError{
		ScimType: errors.ScimTypeInvalidFilter,
		Detail:   msg,
		Status:   http.StatusBadRequest,
	}
}

// NewFilter get the filter from the request, parses it and adds reference schemas to validate against.
func NewFilter(r *http.Request, schemas ...schema.Schema) (Filter, error) {
	if len(schemas) == 0 {
		return Filter{}, nil
	}

	rawFilter := strings.TrimSpace(r.URL.Query().Get("filter"))
	if rawFilter != "" {
		parser := filter.NewParser(strings.NewReader(rawFilter))
		exp, err := parser.Parse()
		if err != nil {
			return Filter{}, err
		}
		return Filter{
			Expression: exp,
			schema:     schemas[0],
		}, nil
	}
	return Filter{}, nil
}

// GetPathFilter returns a parsed filter path.
func GetPathFilter(path string) (filter.Path, error) {
	if path != "" {
		parser := filter.NewParser(strings.NewReader(path))
		return parser.ParsePath()
	}
	return filter.Path{}, nil
}

// Filter represents a parsed SCIM Filter.
type Filter struct {
	filter.Expression
	schema schema.Schema
}
