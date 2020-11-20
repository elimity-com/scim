package scim

import (
	"context"
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

type contextKey struct{ name string }

var requestNoChange = contextKey{"RequestNoChange"}

func tellRequestNoChange(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), requestNoChange, struct{}{}))
}

func isRequestNoChange(r *http.Request) bool {
	_, ok := r.Context().Value(requestNoChange).(struct{})
	return ok
}
