package tbql

import "regexp"

var alphaRegexp = regexp.MustCompile("^[a-zA-Z]$")
var alphaNumRegexp = regexp.MustCompile("^[a-zA-Z0-9]$")
var numRegexp = regexp.MustCompile("^[0-9]$")
var dateRegexp = regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")

var keywords = []string{"AND", "OR", "NOT"}

func keyword(s string) bool {
	for _, k := range keywords {
		if s == k {
			return true
		}
	}
	return false
}

func isAlpha(c byte) bool {
	return alphaRegexp.MatchString(string(c))
}

func isAlphaNum(c byte) bool {
	return alphaNumRegexp.MatchString(string(c))
}

func isHyphen(c byte) bool {
	return c == '-'
}

func isNum(c byte) bool {
	return numRegexp.MatchString(string(c))
}

func isWhitespace(c byte) bool {
	return c == ' '
}

func isBinaryOperator(c byte) bool {
	return c == '&' || c == '|'
}

func isUnaryOperator(c byte) bool {
	return c == '!'
}

func isOpenParen(c byte) bool {
	return c == '('
}

func isCloseParen(c byte) bool {
	return c == ')'
}


