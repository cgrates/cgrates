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
package engine

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// roundingDecimals is the GeneralCfg default these metric tests round against.
const roundingDecimals = 5

func TestStatMetricsGetStringValue(t *testing.T) {
	startTime := time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)
	started := map[string]any{utils.MetaStartTime: startTime}
	sum1, err := NewStatSum(2, "~*opts.*cost", nil)
	if err != nil {
		t.Fatal(err)
	}
	sum2, err := NewStatSum(2, "~*opts.*cost", nil)
	if err != nil {
		t.Fatal(err)
	}
	type step struct {
		add, rem string
		opts     map[string]any
		wantStr  string
		wantErr  string
	}
	tests := []struct {
		name   string
		metric StatMetric
		steps  []step
	}{
		{
			name:   "ASR",
			metric: NewASR(2, "", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: started, wantStr: utils.NotAvailable},
				{add: "EVENT_2", wantStr: "50%"},
				{add: "EVENT_3", wantStr: "33.333%"},
				{rem: "EVENT_3", wantStr: "50%"},
				{add: "EVENT_4", opts: started},
				{add: "EVENT_5", opts: started},
				{rem: "EVENT_1", wantStr: "66.667%"},
				{rem: "EVENT_2", wantStr: "100%"},
				{rem: "EVENT_4"},
				{rem: "EVENT_5", wantStr: utils.NotAvailable},
			},
		},
		{
			name:   "ASR repeated events",
			metric: NewASR(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: started},
				{add: "EVENT_2", wantStr: "50%"},
				{add: "EVENT_2", wantStr: "33.333%"},
				{add: "EVENT_4", wantStr: "25%"},
				{rem: "EVENT_4"},
				{rem: "EVENT_2", wantStr: "50%"},
				{rem: "EVENT_2"},
				{add: "EVENT_1", opts: started, wantStr: "100%"},
			},
		},
		{
			name:   "ACD",
			metric: NewACD(2, "", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 10 * time.Second, utils.MetaStartTime: startTime}, wantStr: utils.NotAvailable},
				{add: "EVENT_2", wantErr: "NOT_FOUND:*usage"},
				{add: "EVENT_3", wantErr: "NOT_FOUND:*usage", wantStr: utils.NotAvailable},
				{wantStr: utils.NotAvailable},
				{rem: "EVENT_1", wantStr: utils.NotAvailable},
				{add: "EVENT_4", opts: map[string]any{utils.MetaUsage: 478433753 * time.Nanosecond, utils.MetaStartTime: startTime}, wantStr: utils.NotAvailable},
				{add: "EVENT_5", opts: map[string]any{utils.MetaUsage: 30*time.Second + 982433452*time.Nanosecond, utils.MetaStartTime: startTime}, wantStr: "15.73s"},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantStr: "15.73s"},
				{rem: "EVENT_5"},
				{rem: "EVENT_4"},
				{rem: "EVENT_5", wantErr: "NOT_FOUND", wantStr: utils.NotAvailable},
			},
		},
		{
			name:   "ACD repeated events",
			metric: NewACD(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 2 * time.Minute}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaUsage: time.Minute}, wantStr: "1m30s"},
				{add: "EVENT_2", opts: map[string]any{utils.MetaUsage: time.Minute}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaUsage: time.Minute}, wantStr: "1m15s"},
				{rem: "EVENT_2", wantStr: "1m20s"},
			},
		},
		{
			name:   "TCD",
			metric: NewTCD(2, "", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 10 * time.Second, utils.MetaStartTime: startTime}, wantStr: utils.NotAvailable},
				{add: "EVENT_2", opts: map[string]any{utils.MetaUsage: 10 * time.Second, utils.MetaStartTime: startTime}},
				{add: "EVENT_3", wantErr: "NOT_FOUND:*usage", wantStr: "20s"},
				{rem: "EVENT_2", wantStr: utils.NotAvailable},
				{rem: "EVENT_1", wantStr: utils.NotAvailable},
				{add: "EVENT_4", opts: map[string]any{utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime}},
				{add: "EVENT_5", opts: map[string]any{utils.MetaUsage: time.Minute + 30*time.Second, utils.MetaStartTime: startTime}, wantStr: "2m30s"},
				{rem: "EVENT_4", wantStr: utils.NotAvailable},
				{rem: "EVENT_5"},
				{rem: "EVENT_3", wantErr: "NOT_FOUND", wantStr: utils.NotAvailable},
			},
		},
		{
			name:   "TCD repeated events",
			metric: NewTCD(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 2 * time.Minute}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaUsage: time.Minute}, wantStr: "3m0s"},
				{add: "EVENT_2", opts: map[string]any{utils.MetaUsage: time.Minute}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaUsage: time.Minute}, wantStr: "5m0s"},
				{rem: "EVENT_2", wantStr: "4m0s"},
			},
		},
		{
			name:   "ACC",
			metric: NewACC(2, "", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: 12.3}, wantStr: utils.NotAvailable},
				{add: "EVENT_2", wantErr: "NOT_FOUND:*cost"},
				{add: "EVENT_3", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: 12.3}, wantStr: "12.3"},
				{rem: "EVENT_3", wantStr: utils.NotAvailable},
				{add: "EVENT_4", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: 5.6}},
				{add: "EVENT_5", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: 1.2}},
				{rem: "EVENT_1", wantStr: "3.4"},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantStr: "3.4"},
				{rem: "EVENT_4"},
				{rem: "EVENT_5", wantStr: utils.NotAvailable},
				{add: "EVENT_5", opts: map[string]any{utils.MetaCost: -1}, wantErr: "NEGATIVE:*cost"},
			},
		},
		{
			name:   "ACC repeated events",
			metric: NewACC(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 12.3}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}, wantStr: "15.3"},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}, wantStr: "16.8"},
				{rem: "EVENT_2", wantStr: "16.3"},
			},
		},
		{
			name:   "TCC",
			metric: NewTCC(2, "", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: 12.3}, wantStr: utils.NotAvailable},
				{add: "EVENT_2", wantErr: "NOT_FOUND:*cost"},
				{add: "EVENT_3", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: 5.7}, wantStr: "18"},
				{rem: "EVENT_3", wantStr: utils.NotAvailable},
				{add: "EVENT_4", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: 5.6}},
				{add: "EVENT_5", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: 1.2}},
				{rem: "EVENT_1", wantStr: "6.8"},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantStr: "6.8"},
				{rem: "EVENT_4"},
				{rem: "EVENT_5", wantStr: utils.NotAvailable},
				{add: "EVENT_5", opts: map[string]any{utils.MetaCost: -1}, wantErr: "NEGATIVE:*cost"},
			},
		},
		{
			name:   "TCC repeated events",
			metric: NewTCC(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 12.3}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}, wantStr: "30.6"},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}, wantStr: "67.2"},
				{rem: "EVENT_2", wantStr: "48.9"},
			},
		},
		{
			name:   "PDD",
			metric: NewPDD(2, "", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaPDD: 5 * time.Second, utils.MetaUsage: 10 * time.Second, utils.MetaStartTime: startTime}, wantStr: utils.NotAvailable},
				{add: "EVENT_2", wantErr: "NOT_FOUND:*pdd"},
				{add: "EVENT_3", wantErr: "NOT_FOUND:*pdd", wantStr: utils.NotAvailable},
				{rem: "EVENT_3", wantErr: "NOT_FOUND", wantStr: utils.NotAvailable},
				{rem: "EVENT_1", wantStr: utils.NotAvailable},
				{add: "EVENT_4", opts: map[string]any{utils.MetaPDD: 10 * time.Second, utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime}, wantStr: utils.NotAvailable},
				{add: "EVENT_5", opts: map[string]any{utils.MetaPDD: 10 * time.Second}, wantStr: "10s"},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantStr: "10s"},
				{rem: "EVENT_5"},
				{rem: "EVENT_4"},
				{rem: "EVENT_5", wantErr: "NOT_FOUND", wantStr: utils.NotAvailable},
			},
		},
		{
			name:   "PDD repeated events",
			metric: NewPDD(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaPDD: 2 * time.Minute}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaPDD: time.Minute}, wantStr: "1m30s"},
				{add: "EVENT_2", opts: map[string]any{utils.MetaPDD: time.Minute}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaPDD: time.Minute}, wantStr: "1m15s"},
				{rem: "EVENT_2", wantStr: "1m20s"},
			},
		},
		{
			name:   "DDC",
			metric: NewDDC(2, "", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaDestination: "1002", utils.MetaStartTime: startTime}, wantStr: utils.NotAvailable},
				{add: "EVENT_2", opts: map[string]any{utils.MetaDestination: "1002", utils.MetaStartTime: startTime}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaDestination: "1001", utils.MetaStartTime: startTime}, wantStr: "2"},
				{rem: "EVENT_1", wantStr: "2"},
				{rem: "EVENT_2", wantStr: utils.NotAvailable},
				{rem: "EVENT_3", wantStr: utils.NotAvailable},
			},
		},
		{
			name:   "DDC repeated events",
			metric: NewDDC(2, "", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaDestination: "1001"}, wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaDestination: "1002"}, wantStr: "2"},
				{rem: "EVENT_1", wantStr: utils.NotAvailable},
			},
		},
		{
			name:   "Sum",
			metric: sum1,
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaDestination: "1002", utils.MetaCost: "20", utils.MetaStartTime: startTime}, wantStr: utils.NotAvailable},
				{add: "EVENT_2", opts: map[string]any{utils.MetaDestination: "1002", utils.MetaCost: "20", utils.MetaStartTime: startTime}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaDestination: "1001", utils.MetaCost: "20", utils.MetaStartTime: startTime}, wantStr: "60"},
				{rem: "EVENT_1", wantStr: "40"},
				{rem: "EVENT_2", wantStr: utils.NotAvailable},
				{rem: "EVENT_3", wantStr: utils.NotAvailable},
			},
		},
		{
			name:   "Sum repeated events",
			metric: sum2,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 12.3}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}, wantStr: "30.6"},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}, wantStr: "67.2"},
				{rem: "EVENT_2", wantStr: "48.9"},
			},
		},
		{
			name:   "Average",
			metric: NewStatAverage(2, "~*opts.*cost", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: "20", utils.MetaStartTime: startTime, utils.MetaDestination: "1002"}, wantStr: utils.NotAvailable},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: "20", utils.MetaStartTime: startTime, utils.MetaDestination: "1002"}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaCost: "20", utils.MetaStartTime: startTime, utils.MetaDestination: "1001"}, wantStr: "20"},
				{rem: "EVENT_1", wantStr: "20"},
				{rem: "EVENT_2", wantStr: utils.NotAvailable},
				{rem: "EVENT_3", wantStr: utils.NotAvailable},
			},
		},
		{
			name:   "Average repeated events",
			metric: NewStatAverage(2, "~*opts.*cost", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 12.3}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}, wantStr: "15.3"},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 18.3}, wantStr: "16.8"},
				{rem: "EVENT_2", wantStr: "16.3"},
			},
		},
		{
			name:   "Distinct",
			metric: NewStatDistinct(2, "~*opts.*cost", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: "20"}, wantStr: utils.NotAvailable},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: "20"}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaCost: "40"}, wantStr: "2"},
				{rem: "EVENT_1", wantStr: "2"},
				{rem: "EVENT_2", wantStr: utils.NotAvailable},
				{rem: "EVENT_3", wantStr: utils.NotAvailable},
			},
		},
		{
			name:   "Distinct repeated events",
			metric: NewStatDistinct(2, "~*opts.*cost", nil),
			steps: []step{
				{wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: "20"}, wantStr: utils.NotAvailable},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: "40"}, wantStr: "2"},
				{rem: "EVENT_1", wantStr: utils.NotAvailable},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.metric
			for i, s := range tt.steps {
				var err error
				switch {
				case s.add != "":
					err = m.AddEvent(s.add, utils.MapStorage{utils.MetaOpts: s.opts})
				case s.rem != "":
					err = m.RemEvent(s.rem)
				}
				switch {
				case s.wantErr != "":
					if err == nil || err.Error() != s.wantErr {
						t.Fatalf("step %d: want error %q, got %v", i, s.wantErr, err)
					}
				case err != nil:
					t.Fatalf("step %d: %v", i, err)
				}
				if s.wantStr != "" {
					if got := m.GetStringValue(roundingDecimals); got != s.wantStr {
						t.Fatalf("step %d: GetStringValue() = %q, want %q", i, got, s.wantStr)
					}
				}
			}
		})
	}
}

