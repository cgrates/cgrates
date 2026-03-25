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
package utils

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/cron"
)

func TestExporterMetricsString(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}
	expected := "{\"field1\":2}"
	if reply := ms.String(); reply != expected {
		t.Errorf("Expected %s \n but received \n %s", expected, reply)
	}
}

func TestExporterMetricsFieldAsInterface(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	input := []string{"field1"}
	expected := 2
	if reply, err := ms.FieldAsInterface(input); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %d \n but received \n %d", expected, reply)
	}
}

func TestExporterMetricsFieldAsString(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	input := []string{"field1"}
	expected := "2"
	if reply, err := ms.FieldAsString(input); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %s \n but received \n %s", expected, reply)
	}
}

func TestExporterMetricsSet(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	expected := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	if err := ms.Set([]string{"field2"}, 3); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ms, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, ms)
	}
}

func TestExporterMetricsGetKeys(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	expected := []string{"*req.field1", "*req.field2"}
	reply := ms.GetKeys(false, 0, MetaReq)
	sort.Strings(reply)
	if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, reply)
	}
}

func TestExporterMetricsRemove(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	expected := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	if err := ms.Remove([]string{"field2"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ms, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, ms)
	}
}

func TestExporterMetricsClonedMapStorage(t *testing.T) {

	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	if reply := ms.ClonedMapStorage(); reflect.DeepEqual(ms, reply) {
		t.Errorf("Expected %v \n but received \n %v", ms, reply)
	}
}

func TestNewExporterMetrics(t *testing.T) {
	tests := []struct {
		name     string
		schedule string
		timezone string
		wantErr  bool
	}{
		{
			name:     "Success without schedule",
			schedule: "",
			timezone: "Local",
			wantErr:  false,
		},
		{
			name:     "Success with schedule and timezone",
			schedule: "@every 1s",
			timezone: "Local",
			wantErr:  false,
		},
		{
			name:     "Empty fields",
			schedule: "",
			timezone: "",
			wantErr:  false,
		},
		{
			name:     "Fail case with wrong schedule format",
			schedule: "tst",
			timezone: "",
			wantErr:  true,
		},
		{
			name:     "Fail case with invalid timezone",
			schedule: "",
			timezone: "invlid",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewExporterMetrics(tt.schedule, tt.timezone)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewExporterMetrics() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewExporterMetrics() succeeded unexpectedly")
			}

			if got == nil {
				t.Errorf("NewExporterMetrics() was expecting a struct but got: %v", got)
			}

			if got.MapStorage == nil {
				t.Errorf("The map storage should have been created")
			}
		})
	}
}

func TestExporterMetricsStopCron(t *testing.T) {

	tests := []struct {
		name            string
		exporterMetrics *ExporterMetrics
	}{
		{
			name: "",
			exporterMetrics: &ExporterMetrics{
				cron: cron.New(),
			},
		},
		{
			name: "Nil",
			exporterMetrics: &ExporterMetrics{
				cron: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.exporterMetrics.StopCron()

			want := tt.exporterMetrics
			if tt.exporterMetrics != want {
				t.Errorf("Got %v, wanted %v", tt.exporterMetrics, want)
			}

		})
	}
}

func TestExporterMetricsIncrementEvents(t *testing.T) {

	tests := []struct {
		name            string
		schedule        string
		timezone        string
		exporterMetrics *ExporterMetrics
	}{
		{
			name: "",
			exporterMetrics: &ExporterMetrics{
				MapStorage: MapStorage{
					NumberOfEvents: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.exporterMetrics.IncrementEvents()
			tt.exporterMetrics.IncrementEvents()
			tt.exporterMetrics.IncrementEvents()

			w := tt.exporterMetrics

			if !reflect.DeepEqual(tt.exporterMetrics, w) {
				t.Errorf("Got %v, wanted %v", tt.exporterMetrics, w)
			}

		})
	}
}
