package scim

import (
	"log"
	"net/http"
)

// Errors are ignored to keep it simple.
func ExampleNewServer() {
	config, _ := NewServiceProviderConfigFromFile("/path/to/config")
	schema, _ := NewSchemaFromFile("/path/to/schema")
	resourceType, _ := NewResourceTypeFromFile("/path/to/resourceType")
	server, _ := NewServer(config, []Schema{schema}, []ResourceType{resourceType})
	log.Fatal(http.ListenAndServe(":8080", server))
}
