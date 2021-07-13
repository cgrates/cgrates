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
	"fmt"
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
	RmtConns           []string // Remote DataDB  connIDs
	RplConns           []string // Replication connIDs
	RplFiltered        bool
	Items              map[string]*ItemOpt
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
		dbcfg.DataDbPort = dbDefaultsCfg.dbPort(dbcfg.DataDbType, port)
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
	if jsnDbCfg.Replication_filtered != nil {
		dbcfg.RplFiltered = *jsnDbCfg.Replication_filtered
	}
	if jsnDbCfg.Remote_conns != nil {
		dbcfg.RmtConns = make([]string, len(*jsnDbCfg.Remote_conns))
		for idx, rmtConn := range *jsnDbCfg.Remote_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if rmtConn == utils.MetaInternal {
				return fmt.Errorf("Remote connection ID needs to be different than *internal")
			} else {
				dbcfg.RmtConns[idx] = rmtConn
			}
		}
	}
	if jsnDbCfg.Replication_conns != nil {
		dbcfg.RplConns = make([]string, len(*jsnDbCfg.Replication_conns))
		for idx, rplConn := range *jsnDbCfg.Replication_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if rplConn == utils.MetaInternal {
				return fmt.Errorf("Replication connection ID needs to be different than *internal")
			}
			dbcfg.RplConns[idx] = rplConn
		}
	}
	if jsnDbCfg.Items != nil {
		for kJsn, vJsn := range *jsnDbCfg.Items {
			val, has := dbcfg.Items[kJsn]
			if val == nil || !has {
				val = new(ItemOpt)
			}
			if err := val.loadFromJsonCfg(vJsn); err != nil {
				return err
			}
			dbcfg.Items[kJsn] = val
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
		Items:              dbcfg.Items,
		RplFiltered:        dbcfg.RplFiltered,
	}
}

func (dbcfg *DataDbCfg) AsMapInterface() map[string]interface{} {
	items := make(map[string]interface{})
	for key, item := range dbcfg.Items {
		items[key] = item.AsMapInterface()
	}
	var queryTimeout string = "0"
	if dbcfg.QueryTimeout != 0 {
		queryTimeout = dbcfg.QueryTimeout.String()
	}
	dbPort, _ := strconv.Atoi(dbcfg.DataDbPort)

	return map[string]interface{}{
		utils.DataDbTypeCfg:          utils.Meta + dbcfg.DataDbType,
		utils.DataDbHostCfg:          dbcfg.DataDbHost,
		utils.DataDbPortCfg:          dbPort,
		utils.DataDbNameCfg:          dbcfg.DataDbName,
		utils.DataDbUserCfg:          dbcfg.DataDbUser,
		utils.DataDbPassCfg:          dbcfg.DataDbPass,
		utils.DataDbSentinelNameCfg:  dbcfg.DataDbSentinelName,
		utils.QueryTimeoutCfg:        queryTimeout,
		utils.RmtConnsCfg:            dbcfg.RmtConns,
		utils.RplConnsCfg:            dbcfg.RplConns,
		utils.ItemsCfg:               items,
		utils.ReplicationFilteredCfg: dbcfg.RplFiltered,
	}
}

type ItemOpt struct {
	Remote    bool
	Replicate bool
	TTL       time.Duration
	Limit     int
	StaticTTL bool
	// used for ArgDispatcher in case we send this to a dispatcher engine
	RouteID string
	APIKey  string
}

func (itm *ItemOpt) AsMapInterface() map[string]interface{} {
	var ttl string = ""
	if itm.TTL != 0 {
		ttl = itm.TTL.String()
	}

	return map[string]interface{}{
		utils.RemoteCfg:    itm.Remote,
		utils.ReplicateCfg: itm.Replicate,
		utils.LimitCfg:     itm.Limit,
		utils.TTLCfg:       ttl,
		utils.StaticTTLCfg: itm.StaticTTL,
	}
}

func (itm *ItemOpt) loadFromJsonCfg(jsonItm *ItemOptJson) (err error) {
	if jsonItm == nil {
		return
	}
	if jsonItm.Remote != nil {
		itm.Remote = *jsonItm.Remote
	}
	if jsonItm.Replicate != nil {
		itm.Replicate = *jsonItm.Replicate
	}
	if jsonItm.Ttl != nil {
		if itm.TTL, err = utils.ParseDurationWithNanosecs(*jsonItm.Ttl); err != nil {
			return err
		}
	}
	if jsonItm.Limit != nil {
		itm.Limit = *jsonItm.Limit
	}
	if jsonItm.Static_ttl != nil {
		itm.StaticTTL = *jsonItm.Static_ttl
	}
	if jsonItm.Route_id != nil {
		itm.RouteID = *jsonItm.Route_id
	}
	if jsonItm.Api_key != nil {
		itm.APIKey = *jsonItm.Api_key
	}
	return
}
