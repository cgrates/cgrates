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
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type StorDBOpts struct {
	SQLMaxOpenConns    int
	SQLMaxIdleConns    int
	SQLConnMaxLifetime time.Duration
	SQLDSNParams       map[string]string
	MongoQueryTimeout  time.Duration
	SSLMode            string
	MySQLLocation      string
}

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
	Opts                *StorDBOpts
}

// loadStorDBCfg loads the StorDB section of the configuration
func (dbcfg *StorDbCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnDataDbCfg := new(DbJsonCfg)
	if err = jsnCfg.GetSection(ctx, StorDBJSON, jsnDataDbCfg); err != nil {
		return
	}
	return dbcfg.loadFromJSONCfg(jsnDataDbCfg)
}

func (dbOpts *StorDBOpts) loadFromJSONCfg(jsnCfg *DBOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.SQLMaxOpenConns != nil {
		dbOpts.SQLMaxOpenConns = *jsnCfg.SQLMaxOpenConns
	}
	if jsnCfg.SQLMaxIdleConns != nil {
		dbOpts.SQLMaxIdleConns = *jsnCfg.SQLMaxIdleConns
	}
	if jsnCfg.SQLConnMaxLifetime != nil {
		if dbOpts.SQLConnMaxLifetime, err = utils.ParseDurationWithNanosecs(*jsnCfg.SQLConnMaxLifetime); err != nil {
			return
		}
	}
	if jsnCfg.MYSQLDSNParams != nil {
		dbOpts.SQLDSNParams = make(map[string]string)
		dbOpts.SQLDSNParams = jsnCfg.MYSQLDSNParams
	}
	if jsnCfg.MongoQueryTimeout != nil {
		if dbOpts.MongoQueryTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.MongoQueryTimeout); err != nil {
			return
		}
	}
	if jsnCfg.SSLMode != nil {
		dbOpts.SSLMode = *jsnCfg.SSLMode
	}
	if jsnCfg.MySQLLocation != nil {
		dbOpts.MySQLLocation = *jsnCfg.MySQLLocation
	}
	return
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
		dbcfg.Port = defaultDBPort(dbcfg.Type, port)
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
	if jsnDbCfg.Items != nil {
		for kJsn, vJsn := range jsnDbCfg.Items {
			val := new(ItemOpt)
			if err = val.loadFromJSONCfg(vJsn); err != nil {
				return
			}
			dbcfg.Items[kJsn] = val
		}
	}
	if jsnDbCfg.Opts != nil {
		err = dbcfg.Opts.loadFromJSONCfg(jsnDbCfg.Opts)
	}
	return
}

func (StorDbCfg) SName() string               { return StorDBJSON }
func (dbcfg StorDbCfg) CloneSection() Section { return dbcfg.Clone() }

func (dbOpts *StorDBOpts) Clone() *StorDBOpts {
	return &StorDBOpts{
		SQLMaxOpenConns:    dbOpts.SQLMaxOpenConns,
		SQLMaxIdleConns:    dbOpts.SQLMaxIdleConns,
		SQLConnMaxLifetime: dbOpts.SQLConnMaxLifetime,
		SQLDSNParams:       dbOpts.SQLDSNParams,
		MongoQueryTimeout:  dbOpts.MongoQueryTimeout,
		SSLMode:            dbOpts.SSLMode,
		MySQLLocation:      dbOpts.MySQLLocation,
	}
}

// Clone returns the cloned object
func (dbcfg StorDbCfg) Clone() (cln *StorDbCfg) {
	cln = &StorDbCfg{
		Type:     dbcfg.Type,
		Host:     dbcfg.Host,
		Port:     dbcfg.Port,
		Name:     dbcfg.Name,
		User:     dbcfg.User,
		Password: dbcfg.Password,

		Items: make(map[string]*ItemOpt),
		Opts:  dbcfg.Opts.Clone(),
	}
	for key, item := range dbcfg.Items {
		cln.Items[key] = item.Clone()
	}
	if dbcfg.StringIndexedFields != nil {
		cln.StringIndexedFields = utils.CloneStringSlice(dbcfg.StringIndexedFields)
	}
	if dbcfg.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = utils.CloneStringSlice(dbcfg.PrefixIndexedFields)
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
func (dbcfg StorDbCfg) AsMapInterface(string) interface{} {
	opts := map[string]interface{}{
		utils.SQLMaxOpenConnsCfg:   dbcfg.Opts.SQLMaxOpenConns,
		utils.SQLMaxIdleConnsCfg:   dbcfg.Opts.SQLMaxIdleConns,
		utils.SQLConnMaxLifetime:   dbcfg.Opts.SQLConnMaxLifetime.String(),
		utils.MYSQLDSNParams:       dbcfg.Opts.SQLDSNParams,
		utils.MongoQueryTimeoutCfg: dbcfg.Opts.MongoQueryTimeout.String(),
		utils.SSLModeCfg:           dbcfg.Opts.SSLMode,
		utils.MysqlLocation:        dbcfg.Opts.MySQLLocation,
	}
	mp := map[string]interface{}{
		utils.DataDbTypeCfg:          utils.Meta + dbcfg.Type,
		utils.DataDbHostCfg:          dbcfg.Host,
		utils.DataDbNameCfg:          dbcfg.Name,
		utils.DataDbUserCfg:          dbcfg.User,
		utils.DataDbPassCfg:          dbcfg.Password,
		utils.StringIndexedFieldsCfg: dbcfg.StringIndexedFields,
		utils.PrefixIndexedFieldsCfg: dbcfg.PrefixIndexedFields,
		utils.RemoteConnsCfg:         dbcfg.RmtConns,
		utils.ReplicationConnsCfg:    dbcfg.RplConns,
		utils.OptsCfg:                opts,
	}
	if dbcfg.Items != nil {
		items := make(map[string]interface{})
		for key, item := range dbcfg.Items {
			items[key] = item.AsMapInterface()
		}
		mp[utils.ItemsCfg] = items
	}
	if dbcfg.Port != utils.EmptyString {
		dbPort, _ := strconv.Atoi(dbcfg.Port)
		mp[utils.DataDbPortCfg] = dbPort
	}
	return mp
}

func diffStorDBOptsJsonCfg(d *DBOptsJson, v1, v2 *StorDBOpts) *DBOptsJson {
	if d == nil {
		d = new(DBOptsJson)
	}
	if v1.SQLMaxOpenConns != v2.SQLMaxOpenConns {
		d.SQLMaxOpenConns = utils.IntPointer(v2.SQLMaxOpenConns)
	}
	if v1.SQLMaxIdleConns != v2.SQLMaxIdleConns {
		d.SQLMaxIdleConns = utils.IntPointer(v2.SQLMaxIdleConns)
	}
	if v1.SQLConnMaxLifetime != v2.SQLConnMaxLifetime {
		d.SQLConnMaxLifetime = utils.StringPointer(v2.SQLConnMaxLifetime.String())
	}
	if !reflect.DeepEqual(v1.SQLDSNParams, v2.SQLDSNParams) {
		d.MYSQLDSNParams = v2.SQLDSNParams
	}
	if v1.MongoQueryTimeout != v2.MongoQueryTimeout {
		d.MongoQueryTimeout = utils.StringPointer(v2.MongoQueryTimeout.String())
	}
	if v1.SSLMode != v2.SSLMode {
		d.SSLMode = utils.StringPointer(v2.SSLMode)
	}
	if v1.MySQLLocation != v2.MySQLLocation {
		d.MySQLLocation = utils.StringPointer(v2.MySQLLocation)
	}
	return d
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
	d.Opts = diffStorDBOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)

	return d
}
