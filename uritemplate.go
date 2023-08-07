// Copyright 2023 Ross Light
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//		 https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package uritemplate

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Expand(template string, data any) (string, error) {
	sb := new(strings.Builder)
	dataValue := reflect.ValueOf(data)
	var firstError error
	for i := 0; i < len(template); {
		c, size := utf8.DecodeRuneInString(template[i:])
		switch {
		case isLiteral(c):
			if literalNeedsPercentEscape(c) {
				percentEscape(sb, template[i:i+size])
			} else {
				sb.WriteString(template[i : i+size])
			}
			i += size
		case c == '{':
			exprLen, err := expandExpression(sb, template[i:], dataValue)
			if err != nil && firstError == nil {
				firstError = fmt.Errorf("expand uri template %q: %w", template, err)
			}
			i += exprLen
		case c == '%':
			seq, _, ok := cutPercentEscape(template[i:])
			if !ok && firstError == nil {
				firstError = fmt.Errorf("expand uri template %q: invalid percent escape %q", template, seq)
			}
			i += len(seq)
		default:
			if firstError == nil {
				firstError = fmt.Errorf("expand uri template %q: illegal character %q", template, c)
			}
			i += size
		}
	}
	return sb.String(), firstError
}

func expandExpression(sb *strings.Builder, expr string, data reflect.Value) (exprLen int, err error) {
	end := strings.IndexByte(expr, '}')
	if end < 0 {
		sb.WriteString(expr)
		return len(expr), errors.New("unterminated expression")
	}
	exprLen = end + 1
	rest := strings.TrimPrefix(expr[:end], "{")

	var op byte
	const reservedOps = "=,!@|"
	if len(rest) > 0 && strings.IndexByte("+#./;?&"+reservedOps, rest[0]) != -1 {
		op = rest[0]
		rest = rest[1:]
	}

	if rest == "" {
		sb.WriteString(expr[:exprLen])
		return exprLen, errors.New("empty expression")
	}
	if strings.IndexByte(reservedOps, op) != -1 {
		sb.WriteString(expr[:exprLen])
		return exprLen, fmt.Errorf("expression %q: unknown operator %q", expr, op)
	}
	varName, modifier, rest := cutVarSpec(rest)
	if varName == "" {
		sb.WriteString(expr[:exprLen])
		return exprLen, fmt.Errorf("expression %q: missing variable name", expr)
	}
	first, err := expandVariable(sb, op, true, data, varName, modifier)
	if err != nil {
		writeRemainingExpression(sb, op, rest)
		return exprLen, fmt.Errorf("expression %q: %v", expr, err)
	}

	for len(rest) > 0 {
		if rest[0] != ',' {
			writeRemainingExpression(sb, op, rest)
			return exprLen, fmt.Errorf("expression %q: unexpected character %q", expr, rest[0])
		}
		rest = rest[1:]

		varName, modifier, rest = cutVarSpec(rest)
		if varName == "" {
			writeRemainingExpression(sb, op, rest)
			return exprLen, fmt.Errorf("expression %q: missing variable name", expr)
		}
		first, err = expandVariable(sb, op, first, data, varName, modifier)
		if err != nil {
			writeRemainingExpression(sb, op, rest)
			return exprLen, fmt.Errorf("expression %q: %v", expr, err)
		}
	}

	return exprLen, nil
}

func writeRemainingExpression(sb *strings.Builder, op byte, rest string) {
	if rest == "" {
		return
	}
	sb.WriteString("{")
	if op != 0 {
		sb.WriteByte(op)
	}
	sb.WriteString(rest)
	sb.WriteString("}")
}

func cutVarSpec(expr string) (varName, modifier, rest string) {
	// Parse varname.
	first, rest := cutVarChar(expr)
	if first == "" {
		return "", "", expr
	}

	for len(rest) > 0 {
		if rest[0] == '.' {
			next, possibleRest := cutVarChar(rest[1:])
			if next == "" {
				return expr[:len(expr)-len(rest)], "", rest
			}
			rest = possibleRest
			continue
		}
		var next string
		next, rest = cutVarChar(rest)
		if next == "" {
			break
		}
	}
	varName = expr[:len(expr)-len(rest)]

	// Parse modifier.
	if len(rest) == 0 {
		return varName, "", ""
	}
	switch rest[0] {
	case '*':
		return varName, rest[:1], rest[1:]
	case ':':
		if len(rest) < 2 || rest[1] == '0' || !isDigit(rune(rest[1])) {
			return varName, "", rest
		}
		n := 2
		for n < 5 && n < len(rest) && isDigit(rune(rest[n])) {
			n++
		}
		return varName, rest[:n], rest[n:]
	default:
		return varName, "", rest
	}
}

func cutVarChar(s string) (vc, rest string) {
	if len(s) == 0 {
		return "", s
	}
	switch {
	case s[0] == '%':
		vc, rest, ok := cutPercentEscape(s)
		if !ok {
			return "", s
		}
		return vc, rest
	case isAlpha(rune(s[0])) || isDigit(rune(s[0])) || s[0] == '_':
		return s[:1], s[1:]
	default:
		return "", s
	}
}

func cutPercentEscape(s string) (pct, rest string, ok bool) {
	if len(s) == 0 || s[0] != '%' {
		return "", s, false
	}
	const escapeLen = len("%FF")
	if len(s) < escapeLen {
		return s, "", false
	}
	return s[:escapeLen], s[escapeLen:], isHex(s[1]) && isHex(s[2])
}

func percentEscape(sb *strings.Builder, s string) {
	for _, b := range []byte(s) {
		sb.WriteByte('%')
		sb.WriteByte(upperHex(b >> 4))
		sb.WriteByte(upperHex(b & 0x0f))
	}
}

func isLiteral(c rune) bool {
	return !strings.ContainsRune(" \"'%<>\\^`{|}", c) &&
		!unicode.IsControl(c)
}

func isAlpha(c rune) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z'
}

func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

func isUnreserved(c rune) bool {
	return isAlpha(c) || isDigit(c) || strings.ContainsRune(`-._~`, c)
}

func isReserved(c rune) bool {
	return strings.ContainsRune(`:/?#[]@!$&'()*+,;=`, c)
}

func literalNeedsPercentEscape(c rune) bool {
	return !isUnreserved(c) && !isReserved(c)
}

func upperHex(x byte) byte {
	if x >= 0xa {
		return 'A' + (x - 0xa)
	}
	return '0' + x
}

func isHex(c byte) bool {
	return isDigit(rune(c)) || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}
