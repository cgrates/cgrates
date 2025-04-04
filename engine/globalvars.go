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
	"net/http"

	"github.com/cgrates/cgrates/config"
)

// this file will contain all the global variable that are used by other subsystems

var (
	httpPstrTransport *http.Transport
	dm                *DataManager
	cdrStorage        CdrStorage
	connMgr           *ConnManager
)

func init() {
	dm = NewDataManager(NewInternalDB(nil, nil, true, false, config.CgrConfig().DataDbCfg().Items), config.CgrConfig().CacheCfg(), connMgr)
	httpPstrTransport = config.CgrConfig().HTTPCfg().ClientOpts
}

// SetDataStorage is the exported method to set the storage getter.
func SetDataStorage(dm2 *DataManager) {
	dm = dm2
}

// SetConnManager is the exported method to set the connectionManager used when operate on an account.
func SetConnManager(conMgr *ConnManager) {
	connMgr = conMgr
}

// SetCdrStorage sets the database for CDR storing, used by *cdrlog in first place
func SetCdrStorage(cStorage CdrStorage) {
	cdrStorage = cStorage
}

// SetHTTPPstrTransport sets the http transport to be used by the HTTP Poster
func SetHTTPPstrTransport(pstrTransport *http.Transport) {
	httpPstrTransport = pstrTransport
}

// GetHTTPPstrTransport gets the http transport to be used by the HTTP Poster
func GetHTTPPstrTransport() *http.Transport {
	return httpPstrTransport
}
