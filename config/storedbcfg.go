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

package config

import (
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// StorDbCfg StroreDb config
type StorDbCfg struct {
	StorDBType            string // Should reflect the database type used to store logs
	StorDBHost            string // The host to connect to. Values that start with / are for UNIX domain sockets.
	StorDBPort            string // Th e port to bind to.
	StorDBName            string // The name of the database to connect to.
	StorDBUser            string // The user to sign in as.
	StorDBPass            string // The user's password.
	StorDBMaxOpenConns    int    // Maximum database connections opened
	StorDBMaxIdleConns    int    // Maximum idle connections to keep opened
	StorDBConnMaxLifetime int
	StorDBCDRSIndexes     []string
}

//loadFromJsonCfg loads StoreDb config from JsonCfg
func (dbcfg *StorDbCfg) loadFromJsonCfg(jsnDbCfg *DbJsonCfg) (err error) {
	if jsnDbCfg == nil {
		return nil
	}
	if jsnDbCfg.Db_type != nil {
		dbcfg.StorDBType = strings.TrimPrefix(*jsnDbCfg.Db_type, "*")
	}
	if jsnDbCfg.Db_host != nil {
		dbcfg.StorDBHost = *jsnDbCfg.Db_host
	}
	if jsnDbCfg.Db_port != nil {
		port := strconv.Itoa(*jsnDbCfg.Db_port)
		if port == "-1" {
			port = utils.MetaDynamic
		}
		dbcfg.StorDBPort = NewDbDefaults().DBPort(dbcfg.StorDBType, port)
	}
	if jsnDbCfg.Db_name != nil {
		dbcfg.StorDBName = *jsnDbCfg.Db_name
	}
	if jsnDbCfg.Db_user != nil {
		dbcfg.StorDBUser = *jsnDbCfg.Db_user
	}
	if jsnDbCfg.Db_password != nil {
		dbcfg.StorDBPass = *jsnDbCfg.Db_password
	}
	if jsnDbCfg.Max_open_conns != nil {
		dbcfg.StorDBMaxOpenConns = *jsnDbCfg.Max_open_conns
	}
	if jsnDbCfg.Max_idle_conns != nil {
		dbcfg.StorDBMaxIdleConns = *jsnDbCfg.Max_idle_conns
	}
	if jsnDbCfg.Conn_max_lifetime != nil {
		dbcfg.StorDBConnMaxLifetime = *jsnDbCfg.Conn_max_lifetime
	}
	if jsnDbCfg.Cdrs_indexes != nil {
		dbcfg.StorDBCDRSIndexes = *jsnDbCfg.Cdrs_indexes
	}
	return nil
}
