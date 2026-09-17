package ntql

import "regexp"

// Regexps for various token types
var alphaRegexp = regexp.MustCompile("^[a-zA-Z]$")
var alphaNumRegexp = regexp.MustCompile("^[a-zA-Z0-9]$")
var numRegexp = regexp.MustCompile("^[0-9]$")
var dateRegexp = regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")
var dateTimeRegexp = regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}$")
var stringRegexp = regexp.MustCompile("^\".*\"$")

// objectType is a list of all the object types a token could possible represent
var objectTypes = []TokenType{TokenInt, TokenString, TokenBool, TokenDate, TokenDateTime, TokenTag}

func isSymbol(c byte) bool {
	return c == '!' || c == '(' || c == ')' || c == '.'
}
