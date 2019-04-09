package scim

import (
	"net/http"
	"strings"
)

type Server struct {
	config        serviceProviderConfig
	schemas       map[string]schema
	resourceTypes map[string]resourceType
}

func NewServer(config ServiceProviderConfig, schemas []Schema, resourceTypes []ResourceType) Server {
	schemasMap := make(map[string]schema)
	for _, s := range schemas {
		schemasMap[s.schema.ID] = s.schema
	}

	resourceTypesMap := make(map[string]resourceType)
	for _, t := range resourceTypes {
		resourceTypesMap[t.resourceType.Name] = t.resourceType
	}

	return Server{
		schemas:       schemasMap,
		resourceTypes: resourceTypesMap,
	}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := r.URL.Path
	switch {
	case path == "/Schemas" && r.Method == "GET":
		s.schemasHandler(w, r)
	case strings.HasPrefix(path, "/Schemas/") && r.Method == "GET":
		s.schemaHandler(w, r, strings.TrimPrefix(path, "/Schemas/"))
	case path == "/ResourceTypes" && r.Method == "GET":
		s.resourceTypesHandler(w, r)
	case strings.HasPrefix(path, "/ResourceTypes/") && r.Method == "GET":
		s.resourceTypeHandler(w, r, strings.TrimPrefix(path, "/ResourceTypes/"))
	case path == "/ServiceProviderConfig":
		s.serviceProviderConfigHandler(w, r)
	default:
		errorHandler(w, r)
	}
}
