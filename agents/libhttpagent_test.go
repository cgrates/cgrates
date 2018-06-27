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

package agents

import (
	"bufio"
	"net/http"
	"strings"
	"testing"
	//"github.com/cgrates/cgrates/utils"
)

func TestHttpUrlDPFieldAsInterface(t *testing.T) {
	br := bufio.NewReader(strings.NewReader(`GET /cdr?request_type=MOSMS_CDR&timestamp=2008-08-15%2017:49:21&message_date=2008-08-15%2017:49:21&transactionid=100744&CDR_ID=123456&carrierid=1&mcc=222&mnc=10&imsi=235180000000000&msisdn=%2B4977000000000&destination=%2B497700000001&message_status=0&IOT=0&service_id=1 HTTP/1.1
Host: api.cgrates.org

`))
	req, err := http.ReadRequest(br)
	if err != nil {
		t.Error(err)
	}
	hU, _ := newHTTPUrlDP(req)
	if data, err := hU.FieldAsString([]string{"request_type"}); err != nil {
		t.Error(err)
	} else if data != "MOSMS_CDR" {
		t.Errorf("expecting: MOSMS_CDR, received: <%s>", data)
	}
	if data, err := hU.FieldAsString([]string{"nonexistent"}); err != nil {
		t.Error(err)
	} else if data != "" {
		t.Errorf("received: <%s>", data)
	}
}
