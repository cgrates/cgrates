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
	"encoding/json"
	"net/http"
	"net/http/httputil"

	"github.com/cgrates/cgrates/utils"
)

// janusAccessControlHeaders will add the necessary access control headers
func janusAccessControlHeaders(w http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Accept-Language, Content-Type")
	}
}

// newJanusHTTPjsonDP is the constructor for janusHTTPjsonDP struct
func newJanusHTTPjsonDP(req *http.Request) (utils.DataProvider, error) {
	jHj := &janusHTTPjsonDP{req: req, cache: utils.MapStorage{}}
	if err := json.NewDecoder(req.Body).Decode(&jHj.reqBody); err != nil {
		return nil, err
	}
	return jHj, nil
}

// janusHTTPjsonDP implements utils.DataProvider, serving as JSON data decoder
// decoded data is only searched once and cached
type janusHTTPjsonDP struct {
	req     *http.Request
	reqBody map[string]interface{} // unmarshal JSON body here
	cache   utils.MapStorage
}

// String is part of utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (jHj *janusHTTPjsonDP) String() string {
	byts, _ := httputil.DumpRequest(jHj.req, true)
	return string(byts)
}

// FieldAsInterface is part of utils.DataProvider interface
func (jHj *janusHTTPjsonDP) FieldAsInterface(fldPath []string) (data any, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	if data, err = jHj.cache.FieldAsInterface(fldPath); err != nil {
		if err != utils.ErrNotFound { // item found in cache
			return
		}
		err = nil // cancel previous err
	} else {
		return // data found in cache
	}
	data = jHj.reqBody[fldPath[0]]
	jHj.cache.Set(fldPath, data)
	return
}

// FieldAsString is part of utils.DataProvider interface
func (jHj *janusHTTPjsonDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface any
	valIface, err = jHj.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}
