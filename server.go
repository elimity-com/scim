package scim

import (
	"net/http"
	"strings"
)

type Server struct {
	schemas []Schema
}

func NewServer(schemas ...Schema) Server {
	return Server{
		schemas: schemas,
	}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	default:
		errorHandler(w, r)
	}
}
