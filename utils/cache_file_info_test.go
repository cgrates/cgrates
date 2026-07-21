// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestMapStringInterfaceToLoadInstances(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]any
		want []*LoadInstance
	}{
		{
			m: map[string]any{
				LoadID:           "load",
				RatingLoadID:     "ratingLoad",
				AccountingLoadID: "account",
				LoadTime:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: nil,
		},
		{
			m: map[string]any{
				"LoadHistory": map[string]any{
					LoadID:           "load",
					RatingLoadID:     "ratingLoad",
					AccountingLoadID: "account",
					LoadTime:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want: nil,
		},
		{
			m: map[string]any{
				"LoadHistory": []any{
					map[string]string{
						LoadID:           "load1",
						RatingLoadID:     "ratingLoad1",
						AccountingLoadID: "account1",
						LoadTime:         "2026-01-01T00:00:00Z",
					},
				},
			},
			want: nil,
		},
		{
			m: map[string]any{
				"LoadHistory": []any{
					map[string]any{
						LoadID:           "load",
						RatingLoadID:     "ratingLoad",
						AccountingLoadID: "account",
						LoadTime:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					map[string]any{
						LoadID:           "load1",
						RatingLoadID:     "ratingLoad1",
						AccountingLoadID: "account1",
						LoadTime:         "2026-01-01T00:00:00Z",
					},
				},
			},
			want: []*LoadInstance{
				{
					LoadID:           "load",
					RatingLoadID:     "ratingLoad",
					AccountingLoadID: "account",
					LoadTime:         time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					LoadID:           "load1",
					RatingLoadID:     "ratingLoad1",
					AccountingLoadID: "account1",
					LoadTime:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapStringInterfaceToLoadInstances(tt.m)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected %v, recieved %v", ToJSON(tt.want), ToJSON(got))
			}
		})
	}
}

func TestLoadInstancesAsMapStringInterface(t *testing.T) {
	tests := []struct {
		name          string
		loadInstances []*LoadInstance
		want          map[string]any
	}{
		{
			loadInstances: []*LoadInstance{},
			want: map[string]any{
				"LoadHistory": []map[string]any{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LoadInstancesAsMapStringInterface(tt.loadInstances)
			if !reflect.DeepEqual(ToJSON(got), ToJSON(tt.want)) {
				t.Errorf("Expected %#+v, recieved %#+v", ToJSON(tt.want), ToJSON(got))
			}
		})
	}
}
