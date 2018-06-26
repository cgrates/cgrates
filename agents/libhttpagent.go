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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// newHADataProvider constructs a DataProvider
func newHADataProvider(dpType string,
	req *http.Request) (dP engine.DataProvider, err error) {
	switch dpType {
	default:
		return nil, fmt.Errorf("unsupported decoder type <%s>", dpType)
	case utils.MetaUrl:
		return newHTTPUrlDP(req)
	}
}

// httpAgentReplyEncoder will encode  []*engine.NMElement
// and write content to http writer
type httpAgentReplyEncoder interface {
	encode(*engine.NavigableMap) error
}

// newHAReplyEncoder constructs a httpAgentReqDecoder based on encoder type
func newHAReplyEncoder(encType string,
	w http.ResponseWriter) (rE httpAgentReplyEncoder, err error) {
	switch encType {
	default:
		return nil, fmt.Errorf("unsupported encoder type <%s>", encType)
	}
}

func newHTTPUrlDP(req *http.Request) (dP engine.DataProvider, err error) {
	dP = &httpUrlDP{req: req}
	return
}

// httpUrlDP implements engine.DataProvider, serving as url data decoder
// decoded data is only searched once and cached
type httpUrlDP struct {
	req   *http.Request
	cache engine.NavigableMap
}

// String is part of engine.DataProvider interface
func (url *httpUrlDP) String() string {
	return utils.ToJSON(url.cache.AsMapStringInterface())
}

// FieldAsInterface is part of engine.DataProvider interface
func (url *httpUrlDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	return
}

// FieldAsString is part of engine.DataProvider interface
func (url *httpUrlDP) FieldAsString(fldPath []string) (data string, err error) {
	return
}

// AsNavigableMap is part of engine.DataProvider interface
func (url *httpUrlDP) AsNavigableMap([]*config.CfgCdrField) (nm *engine.NavigableMap, err error) {
	return
}
