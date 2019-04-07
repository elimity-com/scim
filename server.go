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
	case strings.HasPrefix(path, "/Schemas"):
		path := strings.TrimPrefix(path, "/Schemas")
		if path == "" {
			s.schemasHandler(w, r)
			return
		}
		s.schemaHandler(w, r, path[1:])
	default:
		errorHandler(w, r)
	}
}
