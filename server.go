package scim

import (
	"fmt"
	"net/http"
	"strings"
)

type Server struct {
	config        serviceProviderConfig
	schemas       map[string]schema
	resourceTypes map[string]resourceType
}

// Schemas must contain every schema referenced in by given resource types. Given schemas and resource types must not
// contain duplicates based on their identifier.
func NewServer(config ServiceProviderConfig, schemas []Schema, resourceTypes []ResourceType) (Server, error) {
	schemasMap := make(map[string]schema)
	for _, s := range schemas {
		if _, ok := schemasMap[s.schema.ID]; ok {
			return Server{}, fmt.Errorf("duplicate schema with id: %s", s.schema.ID)
		}
		schemasMap[s.schema.ID] = s.schema
	}

	tmpEndpoints := make(map[string]unitType)
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
	path := r.URL.Path
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
			switch r.Method {
			case http.MethodGet:
				s.resourceGetHandler(w, r, strings.TrimPrefix(path, resourceType.Endpoint+"/"), resourceType)
				return
			case http.MethodPut:
				s.resourcePutHandler(w, r, strings.TrimPrefix(path, resourceType.Endpoint+"/"), resourceType)
				return
			case http.MethodDelete:
				s.resourceDeleteHandler(w, r, strings.TrimPrefix(path, resourceType.Endpoint+"/"), resourceType)
				return
			}
		}
	}

	errorHandler(w, r)
}
