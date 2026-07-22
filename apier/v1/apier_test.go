// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"testing"

	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

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
