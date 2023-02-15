package odin_uri

import "testing"

func TestParseUriHttp(t *testing.T) {
	var httpUri = "http://www.ietf.org/rfc/rfc2396.txt"
	var uri, err = ParseUri(httpUri)
	if err != nil {
		t.Fatal(err)
	}
	if uri.Schema() != "http" {
		t.Fatalf("invalid http schema: %s", uri.Schema())
	}
	if uri.Authority() != "www.ietf.org" {
		t.Fatalf("invalid http authority: %s", uri.Authority())
	}
}