func TestStatMetricsGetValue(t *testing.T) {
	startTime := time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)
	startTime2015 := time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC)
	started := map[string]any{utils.MetaStartTime: startTime}
	sum, err := NewStatSum(2, "~*opts.*cost", nil)
	if err != nil {
		t.Fatal(err)
	}
	type step struct {
		add, rem string
		opts     map[string]any
		wantVal  *utils.Decimal
		wantErr  string
	}
	tests := []struct {
		name   string
		metric StatMetric
		steps  []step
	}{
		{
			name:   "ASR",
			metric: NewASR(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: started, wantVal: utils.DecimalNaN},
				{add: "EVENT_2"},
				{add: "EVENT_3", wantVal: utils.NewDecimalFromFloat64(33.33333333333333)},
				{rem: "EVENT_3", wantVal: utils.NewDecimalFromFloat64(50.0)},
				{add: "EVENT_4", opts: started},
				{add: "EVENT_5", opts: started},
				{rem: "EVENT_1", wantVal: utils.NewDecimalFromFloat64(66.66666666666667)},
				{rem: "EVENT_2", wantVal: utils.NewDecimalFromFloat64(100.0)},
				{rem: "EVENT_4", wantVal: utils.DecimalNaN},
				{rem: "EVENT_5", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "ACD",
			metric: NewACD(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", wantErr: "NOT_FOUND:*usage", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime2015}, wantVal: utils.NewDecimalFromFloat64(35.0 * 1e9)},
				{add: "EVENT_5", opts: map[string]any{utils.MetaUsage: time.Minute + 30*time.Second, utils.MetaStartTime: startTime2015}, wantVal: utils.NewDecimalFromFloat64(53.33333333333333 * 1e9)},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantVal: utils.NewDecimalFromFloat64(53.33333333333333 * 1e9)},
				{rem: "EVENT_4", wantVal: utils.NewDecimalFromFloat64(50.0 * 1e9)},
				{rem: "EVENT_1", wantVal: utils.DecimalNaN},
				{rem: "EVENT_5", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "ACD 2",
			metric: NewACD(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaUsage: 8 * time.Second}},
				{add: "EVENT_3", wantErr: "NOT_FOUND:*usage", wantVal: utils.NewDecimalFromFloat64(float64(9 * time.Second))},
				{rem: "EVENT_1", wantVal: utils.DecimalNaN},
				{rem: "EVENT_2", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime}},
				{add: "EVENT_5", opts: map[string]any{utils.MetaUsage: 4*time.Minute + 30*time.Second, utils.MetaStartTime: startTime}, wantVal: utils.NewDecimalFromFloat64(float64(2*time.Minute + 45*time.Second))},
				{rem: "EVENT_5"},
				{rem: "EVENT_4", wantVal: utils.DecimalNaN},
				{rem: "EVENT_3", wantErr: "NOT_FOUND", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "TCD",
			metric: NewTCD(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", wantErr: "NOT_FOUND:*usage", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime}, wantVal: utils.NewDecimalFromFloat64(70.0 * 1e9)},
				{add: "EVENT_5", opts: map[string]any{utils.MetaUsage: time.Minute + 30*time.Second, utils.MetaStartTime: startTime}, wantVal: utils.NewDecimalFromFloat64(160.0 * 1e9)},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantVal: utils.NewDecimalFromFloat64(160.0 * 1e9)},
				{rem: "EVENT_4", wantVal: utils.NewDecimalFromFloat64(100.0 * 1e9)},
				{rem: "EVENT_1", wantVal: utils.DecimalNaN},
				{rem: "EVENT_5", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "TCD 2",
			metric: NewTCD(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaUsage: 5 * time.Second}},
				{add: "EVENT_3", wantErr: "NOT_FOUND:*usage", wantVal: utils.NewDecimalFromFloat64(float64(15 * time.Second))},
				{rem: "EVENT_1", wantVal: utils.DecimalNaN},
				{rem: "EVENT_2", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime}},
				{add: "EVENT_5", opts: map[string]any{utils.MetaUsage: time.Minute + 30*time.Second, utils.MetaStartTime: startTime}, wantVal: utils.NewDecimalFromFloat64(float64(2*time.Minute + 30*time.Second))},
				{rem: "EVENT_5"},
				{rem: "EVENT_4", wantVal: utils.DecimalNaN},
				{rem: "EVENT_3", wantErr: "NOT_FOUND", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "ACC",
			metric: NewACC(2, "", nil),
			steps: []step{
				{wantVal: utils.DecimalNaN},
				{add: "EVENT_1", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: "12.3"}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", wantErr: "NOT_FOUND:*cost"},
				{add: "EVENT_3", wantErr: "NOT_FOUND:*cost", wantVal: utils.DecimalNaN},
				{rem: "EVENT_3", wantErr: "NOT_FOUND", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: "5.6"}},
				{add: "EVENT_5", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: "1.2"}},
				{rem: "EVENT_1", wantVal: utils.NewDecimalFromFloat64(3.4)},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantVal: utils.NewDecimalFromFloat64(3.4)},
				{rem: "EVENT_4"},
				{rem: "EVENT_5", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "TCC",
			metric: NewTCC(2, "", nil),
			steps: []step{
				{wantVal: utils.DecimalNaN},
				{add: "EVENT_1", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: "12.3"}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", wantErr: "NOT_FOUND:*cost"},
				{add: "EVENT_3", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: 1.2}, wantVal: utils.NewDecimalFromFloat64(13.5)},
				{rem: "EVENT_3", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: "5.6"}},
				{add: "EVENT_5", opts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: "1.2"}},
				{rem: "EVENT_1", wantVal: utils.NewDecimalFromFloat64(6.8)},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantVal: utils.NewDecimalFromFloat64(6.8)},
				{rem: "EVENT_4"},
				{rem: "EVENT_5", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "PDD",
			metric: NewPDD(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaPDD: 5 * time.Second, utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", wantErr: "NOT_FOUND:*pdd", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaPDD: 10 * time.Second, utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime2015}, wantVal: utils.NewDecimalFromFloat64(7.5 * 1e9)},
				{add: "EVENT_5", opts: map[string]any{utils.MetaUsage: time.Minute + 30*time.Second, utils.MetaStartTime: startTime2015}, wantErr: "NOT_FOUND:*pdd", wantVal: utils.NewDecimalFromFloat64(7.5 * 1e9)},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantVal: utils.NewDecimalFromFloat64(7.5 * 1e9)},
				{rem: "EVENT_4", wantVal: utils.DecimalNaN},
				{rem: "EVENT_1", wantVal: utils.DecimalNaN},
				{rem: "EVENT_5", wantErr: "NOT_FOUND", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "PDD 2",
			metric: NewPDD(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaPDD: 9 * time.Second, utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", opts: map[string]any{utils.MetaPDD: 10 * time.Second, utils.MetaStartTime: startTime, utils.MetaUsage: 8 * time.Second}},
				{add: "EVENT_3", wantErr: "NOT_FOUND:*pdd", wantVal: utils.NewDecimalFromFloat64(float64(9*time.Second + 500*time.Millisecond))},
				{rem: "EVENT_1", wantVal: utils.DecimalNaN},
				{rem: "EVENT_2", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaPDD: 8 * time.Second, utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime}},
				{add: "EVENT_5", opts: map[string]any{utils.MetaUsage: 4*time.Minute + 30*time.Second, utils.MetaStartTime: startTime}, wantErr: "NOT_FOUND:*pdd", wantVal: utils.DecimalNaN},
				{rem: "EVENT_5", wantErr: "NOT_FOUND"},
				{rem: "EVENT_4", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "DDC",
			metric: NewDDC(2, "", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaDestination: "1002", utils.MetaPDD: 5 * time.Second, utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", wantErr: "NOT_FOUND:*destination", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaDestination: "1001", utils.MetaPDD: 10 * time.Second, utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime2015}, wantVal: utils.NewDecimalFromFloat64(2)},
				{add: "EVENT_5", opts: map[string]any{utils.MetaDestination: "1003", utils.MetaUsage: time.Minute + 30*time.Second, utils.MetaStartTime: startTime2015}, wantVal: utils.NewDecimalFromFloat64(3)},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantVal: utils.NewDecimalFromFloat64(3)},
				{rem: "EVENT_4", wantVal: utils.NewDecimalFromFloat64(2)},
				{rem: "EVENT_1", wantVal: utils.DecimalNaN},
				{rem: "EVENT_5", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "Sum",
			metric: sum,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaDestination: "1002", utils.MetaPDD: 5 * time.Second, utils.MetaCost: "20", utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", wantErr: "NOT_FOUND", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaDestination: "1001", utils.MetaPDD: 10 * time.Second, utils.MetaCost: "20", utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime2015}, wantVal: utils.NewDecimalFromFloat64(40)},
				{add: "EVENT_5", opts: map[string]any{utils.MetaDestination: "1003", utils.MetaCost: "20", utils.MetaUsage: time.Minute + 30*time.Second, utils.MetaStartTime: startTime2015}, wantVal: utils.NewDecimalFromFloat64(60)},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantVal: utils.NewDecimalFromFloat64(60)},
				{rem: "EVENT_4", wantVal: utils.NewDecimalFromFloat64(40)},
				{rem: "EVENT_1", wantVal: utils.DecimalNaN},
				{rem: "EVENT_5", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "Average",
			metric: NewStatAverage(2, "~*opts.*cost", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: "20", utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second, utils.MetaPDD: 5 * time.Second, utils.MetaDestination: "1002"}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", wantErr: "NOT_FOUND:~*opts.*cost", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaCost: "30", utils.MetaUsage: time.Minute, utils.MetaStartTime: startTime2015, utils.MetaPDD: 10 * time.Second, utils.MetaDestination: "1001"}, wantVal: utils.NewDecimalFromFloat64(25)},
				{add: "EVENT_5", opts: map[string]any{utils.MetaCost: "20", utils.MetaUsage: time.Minute + 30*time.Second, utils.MetaStartTime: startTime2015, utils.MetaDestination: "1003"}, wantVal: utils.NewDecimalFromFloat64(23.33333333333333)},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantVal: utils.NewDecimalFromFloat64(23.33333333333333)},
				{rem: "EVENT_4", wantVal: utils.NewDecimalFromFloat64(20)},
				{rem: "EVENT_1", wantVal: utils.DecimalNaN},
				{rem: "EVENT_5", wantVal: utils.DecimalNaN},
			},
		},
		{
			name:   "Distinct",
			metric: NewStatDistinct(2, "~*opts.*usage", nil),
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 10 * time.Second}, wantVal: utils.DecimalNaN},
				{add: "EVENT_2", wantErr: "NOT_FOUND:~*opts.*usage", wantVal: utils.DecimalNaN},
				{add: "EVENT_4", opts: map[string]any{utils.MetaUsage: time.Minute}, wantVal: utils.NewDecimalFromFloat64(2)},
				{add: "EVENT_5", opts: map[string]any{utils.MetaUsage: time.Minute + 30*time.Second}, wantVal: utils.NewDecimalFromFloat64(3)},
				{rem: "EVENT_2", wantErr: "NOT_FOUND", wantVal: utils.NewDecimalFromFloat64(3)},
				{rem: "EVENT_4", wantVal: utils.NewDecimalFromFloat64(2)},
				{rem: "EVENT_1", wantVal: utils.DecimalNaN},
				{rem: "EVENT_5", wantVal: utils.DecimalNaN},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.metric
			for i, s := range tt.steps {
				var err error
				switch {
				case s.add != "":
					err = m.AddEvent(s.add, utils.MapStorage{utils.MetaOpts: s.opts})
				case s.rem != "":
					err = m.RemEvent(s.rem)
				}
				switch {
				case s.wantErr != "":
					if err == nil || err.Error() != s.wantErr {
						t.Fatalf("step %d: want error %q, got %v", i, s.wantErr, err)
					}
				case err != nil:
					t.Fatalf("step %d: %v", i, err)
				}
				if s.wantVal != nil {
					got := m.GetValue()
					if s.wantVal == utils.DecimalNaN {
						if got != utils.DecimalNaN {
							t.Fatalf("step %d: GetValue() = %v, want NaN", i, got)
						}
					} else if got.Compare(s.wantVal) != 0 {
						t.Fatalf("step %d: GetValue() = %v, want %v", i, got, s.wantVal)
					}
				}
			}
		})
	}
}

