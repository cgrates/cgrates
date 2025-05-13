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
	"github.com/cgrates/ltcache"
)

type StorDBOpts struct {
	InternalDBDumpPath        string        // Path to the dump file
	InternalDBBackupPath      string        // Path where db dump will backup
	InternalDBStartTimeout    time.Duration // Transcache recover from dump files timeout duration
	InternalDBDumpInterval    time.Duration // Regurarly dump database to file
	InternalDBRewriteInterval time.Duration // Regurarly rewrite dump files
	InternalDBFileSizeLimit   int64         // maximum size that can be written in a singular dump file
	SQLMaxOpenConns           int
	SQLMaxIdleConns           int
	SQLConnMaxLifetime        time.Duration
	SQLLogLevel               int
	MongoQueryTimeout         time.Duration
	MongoConnScheme           string
	PgSSLMode                 string
	PgSSLCert                 string
	PgSSLKey                  string
	PgSSLPassword             string
	PgSSLCertMode             string
	PgSSLRootCert             string
	PgSchema                  string
	MySQLLocation             string
	MySQLDSNParams            map[string]string
}

// StorDbCfg StroreDb config
type StorDbCfg struct {
	Type                string // Should reflect the database type used to store logs
	Host                string // The host to connect to. Values that start with / are for UNIX domain sockets.
	Port                string // The port to bind to.
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
	if jsnCfg.InternalDBDumpPath != nil {
		dbOpts.InternalDBDumpPath = *jsnCfg.InternalDBDumpPath
	}
	if jsnCfg.InternalDBBackupPath != nil {
		dbOpts.InternalDBBackupPath = *jsnCfg.InternalDBBackupPath
	}
	if jsnCfg.InternalDBStartTimeout != nil {
		if dbOpts.InternalDBStartTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.InternalDBStartTimeout); err != nil {
			return err
		}
	}
	if jsnCfg.InternalDBDumpInterval != nil {
		if dbOpts.InternalDBDumpInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.InternalDBDumpInterval); err != nil {
			return err
		}
	}
	if jsnCfg.InternalDBRewriteInterval != nil {
		if dbOpts.InternalDBRewriteInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.InternalDBRewriteInterval); err != nil {
			return err
		}
	}
	if jsnCfg.InternalDBFileSizeLimit != nil {
		if dbOpts.InternalDBFileSizeLimit, err = utils.ParseBinarySize(*jsnCfg.InternalDBFileSizeLimit); err != nil {
			return err
		}
	}
	if jsnCfg.SQLMaxOpenConns != nil {
		dbOpts.SQLMaxOpenConns = *jsnCfg.SQLMaxOpenConns
	}
	if jsnCfg.SQLMaxIdleConns != nil {
		dbOpts.SQLMaxIdleConns = *jsnCfg.SQLMaxIdleConns
	}
	if jsnCfg.SQLLogLevel != nil {
		dbOpts.SQLLogLevel = *jsnCfg.SQLLogLevel
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
	if jsnCfg.MongoConnScheme != nil {
		dbOpts.MongoConnScheme = *jsnCfg.MongoConnScheme
	}
	if jsnCfg.PgSSLMode != nil {
		dbOpts.PgSSLMode = *jsnCfg.PgSSLMode
	}
	if jsnCfg.PgSSLCert != nil {
		dbOpts.PgSSLCert = *jsnCfg.PgSSLCert
	}
	if jsnCfg.PgSSLKey != nil {
		dbOpts.PgSSLKey = *jsnCfg.PgSSLKey
	}
	if jsnCfg.PgSSLPassword != nil {
		dbOpts.PgSSLPassword = *jsnCfg.PgSSLPassword
	}
	if jsnCfg.PgSSLCertMode != nil {
		dbOpts.PgSSLCertMode = *jsnCfg.PgSSLCertMode
	}
	if jsnCfg.PgSSLRootCert != nil {
		dbOpts.PgSSLRootCert = *jsnCfg.PgSSLRootCert
	}
	if jsnCfg.PgSchema != nil {
		dbOpts.PgSchema = *jsnCfg.PgSchema
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
		if !strings.HasPrefix(*jsnDbCfg.Db_type, "*") {
			dbcfg.Type = fmt.Sprintf("*%v", *jsnDbCfg.Db_type)
		} else {
			dbcfg.Type = *jsnDbCfg.Db_type
		}
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
	return
}

func (dbOpts *StorDBOpts) Clone() *StorDBOpts {
	return &StorDBOpts{
		InternalDBDumpPath:        dbOpts.InternalDBDumpPath,
		InternalDBBackupPath:      dbOpts.InternalDBBackupPath,
		InternalDBStartTimeout:    dbOpts.InternalDBStartTimeout,
		InternalDBDumpInterval:    dbOpts.InternalDBDumpInterval,
		InternalDBRewriteInterval: dbOpts.InternalDBRewriteInterval,
		InternalDBFileSizeLimit:   dbOpts.InternalDBFileSizeLimit,
		SQLMaxOpenConns:           dbOpts.SQLMaxOpenConns,
		SQLMaxIdleConns:           dbOpts.SQLMaxIdleConns,
		SQLConnMaxLifetime:        dbOpts.SQLConnMaxLifetime,
		SQLLogLevel:               dbOpts.SQLLogLevel,
		MySQLDSNParams:            dbOpts.MySQLDSNParams,
		MongoQueryTimeout:         dbOpts.MongoQueryTimeout,
		MongoConnScheme:           dbOpts.MongoConnScheme,
		PgSSLMode:                 dbOpts.PgSSLMode,
		PgSSLCert:                 dbOpts.PgSSLCert,
		PgSSLKey:                  dbOpts.PgSSLKey,
		PgSSLPassword:             dbOpts.PgSSLPassword,
		PgSSLCertMode:             dbOpts.PgSSLCertMode,
		PgSSLRootCert:             dbOpts.PgSSLRootCert,
		PgSchema:                  dbOpts.PgSchema,
		MySQLLocation:             dbOpts.MySQLLocation,
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
		copy(cln.StringIndexedFields, dbcfg.StringIndexedFields)
	}
	if dbcfg.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = make([]string, len(dbcfg.PrefixIndexedFields))
		copy(cln.PrefixIndexedFields, dbcfg.PrefixIndexedFields)
	}
	if dbcfg.RmtConns != nil {
		cln.RmtConns = make([]string, len(dbcfg.RmtConns))
		copy(cln.RmtConns, dbcfg.RmtConns)
	}
	if dbcfg.RplConns != nil {
		cln.RplConns = make([]string, len(dbcfg.RplConns))
		copy(cln.RplConns, dbcfg.RplConns)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (dbcfg *StorDbCfg) AsMapInterface() (mp map[string]any) {
	opts := map[string]any{
		utils.InternalDBDumpPathCfg:        dbcfg.Opts.InternalDBDumpPath,
		utils.InternalDBBackupPathCfg:      dbcfg.Opts.InternalDBBackupPath,
		utils.InternalDBStartTimeoutCfg:    dbcfg.Opts.InternalDBStartTimeout.String(),
		utils.InternalDBDumpIntervalCfg:    dbcfg.Opts.InternalDBDumpInterval.String(),
		utils.InternalDBRewriteIntervalCfg: dbcfg.Opts.InternalDBRewriteInterval.String(),
		utils.InternalDBFileSizeLimitCfg:   dbcfg.Opts.InternalDBFileSizeLimit,
		utils.SQLMaxOpenConnsCfg:           dbcfg.Opts.SQLMaxOpenConns,
		utils.SQLMaxIdleConnsCfg:           dbcfg.Opts.SQLMaxIdleConns,
		utils.SQLConnMaxLifetime:           dbcfg.Opts.SQLConnMaxLifetime.String(),
		utils.SQLLogLevel:                  dbcfg.Opts.SQLLogLevel,
		utils.MYSQLDSNParams:               dbcfg.Opts.MySQLDSNParams,
		utils.MongoQueryTimeoutCfg:         dbcfg.Opts.MongoQueryTimeout.String(),
		utils.MongoConnSchemeCfg:           dbcfg.Opts.MongoConnScheme,
		utils.PgSSLModeCfg:                 dbcfg.Opts.PgSSLMode,
		utils.PgSchema:                     dbcfg.Opts.PgSchema,
		utils.MysqlLocation:                dbcfg.Opts.MySQLLocation,
	}
	if dbcfg.Opts.PgSSLCert != "" {
		opts[utils.PgSSLCertCfg] = dbcfg.Opts.PgSSLCert
	}
	if dbcfg.Opts.PgSSLKey != "" {
		opts[utils.PgSSLKeyCfg] = dbcfg.Opts.PgSSLKey
	}
	if dbcfg.Opts.PgSSLPassword != "" {
		opts[utils.PgSSLPasswordCfg] = dbcfg.Opts.PgSSLPassword
	}
	if dbcfg.Opts.PgSSLCertMode != "" {
		opts[utils.PgSSLCertModeCfg] = dbcfg.Opts.PgSSLCertMode
	}
	if dbcfg.Opts.PgSSLRootCert != "" {
		opts[utils.PgSSLRootCertCfg] = dbcfg.Opts.PgSSLRootCert
	}
	mp = map[string]any{
		utils.DataDbTypeCfg:          dbcfg.Type,
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
		items := make(map[string]any)
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

// ToTransCacheOpts returns to ltcache.TransCacheOpts from StorDBOpts
func (s *StorDBOpts) ToTransCacheOpts() (tco *ltcache.TransCacheOpts) {
	if s == nil {
		return
	}
	return &ltcache.TransCacheOpts{
		DumpPath:        s.InternalDBDumpPath,
		BackupPath:      s.InternalDBBackupPath,
		StartTimeout:    s.InternalDBStartTimeout,
		DumpInterval:    s.InternalDBDumpInterval,
		RewriteInterval: s.InternalDBRewriteInterval,
		FileSizeLimit:   s.InternalDBFileSizeLimit,
	}
}
