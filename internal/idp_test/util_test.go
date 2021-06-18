package idp_test

import (
	"bytes"
	"encoding/json"

	"github.com/elimity-com/scim"
)

func getNewServer(idpName string) scim.Server {
	switch idpName {
	case "okta":
		return newOktaTestServer()
	case "azuread":
		return newAzureADTestServer()
	default:
		panic("unreachable")
	}
}

func unmarshal(data []byte, v interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	return d.Decode(v)
}
