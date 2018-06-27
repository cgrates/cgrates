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

func newHTTPUrlDP(req *http.Request) (dP engine.DataProvider, err error) {
	dP = &httpUrlDP{req: req, cache: engine.NewNavigableMap(nil)}
	return
}

// httpUrlDP implements engine.DataProvider, serving as url data decoder
// decoded data is only searched once and cached
type httpUrlDP struct {
	req   *http.Request
	cache *engine.NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (hU *httpUrlDP) String() string {
	return utils.ToJSON(hU.cache.AsMapStringInterface())
}

// FieldAsInterface is part of engine.DataProvider interface
func (hU *httpUrlDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	if data, err = hU.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	data = hU.req.FormValue(fldPath[0])
	hU.cache.Set(fldPath, data, false)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (hU *httpUrlDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = hU.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	data, _ = utils.CastFieldIfToString(valIface)
	return
}

// AsNavigableMap is part of engine.DataProvider interface
func (hU *httpUrlDP) AsNavigableMap([]*config.CfgCdrField) (
	nm *engine.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
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
