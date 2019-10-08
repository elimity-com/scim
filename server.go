package scim

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/elimity-com/scim/schema"
)

const (
	defaultStartIndex = 1
	defaultCount      = 10
	maxCount          = 200
)

// Server represents a SCIM server which implements the HTTP-based SCIM protocol that makes managing identities in multi-
// domain scenarios easier to support via a standardized service.
type Server struct {
	Config        ServiceProviderConfig
	ResourceTypes []ResourceType
}

// getSchemas extracts all the schemas from the resources types defined in the server. Duplicate IDs will get overwritten.
func (s Server) getSchemas() map[string]schema.Schema {
	schemas := make(map[string]schema.Schema)
	for _, resourceType := range s.ResourceTypes {
		schemas[resourceType.Schema.ID] = resourceType.Schema
		for _, extension := range resourceType.SchemaExtensions {
			schemas[extension.Schema.ID] = extension.Schema
		}
	}
	return schemas
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
				requestParams, paramsErr := parseRequestParams(r)

				if paramsErr != nil {
					errorHandler(w, r, *paramsErr)
				}

				s.resourcesGetHandler(w, r, resourceType, requestParams)
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

func getPositiveIntQueryParam(r *http.Request, key string, def int) (int, error) {
	strVal := r.URL.Query().Get(key)

	if strVal == "" {
		return def, nil
	}

	if intVal, err := strconv.Atoi(strVal); err == nil {
		return intVal, nil
	}

	return 0, fmt.Errorf("invalid query parameter, \"%s\" must be an integer", key)
}

func parseRequestParams(r *http.Request) (response ListRequestParams, err *scimError) {
	var invalidParams []string

	count, ctErr := getPositiveIntQueryParam(r, "count", defaultCount)
	startIndex, idxErr := getPositiveIntQueryParam(r, "startIndex", defaultStartIndex)

	if ctErr != nil {
		invalidParams = append(invalidParams, "count")
	}

	if idxErr != nil {
		invalidParams = append(invalidParams, "startIndex")
	}

	if len(invalidParams) > 1 {
		badReqErr := scimErrorBadRequest(invalidParams)
		err = &badReqErr
	}

	if count > maxCount {
		count = maxCount
	}

	response = ListRequestParams{
		Count:      count,
		StartIndex: startIndex,
	}

	return
}
