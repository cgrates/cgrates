// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	idb, _ := NewInternalDB(nil, nil, true, nil, config.CgrConfig().DataDbCfg().Items) // make TransCacheOpts nil to not create dump files for the DB
	dm = NewDataManager(idb, config.CgrConfig().CacheCfg(), connMgr)
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
