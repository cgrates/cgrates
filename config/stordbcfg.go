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

type StorDBOpts struct {
	SQLMaxOpenConns    int
	SQLMaxIdleConns    int
	SQLConnMaxLifetime time.Duration
	MongoQueryTimeout  time.Duration
	PgSSLMode          string
	MySQLLocation      string
	MySQLDSNParams     map[string]string
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
	if jsnCfg.MySQLDSNParams != nil {
		dbOpts.MySQLDSNParams = make(map[string]string)
		dbOpts.MySQLDSNParams = jsnCfg.MySQLDSNParams
	}
	if jsnCfg.MongoQueryTimeout != nil {
		if dbOpts.MongoQueryTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.MongoQueryTimeout); err != nil {
			return
		}
	}
	if jsnCfg.PgSSLMode != nil {
		dbOpts.PgSSLMode = *jsnCfg.PgSSLMode
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
				return fmt.Errorf("Remote connection ID needs to be different than *internal")
			}
			dbcfg.RmtConns[i] = item
		}
	}
	if jsnDbCfg.Replication_conns != nil {
		dbcfg.RplConns = make([]string, len(*jsnDbCfg.Replication_conns))
		for i, item := range *jsnDbCfg.Replication_conns {
			if item == utils.MetaInternal {
				return fmt.Errorf("Replication connection ID needs to be different than *internal")
			}
			dbcfg.RplConns[i] = item
		}
	}
	if jsnDbCfg.Items != nil {
		for kJsn, vJsn := range *jsnDbCfg.Items {
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
	return nil
}

func (dbOpts *StorDBOpts) Clone() *StorDBOpts {
	return &StorDBOpts{
		SQLMaxOpenConns:    dbOpts.SQLMaxOpenConns,
		SQLMaxIdleConns:    dbOpts.SQLMaxIdleConns,
		SQLConnMaxLifetime: dbOpts.SQLConnMaxLifetime,
		MySQLDSNParams:     dbOpts.MySQLDSNParams,
		MongoQueryTimeout:  dbOpts.MongoQueryTimeout,
		PgSSLMode:          dbOpts.PgSSLMode,
		MySQLLocation:      dbOpts.MySQLLocation,
	}
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
		Opts:  dbcfg.Opts.Clone(),
	}
	for key, item := range dbcfg.Items {
		cln.Items[key] = item.Clone()
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
func (dbcfg *StorDbCfg) AsMapInterface() (mp map[string]interface{}) {
	opts := map[string]interface{}{
		utils.SQLMaxOpenConnsCfg:   dbcfg.Opts.SQLMaxOpenConns,
		utils.SQLMaxIdleConnsCfg:   dbcfg.Opts.SQLMaxIdleConns,
		utils.SQLConnMaxLifetime:   dbcfg.Opts.SQLConnMaxLifetime.String(),
		utils.MYSQLDSNParams:       dbcfg.Opts.MySQLDSNParams,
		utils.MongoQueryTimeoutCfg: dbcfg.Opts.MongoQueryTimeout.String(),
		utils.PgSSLModeCfg:         dbcfg.Opts.PgSSLMode,
		utils.MysqlLocation:        dbcfg.Opts.MySQLLocation,
	}
	mp = map[string]interface{}{
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
	return
}
