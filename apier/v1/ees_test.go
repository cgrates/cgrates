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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/utils"
)

func TestNewEeSv1(t *testing.T) {
	eeS := &ees.EventExporterS{}
	eeSv1 := NewEeSv1(eeS)
	if eeSv1 == nil {
		t.Fatalf("Expected non-nil EeSv1, got nil")
	}
	if eeSv1.eeS != eeS {
		t.Errorf("Expected eeS field to be set correctly")
	}
}

func TestEeSv1Ping(t *testing.T) {
	eeSv1 := &EeSv1{}
	ctx := context.Background()
	event := &utils.CGREvent{}
	var reply string
	err := eeSv1.Ping(ctx, event, &reply)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if reply != utils.Pong {
		t.Errorf("Expected reply to be %s, got %s", utils.Pong, reply)
	}
}

func TestErSv1NewErSv1AndPing(t *testing.T) {
	mockErS := &ers.ERService{}
	erSv1 := NewErSv1(mockErS)
	if erSv1 == nil {
		t.Fatalf("Expected non-nil ErSv1, got nil")
	}
	if erSv1.erS != mockErS {
		t.Errorf("Expected erS field to be set correctly")
	}
	ctx := context.Background()
	var reply string
	err := erSv1.Ping(ctx, nil, &reply)
	if err != nil {
		t.Fatalf("Expected no error from Ping, got %v", err)
	}
	if reply != utils.Pong {
		t.Errorf("Expected reply to be %s, got %s", utils.Pong, reply)
	}
}
