package scim

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	scim "github.com/di-wu/scim-filter-parser"
	"github.com/elimity-com/scim/schema"
)

const (
	defaultStartIndex = 1
	fallbackCount     = 100
)

// Server represents a SCIM server which implements the HTTP-based SCIM protocol that makes managing identities in multi-
// domain scenarios easier to support via a standardized service.
type Server struct {
	Config        ServiceProviderConfig
	ResourceTypes []ResourceType
}

// getSchemas extracts all the schemas from the resources types defined in the server. Duplicate IDs will be ignored.
func (s Server) getSchemas() []schema.Schema {
	ids := make([]string, 0)
	schemas := make([]schema.Schema, 0)
	for _, resourceType := range s.ResourceTypes {
		if !contains(ids, resourceType.Schema.ID) {
			schemas = append(schemas, resourceType.Schema)
		}
		ids = append(ids, resourceType.Schema.ID)
		for _, extension := range resourceType.SchemaExtensions {
			if !contains(ids, extension.Schema.ID) {
				schemas = append(schemas, extension.Schema)
			}
			ids = append(ids, extension.Schema.ID)
		}
	}
	return schemas
}

// getSchema extracts the schemas from the resources types defined in the server with given id.
func (s Server) getSchema(id string) schema.Schema {
	for _, resourceType := range s.ResourceTypes {
		if resourceType.Schema.ID == id {
			return resourceType.Schema
		}
		for _, extension := range resourceType.SchemaExtensions {
			if extension.Schema.ID == id {
				return extension.Schema
			}
		}
	}
	return schema.Schema{}
}

// ServeHTTP dispatches the request to the handler whose pattern most closely matches the request URL.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/scim+json")
	path := strings.TrimPrefix(r.URL.Path, "/v2")
	switch {
	case path == "/Schemas" && r.Method == http.MethodGet:
		s.schemasHandler(w, r)
		return
	case strings.HasPrefix(path, "/Schemas/") && r.Method == http.MethodGet:
		s.schemaHandler(w, r, strings.TrimPrefix(path, "/Schemas/"))
		return
	case path == "/ResourceTypes" && r.Method == http.MethodGet:
		s.resourceTypesHandler(w, r)
		return
	case strings.HasPrefix(path, "/ResourceTypes/") && r.Method == http.MethodGet:
		s.resourceTypeHandler(w, r, strings.TrimPrefix(path, "/ResourceTypes/"))
		return
	case path == "/ServiceProviderConfig":
		s.serviceProviderConfigHandler(w, r)
		return
	}

	for _, resourceType := range s.ResourceTypes {
		if path == resourceType.Endpoint {
			switch r.Method {
			case http.MethodPost:
				s.resourcePostHandler(w, r, resourceType)
				return
			case http.MethodGet:
				s.resourcesGetHandler(w, r, resourceType)
				return
			}
		}

		if strings.HasPrefix(path, resourceType.Endpoint+"/") {
			id, err := parseIdentifier(path, resourceType.Endpoint)
			if err != nil {
				break
			}

			switch r.Method {
			case http.MethodGet:
				s.resourceGetHandler(w, r, id, resourceType)
				return
			case http.MethodPut:
				s.resourcePutHandler(w, r, id, resourceType)
				return
			case http.MethodPatch:
				s.resourcePatchHandler(w, r, id, resourceType)
				return
			case http.MethodDelete:
				s.resourceDeleteHandler(w, r, id, resourceType)
				return
			}
		}
	}

	errorHandler(w, r, scimError{
		detail: "Specified endpoint does not exist.",
		status: http.StatusNotFound,
	})
}

func parseIdentifier(path, endpoint string) (string, error) {
	return url.PathUnescape(strings.TrimPrefix(path, endpoint+"/"))
}

func getIntQueryParam(r *http.Request, key string, def int) (int, error) {
	strVal := r.URL.Query().Get(key)

	if strVal == "" {
		return def, nil
	}

	if intVal, err := strconv.Atoi(strVal); err == nil {
		return intVal, nil
	}

	return 0, fmt.Errorf("invalid query parameter, \"%s\" must be an integer", key)
}

func (s Server) parseRequestParams(r *http.Request) (ListRequestParams, *scimError) {
	invalidParams := make([]string, 0)

	defaultCount := s.Config.getItemsPerPage()
	count, countErr := getIntQueryParam(r, "count", defaultCount)
	if countErr != nil {
		invalidParams = append(invalidParams, "count")
	}
	startIndex, indexErr := getIntQueryParam(r, "startIndex", defaultStartIndex)
	if indexErr != nil {
		invalidParams = append(invalidParams, "startIndex")
	}

	if len(invalidParams) > 1 {
		err := scimErrorBadParams(invalidParams)
		return ListRequestParams{}, &err
	}

	// Ensure the count isn't more then the allowable max and not less then 1.
	if count > defaultCount || count < 1 {
		count = defaultCount
	}

	if startIndex < 1 {
		startIndex = defaultStartIndex
	}

	filter, filterErr := getFilter(r)
	if filterErr != nil {
		err := scimErrorBadParams([]string{"filter"})
		return ListRequestParams{}, &err
	}

	return ListRequestParams{
		Count:      count,
		Filter:     filter,
		StartIndex: startIndex,
	}, nil
}

func getFilter(r *http.Request) (scim.Expression, error) {
	rawFilter := strings.TrimSpace(r.URL.Query().Get("filter"))
	if rawFilter != "" {
		parser := scim.NewParser(strings.NewReader(rawFilter))
		return parser.Parse()
	}
	return nil, nil
}
