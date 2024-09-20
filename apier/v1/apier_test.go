/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package v1

import (
	"testing"

	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestCheckDefaultTiming(t *testing.T) {
	tests := []struct {
		name      string
		tStr      string
		wantID    string
		wantIsDef bool
	}{
		{"Every Minute", utils.MetaEveryMinute, utils.MetaEveryMinute, true},
		{"Hourly", utils.MetaHourly, utils.MetaHourly, true},
		{"Daily", utils.MetaDaily, utils.MetaDaily, true},
		{"Weekly", utils.MetaWeekly, utils.MetaWeekly, true},
		{"Monthly", utils.MetaMonthly, utils.MetaMonthly, true},
		{"Monthly Estimated", utils.MetaMonthlyEstimated, utils.MetaMonthlyEstimated, true},
		{"Month End", utils.MetaMonthEnd, utils.MetaMonthEnd, true},
		{"Yearly", utils.MetaYearly, utils.MetaYearly, true},
		{"Unknown", "unknown", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, isDef := checkDefaultTiming(tt.tStr)
			if isDef != tt.wantIsDef {
				t.Errorf("checkDefaultTiming(%q) isDefault = %v, want %v", tt.tStr, isDef, tt.wantIsDef)
			}
			if isDef && got.ID != tt.wantID {
				t.Errorf("checkDefaultTiming(%q) got.ID = %v, want %v", tt.tStr, got.ID, tt.wantID)
			}
			if !isDef && got != nil {
				t.Errorf("checkDefaultTiming(%q) expected nil, got non-nil", tt.tStr)
			}
		})
	}
}

func TestGetId(t *testing.T) {
	tests := []struct {
		name       string
		attr       AttrRemoveRatingProfile
		expectedID string
	}{
		{
			name: "All fields provided",
			attr: AttrRemoveRatingProfile{
				Tenant:   "cgrates.org",
				Category: "category1",
				Subject:  "subject1",
			},
			expectedID: "*out:cgrates.org:category1:subject1",
		},
		{
			name: "Empty Tenant and Category",
			attr: AttrRemoveRatingProfile{
				Tenant:   utils.EmptyString,
				Category: utils.EmptyString,
				Subject:  "subject1",
			},
			expectedID: "*out:",
		},
		{
			name: "Tenant and Category are MetaAny",
			attr: AttrRemoveRatingProfile{
				Tenant:   utils.MetaAny,
				Category: utils.MetaAny,
				Subject:  "subject1",
			},
			expectedID: "*out:",
		},
		{
			name: "Only Subject provided",
			attr: AttrRemoveRatingProfile{
				Tenant:   utils.EmptyString,
				Category: utils.EmptyString,
				Subject:  "subject1",
			},
			expectedID: "*out:",
		},
		{
			name: "No fields provided",
			attr: AttrRemoveRatingProfile{
				Tenant:   utils.EmptyString,
				Category: utils.EmptyString,
				Subject:  utils.EmptyString,
			},
			expectedID: "*out:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.attr.GetId()
			if result != tt.expectedID {
				t.Errorf("expected %v, but got %v", tt.expectedID, result)
			}
		})
	}
}

func TestNewSMGenericV1(t *testing.T) {
	Session := &sessions.SessionS{}
	result := NewSMGenericV1(Session)

	if result.Ss != Session {
		t.Error("Expected the SessionS to be the same as the input, but got a different value")
	}

	if result == nil {
		t.Error("Expected result to be a valid SMGenericV1 instance, but got nil")
	}
}

func TestNewErSv1(t *testing.T) {
	erService := &ers.ERService{}
	erSv1 := NewErSv1(erService)

	if erSv1 == nil {
		t.Fatalf("Expected non-nil ErSv1, got nil")
	}

	if erSv1.erS != erService {
		t.Fatalf("Expected erS to be %v, got %v", erService, erSv1.erS)
	}
}
