// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func setInput(value *formValue, input string) {
	value.input.SetValue(input)
	value.set = true
}

func TestFormValueBuildNestedCollections(t *testing.T) {
	root := newValue(utils.FieldDescriptor{Type: "object", Fields: []utils.FieldDescriptor{
		{
			Name: "Items", Type: "[]test.Item", Fields: []utils.FieldDescriptor{
				{Name: "Name", Type: "string"},
				{Name: "Tags", Type: "[]string"},
				{Name: "Attrs", Type: "map[string]any"},
			},
		},
		{Name: "EmptyList", Type: "[]string"},
		{Name: "EmptyMap", Type: "map[string]string"},
		{Name: "Omitted", Type: "[]string"},
	}})
	root.set = true

	items := root.fields[0]
	items.set = true
	item := newValue(elementField(items.field))
	item.set = true
	setInput(item.fields[0], "first")

	tags := item.fields[1]
	tags.set = true
	tag := newValue(elementField(tags.field))
	setInput(tag, "one")
	tags.items = append(tags.items, tag)

	attrs := item.fields[2]
	attrs.set = true
	attr := newEntry(attrs.field)
	setInput(attr.fields[0], "flag")
	setInput(attr.fields[1], `"true"`)
	attrs.items = append(attrs.items, attr)
	items.items = append(items.items, item)
	if got := item.summary(); got != "Name=first" {
		t.Fatalf("item summary = %q", got)
	}

	root.fields[1].set = true
	root.fields[2].set = true
	got, present, err := root.build()
	if err != nil {
		t.Fatal(err)
	}
	if !present {
		t.Fatal("root is unset")
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"EmptyList":[],"EmptyMap":{},"Items":[{"Attrs":{"flag":"true"},"Name":"first","Tags":["one"]}]}`
	if string(encoded) != want {
		t.Fatalf("got %s, want %s", encoded, want)
	}
}

func TestNestedEditingKeepsAnyJSON(t *testing.T) {
	root := newValue(utils.FieldDescriptor{Type: "object", Fields: []utils.FieldDescriptor{
		{Name: "Values", Type: "map[string]any"},
	}})
	root.set = true
	m := model{forms: []formFrame{{path: "test.edit", value: root}}}

	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(model)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = next.(model)
	entry := m.frame().value
	setInput(entry.fields[0], "flag")
	setInput(entry.fields[1], `"true"`)

	next, _ = m.closeFrame()
	m = next.(model)
	next, _ = m.closeFrame()
	m = next.(model)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(model)
	next, _ = m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(model)
	if got := m.frame().value.fields[1].input.Value(); got != `"true"` {
		t.Fatalf("value changed to %q", got)
	}

	got, _, err := root.build()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"Values":{"flag":"true"}}` {
		t.Fatalf("got %s", encoded)
	}
}

func TestMalformedFallbackJSONIsRetained(t *testing.T) {
	for _, typ := range []string{"any", "test.Recursive"} {
		t.Run(typ, func(t *testing.T) {
			value := newValue(utils.FieldDescriptor{Name: "Value", Type: typ})
			setInput(value, `{"broken":`)
			if _, _, err := value.build(); err == nil {
				t.Fatal("expected malformed JSON error")
			}
			if got := value.input.Value(); got != `{"broken":` {
				t.Fatalf("value changed to %q", got)
			}
		})
	}

	nested := newValue(utils.FieldDescriptor{
		Name: "Nested", Type: "test.Nested",
		Fields: []utils.FieldDescriptor{{Name: "Value", Type: "string"}},
	})
	if nested.input != nil || nested.kind != objectKind {
		t.Fatal("described nested struct should use generated fields")
	}
}

func TestEnterDoesNotSubmit(t *testing.T) {
	root := newValue(utils.FieldDescriptor{Type: "object", Fields: []utils.FieldDescriptor{
		{Name: "Value", Type: "string"},
	}})
	root.set = true
	m := model{view: viewForm, forms: []formFrame{{path: "test.edit", value: root}}}
	next, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(model)
	if m.calling {
		t.Fatal("enter submitted the form")
	}
	if !strings.Contains(m.View(), "ctrl+s run") {
		t.Fatal("form does not advertise explicit submission")
	}
}
