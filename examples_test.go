package scim

import (
	"log"
	"net/http"
)

func ExampleNewServer() {
	log.Fatal(http.ListenAndServe(":7643", Server{
		Config:        ServiceProviderConfig{},
		ResourceTypes: nil,
	}))
}

func ExampleNewServer_basePath() {
	http.Handle("/scim/", http.StripPrefix("/scim", Server{
		Config:        ServiceProviderConfig{},
		ResourceTypes: nil,
	}))
	log.Fatal(http.ListenAndServe(":7643", nil))
}
