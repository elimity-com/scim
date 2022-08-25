package scim

import (
	"github.com/elimity-com/scim/logging"
	"log"
	"net/http"
	"os"
)

func ExampleNewServer() {
	log.Fatal(http.ListenAndServe(":7643", NewServer(
		ServiceProviderConfig{},
		nil,
		logging.NewSimpleLogger(os.Stderr),
	)))
}

func ExampleNewServer_basePath() {
	http.Handle("/scim/", http.StripPrefix("/scim", NewServer(
		ServiceProviderConfig{},
		nil,
		logging.NewSimpleLogger(os.Stderr),
	)))
	log.Fatal(http.ListenAndServe(":7643", nil))
}
