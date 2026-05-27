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

package agents

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/sessions"
)

// buildDlgListResponse builds a CGR_DLG_LIST reply with n realistic dialogs.
func buildDlgListResponse(n int) []byte {
	var buf bytes.Buffer
	buf.WriteString(`{"Event":"CGR_DLG_LIST","Jsonrpl_body":{"jsonrpc":"2.0","id":1,"result":[`)
	for i := range n {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(&buf,
			`{"h_entry":%d,"h_id":%d,"ref_count":1,"timestart":%d,"timeout":%d,`+
				`"state":4,"lifetime":3600,"init_ts":%d,"dflags":64,"iflags":0,`+
				`"sflags":0,"toroute_index":0,"toroute_name":"",`+
				`"call-id":"call-bench-%d-1234567890abcdef@bench.local",`+
				`"from_uri":"sip:caller-bench-%d@bench.local",`+
				`"to_uri":"sip:callee-bench-%d@bench.local",`+
				`"caller":{"tag":"ftag-bench-%d-abcdef0123","contact":"<sip:caller-bench-%d@10.0.0.1:5060>","cseq":"1","route_set":"","socket":"udp:10.0.0.1:5060","bind_addr":"udp:10.0.0.1:5060","sdp":""},`+
				`"callee":{"tag":"ttag-bench-%d-abcdef0123","contact":"<sip:callee-bench-%d@10.0.0.2:5060>","cseq":"1","route_set":"","socket":"udp:10.0.0.2:5060","bind_addr":"udp:10.0.0.2:5060","sdp":""},`+
				`"profiles":[],"variables":[]}`,
			i%4096, i, 1715000000+i, 1715003600+i, 1715000000+i,
			i, i, i, i, i, i, i,
		)
	}
	buf.WriteString(`]}}`)
	return buf.Bytes()
}

// BenchmarkV1GetActiveSessionIDs measures parsing a dlg.list reply and collecting its
// session IDs. The reply is built once outside the loop so its cost isn't measured.
func BenchmarkV1GetActiveSessionIDs(b *testing.B) {
	sizes := []int{1000, 10000, 36000}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("%dDialogs", n), func(b *testing.B) {
			response := buildDlgListResponse(n)
			addr := startMockKamailio(b, response, 0)

			ka := dialMockKamailio(b, addr, 10*time.Second)

			b.SetBytes(int64(len(response)))
			b.ReportAllocs()
			for b.Loop() {
				var sIDs []*sessions.SessionID
				if err := ka.V1GetActiveSessionIDs(context.Background(), "", &sIDs); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