func TestStatMetricsGetValue2(t *testing.T) {
	tests := []struct {
		name   string
		metric StatMetric
		want   *utils.Decimal
	}{
		{
			name: "Distinct",
			metric: &StatDistinct{
				FieldValues: map[string]utils.StringSet{
					"FieldValue1": {},
				},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 2,
					},
					"Event2": {},
				},
				MinItems:  3,
				FieldName: "Test_Field_Name",
				Count:     3,
			},
			want: utils.NewDecimal(1, 0),
		},
		{
			name: "Average",
			metric: &StatAverage{
				Metric: &Metric{
					Value:    utils.NewDecimal(10, 0),
					Count:    20,
					MinItems: 10,
					Events: map[string]*DecimalWithCompress{
						"Event1": {},
					},
				},
			},
			want: utils.NewDecimalFromFloat64(0.5),
		},
		{
			name: "Sum",
			metric: &StatSum{
				Metric: &Metric{
					Value:    utils.NewDecimal(10, 0),
					Count:    15,
					MinItems: 20,
					Events: map[string]*DecimalWithCompress{
						"Event1": {},
					},
				},
			},
			want: utils.DecimalNaN,
		},
		{
			name: "ACC",
			metric: &StatACC{
				Metric: &Metric{
					Events: map[string]*DecimalWithCompress{
						"Event1": {},
					},
					MinItems: 3,
					Value:    utils.NewDecimal(0, 0),
					Count:    3,
				},
			},
			want: utils.NewDecimal(0, 0),
		},
		{
			name: "TCC",
			metric: &StatTCC{
				Metric: &Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 3,
					Events: map[string]*DecimalWithCompress{
						"Event1": {},
					},
					MinItems: 3,
				},
			},
			want: utils.NewDecimal(2, 0),
		},
		{
			name: "DDC",
			metric: &StatDDC{
				FieldValues: map[string]utils.StringSet{
					"Field_Value1": {},
				},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 1,
					},
					"Event2": {},
				},
				MinItems: 3,
				Count:    3,
			},
			want: utils.NewDecimal(1, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metric.GetValue()
			if tt.want == utils.DecimalNaN {
				if got != utils.DecimalNaN {
					t.Errorf("GetValue() = %v, want NaN", got)
				}
			} else if got.Compare(tt.want) != 0 {
				t.Errorf("GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatMetricsState(t *testing.T) {
	started := map[string]any{utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}
	asr := &StatASR{Metric: NewMetric(2, nil)}
	acd := &StatACD{Metric: NewMetric(2, nil)}
	tcd := &StatTCD{Metric: NewMetric(2, nil)}
	acc := &StatACC{Metric: NewMetric(2, nil)}
	tcc := &StatTCC{Metric: NewMetric(2, nil)}
	pdd := &StatPDD{Metric: NewMetric(2, nil)}
	sum := &StatSum{
		Metric: NewMetric(2, nil),
		Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
	}
	avg := &StatAverage{Metric: NewMetric(2, nil), FieldName: "~*opts.*cost"}
	type step struct {
		add, rem string
		opts     map[string]any
		wantStr  string
		check    func(t *testing.T)
	}
	tests := []struct {
		name   string
		metric StatMetric
		steps  []step
	}{
		{
			name:   "ASR",
			metric: asr,
			steps: []step{
				{add: "EVENT_1", opts: started},
				{add: "EVENT_2", wantStr: "50%", check: func(t *testing.T) {
					expected := &StatASR{
						Metric: &Metric{
							Value:    utils.NewDecimal(1, 0),
							Count:    2,
							MinItems: 2,
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
								"EVENT_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 1},
							},
						},
					}
					if !reflect.DeepEqual(*expected, *asr) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
					}
				}},
				{add: "EVENT_2"},
				{add: "EVENT_1", wantStr: "25%", check: func(t *testing.T) {
					expected := &StatASR{
						Metric: &Metric{
							Value:    utils.NewDecimal(1, 0),
							Count:    4,
							MinItems: 2,
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 2},
								"EVENT_2": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 2},
							},
						},
					}
					if !reflect.DeepEqual(*expected, *asr) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
					}
				}},
				{rem: "EVENT_1"},
				{rem: "EVENT_2", wantStr: "50%", check: func(t *testing.T) {
					expected := &StatASR{
						Metric: &Metric{
							Value:    utils.NewDecimal(1, 0),
							Count:    2,
							MinItems: 2,
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromFloat64(1), CompressFactor: 1},
								"EVENT_2": {Stat: utils.NewDecimalFromFloat64(0), CompressFactor: 1},
							},
						},
					}
					if !reflect.DeepEqual(*expected, *asr) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
					}
				}},
			},
		},
		{
			name:   "ACD",
			metric: acd,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 2 * time.Minute}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 3 * time.Minute}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaUsage: time.Minute}, check: func(t *testing.T) {
					expected := &StatACD{
						Metric: &Metric{
							Value:    utils.NewDecimal(6*int64(time.Minute), 0),
							Count:    3,
							MinItems: 2,
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
								"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
							},
						},
					}
					if !expected.Equal(acd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
					}
				}},
				{rem: "EVENT_1", check: func(t *testing.T) {
					expected := &StatACD{
						Metric: &Metric{
							Value:    utils.NewDecimal(int64(3*time.Minute+30*time.Second), 0),
							Count:    2,
							MinItems: 2,
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 1},
								"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
							},
						},
					}
					if !expected.Equal(acd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
					}
				}},
			},
		},
		{
			name:   "TCD",
			metric: tcd,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 2 * time.Minute}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 3 * time.Minute}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaUsage: time.Minute}, check: func(t *testing.T) {
					expected := &StatTCD{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
								"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
							},
							Value:    utils.NewDecimal(6*int64(time.Minute), 0),
							Count:    3,
							MinItems: 2,
						},
					}
					if !expected.Equal(tcd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
					}
				}},
				{rem: "EVENT_1", check: func(t *testing.T) {
					expected := &StatTCD{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 1},
								"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
							},
							Value:    utils.NewDecimalFromFloat64(float64(3*time.Minute + 30*time.Second)),
							Count:    2,
							MinItems: 2,
						},
					}
					if !expected.Equal(tcd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
					}
				}},
			},
		},
		{
			name:   "ACC",
			metric: acc,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 18.2}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 6.2}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaCost: 18.3}, check: func(t *testing.T) {
					expected := &StatACC{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.2"), CompressFactor: 2},
								"EVENT_3": {Stat: &utils.Decimal{Big: decimal.WithContext(decimal.Context{Precision: 3}).SetFloat64(18.3)}, CompressFactor: 1},
							},
							MinItems: 2,
							Count:    3,
							Value:    utils.NewDecimalFromStringIgnoreError("42.7"),
						},
					}
					if !expected.Equal(acc.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
					}
				}},
				{rem: "EVENT_1", check: func(t *testing.T) {
					expected := &StatACC{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.2"), CompressFactor: 1},
								"EVENT_3": {Stat: &utils.Decimal{Big: decimal.WithContext(decimal.Context{Precision: 3}).SetFloat64(18.3)}, CompressFactor: 1},
							},
							MinItems: 2,
							Count:    2,
							Value:    utils.SubstractDecimal(utils.NewDecimalFromStringIgnoreError("42.7"), utils.NewDecimalFromFloat64(12.2)),
						},
					}
					if !expected.Equal(acc.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
					}
				}},
			},
		},
		{
			name:   "TCC",
			metric: tcc,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 18.2}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 6.2}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaCost: 18.3}, check: func(t *testing.T) {
					expected := &StatTCC{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.20000000000000"), CompressFactor: 2},
								"EVENT_3": {Stat: utils.NewDecimalFromStringIgnoreError("18.300000000000000710542735760100185871124267578125"), CompressFactor: 1},
							},
							MinItems: 2,
							Count:    3,
							Value:    utils.NewDecimalFromStringIgnoreError("42.700"),
						},
					}
					if !expected.Equal(tcc.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
					}
				}},
				{rem: "EVENT_1", check: func(t *testing.T) {
					expected := &StatTCC{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.20000000000000"), CompressFactor: 1},
								"EVENT_3": {Stat: utils.NewDecimalFromStringIgnoreError("18.300000000000000710542735760100185871124267578125"), CompressFactor: 1},
							},
							MinItems: 2,
							Count:    2,
							Value:    utils.SubstractDecimal(utils.NewDecimalFromStringIgnoreError("42.700"), utils.NewDecimalFromFloat64(12.2)),
						},
					}
					if !expected.Equal(tcc.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
					}
				}},
			},
		},
		{
			name:   "PDD",
			metric: pdd,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaPDD: 2 * time.Minute}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaPDD: 3 * time.Minute}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaPDD: time.Minute}, check: func(t *testing.T) {
					expected := &StatPDD{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
								"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
							},
							Value:    utils.NewDecimal(6*int64(time.Minute), 0),
							Count:    3,
							MinItems: 2,
						},
					}
					if !expected.Equal(pdd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
					}
				}},
				{rem: "EVENT_1", check: func(t *testing.T) {
					expected := &StatPDD{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 1},
								"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
							},
							Value:    utils.NewDecimalFromFloat64(float64(3*time.Minute + 30*time.Second)),
							Count:    2,
							MinItems: 2,
						},
					}
					if !expected.Equal(pdd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
					}
				}},
			},
		},
		{
			name:   "Sum",
			metric: sum,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: utils.NewDecimal(182, 1)}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: utils.NewDecimal(62, 1)}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaCost: utils.NewDecimal(183, 1)}, check: func(t *testing.T) {
					expected := &StatSum{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.2"), CompressFactor: 2},
								"EVENT_3": {Stat: utils.NewDecimalFromStringIgnoreError("18.3"), CompressFactor: 1},
							},
							MinItems: 2,
							Count:    3,
							Value:    utils.NewDecimalFromStringIgnoreError("42.7"),
						},
						Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
					}
					if !expected.Equal(sum.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
					}
				}},
				{rem: "EVENT_1", check: func(t *testing.T) {
					expected := &StatSum{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.2"), CompressFactor: 1},
								"EVENT_3": {Stat: utils.NewDecimalFromStringIgnoreError("18.3"), CompressFactor: 1},
							},
							MinItems: 2,
							Count:    2,
							Value:    utils.SubstractDecimal(utils.NewDecimalFromStringIgnoreError("42.7"), utils.NewDecimalFromFloat64(12.2)),
						},
						Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
					}
					if !expected.Equal(sum.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
					}
				}},
			},
		},
		{
			name:   "Average",
			metric: avg,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 18.2}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 6.2}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaCost: 18.3}, check: func(t *testing.T) {
					expected := &StatAverage{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.20000000000000"), CompressFactor: 2},
								"EVENT_3": {Stat: utils.NewDecimalFromStringIgnoreError("18.300000000000000710542735760100185871124267578125"), CompressFactor: 1},
							},
							MinItems: 2,
							Count:    3,
							Value:    utils.NewDecimalFromStringIgnoreError("42.70000000000000"),
						},
						FieldName: "~*opts.*cost",
					}
					if !expected.Equal(avg.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
					}
				}},
				{rem: "EVENT_1", check: func(t *testing.T) {
					expected := &StatAverage{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("12.20000000000000"), CompressFactor: 1},
								"EVENT_3": {Stat: utils.NewDecimalFromStringIgnoreError("18.300000000000000710542735760100185871124267578125"), CompressFactor: 1},
							},
							MinItems: 2,
							Count:    2,
							Value:    utils.SubstractDecimal(utils.NewDecimalFromStringIgnoreError("42.70000000000000"), utils.NewDecimalFromFloat64(12.2)),
						},
						FieldName: "~*opts.*cost",
					}
					if !reflect.DeepEqual(*expected, *avg) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
					}
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.metric
			for i, s := range tt.steps {
				switch {
				case s.add != "":
					if err := m.AddEvent(s.add, utils.MapStorage{utils.MetaOpts: s.opts}); err != nil {
						t.Fatalf("step %d: %v", i, err)
					}
				case s.rem != "":
					if err := m.RemEvent(s.rem); err != nil {
						t.Fatalf("step %d: %v", i, err)
					}
				}
				if s.wantStr != "" {
					if got := m.GetStringValue(roundingDecimals); got != s.wantStr {
						t.Fatalf("step %d: GetStringValue() = %q, want %q", i, got, s.wantStr)
					}
				}
				if s.check != nil {
					s.check(t)
				}
			}
		})
	}
}

