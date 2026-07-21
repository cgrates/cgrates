// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/rpcconsole"
	"github.com/cgrates/cgrates/utils"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type valueKind uint8

const (
	scalarKind valueKind = iota
	objectKind
	listKind
	mapKind
)

type formValue struct {
	field  utils.FieldDescriptor
	kind   valueKind
	set    bool
	input  *textinput.Model
	fields []*formValue
	items  []*formValue
}

type formFrame struct {
	path  string
	value *formValue
	focus int
}

func newValue(field utils.FieldDescriptor) *formValue {
	value := &formValue{field: field, kind: fieldKind(field)}
	switch value.kind {
	case objectKind:
		value.fields = make([]*formValue, len(field.Fields))
		for i, child := range field.Fields {
			value.fields[i] = newValue(child)
		}
	case scalarKind:
		input := textinput.New()
		input.Prompt = fmt.Sprintf("  %s = ", field.Name)
		input.Placeholder = field.Type
		input.Width = 64
		value.input = &input
	}
	return value
}

func fieldKind(field utils.FieldDescriptor) valueKind {
	switch {
	case strings.HasPrefix(field.Type, "[]"):
		return listKind
	case strings.HasPrefix(field.Type, "map["):
		return mapKind
	case field.Type == "object" || len(field.Fields) != 0:
		return objectKind
	default:
		return scalarKind
	}
}

func elementField(field utils.FieldDescriptor) utils.FieldDescriptor {
	typ := strings.TrimPrefix(field.Type, "[]")
	if typ == field.Type {
		_, typ = mapTypes(typ)
	}
	return utils.FieldDescriptor{Name: "Value", Type: typ, Fields: field.Fields}
}

func mapTypes(typ string) (string, string) {
	rest := strings.TrimPrefix(typ, "map[")
	if end := strings.IndexByte(rest, ']'); end >= 0 {
		return rest[:end], rest[end+1:]
	}
	return "string", "any"
}

func newEntry(field utils.FieldDescriptor) *formValue {
	keyType, valueType := mapTypes(field.Type)
	entry := newValue(utils.FieldDescriptor{Type: "object", Fields: []utils.FieldDescriptor{
		{Name: "Key", Type: keyType},
		{Name: "Value", Type: valueType, Fields: field.Fields},
	}})
	entry.set, entry.fields[1].set = true, true
	return entry
}

func (value *formValue) resize(width int) {
	if value.input != nil {
		value.input.Width = max(width-len(value.input.Prompt)-1, 8)
	}
	for _, child := range value.fields {
		child.resize(width)
	}
	for _, item := range value.items {
		item.resize(width)
	}
}

func (value *formValue) build() (any, bool, error) {
	if !value.set {
		return nil, false, nil
	}
	switch value.kind {
	case objectKind:
		object := make(map[string]any)
		for _, child := range value.fields {
			built, present, err := child.build()
			if err != nil {
				return nil, false, fmt.Errorf("field %s: %w", child.field.Name, err)
			}
			if present {
				object[child.field.Name] = built
			}
		}
		return object, true, nil
	case listKind:
		list := make([]any, len(value.items))
		for i, item := range value.items {
			built, present, err := item.build()
			if err != nil {
				return nil, false, fmt.Errorf("element %d: %w", i+1, err)
			}
			if !present {
				return nil, false, fmt.Errorf("element %d is unset", i+1)
			}
			list[i] = built
		}
		return list, true, nil
	case mapKind:
		object := make(map[string]any, len(value.items))
		for i, entry := range value.items {
			key, item := entry.fields[0], entry.fields[1]
			if !key.set {
				return nil, false, fmt.Errorf("entry %d has no key", i+1)
			}
			if _, err := key.scalar(); err != nil {
				return nil, false, fmt.Errorf("entry %d key: %w", i+1, err)
			}
			name := key.input.Value()
			if _, exists := object[name]; exists {
				return nil, false, fmt.Errorf("duplicate key %q", name)
			}
			built, present, err := item.build()
			if err != nil {
				return nil, false, fmt.Errorf("entry %q: %w", name, err)
			}
			if !present {
				return nil, false, fmt.Errorf("entry %q has no value", name)
			}
			object[name] = built
		}
		return object, true, nil
	default:
		built, err := value.scalar()
		return built, true, err
	}
}

