// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"
)

func Test_tagInternalConns(t *testing.T) {
	tests := []struct {
		name      string
		conns     []string
		subsystem string
		want      []string
	}{
		{
			name:      "Without internal",
			conns:     []string{"conn1", "conn2"},
			subsystem: "*stats",
			want:      []string{"conn1", "conn2"},
		},
		{
			name:      "With internal and conns",
			conns:     []string{"*internal", "conn2"},
			subsystem: "*stats",
			want:      []string{"*internal:*stats", "conn2"},
		},
		{
			name:      "Only with internal",
			conns:     []string{"*internal"},
			subsystem: "*stats",
			want:      []string{"*internal:*stats"},
		},
		{
			name:      "Empty conns",
			conns:     []string{},
			subsystem: "*stats",
			want:      []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := tagInternalConns(tt.conns, tt.subsystem)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tagInternalConns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stripInternalConns(t *testing.T) {
	tests := []struct {
		name  string
		conns []string
		want  []string
	}{
		{
			name:  "with *internal",
			conns: []string{"*internal:*stats"},
			want:  []string{"*internal"},
		},
		{
			name:  "with random subsystem suffix",
			conns: []string{"*test:*stats"},
			want:  []string{"*test:*stats"},
		},
		{
			name:  "nil case",
			conns: nil,
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := stripInternalConns(tt.conns)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stripInternalConns() = %v, want %v", got, tt.want)
			}
		})
	}
}
