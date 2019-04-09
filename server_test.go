package scim

import (
	"log"
	"net/http"
)

func ExampleNewServer() {
	schema, _ := NewSchemaFromFile("/path/to/schema")
	resourceType, _ := NewResourceTypeFromFile("/path/to/resourceType")
	log.Fatal(http.ListenAndServe(":8080", NewServer([]Schema{schema}, []ResourceType{resourceType})))
}
