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
	index := 2
	runes = runes[index:]
	authority := ""
	userInfo, i, uiErr := parseUserInfo(runes)
	if uiErr != nil {
		if runes[3] != '@' {
			return "", errors.New(fmt.Sprintf("invalid authority: %s", uiErr.Error()))
		}
		authority = userInfo + "@"
		runes = runes[i:]
	}
	index += i
	host, end, hostErr := parseHost(runes)
	if hostErr == nil {
		authority = authority + host
	}
	index += end
	if runes[index] == ':' {
		parsePort(runes)
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

func parseHost(runes []rune) (string, int, error) {
	index := 0
	ipLit, end, ipLitErr := parseIpLiteral(runes)
	if ipLitErr == nil {
		index += end
		return ipLit, index, nil
	}
	ipv4, end, ipv4Err := parseIpv4(runes)
	if ipv4Err == nil {
		index += end
		return ipv4, index, nil
	}
	regHost, end, rhErr := parseRegHost(runes)
	if rhErr == nil {
		index += end
		return regHost, index, nil
	}
	return "", -1, errors.New("invalid host")
}

func parseRegHost(runes []rune) (string, int, error) {
	if len(runes) == 0 {
		return "", -1, errors.New("no host name")
	}
	index := 0
	regHost := make([]rune, 0)
	for index < len(runes) {
		r := runes[index]
		index++
		if !isReserved(r) || isSubDelim(r) {
			regHost = append(regHost, r)
		} else {
			pctEncoded, pctErr := parsePctEncoded(runes[index : index+2])
			if pctErr == nil {
				// TODO is this right?
				continue
			} else {
				regHost = append(regHost, pctEncoded...)
			}
		}
	}
	if len(regHost) == 0 {
		return "", -2, errors.New("no host name")
	}
	return string(regHost), index, nil
}

func parsePort(runes []rune) (string, int, error) {
	if len(runes) == 0 {
		return "", -1, errors.New("no port")
	}
	index := 0
	port := make([]rune, 0)
	for i, r := range runes {
		if isDigit(r) {
			if i == 0 && r == '0' {
				return "", -1, errors.New("invalid port: 0 is reserved")
			}
			index++
			port = append(port, r)
			continue
		}
		if i == 0 {
			return "", -1, errors.New("invalid port: must be between 1 and 65535")
		}
		if i >= 5 {
			return "", -1, errors.New("invalid port: exceeds range")
		}
		break
	}
	parsed, _ := strconv.ParseInt(string(port), 10, 32)
	if parsed > 65535 {
		return "", -1, errors.New("invalid port: exceeds range")
	}
	return string(port), index, nil
}

func parseIpLiteral(runes []rune) (string, int, error) {
	index := 0
	if len(runes) == 0 {
		return "", -1, errors.New("no ip literal")
	}
	if runes[index] != '[' {
		return "", -1, errors.New("invalid ip literal: missing '['")
	}
	ipv6, end, ipv6Err := parseIpv4(runes[1:])
	if ipv6Err == nil && runes[end-1] == ']' {
		index += end
		return "[" + ipv6 + "]", index, nil
	}
	ipvf, end, ipvfErr := parseIpvFuture(runes[1:])
	if ipvfErr == nil && runes[end-1] == ']' {
		index += end
		return "[" + ipvf + "]", index, nil
	}
	return "", -1, errors.New("invalid ip-literal")
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
	if len(runes) == 0 {
		return "", -1, errors.New("no ipv future")
	}
	if runes[0] != 'v' {
		return "", -1, errors.New("invalid ipv future: missing leading 'v'")
	}
	sub := runes[1:]
	index := 1
	for i, r := range sub {
		index++
		if r == '.' {
			if i < 1 {
				return "", -1, errors.New("invalid ipv future: misplaced '.'")
			}
			break
		}
		if !isHexDigit(r) {
			return "", -1, errors.New(fmt.Sprintf("invalid ipv future: index %d", index))
		}
	}
	sub = sub[index-1:]
	for i, r := range sub {
		index++
		if isUnreserved(r) || isSubDelim(r) || r == ':' {
			continue
		}
		if i == 0 {
			return "", -1, errors.New(fmt.Sprintf("invalid ipv future: index %d", index))
		}
	}
	return string(runes[0:index]), index, nil
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
		if !isHexDigit(r) {
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

func isHexDigit(value rune) bool {
	return strings.ContainsRune(pctEncoding, value)
}

func isSubDelim(value rune) bool {
	return strings.ContainsRune(subDelims, value)
}

func isReserved(value rune) bool {
	return strings.ContainsRune(genDelims, value) || strings.ContainsRune(subDelims, value)
}

func isUnreserved(value rune) bool {
	return !isReserved(value)
}

func isAlpha(value rune) bool {
	return strings.ContainsRune(alpha, value)
}

func isDigit(value rune) bool {
	return strings.ContainsRune(digit, value)
}
