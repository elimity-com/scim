package scim

import (
	"fmt"
	logger "log"
	"net/http"

	"github.com/elimity-com/scim/schema"
	filter "github.com/scim2/filter-parser/v2"
)

func ExampleApplyPatch() {
	// Define the schema for the resource.
	userSchema := schema.Schema{
		ID: schema.UserSchema,
		Attributes: schema.Attributes{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "userName",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "displayName",
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name:        "emails",
				MultiValued: true,
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{Name: "value"}),
					schema.SimpleStringParams(schema.StringParams{Name: "type"}),
				},
			}),
		},
	}

	// The resource to patch, typically fetched from a data store.
	attrs := ResourceAttributes{
		"userName":    "john",
		"displayName": "John Doe",
		"emails": []interface{}{
			map[string]interface{}{
				"value": "john@work.com",
				"type":  "work",
			},
		},
	}

	// Parse paths for the operations.
	emailsPath, _ := filter.ParsePath([]byte("emails"))
	workEmailValuePath, _ := filter.ParsePath([]byte(`emails[type eq "work"].value`))

	// Apply a sequence of patch operations.
	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationAdd, Path: &emailsPath, Value: []interface{}{
			map[string]interface{}{"value": "john@home.com", "type": "home"},
		}},
		{Op: PatchOperationReplace, Path: &workEmailValuePath, Value: "john@newwork.com"},
	}, userSchema)
	if err != nil {
		panic(err)
	}

	fmt.Println(result["userName"])
	emails := result["emails"].([]interface{})
	fmt.Println(len(emails))
	fmt.Println(emails[0].(map[string]interface{})["value"])
	fmt.Println(emails[1].(map[string]interface{})["value"])

	// Output:
	// john
	// 2
	// john@newwork.com
	// john@home.com
}

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
