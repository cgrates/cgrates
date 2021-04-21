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

// StorDbCfg StroreDb config
type StorDbCfg struct {
	Type                string // Should reflect the database type used to store logs
	Host                string // The host to connect to. Values that start with / are for UNIX domain sockets.
	Port                string // Th e port to bind to.
	Name                string // The name of the database to connect to.
	User                string // The user to sign in as.
	Password            string // The user's password.
	StringIndexedFields []string
	PrefixIndexedFields []string
	RmtConns            []string // Remote DataDB  connIDs
	RplConns            []string // Replication connIDs
	Items               map[string]*ItemOpt
	Opts                map[string]interface{}
}

// loadFromJSONCfg loads StoreDb config from JsonCfg
func (dbcfg *StorDbCfg) loadFromJSONCfg(jsnDbCfg *DbJsonCfg) (err error) {
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
	if jsnDbCfg.String_indexed_fields != nil {
		dbcfg.StringIndexedFields = *jsnDbCfg.String_indexed_fields
	}
	if jsnDbCfg.Prefix_indexed_fields != nil {
		dbcfg.PrefixIndexedFields = *jsnDbCfg.Prefix_indexed_fields
	}
	if jsnDbCfg.Remote_conns != nil {
		dbcfg.RmtConns = make([]string, len(*jsnDbCfg.Remote_conns))
		for i, item := range *jsnDbCfg.Remote_conns {
			if item == utils.MetaInternal {
				return fmt.Errorf("Remote connection ID needs to be different than *internal ")
			}
			dbcfg.RmtConns[i] = item
		}
	}
	if jsnDbCfg.Replication_conns != nil {
		dbcfg.RplConns = make([]string, len(*jsnDbCfg.Replication_conns))
		for i, item := range *jsnDbCfg.Replication_conns {
			if item == utils.MetaInternal {
				return fmt.Errorf("Replication connection ID needs to be different than *internal ")
			}
			dbcfg.RplConns[i] = item
		}
	}
	if jsnDbCfg.Opts != nil {
		for k, v := range jsnDbCfg.Opts {
			dbcfg.Opts[k] = v
		}
	}
	if jsnDbCfg.Items != nil {
		for kJsn, vJsn := range jsnDbCfg.Items {
			val := new(ItemOpt)
			val.loadFromJSONCfg(vJsn) //To review if the function signature changes
			dbcfg.Items[kJsn] = val
		}
	}
	return nil
}

// Clone returns the cloned object
func (dbcfg *StorDbCfg) Clone() (cln *StorDbCfg) {
	cln = &StorDbCfg{
		Type:     dbcfg.Type,
		Host:     dbcfg.Host,
		Port:     dbcfg.Port,
		Name:     dbcfg.Name,
		User:     dbcfg.User,
		Password: dbcfg.Password,

		Items: make(map[string]*ItemOpt),
		Opts:  make(map[string]interface{}),
	}
	for key, item := range dbcfg.Items {
		cln.Items[key] = item.Clone()
	}
	for key, val := range dbcfg.Opts {
		cln.Opts[key] = val
	}
	if dbcfg.StringIndexedFields != nil {
		cln.StringIndexedFields = make([]string, len(dbcfg.StringIndexedFields))
		for i, idx := range dbcfg.StringIndexedFields {
			cln.StringIndexedFields[i] = idx
		}
	}
	if dbcfg.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = make([]string, len(dbcfg.PrefixIndexedFields))
		for i, idx := range dbcfg.PrefixIndexedFields {
			cln.PrefixIndexedFields[i] = idx
		}
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
func (dbcfg *StorDbCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.DataDbTypeCfg:          utils.Meta + dbcfg.Type,
		utils.DataDbHostCfg:          dbcfg.Host,
		utils.DataDbNameCfg:          dbcfg.Name,
		utils.DataDbUserCfg:          dbcfg.User,
		utils.DataDbPassCfg:          dbcfg.Password,
		utils.StringIndexedFieldsCfg: dbcfg.StringIndexedFields,
		utils.PrefixIndexedFieldsCfg: dbcfg.PrefixIndexedFields,
		utils.RemoteConnsCfg:         dbcfg.RmtConns,
		utils.ReplicationConnsCfg:    dbcfg.RplConns,
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
	if dbcfg.Port != utils.EmptyString {
		dbPort, _ := strconv.Atoi(dbcfg.Port)
		initialMP[utils.DataDbPortCfg] = dbPort
	}
	return
}

func diffStorDBDbJsonCfg(d *DbJsonCfg, v1, v2 *StorDbCfg) *DbJsonCfg {
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

	if !utils.SliceStringEqual(v1.RplConns, v2.RplConns) {
		d.Replication_conns = &v2.RplConns
	}

	if !utils.SliceStringEqual(v1.StringIndexedFields, v2.StringIndexedFields) {
		d.String_indexed_fields = &v2.StringIndexedFields
	}
	if !utils.SliceStringEqual(v1.PrefixIndexedFields, v2.PrefixIndexedFields) {
		d.Prefix_indexed_fields = &v2.PrefixIndexedFields
	}

	d.Items = diffMapItemOptJson(d.Items, v1.Items, v2.Items)
	d.Opts = diffMap(d.Opts, v1.Opts, v2.Opts)

	return d
}
