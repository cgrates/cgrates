/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package config

import (
	"github.com/cgrates/cgrates/utils"
	"testing"
)

func TestAppendDerivedChargers(t *testing.T) {
	var err error
	dcs := make(DerivedChargers, 0)
	if _, err := dcs.Append(&DerivedCharger{RunId: utils.DEFAULT_RUNID}); err == nil {
		t.Error("Failed to detect using of the default runid")
	}
	if dcs, err = dcs.Append(&DerivedCharger{RunId: "FIRST_RUNID"}); err != nil {
		t.Error("Failed to add runid")
	} else if len(dcs) != 1 {
		t.Error("Unexpected number of items inside DerivedChargers configuration", len(dcs))
	}
	if dcs, err = dcs.Append(&DerivedCharger{RunId: "SECOND_RUNID"}); err != nil {
		t.Error("Failed to add runid")
	} else if len(dcs) != 2 {
		t.Error("Unexpected number of items inside DerivedChargers configuration", len(dcs))
	}
	if _, err := dcs.Append(&DerivedCharger{RunId: "SECOND_RUNID"}); err == nil {
		t.Error("Failed to detect duplicate runid")
	}
}
