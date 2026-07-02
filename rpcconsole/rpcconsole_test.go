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

package rpcconsole_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/rpcconsole"
	"github.com/cgrates/cgrates/utils"
)

func TestBuildParams(t *testing.T) {
	md := &utils.MethodDescriptor{
		Args: []utils.FieldDescriptor{
			{Name: "Duration", Type: "duration"},
			{Name: "Tenant", Type: "string"},
			{Name: "Event", Type: "map[string]any"},
		},
	}
	got, err := rpcconsole.BuildParams(`Duration=1ms Tenant=cgrates.org Event={"a":{"b":1}} Count=5`, md)
	if err != nil {
		t.Fatal(err)
	}
	if got["Duration"] != time.Millisecond.Nanoseconds() {
		t.Errorf("Duration = %v, want %d", got["Duration"], time.Millisecond.Nanoseconds())
	}
	if got["Tenant"] != "cgrates.org" {
		t.Errorf("Tenant = %v, want cgrates.org", got["Tenant"])
	}
	// the nested object must survive whole, not be truncated at the first }
	if ev, _ := got["Event"].(json.RawMessage); string(ev) != `{"a":{"b":1}}` {
		t.Errorf("Event = %s, want {\"a\":{\"b\":1}}", ev)
	}
	if n, _ := got["Count"].(json.RawMessage); string(n) != "5" {
		t.Errorf("Count = %s, want 5", n)
	}
}

func TestFormat(t *testing.T) {
	md := &utils.MethodDescriptor{
		Result: []utils.FieldDescriptor{
			{Name: "Usage", Type: "duration"},
			{Name: "Account", Type: "string"},
			{Name: "Nested", Type: "object", Fields: []utils.FieldDescriptor{
				{Name: "TTL", Type: "duration"},
			}},
		},
	}
	reply := map[string]any{
		"Usage":   int64(90 * time.Minute),
		"Account": "1001",
		"Nested":  map[string]any{"TTL": int64(time.Second), "Usage": int64(5)},
	}
	// the nested TTL is rewritten, but Nested.Usage keeps its number though it shares a name.
	out := rpcconsole.Format(reply, md)
	for _, want := range []string{`"Usage": "1h30m0s"`, `"TTL": "1s"`, `"Account": "1001"`, `"Usage": 5`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in:\n%s", want, out)
		}
	}
}