func (value *formValue) scalar() (any, error) {
	typ, raw := value.field.Type, value.input.Value()
	jsonType := typ != "string" && typ != "time" && typ != "duration"
	if jsonType && !strings.HasPrefix(raw, "@") && !json.Valid([]byte(raw)) {
		return nil, fmt.Errorf("expected JSON for %s", typ)
	}
	parsed, err := rpcconsole.ParseValue(typ, raw)
	if err != nil {
		return nil, err
	}
	if _, ok := parsed.(json.RawMessage); jsonType && !ok {
		return nil, fmt.Errorf("expected JSON for %s", typ)
	}
	return parsed, nil
}

func (frame *formFrame) count() int {
	switch frame.value.kind {
	case objectKind:
		return len(frame.value.fields)
	case listKind, mapKind:
		return len(frame.value.items)
	default:
		return 1
	}
}

func (frame *formFrame) current() *formValue {
	if frame.count() == 0 {
		return nil
	}
	switch frame.value.kind {
	case objectKind:
		return frame.value.fields[frame.focus]
	case listKind, mapKind:
		return frame.value.items[frame.focus]
	default:
		return frame.value
	}
}

func (frame *formFrame) focusInput(focus bool) tea.Cmd {
	current := frame.current()
	if current == nil || current.input == nil {
		return nil
	}
	if !focus {
		current.input.Blur()
		return nil
	}
	return current.input.Focus()
}

func (m *model) frame() *formFrame { return &m.forms[len(m.forms)-1] }

func (m model) openForm(method string) (tea.Model, tea.Cmd) {
	m.method, m.md, m.status = method, m.client.Describe(method), ""
	var fields []utils.FieldDescriptor
	if m.md != nil {
		fields = m.md.Args
	}
	root := newValue(utils.FieldDescriptor{Type: "object", Fields: fields})
	root.set = true
	root.resize(m.width)
	m.forms = []formFrame{{path: method, value: root}}
	m.view = viewForm
	return m, m.frame().focusInput(true)
}

func (m model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m.updateInput(msg)
	}
	if m.calling {
		return m, nil
	}
	switch key.String() {
	case "esc":
		return m.closeFrame()
	case "ctrl+s":
		return m.submit()
	}
	frame := m.frame()
	if frame.value.kind == listKind || frame.value.kind == mapKind {
		return m.updateCollection(key)
	}
	switch key.String() {
	case "enter":
		current := frame.current()
		if current == nil {
			return m, nil
		}
		if current.kind != scalarKind {
			return m.openValue(current, frame.path+" > "+current.field.Name)
		}
		if frame.value.kind == scalarKind || len(m.forms) > 1 && frame.focus == frame.count()-1 {
			return m.closeFrame()
		}
		return m.move(frame.focus + 1)
	case "tab", "down":
		return m.move(frame.focus + 1)
	case "shift+tab", "up":
		return m.move(frame.focus - 1)
	}
	return m.updateInput(msg)
}

func (m model) updateCollection(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	frame := m.frame()
	switch key.String() {
	case "up", "k":
		return m.move(frame.focus - 1)
	case "down", "j":
		return m.move(frame.focus + 1)
	case "a":
		var item *formValue
		if frame.value.kind == mapKind {
			item = newEntry(frame.value.field)
		} else {
			item = newValue(elementField(frame.value.field))
			item.set = true
		}
		item.resize(m.width)
		frame.value.items = append(frame.value.items, item)
		frame.focus = len(frame.value.items) - 1
		return m.openValue(item, itemPath(frame))
	case "enter", "e":
		if frame.count() != 0 {
			return m.openValue(frame.current(), itemPath(frame))
		}
	case "d", "x":
		if frame.count() != 0 {
			frame.value.items = append(frame.value.items[:frame.focus], frame.value.items[frame.focus+1:]...)
			if frame.focus >= frame.count() && frame.focus > 0 {
				frame.focus--
			}
		}
	}
	return m, nil
}

func (m model) openValue(value *formValue, path string) (tea.Model, tea.Cmd) {
	value.set = true
	m.forms = append(m.forms, formFrame{path: path, value: value})
	m.status = ""
	return m, m.frame().focusInput(true)
}

func itemPath(frame *formFrame) string {
	return fmt.Sprintf("%s[%d]", frame.path, frame.focus+1)
}

func (m model) closeFrame() (tea.Model, tea.Cmd) {
	m.frame().focusInput(false)
	if len(m.forms) == 1 {
		m.view = viewBrowse
	} else {
		m.forms = m.forms[:len(m.forms)-1]
	}
	m.status = ""
	return m, nil
}

func (m model) move(index int) (tea.Model, tea.Cmd) {
	frame, count := m.frame(), m.frame().count()
	if count == 0 {
		return m, nil
	}
	frame.focusInput(false)
	frame.focus = (index + count) % count
	return m, frame.focusInput(true)
}

