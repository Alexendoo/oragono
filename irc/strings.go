// Copyright (c) 2012-2014 Jeremy Latt
// Copyright (c) 2014-2015 Edmund Huber
// Copyright (c) 2016-2017 Daniel Oaks <daniel@danieloaks.net>
// released under the MIT license

package irc

import (
	"bytes"
	"strings"

	"golang.org/x/text/secure/precis"
)

const (
	casemappingName = "rfc8265"
)

func casefoldInternal(str string) (string, error) {
	var err error
	oldStr := str
	// follow the stabilizing rules laid out here:
	// https://tools.ietf.org/html/draft-ietf-precis-7564bis-10.html#section-7
	for i := 0; i < 4; i++ {
		str, err = precis.UsernameCaseMapped.CompareKey(str)
		if err != nil {
			return "", err
		}
		if oldStr == str {
			break
		}
		oldStr = str
	}
	if oldStr != str {
		return "", errCouldNotStabilize
	}
	return str, nil
}

func isEmoji(x rune) bool {
	return emojiSet[x]
}

// Casefold returns a casefolded string, without doing any name or channel character checks.
func Casefold(str string) (string, error) {
	var result bytes.Buffer
	var curWord bytes.Buffer
	for _, r := range str {
		if isEmoji(r) {
			prevWord, err := casefoldInternal(curWord.String())
			if err != nil {
				return "", err
			}
			result.WriteString(prevWord)
			curWord.Reset()
			result.WriteRune(r)
		} else {
			curWord.WriteRune(r)
		}
	}
	prevWord, err := casefoldInternal(curWord.String())
	if err != nil {
		return "", err
	}
	result.WriteString(prevWord)
	return result.String(), nil
}

// CasefoldChannel returns a casefolded version of a channel name.
func CasefoldChannel(name string) (string, error) {
	lowered, err := Casefold(name)

	if err != nil {
		return "", err
	} else if len(lowered) == 0 {
		return "", errStringIsEmpty
	}

	if lowered[0] != '#' {
		return "", errInvalidCharacter
	}

	// space can't be used
	// , is used as a separator
	// * is used in mask matching
	// ? is used in mask matching
	if strings.ContainsAny(lowered, " ,*?") {
		return "", errInvalidCharacter
	}

	return lowered, err
}

// CasefoldName returns a casefolded version of a nick/user name.
func CasefoldName(name string) (string, error) {
	lowered, err := Casefold(name)

	if err != nil {
		return "", err
	} else if len(lowered) == 0 {
		return "", errStringIsEmpty
	}

	// space can't be used
	// , is used as a separator
	// * is used in mask matching
	// ? is used in mask matching
	// . denotes a server name
	// ! separates nickname from username
	// @ separates username from hostname
	// : means trailing
	// # is a channel prefix
	// ~&@%+ are channel membership prefixes
	// - I feel like disallowing
	if strings.ContainsAny(lowered, " ,*?.!@:") || strings.ContainsAny(string(lowered[0]), "#~&@%+-") {
		return "", errInvalidCharacter
	}

	return lowered, err
}
