package scim

import (
	"io/ioutil"
	"log"
	"net/http"
)

// Errors are ignored to keep it simple.
func ExampleNewServer() {
	rawConfig, _ := ioutil.ReadFile("/path/to/config")
	config, _ := NewServiceProviderConfig(rawConfig)
	rawSchema, _ := ioutil.ReadFile("/path/to/schema")
	schema, _ := NewSchema(rawSchema)
	rawResourceType, _ := ioutil.ReadFile("/path/to/resourceType")
	resourceType, _ := NewResourceType(rawResourceType, nil)

	server, _ := NewServer(config, []Schema{schema}, []ResourceType{resourceType})
	log.Fatal(http.ListenAndServe(":8080", server))
}

// Errors are ignored to keep it simple.
func ExampleNewServer_basePath() {
	rawConfig, _ := ioutil.ReadFile("/path/to/config")
	config, _ := NewServiceProviderConfig(rawConfig)
	rawSchema, _ := ioutil.ReadFile("/path/to/schema")
	schema, _ := NewSchema(rawSchema)
	rawResourceType, _ := ioutil.ReadFile("/path/to/resourceType")
	resourceType, _ := NewResourceType(rawResourceType, nil)

	server, _ := NewServer(config, []Schema{schema}, []ResourceType{resourceType})
	http.Handle("/scim/", http.StripPrefix("/scim", server))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
