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
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// ArgSignature returns the method's call form, "Method Key= Key= ...". Empty when md is nil.
func ArgSignature(md *utils.MethodDescriptor) string {
	if md == nil {
		return ""
	}
	parts := make([]string, 0, len(md.Args)+1)
	parts = append(parts, md.Method)
	for _, f := range md.Args {
		parts = append(parts, f.Name+"=")
	}
	return strings.Join(parts, " ")
}

// ArgTree renders md's arguments as an indented tree, expanding nested fields to
// full depth. The descriptor already breaks cycles, so the recursion terminates.
func ArgTree(md *utils.MethodDescriptor) string {
	if md == nil {
		return ""
	}
	var b strings.Builder
	writeFieldTree(&b, md.Args, 1, -1)
	return b.String()
}

// InnerFields renders one level of fd's nested fields, the hint shown next to a
// value being completed.
func InnerFields(fd utils.FieldDescriptor) string {
	var b strings.Builder
	writeFieldTree(&b, fd.Fields, 1, 1)
	return b.String()
}

// writeFieldTree writes fields indented by depth. A negative maxDepth means no limit.
func writeFieldTree(b *strings.Builder, fields []utils.FieldDescriptor, depth, maxDepth int) {
	if maxDepth >= 0 && depth > maxDepth {
		return
	}
	indent := strings.Repeat("  ", depth)
	for _, f := range fields {
		b.WriteString(indent)
		b.WriteString(f.Name)
		b.WriteString(": ")
		b.WriteString(f.Type)
		b.WriteString("\n")
		writeFieldTree(b, f.Fields, depth+1, maxDepth)
	}
}

// Format renders an RPC reply as indented JSON, rewriting fields the descriptor
// types as durations from a nanosecond count to a string. md may be nil.
func Format(reply any, md *utils.MethodDescriptor) string {
	b, err := json.Marshal(reply)
	if err != nil {
		return fmt.Sprintf("%v", reply)
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return string(b)
	}
	if md != nil {
		formatDurations(v, md.Result)
	}
	out, _ := json.MarshalIndent(v, "", " ")
	return string(out)
}

// formatDurations rewrites each field the result descriptors type as a duration
// from a nanosecond count to a string.
func formatDurations(v any, fields []utils.FieldDescriptor) {
	switch val := v.(type) {
	case map[string]any:
		for _, f := range fields {
			child, ok := val[f.Name]
			if !ok {
				continue
			}
			if f.Type == "duration" {
				if d, err := utils.IfaceAsDuration(child); err == nil {
					val[f.Name] = d.String()
				}
				continue
			}
			formatDurations(child, f.Fields)
		}
	case []any:
		for _, child := range val {
			formatDurations(child, fields)
		}
	}
}
