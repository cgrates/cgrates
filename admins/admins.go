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

package admins

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

func NewAdminS(cfg *config.CGRConfig, dm *engine.DataManager, connMgr *engine.ConnManager, fltrS *engine.FilterS,
	storDBChan chan engine.StorDB) *AdminS {
	storDB := <-storDBChan
	return &AdminS{
		cfg:     cfg,
		dm:      dm,
		storDB:  storDB,
		connMgr: connMgr,
		fltrS:   fltrS,

		// TODO: Might be a good idea to pass the storDB channel to AdminSv1
		// to be able to close the service the moment storDB is down (inside
		//  a ListenAndServe goroutine maybe)

		// storDBChan: storDBChan,
	}
}

type AdminS struct {
	cfg     *config.CGRConfig
	dm      *engine.DataManager
	storDB  engine.StorDB
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
}
