package idp_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/elimity-com/scim"
)

func getNewServer(t *testing.T, idpName string) scim.Server {
	switch idpName {
	case "okta":
		return newOktaTestServer(t)
	case "azuread":
		return newAzureADTestServer(t)
	default:
		panic("unreachable")
	}
}

func unmarshal(data []byte, v interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	return d.Decode(v)
}
