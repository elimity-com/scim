package main

import (
	"github.com/elimity-com/scim"
	"log"
	"net/http"
)

func main() {
	user, _ := scim.NewSchemaFromFile("testdata/simple_user_schema.json")
	log.Fatal(http.ListenAndServe(":8080", http.StripPrefix("/scim/v2", scim.NewServer(user))))
}
