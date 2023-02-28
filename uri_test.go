package odin_uri

import (
	"bytes"
	"fmt"
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

func TestParseAuthority(t *testing.T) {
	runes := bytes.Runes([]byte("//google.com"))
	auth, _, err := parseAuthority(runes)
	if err != nil {
		t.Fatal(err)
		return
	}
	if auth != "google.com" {
		t.Fatal(fmt.Sprintf("%s != google.com", auth))
		return
	}
	runes = bytes.Runes([]byte("//test@google.com"))
	auth, _, err = parseAuthority(runes)
	if err != nil {
		t.Fatal(err)
		return
	}
	if auth != "test@google.com" {
		t.Fatal(fmt.Sprintf("%s != test@google.com", auth))
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
	host, _, err := parseHost(runes)
	if err != nil {
		t.Fatal(err)
		return
	}
	runes = bytes.Runes([]byte("testhost"))
	host, _, err = parseHost(runes)
	if err != nil {
		t.Fatal(err)
		return
	}
	if host != "testhost" {
		t.Fatalf("%s != %s", host, "testhost")
	}
}

func TestParseIpvFuture(t *testing.T) {
	runes := bytes.Runes([]byte("v8.2"))
	ipvf, _, ipvfErr := parseIpvFuture(runes)
	if ipvfErr != nil {
		t.Fatal(ipvfErr)
		return
	}
	if ipvf != "v8.2" {
		t.Fatalf("%s != %s", ipvf, "v8.2")
		return
	}
	runes = bytes.Runes([]byte("v1.3fh3:e234"))
	ipvf, _, ipvfErr = parseIpvFuture(runes)
	if ipvfErr != nil {
		t.Fatal(ipvfErr)
		return
	}
	if ipvf != "v1.3fh3:e234" {
		t.Fatalf("%s != %s", ipvf, "v1.3fh3:e234")
	}
}

func TestParsePort(t *testing.T) {
	runes := bytes.Runes([]byte("0"))
	_, _, err := parsePort(runes)
	if err == nil {
		t.Fatal(err)
		return
	}
	runes = bytes.Runes([]byte("1"))
	port, _, err := parsePort(runes)
	if err != nil {
		t.Fatal(fmt.Sprintf("'1' should be valid port: %s", err.Error()))
		return
	}
	if port != "1" {
		t.Fatal(fmt.Sprintf("%s != 1", port))
		return
	}
	runes = bytes.Runes([]byte("65536"))
	_, _, err = parsePort(runes)
	if err == nil {
		t.Fatal("'65536' should be invalid port")
		return
	}
	runes = bytes.Runes([]byte("123456"))
	_, _, err = parsePort(runes)
	if err == nil {
		t.Fatal("'123456' should be an invalid port")
	}
}
