package scim

import (
	"log"
	"net/http"
)

func ExampleNewServer() {
	config, _ := NewServiceProviderConfigFromFile("/path/to/config")
	schema, _ := NewSchemaFromFile("/path/to/schema")
	resourceType, _ := NewResourceTypeFromFile("/path/to/resourceType")
	log.Fatal(http.ListenAndServe(":8080", NewServer(config, []Schema{schema}, []ResourceType{resourceType})))
}
