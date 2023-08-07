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
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

func expandVariable(sb *strings.Builder, op byte, first bool, data reflect.Value, varName, modifier string) (stillFirst bool, err error) {
	vk, val := kindOf(lookupKey(data, varName))
	if vk == 0 {
		return first, nil
	}

	sep := opSep(op)
	if first {
		if op != 0 && op != '+' {
			sb.WriteByte(op)
		}
	} else {
		sb.WriteByte(sep)
	}

	switch {
	case vk == scalarKind:
		s, err := coerceString(val)
		writeVarNamePrefix(sb, op, varName, s == "")
		if err != nil {
			return false, err
		}
		s = modify(s, modifier)
		writeValue(sb, op, s)
	case vk == listKind && modifier != "*":
		empty := isEmpty(val)
		writeVarNamePrefix(sb, op, varName, empty)
		if !empty {
			for i, n, defined := 0, val.Len(), false; i < n; i++ {
				elemValue, _ := followIndirection(val.Index(i))
				if !elemValue.IsValid() {
					continue
				}
				s, err := coerceString(elemValue)
				if err != nil {
					return false, err
				}

				if defined {
					sb.WriteByte(',')
				}
				writeValue(sb, op, s)
				defined = true
			}
		}
	case vk == mapKind && modifier != "*":
		empty := isEmpty(val)
		writeVarNamePrefix(sb, op, varName, empty)
		if !empty {
			defined := false
			var err error
			iterateMap(val, func(k string, elemValue reflect.Value) bool {
				elemValue, _ = followIndirection(elemValue)
				if !elemValue.IsValid() {
					return true
				}
				var s string
				s, err = coerceString(elemValue)
				if err != nil {
					return false
				}

				if defined {
					sb.WriteByte(',')
				}
				writeValue(sb, op, k)
				sb.WriteByte(',')
				writeValue(sb, op, s)
				defined = true
				return true
			})
			if err != nil {
				return false, err
			}
		}
	case vk == listKind && modifier == "*":
		for i, n, defined := 0, val.Len(), false; i < n; i++ {
			elemValue, _ := followIndirection(val.Index(i))
			if !elemValue.IsValid() {
				continue
			}
			s, err := coerceString(elemValue)
			if err != nil {
				return false, err
			}

			if defined {
				sb.WriteByte(sep)
			}
			writeVarNamePrefix(sb, op, varName, s == "")
			writeValue(sb, op, s)
			defined = true
		}
	case vk == mapKind && modifier == "*":
		defined := false
		var err error
		iterateMap(val, func(k string, elemValue reflect.Value) bool {
			elemValue, _ = followIndirection(elemValue)
			if !elemValue.IsValid() {
				return true
			}
			var s string
			s, err = coerceString(elemValue)
			if err != nil {
				return false
			}

			if defined {
				sb.WriteByte(sep)
			}
			if opUsesNames(op) {
				writeVarNamePrefix(sb, op, k, s == "")
			} else {
				writeValue(sb, op, k)
				sb.WriteString("=")
			}
			writeValue(sb, op, s)
			defined = true
			return true
		})
		if err != nil {
			return false, err
		}
	default:
		panic("unreachable")
	}

	return false, nil
}

var keyStringPool = sync.Pool{
	New: func() any {
		v := reflect.New(stringType)
		return &v
	},
}

func lookupKey(composite reflect.Value, key string) reflect.Value {
	if !composite.IsValid() {
		return reflect.Value{}
	}
	for {
		if k := composite.Kind(); k != reflect.Pointer && k != reflect.Interface {
			break
		}
		if composite.IsNil() {
			return reflect.Value{}
		}
		composite = composite.Elem()
	}

	switch composite.Kind() {
	case reflect.Map:
		keyType := composite.Type().Key()
		if keyType.Kind() != reflect.String {
			return reflect.Value{}
		}
		keyReflectPtrValue := keyStringPool.Get().(*reflect.Value)
		keyReflectValue := keyReflectPtrValue.Elem()
		keyReflectValue.SetString(key)
		var result reflect.Value
		if keyType == stringType {
			result = composite.MapIndex(keyReflectValue)
		} else {
			result = composite.MapIndex(keyReflectValue.Convert(keyType))
		}
		keyReflectValue.SetString("")
		keyStringPool.Put(keyReflectPtrValue)
		return result
	case reflect.Struct:
		sd := describeStruct(composite.Type())
		i, ok := sd.indexLookup[key]
		if !ok {
			return reflect.Value{}
		}
		return composite.Field(i)
	default:
		return reflect.Value{}
	}
}

func writeVarNamePrefix(sb *strings.Builder, op byte, varName string, empty bool) {
	if !opUsesNames(op) {
		return
	}
	for len(varName) > 0 {
		if pct, rest, ok := cutPercentEscape(varName); ok {
			sb.WriteString(pct)
			varName = rest
			continue
		}
		c, size := utf8.DecodeRuneInString(varName)
		if literalNeedsPercentEscape(c) {
			percentEscape(sb, varName[:size])
		} else {
			sb.WriteString(varName[:size])
		}
		varName = varName[size:]
	}
	if !empty || op == '?' || op == '&' {
		sb.WriteString("=")
	}
}

func opUsesNames(op byte) bool {
	return op == ';' || op == '?' || op == '&'
}

func opSep(op byte) byte {
	switch op {
	case 0, '+', '#':
		return ','
	case '.', '/', ';':
		return op
	case '?', '&':
		return '&'
	default:
		panic("unreachable")
	}
}

