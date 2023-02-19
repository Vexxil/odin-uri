package odin_uri

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const genDelims = ":/?#[]@"
const subDelims = "!$&'()*+,;="
const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digit = "0123456789"
const pctEncoding = digit + "abcdefABCDEF"

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
	runes = runes[2:]
	authority := ""
	if userInfo, i, uiErr := parseUserInfo(runes); uiErr != nil {
		if runes[3] != '@' {
			return "", errors.New(fmt.Sprintf("invalid authority: %s", uiErr.Error()))
		}
		authority = userInfo + "@"
		runes = runes[i:]
	}
	host, hostErr := parseHost(runes)
	if hostErr == nil {
		authority = authority + host
	}

	return "", errors.New("invalid authority")
}

func parsePath(runes []rune) (string, error) {
	panic("todo")
}

func parseUserInfo(runes []rune) (string, int, error) {
	userInfo := make([]rune, 0)
	rl := len(runes)
	i := 0
	for i < rl {
		r := runes[i]
		if r == ':' {
			userInfo = append(userInfo, ':')
		} else if isSubDelim(r) || !isReserved(r) {
			userInfo = append(userInfo, r)
		} else {
			pctEncoded, pctErr := parsePctEncoded(runes[i:])
			if pctErr != nil {
				return "", -1, errors.New("invalid user-info")
			}
			userInfo = append(userInfo, pctEncoded...)
			i += 3
			continue
		}
		i++
	}
	return string(userInfo), i, nil
}

func parseHost(runes []rune) (string, error) {
	ipLit, ipLitErr := parseIpLiteral(runes)
	if ipLitErr == nil {
		return ipLit, nil
	}
	ipv4, _, ipv4Err := parseIpv4(runes)
	if ipv4Err == nil {
		return ipv4, nil
	}
	regHost, rhErr := parseRegHost(runes)
	if rhErr == nil {
		return regHost, nil
	}
	return "", errors.New("invalid host")
}

func parseRegHost(runes []rune) (string, error) {
	if len(runes) == 0 {
		return "", errors.New("no host name")
	}
	regHost := make([]rune, 0)
	i := 0
	for i < len(runes) {
		r := runes[i]
		i++
		if !isReserved(r) || isSubDelim(r) {
			regHost = append(regHost, r)
		} else {
			pctEncoded, pctErr := parsePctEncoded(runes[i : i+2])
			if pctErr == nil {
				// TODO is this right?
				continue
			} else {
				regHost = append(regHost, pctEncoded...)
			}
		}
	}
	if len(regHost) == 0 {
		return "", errors.New("no host name")
	}
	return string(regHost), nil
}

func parsePort() {

}

func parseIpLiteral(runes []rune) (string, error) {
	if len(runes) == 0 {
		return "", errors.New("no ip literal")
	}
	if runes[0] != '[' {
		return "", errors.New("invalid ip literal: missing '['")
	}
	ipv6, end, ipv6Err := parseIpv4(runes[1:])
	if ipv6Err == nil && runes[end-1] == ']' {
		return "[" + ipv6 + "]", nil
	}
	ipvf, end, ipvfErr := parseIpvFuture(runes[1:])
	if ipvfErr == nil && runes[end-1] == ']' {
		return "[" + ipvf + "]", nil
	}
	return "", errors.New("invalid ip-literal")
}

func parseIpv4(runes []rune) (string, int, error) {
	if len(runes) == 0 {
		return "", -1, errors.New("no ipv4")
	}

	ipv4 := make([]rune, 0)
	octet := make([]rune, 0)
	count := 0
	octetStart := false
	end := 0
	for i, r := range runes {
		end = i
		if r == '.' {
			ipv4 = append(ipv4, octet...)
			ipv4 = append(ipv4, r)
			count = count + 1
			octet = make([]rune, 0)
			continue
		} else {
			if !isDigit(r) {
				return "", -1, errors.New(fmt.Sprintf("invalid ip literal: invalid character %c", r))
			}
			if octetStart {
				octetLen := len(octet)
				if octetLen == 3 {
					return "", -1, errors.New("invalid ip literal: octet length")
				}
				octet = append(octet, r)
				parseInt, err := strconv.ParseInt(string(octet), 10, 32)
				if err != nil {
					return "", -1, errors.New(fmt.Sprintf("invalid ip literal: invalid octet %s", string(octet)))
				}
				if parseInt > 255 {
					return "", -1, errors.New(fmt.Sprintf("invalid ip literal: invalid octet %d", parseInt))
				}
			} else {
				if r == '1' || r == '2' {
					octet = append(octet, r)
				} else {
					return "", -1, errors.New("invalid ip literal")
				}
				octetStart = true
			}
		}
		if count == 4 {
			break
		}
	}
	return string(ipv4), end, nil
}

func parseIpv6(runes []rune) (string, int, error) {
	panic("not implemented")
}

func parseIpvFuture(runes []rune) (string, int, error) {
	panic("")
}

func parsePctEncoded(runes []rune) ([]rune, error) {
	var res []rune
	if len(runes) == 0 {
		return res, errors.New("no pct-encoding")
	}
	if runes[0] != '%' {
		return res, errors.New("invalid pct-encoding: missing '%'")
	}
	if len(runes) < 3 {
		return res, errors.New("invalid pct-encoding: too short")
	}
	res = append(res, '%')
	for _, r := range runes[1:2] {
		if !strings.ContainsRune(pctEncoding, r) {
			return []rune{}, errors.New(fmt.Sprintf("invalid pct-encoding: %c is not in range", r))
		}
		res = append(res, r)
	}

	return res, nil
}

func runeIndex(runes []rune, value rune) int {
	for i, r := range runes {
		if r == value {
			return i
		}
	}
	return -1
}

func isSubDelim(value rune) bool {
	return strings.ContainsRune(subDelims, value)
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
