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
	"fmt"
	"net/http"

	"github.com/cgrates/cgrates/engine"
)

// httpReplyField is one field written in HTTP reply
type httpReplyField struct {
	fldPath string
	fldVal  string
}

func newHTTPReplyFields() *httpReplyFields {
	return &httpReplyFields{indexed: make(map[string]*httpReplyField),
		ordered: make([]*httpReplyField, 0)}
}

// httpReplyFields is the reply which will be written to HTTP
// both flds and ordered are pointig towards same httpReplyField
type httpReplyFields struct {
	indexed map[string]*httpReplyField // map[fldPath]*httpReplyField
	ordered []*httpReplyField          // keep order for export
}

// newHAReqDecoder produces decoders
func newHADataProvider(dpType string,
	req *http.Request) (dP engine.DataProvider, err error) {
	switch dpType {
	default:
		return nil, fmt.Errorf("unsupported decoder type <%s>", dpType)
	}
}

// newHAReplyEncoder constructs a httpAgentReqDecoder based on encoder type
func newHAReplyEncoder(encType string,
	w http.ResponseWriter) (rE httpAgentReplyEncoder, err error) {
	switch encType {
	default:
		return nil, fmt.Errorf("unsupported encoder type <%s>", encType)
	}
}

// httpAgentReplyEncoder will encode fields from httpReplyFields
// and write content to http writer
type httpAgentReplyEncoder interface {
	encode(*httpReplyFields) error
}
