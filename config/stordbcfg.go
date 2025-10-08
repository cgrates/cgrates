/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
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
	SQLLogLevel               int
	SQLConnMaxLifetime        time.Duration
	SQLDSNParams              map[string]string
	PgSSLMode                 string
	PgSSLCert                 string
	PgSSLKey                  string
	PgSSLPassword             string
	PgSSLCertMode             string
	PgSSLRootCert             string
	MySQLLocation             string
	MongoQueryTimeout         time.Duration
	MongoConnScheme           string
}

// StorDbCfg StroreDb config
type StorDbCfg struct {
	Type                string // should reflect the database type used to store logs
	Host                string // the host to connect to. Values that start with / are for UNIX domain sockets
	Port                string // the port to bind to
	Name                string // the name of the database to connect to
	User                string // the user to sign in as
	Password            string // the user's password
	StringIndexedFields []string
	PrefixIndexedFields []string
	RmtConns            []string // remote DataDB connIDs
	RplConns            []string // replication connIDs
	Items               map[string]*ItemOpts
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
	if jsnCfg.MYSQLDSNParams != nil {
		dbOpts.SQLDSNParams = make(map[string]string)
		dbOpts.SQLDSNParams = jsnCfg.MYSQLDSNParams
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
	if jsnCfg.MySQLLocation != nil {
		dbOpts.MySQLLocation = *jsnCfg.MySQLLocation
	}
	if jsnCfg.MongoQueryTimeout != nil {
		if dbOpts.MongoQueryTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.MongoQueryTimeout); err != nil {
			return
		}
	}
	if jsnCfg.MongoConnScheme != nil {
		dbOpts.MongoConnScheme = *jsnCfg.MongoConnScheme
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
			val := new(ItemOpts)
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
		InternalDBDumpPath:        dbOpts.InternalDBDumpPath,
		InternalDBBackupPath:      dbOpts.InternalDBBackupPath,
		InternalDBStartTimeout:    dbOpts.InternalDBStartTimeout,
		InternalDBDumpInterval:    dbOpts.InternalDBDumpInterval,
		InternalDBRewriteInterval: dbOpts.InternalDBRewriteInterval,
		InternalDBFileSizeLimit:   dbOpts.InternalDBFileSizeLimit,
		SQLMaxOpenConns:           dbOpts.SQLMaxOpenConns,
		SQLMaxIdleConns:           dbOpts.SQLMaxIdleConns,
		SQLLogLevel:               dbOpts.SQLLogLevel,
		SQLConnMaxLifetime:        dbOpts.SQLConnMaxLifetime,
		SQLDSNParams:              dbOpts.SQLDSNParams,
		PgSSLMode:                 dbOpts.PgSSLMode,
		PgSSLCert:                 dbOpts.PgSSLCert,
		PgSSLKey:                  dbOpts.PgSSLKey,
		PgSSLPassword:             dbOpts.PgSSLPassword,
		PgSSLCertMode:             dbOpts.PgSSLCertMode,
		PgSSLRootCert:             dbOpts.PgSSLRootCert,
		MySQLLocation:             dbOpts.MySQLLocation,
		MongoQueryTimeout:         dbOpts.MongoQueryTimeout,
		MongoConnScheme:           dbOpts.MongoConnScheme,
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

		Items: make(map[string]*ItemOpts),
		Opts:  dbcfg.Opts.Clone(),
	}
	for key, item := range dbcfg.Items {
		cln.Items[key] = item.Clone()
	}
	if dbcfg.StringIndexedFields != nil {
		cln.StringIndexedFields = slices.Clone(dbcfg.StringIndexedFields)
	}
	if dbcfg.PrefixIndexedFields != nil {
		cln.PrefixIndexedFields = slices.Clone(dbcfg.PrefixIndexedFields)
	}
	if dbcfg.RmtConns != nil {
		cln.RmtConns = slices.Clone(dbcfg.RmtConns)
	}
	if dbcfg.RplConns != nil {
		cln.RplConns = slices.Clone(dbcfg.RplConns)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (dbcfg StorDbCfg) AsMapInterface() any {
	opts := map[string]any{
		utils.InternalDBDumpPathCfg:        dbcfg.Opts.InternalDBDumpPath,
		utils.InternalDBBackupPathCfg:      dbcfg.Opts.InternalDBBackupPath,
		utils.InternalDBStartTimeoutCfg:    dbcfg.Opts.InternalDBStartTimeout.String(),
		utils.InternalDBDumpIntervalCfg:    dbcfg.Opts.InternalDBDumpInterval.String(),
		utils.InternalDBRewriteIntervalCfg: dbcfg.Opts.InternalDBRewriteInterval.String(),
		utils.InternalDBFileSizeLimitCfg:   dbcfg.Opts.InternalDBFileSizeLimit,
		utils.SQLMaxOpenConnsCfg:           dbcfg.Opts.SQLMaxOpenConns,
		utils.SQLMaxIdleConnsCfg:           dbcfg.Opts.SQLMaxIdleConns,
		utils.SQLLogLevelCfg:               dbcfg.Opts.SQLLogLevel,
		utils.SQLConnMaxLifetime:           dbcfg.Opts.SQLConnMaxLifetime.String(),
		utils.MYSQLDSNParams:               dbcfg.Opts.SQLDSNParams,
		utils.PgSSLModeCfg:                 dbcfg.Opts.PgSSLMode,
		utils.MysqlLocation:                dbcfg.Opts.MySQLLocation,
		utils.MongoQueryTimeoutCfg:         dbcfg.Opts.MongoQueryTimeout.String(),
		utils.MongoConnSchemeCfg:           dbcfg.Opts.MongoConnScheme,
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
	mp := map[string]any{
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
	return mp
}

func diffStorDBOptsJsonCfg(d *DBOptsJson, v1, v2 *StorDBOpts) *DBOptsJson {
	if d == nil {
		d = new(DBOptsJson)
	}
	if v1.InternalDBDumpPath != v2.InternalDBDumpPath {
		d.InternalDBDumpPath = utils.StringPointer(v2.InternalDBDumpPath)
	}
	if v1.InternalDBBackupPath != v2.InternalDBBackupPath {
		d.InternalDBBackupPath = utils.StringPointer(v2.InternalDBBackupPath)
	}
	if v1.InternalDBStartTimeout != v2.InternalDBStartTimeout {
		d.InternalDBStartTimeout = utils.StringPointer(v2.InternalDBStartTimeout.String())
	}
	if v1.InternalDBDumpInterval != v2.InternalDBDumpInterval {
		d.InternalDBDumpInterval = utils.StringPointer(v2.InternalDBDumpInterval.String())
	}
	if v1.InternalDBRewriteInterval != v2.InternalDBRewriteInterval {
		d.InternalDBRewriteInterval = utils.StringPointer(v2.InternalDBRewriteInterval.String())
	}
	if v1.InternalDBFileSizeLimit != v2.InternalDBFileSizeLimit {
		d.InternalDBFileSizeLimit = utils.StringPointer(fmt.Sprint(v2.InternalDBFileSizeLimit))
	}
	if v1.SQLMaxOpenConns != v2.SQLMaxOpenConns {
		d.SQLMaxOpenConns = utils.IntPointer(v2.SQLMaxOpenConns)
	}
	if v1.SQLMaxIdleConns != v2.SQLMaxIdleConns {
		d.SQLMaxIdleConns = utils.IntPointer(v2.SQLMaxIdleConns)
	}
	if v1.SQLLogLevel != v2.SQLLogLevel {
		d.SQLLogLevel = utils.IntPointer(v2.SQLLogLevel)
	}
	if v1.SQLConnMaxLifetime != v2.SQLConnMaxLifetime {
		d.SQLConnMaxLifetime = utils.StringPointer(v2.SQLConnMaxLifetime.String())
	}
	if !reflect.DeepEqual(v1.SQLDSNParams, v2.SQLDSNParams) {
		d.MYSQLDSNParams = v2.SQLDSNParams
	}
	if v1.PgSSLMode != v2.PgSSLMode {
		d.PgSSLMode = utils.StringPointer(v2.PgSSLMode)
	}
	if v1.PgSSLCert != v2.PgSSLCert {
		d.PgSSLCert = utils.StringPointer(v2.PgSSLCert)
	}
	if v1.PgSSLKey != v2.PgSSLKey {
		d.PgSSLKey = utils.StringPointer(v2.PgSSLKey)
	}
	if v1.PgSSLPassword != v2.PgSSLPassword {
		d.PgSSLPassword = utils.StringPointer(v2.PgSSLPassword)
	}
	if v1.PgSSLCertMode != v2.PgSSLCertMode {
		d.PgSSLCertMode = utils.StringPointer(v2.PgSSLCertMode)
	}
	if v1.PgSSLRootCert != v2.PgSSLRootCert {
		d.PgSSLRootCert = utils.StringPointer(v2.PgSSLRootCert)
	}
	if v1.MySQLLocation != v2.MySQLLocation {
		d.MySQLLocation = utils.StringPointer(v2.MySQLLocation)
	}
	if v1.MongoQueryTimeout != v2.MongoQueryTimeout {
		d.MongoQueryTimeout = utils.StringPointer(v2.MongoQueryTimeout.String())
	}
	if v1.MongoConnScheme != v2.MongoConnScheme {
		d.MongoConnScheme = utils.StringPointer(v2.MongoConnScheme)
	}
	return d
}

func diffStorDBJsonCfg(d *DbJsonCfg, v1, v2 *StorDbCfg) *DbJsonCfg {
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
	if !slices.Equal(v1.RmtConns, v2.RmtConns) {
		d.Remote_conns = &v2.RmtConns
	}

	if !slices.Equal(v1.RplConns, v2.RplConns) {
		d.Replication_conns = &v2.RplConns
	}

	if !slices.Equal(v1.StringIndexedFields, v2.StringIndexedFields) {
		d.String_indexed_fields = &v2.StringIndexedFields
	}
	if !slices.Equal(v1.PrefixIndexedFields, v2.PrefixIndexedFields) {
		d.Prefix_indexed_fields = &v2.PrefixIndexedFields
	}

	d.Items = diffMapItemOptJson(d.Items, v1.Items, v2.Items)
	d.Opts = diffStorDBOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)

	return d
}

// ToTransCacheOpts returns to ltcache.TransCacheOpts from StorDBOpts
func (s *StorDBOpts) ToTransCacheOpts() *ltcache.TransCacheOpts {
	if s == nil {
		return nil
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
