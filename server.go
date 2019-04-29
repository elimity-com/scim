package scim

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Server represents a SCIM server which implements the HTTP-based SCIM protocol that makes managing identities in multi-
// domain scenarios easier to support via a standardized service.
type Server struct {
	config        serviceProviderConfig
	schemas       map[string]schema
	resourceTypes map[string]resourceType
}

// NewServer returns a SCIM server with given config, resource types and matching schemas. Given schemas must contain
// every schema referenced in by given resource types. The given schemas and resource types cannot contain duplicates
// based on their identifier and no duplicate endpoints can be defined. An error is returned otherwise.
func NewServer(config ServiceProviderConfig, schemas []Schema, resourceTypes []ResourceType) (Server, error) {
	schemasMap := make(map[string]schema)
	for _, s := range schemas {
		if _, ok := schemasMap[s.schema.ID]; ok {
			return Server{}, fmt.Errorf("duplicate schema with id: %s", s.schema.ID)
		}
		schemasMap[s.schema.ID] = s.schema
	}

	tmpEndpoints := map[string]unitType{
		"/":                      unit,
		"/Schemas":               unit,
		"/ResourceTypes":         unit,
		"/ServiceProviderConfig": unit,
	}
	resourceTypesMap := make(map[string]resourceType)
	for _, t := range resourceTypes {
		if _, ok := schemasMap[t.resourceType.Schema]; !ok {
			return Server{}, fmt.Errorf(
				"schemas does not contain a schema with id: %s, referenced by resource type: %s",
				t.resourceType.Schema, t.resourceType.Name,
			)
		}
		for idx, extension := range t.resourceType.SchemaExtensions {
			if _, ok := schemasMap[extension.Schema]; !ok {
				return Server{}, fmt.Errorf(
					"schemas does not contain a schema with id: %s, referenced by resource type extension with index: %d",
					extension.Schema, idx,
				)
			}
		}

		if _, ok := resourceTypesMap[t.resourceType.Name]; ok {
			return Server{}, fmt.Errorf("duplicate resource type with name: %s", t.resourceType.Name)
		}

		if !strings.HasPrefix(t.resourceType.Endpoint, "/") {
			return Server{}, fmt.Errorf(
				"endpoint does not start with a (forward) slash: %s",
				t.resourceType.Endpoint,
			)
		}
		if _, ok := tmpEndpoints[t.resourceType.Endpoint]; ok {
			return Server{}, fmt.Errorf(
				"duplicate endpoints in resource types: %s",
				t.resourceType.Endpoint,
			)
		}

		tmpEndpoints[t.resourceType.Endpoint] = unit
		resourceTypesMap[t.resourceType.Name] = t.resourceType
	}

	return Server{
		config:        config.config,
		schemas:       schemasMap,
		resourceTypes: resourceTypesMap,
	}, nil
}

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

	for _, resourceType := range s.resourceTypes {
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