func TestStatMetricsCompress(t *testing.T) {
	started := map[string]any{utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}
	asr := &StatASR{Metric: NewMetric(2, nil)}
	acd := &StatACD{Metric: NewMetric(2, nil)}
	tcd := &StatTCD{Metric: NewMetric(2, nil)}
	acc := &StatACC{Metric: NewMetric(2, nil)}
	tcc := &StatTCC{Metric: NewMetric(2, nil)}
	pdd := &StatPDD{Metric: NewMetric(2, nil)}
	ddc := &StatDDC{
		Events:      make(map[string]map[string]uint64),
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    2,
	}
	sum := &StatSum{
		Metric: NewMetric(2, nil),
		Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
	}
	avg := &StatAverage{Metric: NewMetric(2, nil), FieldName: "~*opts.*cost"}
	dst := &StatDistinct{
		Events:      make(map[string]map[string]uint64),
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    2,
		FieldName:   utils.DynamicDataPrefix + utils.MetaOpts + utils.NestingSep + utils.MetaDestination,
	}
	type step struct {
		add      string
		opts     map[string]any
		compress uint64
		wantIDs  []string
		wantStr  string
		check    func(t *testing.T)
	}
	tests := []struct {
		name   string
		metric StatMetric
		steps  []step
	}{
		{
			name:   "ASR",
			metric: asr,
			steps: []step{
				{add: "EVENT_1", opts: started},
				{add: "EVENT_2"},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_2"}, wantStr: "50%", check: func(t *testing.T) {
					expected := &StatASR{
						Metric: &Metric{
							Value:    utils.NewDecimal(1, 0),
							Count:    2,
							MinItems: 2,
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(1, 0), CompressFactor: 1},
								"EVENT_2": {Stat: utils.NewDecimal(0, 0), CompressFactor: 1},
							},
						},
					}
					if !reflect.DeepEqual(*expected, *asr) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
					}
				}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, wantStr: "50%", check: func(t *testing.T) {
					expected := &StatASR{
						Metric: &Metric{
							Value:    utils.NewDecimal(1, 0),
							Count:    2,
							MinItems: 2,
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimalFromFloat64(0.5), CompressFactor: 2},
							},
						},
					}
					if !reflect.DeepEqual(*expected, *asr) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
					}
				}},
				{add: "EVENT_2"},
				{add: "EVENT_1"},
				{compress: 1, wantIDs: []string{"EVENT_3"}, wantStr: "25%", check: func(t *testing.T) {
					expected := &StatASR{
						Metric: &Metric{
							Value:    utils.NewDecimal(1, 0),
							Count:    4,
							MinItems: 2,
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimalFromFloat64(0.25), CompressFactor: 4},
							},
						},
					}
					if !reflect.DeepEqual(*expected, *asr) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(asr))
					}
				}},
			},
		},
		{
			name:   "ACD",
			metric: acd,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 2 * time.Minute}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 3 * time.Minute}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaUsage: time.Minute}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_3"}, check: func(t *testing.T) {
					expected := &StatACD{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
								"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
							},
							Value:    utils.NewDecimal(6*int64(time.Minute), 0),
							Count:    3,
							MinItems: 2,
						},
					}
					if !expected.Equal(acd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
					}
				}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatACD{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 3},
							},
							Value:    utils.NewDecimal(6*int64(time.Minute), 0),
							Count:    3,
							MinItems: 2,
						},
					}
					if !expected.Equal(acd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acd))
					}
				}},
			},
		},
		{
			name:   "TCD",
			metric: tcd,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 2 * time.Minute}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaUsage: 3 * time.Minute}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaUsage: time.Minute}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_3"}, check: func(t *testing.T) {
					expected := &StatTCD{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
								"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
							},
							Value:    utils.NewDecimal(6*int64(time.Minute), 0),
							Count:    3,
							MinItems: 2,
						},
					}
					if !expected.Equal(tcd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
					}
				}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatTCD{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 3},
							},
							Value:    utils.NewDecimal(6*int64(time.Minute), 0),
							Count:    3,
							MinItems: 2,
						},
					}
					if !expected.Equal(tcd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcd))
					}
				}},
			},
		},
		{
			name:   "ACC",
			metric: acc,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 18.2}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 6.2}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_2"}, check: func(t *testing.T) {
					expected := &StatACC{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("18.199999999999999289457264239899814128875732421875"), CompressFactor: 1},
								"EVENT_2": {Stat: utils.NewDecimalFromStringIgnoreError("6.20000000000000017763568394002504646778106689453125"), CompressFactor: 1},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromStringIgnoreError("24.4"),
							Count:    2,
						},
					}
					if !expected.Equal(acc.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
					}
				}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatACC{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromFloat64(24.4),
							Count:    2,
						},
					}
					if !expected.Equal(acc.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
					}
				}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 6.2}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 18.3}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatACC{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.22500000000000), CompressFactor: 4},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromFloat64(48.9),
							Count:    4,
						},
					}
					if !expected.Equal(acc.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(acc))
					}
				}},
			},
		},
		{
			name:   "TCC",
			metric: tcc,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 18.2}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 6.2}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_2"}, check: func(t *testing.T) {
					expected := &StatTCC{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("18.199999999999999289457264239899814128875732421875"), CompressFactor: 1},
								"EVENT_2": {Stat: utils.NewDecimalFromStringIgnoreError("6.20000000000000017763568394002504646778106689453125"), CompressFactor: 1},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromStringIgnoreError("24.400"),
							Count:    2,
						},
					}
					if !expected.Equal(tcc.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
					}
				}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatTCC{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromFloat64(24.4),
							Count:    2,
						},
					}
					if !expected.Equal(tcc.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
					}
				}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 6.2}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 18.3}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatTCC{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.225), CompressFactor: 4},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromFloat64(48.9),
							Count:    4,
						},
					}
					if !expected.Equal(tcc.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(tcc))
					}
				}},
			},
		},
		{
			name:   "PDD",
			metric: pdd,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaPDD: 2 * time.Minute}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaPDD: 3 * time.Minute}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaPDD: time.Minute}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_3"}, check: func(t *testing.T) {
					expected := &StatPDD{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimal(int64(2*time.Minute+30*time.Second), 0), CompressFactor: 2},
								"EVENT_3": {Stat: utils.NewDecimal(int64(time.Minute), 0), CompressFactor: 1},
							},
							Value:    utils.NewDecimal(6*int64(time.Minute), 0),
							Count:    3,
							MinItems: 2,
						},
					}
					if !expected.Equal(pdd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
					}
				}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatPDD{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimal(int64(2*time.Minute), 0), CompressFactor: 3},
							},
							Value:    utils.NewDecimal(6*int64(time.Minute), 0),
							Count:    3,
							MinItems: 2,
						},
					}
					if !expected.Equal(pdd.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(pdd))
					}
				}},
			},
		},
		{
			name:   "DDC",
			metric: ddc,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaDestination: "1001"}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaDestination: "1001"}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaDestination: "1002"}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_3"}, check: func(t *testing.T) {
					expected := &StatDDC{
						Events: map[string]map[string]uint64{
							"EVENT_1": {
								"1001": 2,
							},
							"EVENT_3": {
								"1002": 1,
							},
						},
						FieldValues: map[string]utils.StringSet{
							"1001": {
								"EVENT_1": {},
							},
							"1002": {
								"EVENT_3": {},
							},
						},
						MinItems: 2,
						Count:    3,
					}
					if !reflect.DeepEqual(*expected, *ddc) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
					}
				}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_3"}, check: func(t *testing.T) {
					expected := &StatDDC{
						Events: map[string]map[string]uint64{
							"EVENT_1": {
								"1001": 2,
							},
							"EVENT_3": {
								"1002": 1,
							},
						},
						FieldValues: map[string]utils.StringSet{
							"1001": {
								"EVENT_1": {},
							},
							"1002": {
								"EVENT_3": {},
							},
						},
						MinItems: 2,
						Count:    3,
					}
					if !reflect.DeepEqual(*expected, *ddc) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(ddc))
					}
				}},
			},
		},
		{
			name:   "Sum",
			metric: sum,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: utils.NewDecimal(182, 1)}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: utils.NewDecimal(62, 1)}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_2"}, check: func(t *testing.T) {
					expected := &StatSum{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("18.2"), CompressFactor: 1},
								"EVENT_2": {Stat: utils.NewDecimalFromStringIgnoreError("6.2"), CompressFactor: 1},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromStringIgnoreError("24.4"),
							Count:    2,
						},
						Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
					}
					if !expected.Equal(sum.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
					}
				}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatSum{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromFloat64(24.4),
							Count:    2,
						},
						Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
					}
					if !expected.Equal(sum.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
					}
				}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: utils.NewDecimal(62, 1)}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: utils.NewDecimal(183, 1)}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatSum{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.225), CompressFactor: 4},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromFloat64(48.9),
							Count:    4,
						},
						Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
					}
					if !expected.Equal(sum.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(sum))
					}
				}},
			},
		},
		{
			name:   "Average",
			metric: avg,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 18.2}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 6.2}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_2"}, check: func(t *testing.T) {
					expected := &StatAverage{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_1": {Stat: utils.NewDecimalFromStringIgnoreError("18.199999999999999289457264239899814128875732421875"), CompressFactor: 1},
								"EVENT_2": {Stat: utils.NewDecimalFromStringIgnoreError("6.20000000000000017763568394002504646778106689453125"), CompressFactor: 1},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromStringIgnoreError("24.40000000000000"),
							Count:    2,
						},
						FieldName: "~*opts.*cost",
					}
					if !expected.Equal(avg.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
					}
				}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatAverage{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.2), CompressFactor: 2},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromFloat64(24.4),
							Count:    2,
						},
						FieldName: "~*opts.*cost",
					}
					if !expected.Equal(avg.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
					}
				}},
				{add: "EVENT_2", opts: map[string]any{utils.MetaCost: 6.2}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaCost: 18.3}},
				{compress: 1, wantIDs: []string{"EVENT_3"}, check: func(t *testing.T) {
					expected := &StatAverage{
						Metric: &Metric{
							Events: map[string]*DecimalWithCompress{
								"EVENT_3": {Stat: utils.NewDecimalFromFloat64(12.22500000000000), CompressFactor: 4},
							},
							MinItems: 2,
							Value:    utils.NewDecimalFromFloat64(48.90000000000000),
							Count:    4,
						},
						FieldName: "~*opts.*cost",
					}
					if !expected.Equal(avg.Metric) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(avg))
					}
				}},
			},
		},
		{
			name:   "Distinct",
			metric: dst,
			steps: []step{
				{add: "EVENT_1", opts: map[string]any{utils.MetaDestination: "1001"}},
				{add: "EVENT_1", opts: map[string]any{utils.MetaDestination: "1001"}},
				{add: "EVENT_3", opts: map[string]any{utils.MetaDestination: "1002"}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_3"}, check: func(t *testing.T) {
					expected := &StatDistinct{
						Events: map[string]map[string]uint64{
							"EVENT_1": {
								"1001": 2,
							},
							"EVENT_3": {
								"1002": 1,
							},
						},
						FieldValues: map[string]utils.StringSet{
							"1001": {
								"EVENT_1": {},
							},
							"1002": {
								"EVENT_3": {},
							},
						},
						MinItems:  2,
						FieldName: utils.DynamicDataPrefix + utils.MetaOpts + utils.NestingSep + utils.MetaDestination,
						Count:     3,
					}
					if !reflect.DeepEqual(*expected, *dst) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(dst))
					}
				}},
				{compress: 10, wantIDs: []string{"EVENT_1", "EVENT_3"}, check: func(t *testing.T) {
					expected := &StatDistinct{
						Events: map[string]map[string]uint64{
							"EVENT_1": {
								"1001": 2,
							},
							"EVENT_3": {
								"1002": 1,
							},
						},
						FieldValues: map[string]utils.StringSet{
							"1001": {
								"EVENT_1": {},
							},
							"1002": {
								"EVENT_3": {},
							},
						},
						MinItems:  2,
						FieldName: utils.DynamicDataPrefix + utils.MetaOpts + utils.NestingSep + utils.MetaDestination,
						Count:     3,
					}
					if !reflect.DeepEqual(*expected, *dst) {
						t.Errorf("Expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(dst))
					}
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.metric
			for i, s := range tt.steps {
				switch {
				case s.add != "":
					if err := m.AddEvent(s.add, utils.MapStorage{utils.MetaOpts: s.opts}); err != nil {
						t.Fatalf("step %d: %v", i, err)
					}
				case s.compress != 0:
					rply := m.Compress(s.compress, "EVENT_3")
					sort.Strings(rply)
					if !reflect.DeepEqual(s.wantIDs, rply) {
						t.Fatalf("step %d: Compress() = %s, want %s", i, utils.ToJSON(rply), utils.ToJSON(s.wantIDs))
					}
				}
				if s.wantStr != "" {
					if got := m.GetStringValue(roundingDecimals); got != s.wantStr {
						t.Fatalf("step %d: GetStringValue() = %q, want %q", i, got, s.wantStr)
					}
				}
				if s.check != nil {
					s.check(t)
				}
			}
		})
	}
}

