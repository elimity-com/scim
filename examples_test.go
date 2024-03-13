package scim

import (
	logger "log"
	"net/http"
)

func ExampleNewServer() {
	server := NewServer()
	logger.Fatal(http.ListenAndServe(":7643", server))
}

func ExampleNewServer_basePath() {
	server := NewServer()
	// You can host the SCIM server on a custom path, make sure to strip the prefix, so only `/v2/` is left.
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
	server := NewServer()
	logger.Fatal(http.ListenAndServe(":7643", loggingMiddleware(server)))
}
