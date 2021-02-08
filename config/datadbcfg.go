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

	"github.com/cgrates/cgrates/utils"
)

// DataDbCfg Database config
type DataDbCfg struct {
	DataDbType string
	DataDbHost string   // The host to connect to. Values that start with / are for UNIX domain sockets.
	DataDbPort string   // The port to bind to.
	DataDbName string   // The name of the database to connect to.
	DataDbUser string   // The user to sign in as.
	DataDbPass string   // The user's password.
	RmtConns   []string // Remote DataDB  connIDs
	RplConns   []string // Replication connIDs
	Items      map[string]*ItemOpt
	Opts       map[string]interface{}
}

// loadFromJSONCfg loads Database config from JsonCfg
func (dbcfg *DataDbCfg) loadFromJSONCfg(jsnDbCfg *DbJsonCfg) (err error) {
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
	if jsnDbCfg.Remote_conns != nil {
		dbcfg.RmtConns = make([]string, len(*jsnDbCfg.Remote_conns))
		for idx, rmtConn := range *jsnDbCfg.Remote_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if rmtConn == utils.MetaInternal {
				return fmt.Errorf("Remote connection ID needs to be different than *internal")
			}
			dbcfg.RmtConns[idx] = rmtConn
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
			val.loadFromJSONCfg(vJsn) //To review if the function signature changes
			dbcfg.Items[kJsn] = val
		}
	}
	if jsnDbCfg.Opts != nil {
		for k, v := range jsnDbCfg.Opts {
			dbcfg.Opts[k] = v
		}
	}
	return nil
}

// Clone returns the cloned object
func (dbcfg *DataDbCfg) Clone() (cln *DataDbCfg) {
	cln = &DataDbCfg{
		DataDbType: dbcfg.DataDbType,
		DataDbHost: dbcfg.DataDbHost,
		DataDbPort: dbcfg.DataDbPort,
		DataDbName: dbcfg.DataDbName,
		DataDbUser: dbcfg.DataDbUser,
		DataDbPass: dbcfg.DataDbPass,
		Items:      make(map[string]*ItemOpt),
		Opts:       make(map[string]interface{}),
	}
	for k, itm := range dbcfg.Items {
		cln.Items[k] = itm.Clone()
	}
	for k, v := range dbcfg.Opts {
		cln.Opts[k] = v
	}

	if dbcfg.RmtConns != nil {
		cln.RmtConns = make([]string, len(dbcfg.RmtConns))
		for i, conn := range dbcfg.RmtConns {
			cln.RmtConns[i] = conn
		}
	}
	if dbcfg.RplConns != nil {
		cln.RplConns = make([]string, len(dbcfg.RplConns))
		for i, conn := range dbcfg.RplConns {
			cln.RplConns[i] = conn
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (dbcfg *DataDbCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.DataDbTypeCfg: utils.Meta + dbcfg.DataDbType,
		utils.DataDbHostCfg: dbcfg.DataDbHost,
		utils.DataDbNameCfg: dbcfg.DataDbName,
		utils.DataDbUserCfg: dbcfg.DataDbUser,
		utils.DataDbPassCfg: dbcfg.DataDbPass,
		utils.RmtConnsCfg:   dbcfg.RmtConns,
		utils.RplConnsCfg:   dbcfg.RplConns,
	}
	opts := make(map[string]interface{})
	for k, v := range dbcfg.Opts {
		opts[k] = v
	}
	initialMP[utils.OptsCfg] = opts
	if dbcfg.Items != nil {
		items := make(map[string]interface{})
		for key, item := range dbcfg.Items {
			items[key] = item.AsMapInterface()
		}
		initialMP[utils.ItemsCfg] = items
	}
	if dbcfg.DataDbPort != "" {
		initialMP[utils.DataDbPortCfg], _ = strconv.Atoi(dbcfg.DataDbPort)
	}
	return
}

// ItemOpt the options for the stored items
type ItemOpt struct {
	Remote    bool
	Replicate bool
	// used for ArgDispatcher in case we send this to a dispatcher engine
	RouteID string
	APIKey  string
}

// AsMapInterface returns the config as a map[string]interface{}
func (itm *ItemOpt) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.RemoteCfg:    itm.Remote,
		utils.ReplicateCfg: itm.Replicate,
	}
	if itm.APIKey != utils.EmptyString {
		initialMP[utils.APIKeyCfg] = itm.APIKey
	}
	if itm.RouteID != utils.EmptyString {
		initialMP[utils.RouteIDCfg] = itm.RouteID
	}
	return
}

func (itm *ItemOpt) loadFromJSONCfg(jsonItm *ItemOptJson) {
	if jsonItm == nil {
		return
	}
	if jsonItm.Remote != nil {
		itm.Remote = *jsonItm.Remote
	}
	if jsonItm.Replicate != nil {
		itm.Replicate = *jsonItm.Replicate
	}
	if jsonItm.Route_id != nil {
		itm.RouteID = *jsonItm.Route_id
	}
	if jsonItm.Api_key != nil {
		itm.APIKey = *jsonItm.Api_key
	}
	return
}

// Clone returns a deep copy of ItemOpt
func (itm *ItemOpt) Clone() *ItemOpt {
	return &ItemOpt{
		Remote:    itm.Remote,
		Replicate: itm.Replicate,
		APIKey:    itm.APIKey,
		RouteID:   itm.RouteID,
	}
}