func TestStatMetricsGetCompressFactor(t *testing.T) {
	started := map[string]any{utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}
	sum, err := NewStatSum(2, "~*opts.*cost", nil)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name                string
		metric              StatMetric
		ev1, ev2, ev4       string
		opts1, opts2, opts4 map[string]any
		cf1, cf2, cf3       map[string]uint64
	}{
		{
			name:   "ASR",
			metric: NewASR(2, "", nil),
			ev1:    "EVENT_1", opts1: started,
			ev2: "EVENT_2",
			ev4: "EVENT_1",
			cf1: map[string]uint64{"EVENT_1": 1, "EVENT_2": 1},
			cf2: map[string]uint64{"EVENT_1": 1, "EVENT_2": 2},
			cf3: map[string]uint64{"EVENT_1": 2, "EVENT_2": 3},
		},
		{
			name:   "ACD",
			metric: NewACD(2, "", nil),
			ev1:    "EVENT_1", opts1: map[string]any{utils.MetaUsage: time.Minute},
			ev2: "EVENT_2", opts2: map[string]any{utils.MetaUsage: time.Minute},
			ev4: "EVENT_2", opts4: map[string]any{utils.MetaUsage: 2 * time.Minute},
			cf1: map[string]uint64{"EVENT_1": 1, "EVENT_2": 1},
			cf2: map[string]uint64{"EVENT_1": 1, "EVENT_2": 2},
			cf3: map[string]uint64{"EVENT_1": 1, "EVENT_2": 3},
		},
		{
			name:   "TCD",
			metric: NewTCD(2, "", nil),
			ev1:    "EVENT_1", opts1: map[string]any{utils.MetaUsage: time.Minute},
			ev2: "EVENT_2", opts2: map[string]any{utils.MetaUsage: time.Minute},
			ev4: "EVENT_2", opts4: map[string]any{utils.MetaUsage: 2 * time.Minute},
			cf1: map[string]uint64{"EVENT_1": 1, "EVENT_2": 1},
			cf2: map[string]uint64{"EVENT_1": 1, "EVENT_2": 2},
			cf3: map[string]uint64{"EVENT_1": 1, "EVENT_2": 3},
		},
		{
			name:   "ACC",
			metric: NewACC(2, "", nil),
			ev1:    "EVENT_1", opts1: map[string]any{utils.MetaCost: 18.2},
			ev2: "EVENT_2", opts2: map[string]any{utils.MetaCost: 18.2},
			ev4: "EVENT_1", opts4: map[string]any{utils.MetaCost: 18.2},
			cf1: map[string]uint64{"EVENT_1": 1, "EVENT_2": 1},
			cf2: map[string]uint64{"EVENT_1": 1, "EVENT_2": 2},
			cf3: map[string]uint64{"EVENT_1": 2, "EVENT_2": 3},
		},
		{
			name:   "TCC",
			metric: NewTCC(2, "", nil),
			ev1:    "EVENT_1", opts1: map[string]any{utils.MetaCost: 18.2},
			ev2: "EVENT_2", opts2: map[string]any{utils.MetaCost: 18.2},
			ev4: "EVENT_1", opts4: map[string]any{utils.MetaCost: 18.2},
			cf1: map[string]uint64{"EVENT_1": 1, "EVENT_2": 1},
			cf2: map[string]uint64{"EVENT_1": 1, "EVENT_2": 2},
			cf3: map[string]uint64{"EVENT_1": 2, "EVENT_2": 3},
		},
		{
			name:   "PDD",
			metric: NewPDD(2, "", nil),
			ev1:    "EVENT_1", opts1: map[string]any{utils.MetaPDD: time.Minute},
			ev2: "EVENT_2", opts2: map[string]any{utils.MetaPDD: time.Minute},
			ev4: "EVENT_2", opts4: map[string]any{utils.MetaPDD: 2 * time.Minute},
			cf1: map[string]uint64{"EVENT_1": 1, "EVENT_2": 1},
			cf2: map[string]uint64{"EVENT_1": 1, "EVENT_2": 2},
			cf3: map[string]uint64{"EVENT_1": 1, "EVENT_2": 3},
		},
		{
			name:   "DDC",
			metric: NewDDC(2, "", nil),
			ev1:    "EVENT_1", opts1: map[string]any{utils.MetaDestination: "1002"},
			ev2: "EVENT_2", opts2: map[string]any{utils.MetaDestination: "1001"},
			ev4: "EVENT_2", opts4: map[string]any{utils.MetaDestination: "1001"},
			cf1: map[string]uint64{"EVENT_1": 1, "EVENT_2": 1},
			cf2: map[string]uint64{"EVENT_1": 1, "EVENT_2": 2},
			cf3: map[string]uint64{"EVENT_1": 1, "EVENT_2": 3},
		},
		{
			name:   "Sum",
			metric: sum,
			ev1:    "EVENT_1", opts1: map[string]any{utils.MetaCost: 18.2},
			ev2: "EVENT_2", opts2: map[string]any{utils.MetaCost: 18.2},
			ev4: "EVENT_1", opts4: map[string]any{utils.MetaCost: 18.2},
			cf1: map[string]uint64{"EVENT_1": 1, "EVENT_2": 1},
			cf2: map[string]uint64{"EVENT_1": 1, "EVENT_2": 2},
			cf3: map[string]uint64{"EVENT_1": 2, "EVENT_2": 3},
		},
		{
			name:   "Average",
			metric: NewStatAverage(2, "~*opts.*cost", nil),
			ev1:    "EVENT_1", opts1: map[string]any{utils.MetaCost: 18.2},
			ev2: "EVENT_2", opts2: map[string]any{utils.MetaCost: 18.2},
			ev4: "EVENT_1", opts4: map[string]any{utils.MetaCost: 18.2},
			cf1: map[string]uint64{"EVENT_1": 1, "EVENT_2": 1},
			cf2: map[string]uint64{"EVENT_1": 1, "EVENT_2": 2},
			cf3: map[string]uint64{"EVENT_1": 2, "EVENT_2": 3},
		},
		{
			name:   "Distinct",
			metric: NewStatDistinct(2, utils.DynamicDataPrefix+utils.MetaOpts+utils.NestingSep+utils.MetaDestination, nil),
			ev1:    "EVENT_1", opts1: map[string]any{utils.MetaDestination: "1002"},
			ev2: "EVENT_2", opts2: map[string]any{utils.MetaDestination: "1001"},
			ev4: "EVENT_2", opts4: map[string]any{utils.MetaDestination: "1001"},
			cf1: map[string]uint64{"EVENT_1": 1, "EVENT_2": 1},
			cf2: map[string]uint64{"EVENT_1": 1, "EVENT_2": 2},
			cf3: map[string]uint64{"EVENT_1": 1, "EVENT_2": 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.metric
			if err := m.AddEvent(tt.ev1, utils.MapStorage{utils.MetaOpts: tt.opts1}); err != nil {
				t.Fatal(err)
			}
			if err := m.AddEvent(tt.ev2, utils.MapStorage{utils.MetaOpts: tt.opts2}); err != nil {
				t.Fatal(err)
			}
			CF := m.GetCompressFactor(make(map[string]uint64))
			if !reflect.DeepEqual(tt.cf1, CF) {
				t.Errorf("Expected: %s , received: %s", utils.ToJSON(tt.cf1), utils.ToJSON(CF))
			}
			if err := m.AddEvent(tt.ev2, utils.MapStorage{utils.MetaOpts: tt.opts2}); err != nil {
				t.Fatal(err)
			}
			CF = m.GetCompressFactor(make(map[string]uint64))
			if !reflect.DeepEqual(tt.cf2, CF) {
				t.Errorf("Expected: %s , received: %s", utils.ToJSON(tt.cf2), utils.ToJSON(CF))
			}
			if err := m.AddEvent(tt.ev4, utils.MapStorage{utils.MetaOpts: tt.opts4}); err != nil {
				t.Fatal(err)
			}
			CF["EVENT_2"] = 3
			CF = m.GetCompressFactor(CF)
			if !reflect.DeepEqual(tt.cf3, CF) {
				t.Errorf("Expected: %s , received: %s", utils.ToJSON(tt.cf3), utils.ToJSON(CF))
			}
		})
	}
}

