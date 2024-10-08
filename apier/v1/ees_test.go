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

	"github.com/cgrates/cgrates/ees"
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
