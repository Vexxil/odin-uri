package odin_uri

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const genDelims = ":/?#[]@"
const subDelims = "!$&'()*+,;="
const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digit = "0123456789"

type URI interface {
	Schema() string
	Authority() string
}

type uri struct {
	schema    string
	authority string
}

func (u uri) Schema() string {
	return u.schema
}

func (u uri) Authority() string {
	return u.authority
}

func ParseUri(value string) (URI, error) {

	if len(value) < 3 {
		return nil, errors.New("URI too short")
	}
	var runes = bytes.Runes([]byte(value))
	var colonIndex = runeIndex(runes, ':')

	if colonIndex == -1 {
		return nil, errors.New("missing ':' delimiter")
	}

	var schema, schemaErr = parseSchema(runes[0:colonIndex])
	if schemaErr != nil {
		return nil, schemaErr
	}

	var authority, authorityErr = parseAuthority(runes[colonIndex+1:])
	if authorityErr != nil {
		if authorityErr.Error() == "no authority" {

		} else {
			return nil, authorityErr
		}
	}

	return uri{schema, authority}, nil
}

func parseSchema(runes []rune) (string, error) {
	if !isAlpha(runes[0]) {
		return "", errors.New("schema must start with alpha")
	}
	for i, r := range runes[1:] {
		if !isAlpha(r) && !isDigit(r) && r != '+' && r != '-' && r != '.' {
			return "", errors.New(fmt.Sprintf("invalid character is schame index %d: '%c'", i, r))
		}
	}
	return string(runes), nil
}

func parseAuthority(runes []rune) (string, error) {
	if string(runes[0:2]) != "//" {
		return "", errors.New("no authority")
	}
	// TODO do proper parse
	for i, r := range runes[2:] {
		if r == '/' {
			return string(runes[2 : i+2]), nil
		}
	}
	return "", errors.New("invalid authority")
}

func parsePath(runes []rune) (string, error) {
	panic("todo")
}

func parseUserInfo() {

}

func parseHost() {

}

func parsePort() {

}

func parseIpLiteral() {

}

func parseIpv4() {

}

func parseIpv6() {

}

func parseIpvFuture() {

}

func parsePctEncoded() {

}

func runeIndex(runes []rune, value rune) int {
	for i, r := range runes {
		if r == value {
			return i
		}
	}
	return -1
}

func isReserved(value rune) bool {
	return strings.ContainsRune(genDelims, value) || strings.ContainsRune(subDelims, value)
}

func isAlpha(value rune) bool {
	return strings.ContainsRune(alpha, value)
}

func isDigit(value rune) bool {
	return strings.ContainsRune(digit, value)
}
