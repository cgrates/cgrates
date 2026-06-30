/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package rpcconsole

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// BuildParams turns a "Key=Value ..." string into the JSON params map, parsing each
// value by its type (durations to nanoseconds, bare strings unquoted, valid JSON kept
// as is). A @path value reads the file. If the whole argument is a {...} object or
// @path, it is decoded as the full request instead.
func BuildParams(argStr string, md *utils.MethodDescriptor) (map[string]any, error) {
	argStr = strings.TrimSpace(argStr)
	if strings.HasPrefix(argStr, "{") || strings.HasPrefix(argStr, "@") {
		return wholeRequest(argStr)
	}
	params := map[string]any{}
	for _, tok := range SplitArgs(argStr) {
		key, raw, ok := strings.Cut(tok, "=")
		if !ok {
			return nil, fmt.Errorf("invalid argument %q, expected Key=Value", tok)
		}
		val, err := parseValue(FieldType(md, key), raw)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", key, err)
		}
		params[key] = val
	}
	return params, nil
}

// wholeRequest decodes the whole argument as one JSON object, given as a {...}
// literal or an @path file.
func wholeRequest(argStr string) (map[string]any, error) {
	data := []byte(argStr)
	if argStr[0] == '@' {
		var err error
		if data, err = os.ReadFile(argStr[1:]); err != nil {
			return nil, err
		}
	}
	var params map[string]any
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, fmt.Errorf("invalid request object: %w", err)
	}
	return params, nil
}

func parseValue(typ, raw string) (any, error) {
	if strings.HasPrefix(raw, "@") {
		data, err := os.ReadFile(raw[1:])
		if err != nil {
			return nil, err
		}
		if json.Valid(data) {
			return json.RawMessage(data), nil
		}
		return strings.TrimRight(string(data), "\r\n"), nil
	}
	switch typ {
	case "duration":
		d, err := utils.ParseDurationWithNanosecs(strings.Trim(raw, `"`))
		if err != nil {
			return nil, err
		}
		return d.Nanoseconds(), nil
	case "string":
		return strings.Trim(raw, `"`), nil
	default:
		if json.Valid([]byte(raw)) {
			return json.RawMessage(raw), nil
		}
		return raw, nil
	}
}

// SplitArgs splits a "Key=Value ..." line on whitespace outside any {...}, [...] or
// "..." value, so those keep their spaces. Completion tokenizes the same way.
func SplitArgs(s string) []string {
	var tokens []string
	depth := 0
	inStr := false
	start := -1
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			switch c {
			case '\\':
				i++
			case '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{', '[':
			depth++
		case '}', ']':
			if depth > 0 {
				depth--
			}
		case ' ', '\t':
			if depth == 0 {
				if start >= 0 {
					tokens = append(tokens, s[start:i])
					start = -1
				}
				continue
			}
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		tokens = append(tokens, s[start:])
	}
	return tokens
}

// FieldType returns the type of md's argument field named name, or "" when md is
// nil or has no such field.
func FieldType(md *utils.MethodDescriptor, name string) string {
	if fd := ArgField(md, name); fd != nil {
		return fd.Type
	}
	return ""
}

// ArgField returns md's argument field named name, or nil if there is none.
func ArgField(md *utils.MethodDescriptor, name string) *utils.FieldDescriptor {
	if md == nil {
		return nil
	}
	for i := range md.Args {
		if md.Args[i].Name == name {
			return &md.Args[i]
		}
	}
	return nil
}