func TestStatMetricsGetCompressFactor2(t *testing.T) {
	tests := []struct {
		name   string
		metric StatMetric
		give   map[string]uint64
		want   map[string]uint64
	}{
		{
			name: "ACD",
			metric: &StatACD{
				Metric: &Metric{
					Events: map[string]*DecimalWithCompress{
						"Event1": {
							Stat:           utils.NewDecimal(int64(time.Second), 0),
							CompressFactor: 200000000,
						},
					},
					MinItems: 3,
					Count:    3,
				},
			},
			give: map[string]uint64{"Event1": 1000000},
			want: map[string]uint64{"Event1": 200000000},
		},
		{
			name: "TCD",
			metric: &StatTCD{
				Metric: &Metric{
					Events: map[string]*DecimalWithCompress{
						"Event1": {
							Stat:           utils.NewDecimal(int64(time.Second), 0),
							CompressFactor: 200000000,
						},
					},
					MinItems: 3,
					Count:    3,
				},
			},
			give: map[string]uint64{"Event1": 1000000},
			want: map[string]uint64{"Event1": 200000000},
		},
		{
			name: "PDD",
			metric: &StatPDD{
				Metric: &Metric{
					Events: map[string]*DecimalWithCompress{
						"Event1": {
							Stat:           utils.NewDecimal(int64(time.Second), 0),
							CompressFactor: 200000000,
						},
					},
					MinItems: 3,
					Count:    3,
				},
			},
			give: map[string]uint64{"Event1": 1000000},
			want: map[string]uint64{"Event1": 200000000},
		},
		{
			name: "DDC",
			metric: &StatDDC{
				Events: map[string]map[string]uint64{
					"Event1": {
						"Event1": 200000000,
					},
				},
				MinItems: 3,
				Count:    3,
			},
			give: map[string]uint64{"Event1": 1000000},
			want: map[string]uint64{"Event1": 200000000},
		},
		{
			name: "Distinct",
			metric: &StatDistinct{
				FieldValues: map[string]utils.StringSet{},
				Events: map[string]map[string]uint64{
					"Event1": {
						"1": 10000000000,
					},
					"Event2": {
						"2": 20000000000,
					},
				},
				MinItems:  3,
				FieldName: "Test_Field_Name",
				Count:     3,
			},
			give: map[string]uint64{"Event1": 1},
			want: map[string]uint64{"Event1": 10000000000, "Event2": 20000000000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.metric.GetCompressFactor(tt.give); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected: %s , received: %s", utils.ToJSON(tt.want), utils.ToJSON(got))
			}
		})
	}
}

var jMarshaler utils.JSONMarshaler

