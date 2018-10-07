// +build integration

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
	"encoding/json"
	"flag"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Arguments received via test command
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")

// Sample HttpJsonPost, more for usage purposes
func TestHttpJsonPost(t *testing.T) {

	cdrOut := &ExternalCDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1",
		Source:     utils.UNIT_TEST, RequestType: utils.META_RATED, Tenant: "cgrates.org",
		Category: "call", Account: "account1", Subject: "tgooiscs0014", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String(), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String(),
		RunID: utils.DEFAULT_RUNID,
		Usage: "0.00000001", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	jsn, _ := json.Marshal(cdrOut)
	if _, err := HttpJsonPost("http://localhost:8000", false, jsn); err == nil {
		t.Error(err)
	}
}