func coerceString(val reflect.Value) (string, error) {
	if !val.IsValid() {
		return "", errors.New("undefined value")
	}
	typ := val.Type()
	switch {
	case typ.Implements(textMarshalerType):
		data, err := val.Interface().(encoding.TextMarshaler).MarshalText()
		return string(data), err
	case typ.Kind() == reflect.String && !(typ.Implements(stringerType) || typ.Implements(errorType) || typ.Implements(formatterType)):
		return val.String(), nil
	default:
		return fmt.Sprint(val), nil
	}
}

func writeValue(sb *strings.Builder, op byte, s string) {
	if op == '+' || op == '#' {
		for len(s) > 0 {
			if pct, _, ok := cutPercentEscape(s); ok {
				sb.WriteString(pct)
				s = s[len(pct):]
				continue
			}
			c, size := utf8.DecodeRuneInString(s)
			if isUnreserved(c) || isReserved(c) {
				sb.WriteString(s[:size])
			} else {
				percentEscape(sb, s[:size])
			}
			s = s[size:]
		}
	} else {
		for len(s) > 0 {
			c, size := utf8.DecodeRuneInString(s)
			if isUnreserved(c) {
				sb.WriteString(s[:size])
			} else {
				percentEscape(sb, s[:size])
			}
			s = s[size:]
		}
	}
}

func modify(s string, modifier string) string {
	if !strings.HasPrefix(modifier, ":") {
		return s
	}
	n, err := strconv.Atoi(modifier[1:])
	if err != nil || n <= 0 {
		return s
	}
	pos := 0
	for i := 0; i < n && pos < len(s); i++ {
		_, size := utf8.DecodeRuneInString(s[pos:])
		pos += size
	}
	return s[:pos]
}

// isEmpty reports whether v is undefined, an empty string,
// or an array/slice/map with no defined elements.
func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.String:
		return v.Len() == 0
	case reflect.Map:
		found := false
		iterateMap(v, func(k string, elem reflect.Value) bool {
			elem, _ = followIndirection(elem)
			if elem.IsValid() {
				found = true
				return false
			}
			return true
		})
		return !found
	case reflect.Slice, reflect.Array:
		for i, n := 0, v.Len(); i < n; i++ {
			elem, _ := followIndirection(v.Index(i))
			if elem.IsValid() {
				return false
			}
		}
		return true
	default:
		return false
	}
}

type varKind int

const (
	scalarKind varKind = 1 + iota
	mapKind
	listKind
)

func kindOf(v reflect.Value) (varKind, reflect.Value) {
	v, scalar := followIndirection(v)
	switch {
	case !v.IsValid():
		return 0, reflect.Value{}
	case !scalar && ((v.Kind() == reflect.Map && v.Type().Key().Kind() == reflect.String) || v.Kind() == reflect.Struct):
		return mapKind, v
	case !scalar && (v.Kind() == reflect.Slice || v.Kind() == reflect.Array):
		return listKind, v
	default:
		return scalarKind, v
	}
}

func followIndirection(v reflect.Value) (_ reflect.Value, scalar bool) {
	for {
		if !v.IsValid() {
			return reflect.Value{}, false
		}

		typ := v.Type()
		k := typ.Kind()
		switch {
		case typ.Implements(stringerType) || typ.Implements(errorType) || typ.Implements(textMarshalerType) || typ.Implements(formatterType):
			return v, true
		case k != reflect.Pointer && k != reflect.Interface:
			return v, false
		}
		if v.IsNil() {
			return reflect.Value{}, false
		}
		v = v.Elem()
	}
}

func iterateMap(m reflect.Value, f func(k string, v reflect.Value) bool) {
	switch m.Kind() {
	case reflect.Map:
		keys := m.MapKeys()
		// For mapKind, keys are guaranteed to have an underlying type of string.
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			if !f(k.String(), m.MapIndex(k)) {
				break
			}
		}
	case reflect.Struct:
		sd := describeStruct(m.Type())
		for i, name := range sd.fieldNames {
			if name == "" {
				continue
			}
			if !f(name, m.Field(i)) {
				break
			}
		}
	default:
		panic("unreachable")
	}
}

var descriptors sync.Map

type structDescriptor struct {
	fieldNames  []string
	indexLookup map[string]int
}

func describeStruct(t reflect.Type) structDescriptor {
	if sd, ok := descriptors.Load(t); ok {
		return sd.(structDescriptor)
	}
	sd := structDescriptor{
		fieldNames:  make([]string, t.NumField()),
		indexLookup: make(map[string]int),
	}
	for i := range sd.fieldNames {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		var fieldName string
		if tag := field.Tag.Get("uritemplate"); tag == "-" {
			continue
		} else if tag != "" {
			fieldName = tag
		} else {
			_, firstRuneSize := utf8.DecodeRuneInString(field.Name)
			fieldName = strings.ToLower(field.Name[:firstRuneSize]) + field.Name[firstRuneSize:]
		}
		sd.fieldNames[i] = fieldName
		sd.indexLookup[fieldName] = i
	}
	descriptors.Store(t, sd)
	return sd
}

var (
	errorType         = reflect.TypeOf((*error)(nil)).Elem()
	formatterType     = reflect.TypeOf((*fmt.Formatter)(nil)).Elem()
	stringType        = reflect.TypeOf((*string)(nil)).Elem()
	stringerType      = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)
