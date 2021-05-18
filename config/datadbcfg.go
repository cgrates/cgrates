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
	Type        string
	Host        string   // The host to connect to. Values that start with / are for UNIX domain sockets.
	Port        string   // The port to bind to.
	Name        string   // The name of the database to connect to.
	User        string   // The user to sign in as.
	Password    string   // The user's password.
	RmtConns    []string // Remote DataDB  connIDs
	RmtConnID   string
	RplConns    []string // Replication connIDs
	RplFiltered bool
	RplCache    string
	Items       map[string]*ItemOpt
	Opts        map[string]interface{}
}

// loadFromJSONCfg loads Database config from JsonCfg
func (dbcfg *DataDbCfg) loadFromJSONCfg(jsnDbCfg *DbJsonCfg) (err error) {
	if jsnDbCfg == nil {
		return nil
	}
	if jsnDbCfg.Db_type != nil {
		dbcfg.Type = strings.TrimPrefix(*jsnDbCfg.Db_type, "*")
	}
	if jsnDbCfg.Db_host != nil {
		dbcfg.Host = *jsnDbCfg.Db_host
	}
	if jsnDbCfg.Db_port != nil {
		port := strconv.Itoa(*jsnDbCfg.Db_port)
		if port == "-1" {
			port = utils.MetaDynamic
		}
		dbcfg.Port = dbDefaultsCfg.dbPort(dbcfg.Type, port)
	}
	if jsnDbCfg.Db_name != nil {
		dbcfg.Name = *jsnDbCfg.Db_name
	}
	if jsnDbCfg.Db_user != nil {
		dbcfg.User = *jsnDbCfg.Db_user
	}
	if jsnDbCfg.Db_password != nil {
		dbcfg.Password = *jsnDbCfg.Db_password
	}
	if jsnDbCfg.Remote_conns != nil {
		dbcfg.RmtConns = make([]string, len(*jsnDbCfg.Remote_conns))
		for idx, rmtConn := range *jsnDbCfg.Remote_conns {
			if rmtConn == utils.MetaInternal {
				return fmt.Errorf("Remote connection ID needs to be different than <%s> ", utils.MetaInternal)
			}
			dbcfg.RmtConns[idx] = rmtConn
		}
	}
	if jsnDbCfg.Replication_conns != nil {
		dbcfg.RplConns = make([]string, len(*jsnDbCfg.Replication_conns))
		for idx, rplConn := range *jsnDbCfg.Replication_conns {
			if rplConn == utils.MetaInternal {
				return fmt.Errorf("Remote connection ID needs to be different than <%s> ", utils.MetaInternal)
			}
			dbcfg.RplConns[idx] = rplConn
		}
	}
	if jsnDbCfg.Items != nil {
		for kJsn, vJsn := range jsnDbCfg.Items {
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
	if jsnDbCfg.Replication_filtered != nil {
		dbcfg.RplFiltered = *jsnDbCfg.Replication_filtered
	}
	if jsnDbCfg.Remote_conn_id != nil {
		dbcfg.RmtConnID = *jsnDbCfg.Remote_conn_id
	}
	if jsnDbCfg.Replication_cache != nil {
		dbcfg.RplCache = *jsnDbCfg.Replication_cache
	}
	return
}

// Clone returns the cloned object
func (dbcfg *DataDbCfg) Clone() (cln *DataDbCfg) {
	cln = &DataDbCfg{
		Type:        dbcfg.Type,
		Host:        dbcfg.Host,
		Port:        dbcfg.Port,
		Name:        dbcfg.Name,
		User:        dbcfg.User,
		Password:    dbcfg.Password,
		RplFiltered: dbcfg.RplFiltered,
		RplCache:    dbcfg.RplCache,
		RmtConnID:   dbcfg.RmtConnID,
		Items:       make(map[string]*ItemOpt),
		Opts:        make(map[string]interface{}),
	}
	for k, itm := range dbcfg.Items {
		cln.Items[k] = itm.Clone()
	}
	for k, v := range dbcfg.Opts {
		cln.Opts[k] = v
	}

	if dbcfg.RmtConns != nil {
		cln.RmtConns = utils.CloneStringSlice(dbcfg.RmtConns)
	}
	if dbcfg.RplConns != nil {
		cln.RplConns = utils.CloneStringSlice(dbcfg.RplConns)
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (dbcfg *DataDbCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.DataDbTypeCfg:          utils.Meta + dbcfg.Type,
		utils.DataDbHostCfg:          dbcfg.Host,
		utils.DataDbNameCfg:          dbcfg.Name,
		utils.DataDbUserCfg:          dbcfg.User,
		utils.DataDbPassCfg:          dbcfg.Password,
		utils.RemoteConnsCfg:         dbcfg.RmtConns,
		utils.RemoteConnIDCfg:        dbcfg.RmtConnID,
		utils.ReplicationConnsCfg:    dbcfg.RplConns,
		utils.ReplicationFilteredCfg: dbcfg.RplFiltered,
		utils.ReplicationCache:       dbcfg.RplCache,
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
	if dbcfg.Port != "" {
		initialMP[utils.DataDbPortCfg], _ = strconv.Atoi(dbcfg.Port)
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

func (itm *ItemOpt) Equals(itm2 *ItemOpt) bool {
	return (itm == nil && itm2 == nil) ||
		(itm != nil && itm2 != nil &&
			itm.Remote == itm2.Remote &&
			itm.Replicate == itm2.Replicate &&
			itm.RouteID == itm2.RouteID &&
			itm.APIKey == itm2.APIKey)
}

type ItemOptJson struct {
	Remote    *bool
	Replicate *bool
	// used for ArgDispatcher in case we send this to a dispatcher engine
	Route_id *string
	Api_key  *string
}

func diffItemOptJson(d *ItemOptJson, v1, v2 *ItemOpt) *ItemOptJson {
	if d == nil {
		d = new(ItemOptJson)
	}
	if v2.Remote != v1.Remote {
		d.Remote = utils.BoolPointer(v2.Remote)
	}
	if v2.Replicate != v1.Replicate {
		d.Replicate = utils.BoolPointer(v2.Replicate)
	}
	if v2.RouteID != v1.RouteID {
		d.Route_id = utils.StringPointer(v2.RouteID)
	}
	if v2.APIKey != v1.APIKey {
		d.Api_key = utils.StringPointer(v2.APIKey)
	}
	return d
}

func diffMapItemOptJson(d map[string]*ItemOptJson, v1, v2 map[string]*ItemOpt) map[string]*ItemOptJson {
	if d == nil {
		d = make(map[string]*ItemOptJson)
	}
	for k, val2 := range v2 {
		if val1, has := v1[k]; !has {
			d[k] = diffItemOptJson(d[k], new(ItemOpt), val2)
		} else if !val1.Equals(val2) {
			d[k] = diffItemOptJson(d[k], val1, val2)
		}
	}
	return d
}

// Database config
type DbJsonCfg struct {
	Db_type               *string
	Db_host               *string
	Db_port               *int
	Db_name               *string
	Db_user               *string
	Db_password           *string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Remote_conns          *[]string
	Remote_conn_id        *string
	Replication_conns     *[]string
	Replication_filtered  *bool
	Replication_cache     *string
	Items                 map[string]*ItemOptJson
	Opts                  map[string]interface{}
}

func diffDataDbJsonCfg(d *DbJsonCfg, v1, v2 *DataDbCfg) *DbJsonCfg {
	if d == nil {
		d = new(DbJsonCfg)
	}
	if v1.Type != v2.Type {
		d.Db_type = utils.StringPointer(v2.Type)
	}
	if v1.Host != v2.Host {
		d.Db_host = utils.StringPointer(v2.Host)
	}
	if v1.Port != v2.Port {
		port, _ := strconv.Atoi(v2.Port)
		d.Db_port = utils.IntPointer(port)
	}
	if v1.Name != v2.Name {
		d.Db_name = utils.StringPointer(v2.Name)
	}
	if v1.User != v2.User {
		d.Db_user = utils.StringPointer(v2.User)
	}
	if v1.Password != v2.Password {
		d.Db_password = utils.StringPointer(v2.Password)
	}
	if !utils.SliceStringEqual(v1.RmtConns, v2.RmtConns) {
		d.Remote_conns = &v2.RmtConns
	}
	if v1.RmtConnID != v2.RmtConnID {
		d.Remote_conn_id = utils.StringPointer(v2.RmtConnID)
	}
	if !utils.SliceStringEqual(v1.RplConns, v2.RplConns) {
		d.Replication_conns = &v2.RplConns
	}
	if v1.RplFiltered != v2.RplFiltered {
		d.Replication_filtered = utils.BoolPointer(v2.RplFiltered)
	}
	if v1.RplCache != v2.RplCache {
		d.Replication_cache = utils.StringPointer(v2.RplCache)
	}
	d.Items = diffMapItemOptJson(d.Items, v1.Items, v2.Items)
	d.Opts = diffMap(d.Opts, v1.Opts, v2.Opts)
	return d
}
