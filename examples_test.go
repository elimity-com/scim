package scim

import (
	logger "log"
	"net/http"
)

func ExampleNewServer() {
	args := &ServerArgs{
		ServiceProviderConfig: &ServiceProviderConfig{},
		ResourceTypes:         []ResourceType{},
	}
	server, err := NewServer(args)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Fatal(http.ListenAndServe(":7643", server))
}

func ExampleNewServer_basePath() {
	args := &ServerArgs{
		ServiceProviderConfig: &ServiceProviderConfig{},
		ResourceTypes:         []ResourceType{},
	}
	server, err := NewServer(args)
	if err != nil {
		logger.Fatal(err)
	}

	// The server never sees the prefix it is mounted under, so strip it. What
	// is left is the resource endpoint, optionally preceded by "/v2".
	http.Handle("/scim/", http.StripPrefix("/scim", server))
	logger.Fatal(http.ListenAndServe(":7643", nil))
}

func ExampleNewServer_logger() {
	loggingMiddleware := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			logger.Println(r.Method, r.URL.Path)

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
	args := &ServerArgs{
		ServiceProviderConfig: &ServiceProviderConfig{},
		ResourceTypes:         []ResourceType{},
	}
	server, err := NewServer(args)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Fatal(http.ListenAndServe(":7643", loggingMiddleware(server)))
}

// A server mounted under a prefix needs a base URL to report locations that
// clients can follow, as it cannot derive the prefix from the request.
func ExampleNewServer_nested() {
	args := &ServerArgs{
		ServiceProviderConfig: &ServiceProviderConfig{},
		ResourceTypes: []ResourceType{{
			Name:     "User",
			Endpoint: "/Users", // Not "/v2/Users", and not "/scim/v2/Users".
		}},
	}
	server, err := NewServer(args, WithBaseURL("https://example.com/scim/v2"))
	if err != nil {
		logger.Fatal(err)
	}

	// Serves both /scim/v2/Users/{id} and /scim/Users/{id}, and reports
	// locations of "https://example.com/scim/v2/Users/{id}".
	http.Handle("/scim/", http.StripPrefix("/scim", server))
	logger.Fatal(http.ListenAndServe(":7643", nil))
}

// WithBaseURL makes locations absolute. The base URL is what a client sees, so
// it carries the mount prefix and the version identifier. Without one,
// locations stay relative to the service provider root (RFC 7643 Section 2.3.7).
func ExampleWithBaseURL() {
	args := &ServerArgs{
		ServiceProviderConfig: &ServiceProviderConfig{},
		ResourceTypes:         []ResourceType{{Name: "User", Endpoint: "/Users"}},
	}

	// NewServer reports a base URL it cannot build on.
	server, err := NewServer(args, WithBaseURL("https://example.com/v2"))
	if err != nil {
		logger.Fatal(err)
	}
	logger.Fatal(http.ListenAndServe(":7643", server))
}