func TestStatMetricsMarshal(t *testing.T) {
	startTime := time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)
	asr, err := NewStatMetric(utils.MetaASR, 2, []string{"*string:Account:1001"})
	if err != nil {
		t.Error(err)
	}
	sum, err := NewStatSum(2, "~*opts.*cost", nil)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		metric StatMetric
		opts   map[string]any
		want   []byte
		into   any
	}{
		{
			name:   "ASR",
			metric: asr,
			opts:   map[string]any{utils.MetaStartTime: startTime},
			want:   []byte(`{"Value":1,"Count":1,"Events":{"EVENT_1":{"Stat":1,"CompressFactor":1}},"MinItems":2,"FilterIDs":["*string:Account:1001"]}`),
			into:   new(StatASR),
		},
		{
			name:   "ACD",
			metric: NewACD(2, "", nil),
			opts:   map[string]any{utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second},
			want:   []byte(`{"Value":10000000000,"Count":1,"Events":{"EVENT_1":{"Stat":10000000000,"CompressFactor":1}},"MinItems":2,"FilterIDs":null}`),
			into:   new(StatACD),
		},
		{
			name:   "TCD",
			metric: NewTCD(2, "", nil),
			opts:   map[string]any{utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second},
			want:   []byte(`{"Value":10000000000,"Count":1,"Events":{"EVENT_1":{"Stat":10000000000,"CompressFactor":1}},"MinItems":2,"FilterIDs":null}`),
			into:   new(StatTCD),
		},
		{
			name:   "ACC",
			metric: NewACC(2, "", nil),
			opts:   map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: "12.3"},
			want:   []byte(`{"Value":12.3,"Count":1,"Events":{"EVENT_1":{"Stat":12.3,"CompressFactor":1}},"MinItems":2,"FilterIDs":null}`),
			into:   new(StatACC),
		},
		{
			name:   "TCC",
			metric: NewTCC(2, "", nil),
			opts:   map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: "12.3"},
			want:   []byte(`{"Value":12.3,"Count":1,"Events":{"EVENT_1":{"Stat":12.3,"CompressFactor":1}},"MinItems":2,"FilterIDs":null}`),
			into:   new(StatTCC),
		},
		{
			name:   "PDD",
			metric: NewPDD(2, "", nil),
			opts:   map[string]any{utils.MetaPDD: 5 * time.Second, utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second},
			want:   []byte(`{"Value":5000000000,"Count":1,"Events":{"EVENT_1":{"Stat":5000000000,"CompressFactor":1}},"MinItems":2,"FilterIDs":null}`),
			into:   new(StatPDD),
		},
		{
			name:   "DDC",
			metric: NewDDC(2, "", nil),
			opts:   map[string]any{utils.MetaDestination: "1002", utils.MetaPDD: 5 * time.Second, utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second},
			want:   []byte(`{"FieldValues":{"1002":{"EVENT_1":{}}},"Events":{"EVENT_1":{"1002":1}},"MinItems":2,"Count":1,"FilterIDs":null}`),
			into:   new(StatDDC),
		},
		{
			name:   "Sum",
			metric: sum,
			opts:   map[string]any{utils.MetaDestination: "1002", utils.MetaPDD: 5 * time.Second, utils.MetaCost: "20", utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second},
			want:   []byte(`{"Value":20,"Count":1,"Events":{"EVENT_1":{"Stat":20,"CompressFactor":1}},"MinItems":2,"FilterIDs":null,"Fields":[{"Rules":"~*opts.*cost","Path":"~*opts.*cost"}]}`),
			into:   new(StatSum),
		},
		{
			name:   "Average",
			metric: NewStatAverage(2, "~*opts.*cost", nil),
			opts:   map[string]any{utils.MetaDestination: "1002", utils.MetaPDD: 5 * time.Second, utils.MetaCost: "20", utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second},
			want:   []byte(`{"Value":20,"Count":1,"Events":{"EVENT_1":{"Stat":20,"CompressFactor":1}},"MinItems":2,"FilterIDs":null,"FieldName":"~*opts.*cost"}`),
			into:   new(StatAverage),
		},
		{
			name:   "Distinct",
			metric: NewStatDistinct(2, "~*opts.*usage", nil),
			opts:   map[string]any{utils.MetaDestination: "1002", utils.MetaPDD: 5 * time.Second, utils.MetaCost: "20", utils.MetaStartTime: startTime, utils.MetaUsage: 10 * time.Second},
			want:   []byte(`{"FieldValues":{"10s":{"EVENT_1":{}}},"Events":{"EVENT_1":{"10s":1}},"MinItems":2,"FieldName":"~*opts.*usage","Count":1,"FilterIDs":null}`),
			into:   new(StatDistinct),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.metric.AddEvent("EVENT_1", utils.MapStorage{utils.MetaOpts: tt.opts}); err != nil {
				t.Fatal(err)
			}
			b, err := jMarshaler.Marshal(tt.metric)
			if err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(tt.want, b) {
				t.Errorf("Expected: %s , received: %s", string(tt.want), string(b))
			} else if err := jMarshaler.Unmarshal(b, tt.into); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestStatMetricsNewStatMetricError(t *testing.T) {
	_, err := NewStatMetric("", 0, []string{})
	if err == nil || err.Error() != "unsupported metric type <>" {
		t.Errorf("\nExpecting <unsupported metric type>,\nRecevied  <%+v>", err)
	}

}

func TestStatMetricsGetMinItems(t *testing.T) {
	tests := []struct {
		name   string
		metric StatMetric
		want   uint64
	}{
		{name: "ASR", metric: NewASR(2, "", nil), want: 2},
		{name: "ACD", metric: &StatACD{Metric: NewMetric(2, nil)}, want: 2},
		{name: "TCD", metric: &StatTCD{Metric: NewMetric(2, nil)}, want: 2},
		{name: "ACC", metric: &StatACC{Metric: NewMetric(2, nil)}, want: 2},
		{name: "TCC", metric: &StatTCC{Metric: NewMetric(2, nil)}, want: 2},
		{name: "PDD", metric: &StatPDD{Metric: NewMetric(2, nil)}, want: 2},
		{
			name: "DDC",
			metric: &StatDDC{
				Count: 15,
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 2,
					},
					"Event2": {},
				},
				MinItems: 20,
			},
			want: 20,
		},
		{name: "Sum", metric: &StatSum{Metric: NewMetric(20, nil)}, want: 20},
		{
			name: "Average",
			metric: &StatAverage{
				Metric:    NewMetric(10, nil),
				FieldName: "Test_Field_Name",
			},
			want: 10,
		},
		{
			name: "Distinct",
			metric: &StatDistinct{
				FieldValues: map[string]utils.StringSet{},
				Events:      map[string]map[string]uint64{},
				MinItems:    3,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.metric.GetMinItems(); got != tt.want {
				t.Errorf("GetMinItems() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestStatMetricsRemEvent(t *testing.T) {
	tests := []struct {
		name    string
		metric  StatMetric
		remID   string
		wantErr error
		want    StatMetric
	}{
		{
			name: "Distinct not found",
			metric: &StatDistinct{
				FieldValues: map[string]utils.StringSet{},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 1,
					},
					"Event2": {},
				},
				MinItems:  3,
				FieldName: "Test_Field_Name",
				Count:     3,
			},
			remID:   "Event2",
			wantErr: utils.ErrNotFound,
		},
		{
			name: "Distinct",
			metric: &StatDistinct{
				FieldValues: map[string]utils.StringSet{},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 1,
					},
					"Event2": {},
				},
				MinItems:  3,
				FieldName: "Test_Field_Name",
				Count:     3,
			},
			remID: "Event1",
			want: &StatDistinct{
				FieldValues: map[string]utils.StringSet{},
				Events: map[string]map[string]uint64{
					"Event1": {},
					"Event2": {},
				},
				MinItems:  3,
				FieldName: "Test_Field_Name",
				Count:     2,
			},
		},
		{
			name: "Distinct compressed",
			metric: &StatDistinct{
				FieldValues: map[string]utils.StringSet{
					"FieldValue1": {},
				},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 2,
					},
					"Event2": {},
				},
				MinItems:  3,
				FieldName: "Test_Field_Name",
				Count:     3,
			},
			remID: "Event1",
			want: &StatDistinct{
				FieldValues: map[string]utils.StringSet{
					"FieldValue1": {},
				},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 1,
					},
					"Event2": {},
				},
				MinItems:  3,
				FieldName: "Test_Field_Name",
				Count:     2,
			},
		},
		{
			name: "DDC not found",
			metric: &StatDDC{
				FieldValues: map[string]utils.StringSet{},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 1,
					},
					"Event2": {},
				},
				MinItems: 3,
				Count:    3,
			},
			remID:   "Event2",
			wantErr: utils.ErrNotFound,
		},
		{
			name: "DDC",
			metric: &StatDDC{
				FieldValues: map[string]utils.StringSet{},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 1,
					},
					"Event2": {},
				},
				MinItems: 3,
				Count:    3,
			},
			remID: "Event1",
			want: &StatDDC{
				FieldValues: map[string]utils.StringSet{},
				Events: map[string]map[string]uint64{
					"Event1": {},
					"Event2": {},
				},
				MinItems: 3,
				Count:    2,
			},
		},
		{
			name: "DDC compressed",
			metric: &StatDDC{
				FieldValues: map[string]utils.StringSet{
					"FieldValue1": {},
				},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 2,
					},
					"Event2": {},
				},
				MinItems: 3,
				Count:    3,
			},
			remID: "Event1",
			want: &StatDDC{
				FieldValues: map[string]utils.StringSet{
					"FieldValue1": {},
				},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 1,
					},
					"Event2": {},
				},
				MinItems: 3,
				Count:    2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metric.RemEvent(tt.remID)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("RemEvent(%q) = %v, want %v", tt.remID, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("RemEvent(%q): %v", tt.remID, err)
			}
			if !reflect.DeepEqual(tt.want, tt.metric) {
				t.Errorf("Expecting <%+v>,\n Recevied <%+v>", tt.want, tt.metric)
			}
		})
	}
}

func TestStatMetricsAddEventErr(t *testing.T) {
	startTime := time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)
	tests := []struct {
		name    string
		metric  StatMetric
		evID    string
		dp      utils.DataProvider
		wantErr string
	}{
		{
			name:    "ASR unsupported time format",
			metric:  &StatASR{Metric: NewMetric(2, nil)},
			evID:    "EVENT_1",
			dp:      utils.MapStorage{utils.MetaOpts: map[string]any{utils.MetaStartTime: "10"}},
			wantErr: "Unsupported time format",
		},
		{
			name:    "ASR bad time type",
			metric:  &StatASR{Metric: NewMetric(2, nil)},
			evID:    "EVENT_1",
			dp:      utils.MapStorage{utils.MetaOpts: utils.MapStorage{utils.MetaStartTime: false}},
			wantErr: "cannot convert field: false to time.Time",
		},
		{
			name:    "ASR field error",
			metric:  &StatASR{Metric: NewMetric(2, nil)},
			evID:    "EVENT_1",
			dp:      new(mockDP),
			wantErr: utils.ErrAccountNotFound.Error(),
		},
		{
			name:    "ACC bad cost",
			metric:  NewACC(2, "", nil),
			evID:    "EVENT_1",
			dp:      utils.MapStorage{utils.MetaOpts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: "wrong input"}},
			wantErr: "can't convert <wrong input> to decimal",
		},
		{
			name:    "TCC bad cost",
			metric:  NewTCC(2, "", nil),
			evID:    "EVENT_1",
			dp:      utils.MapStorage{utils.MetaOpts: map[string]any{utils.MetaStartTime: startTime, utils.MetaCost: "wrong input"}},
			wantErr: "can't convert <wrong input> to decimal",
		},
		{
			name: "Distinct bad field name",
			metric: &StatDistinct{
				FieldValues: map[string]utils.StringSet{
					"FieldValue1": {},
				},
				Events: map[string]map[string]uint64{
					"Event1": {
						"FieldValue1": 2,
					},
					"Event2": {},
				},
				MinItems:  3,
				FieldName: "Test_Field_Name",
				Count:     3,
			},
			evID:    "Event1",
			dp:      utils.MapStorage{utils.MetaOpts: map[string]any{utils.MetaStartTime: startTime}},
			wantErr: "invalid format for field <Test_Field_Name>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.metric.AddEvent(tt.evID, tt.dp); err == nil || err.Error() != tt.wantErr {
				t.Errorf("AddEvent(%q) = %v, want %q", tt.evID, err, tt.wantErr)
			}
		})
	}
}

