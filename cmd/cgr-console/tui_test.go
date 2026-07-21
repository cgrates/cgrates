// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
	tea "github.com/charmbracelet/bubbletea"
)

type testClient struct{}

func (testClient) Methods() []string { return nil }
func (testClient) Describe(string) *utils.MethodDescriptor {
	return nil
}
func (testClient) Call(string, any) (any, error) { return nil, nil }

func TestResultViewportScrollsHorizontally(t *testing.T) {
	m := newModel(testClient{}, "")
	m.result.Width, m.result.Height = 8, 1
	m.result.SetContent("0123456789abcdef")
	before := m.result.View()
	next, _ := m.updateResult(tea.KeyMsg{Type: tea.KeyRight})
	m = next.(model)
	if after := m.result.View(); after == before {
		t.Fatalf("right did not scroll %q", after)
	}
}

func TestResizeClampsResultViewport(t *testing.T) {
	m := newModel(testClient{}, "")
	m.result.Width, m.result.Height = 40, 5
	m.result.SetContent(strings.Repeat("line\n", 100))
	m.result.GotoBottom()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 50})
	m = next.(model)
	if m.result.PastBottom() {
		t.Fatalf("offset %d is past bottom after resize", m.result.YOffset)
	}
	if m.result.Height != 48 {
		t.Fatalf("height = %d, want 48", m.result.Height)
	}
}
