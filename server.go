package scim

import (
	"fmt"
	"net/http"
)

type Server struct {
	schemas []Schema
	router  router
}

func NewServer(schemas ...Schema) Server {
	router := newRouter(errorHandler)
	server := Server{
		schemas: schemas,
		router:  router,
	}

	router.Handle("GET", "/Schemas", server.schemasHandler)
	router.Handle("GET", "/Schemas/{id}", server.schemaHandler)

	return server
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	s.router.ServeHTTP(w, r)
}