func TestStatMetricsStatACDAddEventErr(t *testing.T) {
	acd := NewMetric(2, nil)
	err := acd.addEvent("EVENT_1", false)
	if err == nil || err.Error() != "cannot convert field: bool to decimal.Big" {
		t.Errorf("\nExpecting <cannot convert field: false to time.Duration>,\n Recevied <%+v>", err)
	}
}

type mockDP struct{}

func (mockDP) String() string {
	return ""
}

func (mockDP) FieldAsInterface(fldPath []string) (any, error) {
	return nil, utils.ErrAccountNotFound
}

func (mockDP) FieldAsString([]string) (string, error) {
	return "", nil
}

func TestStatMetricsClone(t *testing.T) {
	tests := []struct {
		name   string
		metric StatMetric
	}{
		{name: "ASR", metric: &StatASR{Metric: NewMetric(2, nil)}},
		{name: "ACD", metric: &StatACD{Metric: NewMetric(2, nil)}},
		{name: "TCD", metric: &StatTCD{Metric: NewMetric(2, nil)}},
		{name: "ACC", metric: &StatACC{Metric: NewMetric(2, nil)}},
		{name: "TCC", metric: &StatTCC{Metric: NewMetric(2, nil)}},
		{name: "PDD", metric: &StatPDD{Metric: NewMetric(2, nil)}},
		{
			name: "DDC",
			metric: &StatDDC{
				Events: map[string]map[string]uint64{
					"EVENT_1": {
						"1001": 2,
					},
					"EVENT_3": {
						"1002": 1,
					},
				},
				FieldValues: map[string]utils.StringSet{
					"1001": {
						"EVENT_1": {},
					},
					"1002": {
						"EVENT_3": {},
					},
				},
				MinItems: 2,
				Count:    3,
			},
		},
		{
			name: "Sum",
			metric: &StatSum{
				Metric: NewMetric(2, nil),
				Fields: utils.NewRSRParsersMustCompile("~*opts.*cost", utils.InfieldSep),
			},
		},
		{name: "Average", metric: &StatAverage{Metric: NewMetric(2, nil), FieldName: "~*opts.*cost"}},
		{
			name: "Distinct",
			metric: &StatDistinct{
				Events: map[string]map[string]uint64{
					"EVENT_1": {
						"1001": 2,
					},
					"EVENT_3": {
						"1002": 1,
					},
				},
				FieldValues: map[string]utils.StringSet{
					"1001": {
						"EVENT_1": {},
					},
					"1002": {
						"EVENT_3": {},
					},
				},
				MinItems: 2,
				Count:    3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rcv := tt.metric.Clone(); !reflect.DeepEqual(rcv, tt.metric) {
				t.Errorf("Clone() = %+v, want %+v", rcv, tt.metric)
			}
		})
	}
}

func TestDDCGetFilterIDs(t *testing.T) {

	ddc := NewDDC(2, "", []string{"flt1", "flt2"})

	exp := &StatDDC{
		FilterIDs: []string{"flt1", "flt2"},
	}

	if rcv := ddc.GetFilterIDs(); !reflect.DeepEqual(rcv, exp.FilterIDs) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", utils.ToJSON(exp.FilterIDs), utils.ToJSON(rcv))
	}

}

func TestMetricClone(t *testing.T) {

	sum := &Metric{
		Value: utils.NewDecimal(2, 0),
		Events: map[string]*DecimalWithCompress{
			"Event1": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 200000000,
			},
		},
		MinItems: 3,
		Count:    3,
	}

	if rcv := sum.Clone(); !reflect.DeepEqual(rcv, sum) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", sum, rcv)
	}
}

func TestMetricEqualFalse(t *testing.T) {

	sum := &Metric{
		Value: utils.NewDecimal(2, 0),
		Events: map[string]*DecimalWithCompress{
			"Event1": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 200000000,
			},
		},
		MinItems: 3,
		Count:    3,
	}

	sum2 := &Metric{
		Value: utils.NewDecimal(2, 0),
		Events: map[string]*DecimalWithCompress{
			"Event1": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 200000000,
			},
			"Event2": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 200000000,
			},
		},
		MinItems: 3,
		Count:    3,
	}

	if rcv := sum.Equal(sum2); rcv {
		t.Errorf("Expecting to not be equal, Recevied equal <%v>", rcv)
	}
}

func TestMetricEqualEventFalse(t *testing.T) {

	sum := &Metric{
		Value: utils.NewDecimal(2, 0),
		Events: map[string]*DecimalWithCompress{
			"even1": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 200000000,
			},
		},
		MinItems: 3,
		Count:    3,
	}

	sum2 := &Metric{
		Value: utils.NewDecimal(2, 0),
		Events: map[string]*DecimalWithCompress{
			"even1": {
				Stat:           utils.NewDecimal(int64(time.Second), 0),
				CompressFactor: 1,
			},
		},
		MinItems: 3,
		Count:    3,
	}

	if rcv := sum.Equal(sum2); rcv {
		t.Errorf("Expecting to not be equal, Recevied equal <%v>", rcv)
	}
}

func TestStatDistinctGetFilterIDs(t *testing.T) {

	dst := NewStatDistinct(2, "", []string{"flt1", "flt2"})

	exp := &StatDistinct{
		FilterIDs: []string{"flt1", "flt2"},
	}

	if rcv := dst.GetFilterIDs(); !reflect.DeepEqual(rcv, exp.FilterIDs) {
		t.Errorf("Expecting <%+v>,\n Recevied <%+v>", utils.ToJSON(exp.FilterIDs), utils.ToJSON(rcv))
	}

}

func TestMetricAddOneEvent(t *testing.T) {
	tests := []struct {
		name        string
		initialVal  *utils.Decimal
		initialCnt  uint64
		input       any
		expectErr   bool
		expectVal   *decimal.Big
		expectCount uint64
	}{
		{
			name:        "Int input",
			input:       42,
			expectErr:   false,
			expectVal:   utils.NewDecimal(42, 0).Big,
			expectCount: 1,
		},
		{
			name:        "Duration input",
			input:       time.Duration(5),
			expectErr:   false,
			expectVal:   utils.NewDecimal(5, 0).Big,
			expectCount: 1,
		},
		{
			name:        "Add to existing value",
			initialVal:  &utils.Decimal{Big: utils.NewDecimal(10, 0).Big},
			initialCnt:  1,
			input:       15,
			expectErr:   false,
			expectVal:   utils.NewDecimal(25, 0).Big,
			expectCount: 2,
		},
		{
			name:        "Invalid type input",
			input:       struct{}{},
			expectErr:   true,
			expectVal:   nil,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				Value: tt.initialVal,
				Count: tt.initialCnt,
			}

			err := m.addOneEvent(tt.input)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectVal == nil {
				if m.Value != nil {
					t.Errorf("Expected nil Value, got: %v", m.Value.Big)
				}
			} else {
				if m.Value == nil || m.Value.Big.Cmp(tt.expectVal) != 0 {
					t.Errorf("Expected Value: %v, got: %v", tt.expectVal, m.Value.Big)
				}
			}

			if m.Count != tt.expectCount {
				t.Errorf("Expected Count: %d, got: %d", tt.expectCount, m.Count)
			}
		})
	}
}

func TestStatDDCClones(t *testing.T) {
	original := &StatDDC{
		FieldValues: map[string]utils.StringSet{
			"field1": utils.NewStringSet([]string{"ID", "ID1"}),
		},
		Events: map[string]map[string]uint64{
			"cgrates.org": {
				"val1": 5,
			},
		},
		MinItems:  2,
		Count:     10,
		FilterIDs: []string{"f1", "f2"},
	}

	cloned := original.Clone().(*StatDDC)

	if !reflect.DeepEqual(original, cloned) {
		t.Errorf("Cloned StatDDC is not equal to original\nOriginal: %+v\nCloned: %+v", original, cloned)
	}

	cloned.Count = 20
	cloned.Events["cgrates.org"]["val1"] = 99
	cloned.FieldValues["field1"].Add("c")
	cloned.FilterIDs[0] = "modified"

	if reflect.DeepEqual(original, cloned) {
		t.Error("Original StatDDC changed after modifying clone")
	}

	t.Run("nil receiver returns nil", func(t *testing.T) {
		var ddc *StatDDC
		if ddc.Clone() != nil {
			t.Error("Expected nil Clone result from nil receiver, got non-nil")
		}
	})
}

func TestStatDistinctClones(t *testing.T) {
	original := &StatDistinct{
		FieldValues: map[string]utils.StringSet{
			"field1": utils.NewStringSet([]string{"ID", "ID1"}),
		},
		Events: map[string]map[string]uint64{
			"cgrates.org": {
				"val1": 5,
			},
		},
		MinItems:  2,
		Count:     10,
		FieldName: "testField",
		FilterIDs: []string{"f1", "f2"},
	}

	cloned := original.Clone().(*StatDistinct)

	if !reflect.DeepEqual(original, cloned) {
		t.Errorf("Cloned StatDistinct is not equal to original\nOriginal: %+v\nCloned: %+v", original, cloned)
	}

	cloned.Count = 20
	cloned.Events["cgrates.org"]["val1"] = 99
	cloned.FieldValues["field1"].Add("c")
	cloned.FilterIDs[0] = "modified"
	cloned.FieldName = "modifiedField"

	if reflect.DeepEqual(original, cloned) {
		t.Error("Original StatDistinct changed after modifying clone")
	}

	t.Run("nil receiver returns nil", func(t *testing.T) {
		var dst *StatDistinct
		if dst.Clone() != nil {
			t.Error("Expected nil Clone result from nil receiver, got non-nil")
		}
	})
}
