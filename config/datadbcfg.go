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
	"time"

	"github.com/cgrates/cgrates/utils"
)

// DataDbCfg Database config
type DataDbCfg struct {
	DataDbType         string
	DataDbHost         string // The host to connect to. Values that start with / are for UNIX domain sockets.
	DataDbPort         string // The port to bind to.
	DataDbName         string // The name of the database to connect to.
	DataDbUser         string // The user to sign in as.
	DataDbPass         string // The user's password.
	DataDbSentinelName string
	QueryTimeout       time.Duration
	RmtDataDBCfgs      []*DataDbCfg  // Remote DataDB  configurations
	RplConns           []*RemoteHost // Replication conns
}

//loadFromJsonCfg loads Database config from JsonCfg
func (dbcfg *DataDbCfg) loadFromJsonCfg(jsnDbCfg *DbJsonCfg) (err error) {
	if jsnDbCfg == nil {
		return nil
	}
	if jsnDbCfg.Db_type != nil {
		dbcfg.DataDbType = strings.TrimPrefix(*jsnDbCfg.Db_type, "*")
	}
	if jsnDbCfg.Db_host != nil {
		dbcfg.DataDbHost = *jsnDbCfg.Db_host
	}
	if jsnDbCfg.Db_port != nil {
		port := strconv.Itoa(*jsnDbCfg.Db_port)
		if port == "-1" {
			port = utils.MetaDynamic
		}
		dbcfg.DataDbPort = NewDbDefaults().DBPort(dbcfg.DataDbType, port)
	}
	if jsnDbCfg.Db_name != nil {
		dbcfg.DataDbName = *jsnDbCfg.Db_name
	}
	if jsnDbCfg.Db_user != nil {
		dbcfg.DataDbUser = *jsnDbCfg.Db_user
	}
	if jsnDbCfg.Db_password != nil {
		dbcfg.DataDbPass = *jsnDbCfg.Db_password
	}
	if jsnDbCfg.Redis_sentinel != nil {
		dbcfg.DataDbSentinelName = *jsnDbCfg.Redis_sentinel
	}
	if jsnDbCfg.Query_timeout != nil {
		if dbcfg.QueryTimeout, err = utils.ParseDurationWithNanosecs(*jsnDbCfg.Query_timeout); err != nil {
			return err
		}
	}
	if jsnDbCfg.Remote_conns != nil {
		dbcfg.RmtDataDBCfgs = make([]*DataDbCfg, len(*jsnDbCfg.Remote_conns))
		for i, cfg := range *jsnDbCfg.Remote_conns {
			dbcfg.RmtDataDBCfgs[i] = newDefaultDataDbCfg()
			if err = dbcfg.RmtDataDBCfgs[i].loadFromJsonCfg(cfg); err != nil {
				return
			}
		}
	}
	if jsnDbCfg.Replication_conns != nil {
		dbcfg.RplConns = make([]*RemoteHost, len(*jsnDbCfg.Replication_conns))
		for idx, jsnRplCfg := range *jsnDbCfg.Replication_conns {
			dbcfg.RplConns[idx] = NewDfltRemoteHost()
			dbcfg.RplConns[idx].loadFromJsonCfg(jsnRplCfg)
		}
	}
	return nil
}

// Clone returns the cloned object
func (dbcfg *DataDbCfg) Clone() *DataDbCfg {
	return &DataDbCfg{
		DataDbType:         dbcfg.DataDbType,
		DataDbHost:         dbcfg.DataDbHost,
		DataDbPort:         dbcfg.DataDbPort,
		DataDbName:         dbcfg.DataDbName,
		DataDbUser:         dbcfg.DataDbUser,
		DataDbPass:         dbcfg.DataDbPass,
		DataDbSentinelName: dbcfg.DataDbSentinelName,
		QueryTimeout:       dbcfg.QueryTimeout,
	}
}

func newDefaultDataDbCfg() *DataDbCfg {
	return &DataDbCfg{
		DataDbType:         utils.REDIS,
		DataDbHost:         "127.0.0.1",
		DataDbPort:         "6379",
		DataDbName:         "10",
		DataDbUser:         "cgrates",
		DataDbPass:         "",
		DataDbSentinelName: "",
		QueryTimeout:       10 * time.Second,
		RmtDataDBCfgs:      nil,
		RplConns:           nil,
	}
}
