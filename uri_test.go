package odin_uri

import (
	"bytes"
	"testing"
)

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

func TestParsePctEncoding(t *testing.T) {
	runes := []rune{'%', '1', 'f'}
	_, err := parsePctEncoded(runes)
	if err != nil {
		t.Fatal(err)
		return
	}
	badRunes := []rune{'%', 'g', '3'}
	_, err = parsePctEncoded(badRunes)
	if err == nil {
		t.Fatalf("%s", "'%g3' should fail as pct encoding")
	}
}

func TestParseIpv4(t *testing.T) {
	runes := bytes.Runes([]byte("192.168.0.1"))
	_, _, err := parseIpv4(runes)
	if err != nil {
		t.Fatal(err)
		return
	}
	runes = bytes.Runes([]byte("111.111.111.111"))
	_, _, err = parseIpv4(runes)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestParseHost(t *testing.T) {
	runes := bytes.Runes([]byte("123.12.12.123"))
	host, err := parseHost(runes)
	if err != nil {
		t.Fatal(err)
		return
	}
	runes = bytes.Runes([]byte("testhost"))
	host, err = parseHost(runes)
	if err != nil {
		t.Fatal(err)
		return
	}
	if host != "testhost" {
		t.Fatalf("%s != %s", host, "testhost")
	}
}
