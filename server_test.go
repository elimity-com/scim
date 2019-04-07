package scim

import (
	"log"
	"net/http"
)

func ExampleNewServer() {
	log.Fatal(http.ListenAndServe(":8080", NewServer()))
}

func ExampleNewServer_schema() {
	schema, _ := NewSchemaFromFile("/path/to/schema")
	log.Fatal(http.ListenAndServe(":8080", NewServer(schema)))
}
