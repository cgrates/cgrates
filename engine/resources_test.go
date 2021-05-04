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

package engine

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestResourcesRemoveExpiredUnitsResetTotalUsage(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	r := &Resource{
		TTLIdx: []string{"ResGroup1", "ResGroup2", "ResGroup3"},
		Usages: map[string]*ResourceUsage{
			"ResGroup2": {
				Tenant:     "cgrates.org",
				ID:         "RU_2",
				Units:      11,
				ExpiryTime: time.Date(2021, 5, 3, 13, 0, 0, 0, time.UTC),
			},
			"ResGroup3": {
				Tenant: "cgrates.org",
				ID:     "RU_3",
			},
		},
		tUsage: utils.Float64Pointer(10),
	}

	exp := &Resource{
		TTLIdx: []string{"ResGroup3"},
		Usages: map[string]*ResourceUsage{
			"ResGroup3": {
				Tenant: "cgrates.org",
				ID:     "RU_3",
			},
		},
	}

	explog := "CGRateS <> [WARNING] resetting total usage for resourceID: , usage smaller than 0: -1.000000\n"
	r.removeExpiredUnits()

	if !reflect.DeepEqual(r, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, r)
	}

	rcvlog := buf.String()[20:]
	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}

	utils.Logger.SetLogLevel(0)
}