func (m model) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	current := m.frame().current()
	if current == nil || current.input == nil {
		return m, nil
	}
	before := current.input.Value()
	updated, cmd := current.input.Update(msg)
	*current.input = updated
	if current.input.Value() != before {
		current.set = true
	}
	return m, cmd
}

func (m model) submit() (tea.Model, tea.Cmd) {
	params, _, err := m.forms[0].value.build()
	if err != nil {
		m.status = errStyle.Render(err.Error())
		return m, nil
	}
	m.calling, m.status = true, ""
	client, method, descriptor := m.client, m.method, m.md
	return m, tea.Batch(m.spin.Tick, func() tea.Msg {
		reply, err := client.Call(method, params)
		if err != nil {
			return callDoneMsg{err: err}
		}
		return callDoneMsg{output: rpcconsole.Format(reply, descriptor)}
	})
}

func (m model) viewForm() string {
	frame := m.frame()
	rows := make([]string, 0, frame.count())
	switch frame.value.kind {
	case objectKind:
		for i, value := range frame.value.fields {
			if value.input != nil {
				rows = append(rows, value.input.View())
			} else {
				row := value.field.Name + " = " + value.summary()
				switch {
				case i == frame.focus:
					row = selStyle.Render("> " + row)
				case !value.set:
					row = hintStyle.Render("  " + row)
				default:
					row = "  " + row
				}
				rows = append(rows, row)
			}
		}
	case listKind, mapKind:
		for i, value := range frame.value.items {
			line := value.summary()
			if frame.value.kind == mapKind {
				key := "<key>"
				if value.fields[0].set {
					key = value.fields[0].input.Value()
				}
				line = key + " = " + value.fields[1].summary()
			} else {
				line = fmt.Sprintf("%d. %s", i+1, line)
			}
			line = clip(line, max(m.width-4, 20))
			if i == frame.focus {
				line = selStyle.Render("> " + line)
			} else {
				line = "  " + line
			}
			rows = append(rows, line)
		}
	default:
		rows = append(rows, frame.value.input.View())
	}
	if len(rows) == 0 {
		if frame.value.kind == objectKind {
			rows = append(rows, hintStyle.Render("  no arguments"))
		} else {
			rows = append(rows, hintStyle.Render("  empty: press a to add"))
		}
	}
	status := ""
	if m.calling {
		status = "\n" + m.spin.View() + "calling...\n"
	} else if m.status != "" {
		status = "\n" + m.status + "\n"
	}
	help := "\n" + m.help.ShortHelpView(helpKeys(
		"enter", "edit", "ctrl+s", "run", "tab/arrows", "move",
		"esc", "back", "ctrl+c", "quit"))
	if frame.value.kind == listKind || frame.value.kind == mapKind {
		help = "\n" + m.help.ShortHelpView(helpKeys(
			"enter", "edit", "a", "add", "d", "delete", "arrows", "move",
			"ctrl+s", "run", "esc", "done"))
	}
	return titleStyle.Render(frame.path) + "\n\n" +
		m.window(rows, frame.focus, status+help) + status + help
}

func (m model) window(rows []string, at int, tail string) string {
	budget := m.height - 3 - strings.Count(tail, "\n")
	if m.height == 0 || len(rows) <= budget {
		return strings.Join(rows, "\n") + "\n"
	}
	budget = max(budget-2, 1)
	start := max(at-budget+1, 0)
	start = min(start, len(rows)-budget)
	end := start + budget
	output := strings.Join(rows[start:end], "\n") + "\n"
	if start > 0 {
		output = hintStyle.Render(fmt.Sprintf("  up: %d more", start)) + "\n" + output
	}
	if end < len(rows) {
		output += hintStyle.Render(fmt.Sprintf("  down: %d more", len(rows)-end)) + "\n"
	}
	return output
}

func (value *formValue) summary() string {
	if !value.set {
		return value.field.Type
	}
	switch value.kind {
	case objectKind:
		var fields []string
		for _, field := range value.fields {
			if field.kind == scalarKind && field.set {
				fields = append(fields, field.field.Name+"="+field.summary())
			}
		}
		if len(fields) != 0 {
			return strings.Join(fields, ", ")
		}
		return fmt.Sprintf("{%d fields}", len(value.fields))
	case listKind:
		return fmt.Sprintf("[%d items]", len(value.items))
	case mapKind:
		return fmt.Sprintf("{%d entries}", len(value.items))
	default:
		if value.input.Value() == "" {
			return `""`
		}
		return value.input.Value()
	}
}

func clip(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:max(width-3, 0)]) + "..."
}
