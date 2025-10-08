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

package engine

import (
	"net/http"

	"github.com/cgrates/cgrates/config"
)

// this file will contain all the global variable that are used by other subsystems

var (
	connMgr           *ConnManager
	httpPstrTransport = config.CgrConfig().HTTPCfg().ClientOpts
)

// SetConnManager is the exported method to set the connectionManager used when operate on an account.
func SetConnManager(cm *ConnManager) {
	connMgr = cm
}

// SetHTTPPstrTransport sets the http transport to be used by the HTTP Poster
func SetHTTPPstrTransport(pstrTransport *http.Transport) {
	httpPstrTransport = pstrTransport
}

// HTTPPstrTransport gets the http transport to be used by the HTTP Poster
func HTTPPstrTransport() *http.Transport {
	return httpPstrTransport
}
