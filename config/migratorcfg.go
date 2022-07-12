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
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// MigratorCgrCfg the migrator config section
type MigratorCgrCfg struct {
	OutDataDBType     string
	OutDataDBHost     string
	OutDataDBPort     string
	OutDataDBName     string
	OutDataDBUser     string
	OutDataDBPassword string
	OutDataDBEncoding string
	OutStorDBType     string
	OutStorDBHost     string
	OutStorDBPort     string
	OutStorDBName     string
	OutStorDBUser     string
	OutStorDBPassword string
	UsersFilters      []string
	OutDataDBOpts     *DataDBOpts
	OutStorDBOpts     *StorDBOpts
}

func (mg *MigratorCgrCfg) loadFromJSONCfg(jsnCfg *MigratorCfgJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Out_dataDB_type != nil {
		mg.OutDataDBType = strings.TrimPrefix(*jsnCfg.Out_dataDB_type, "*")
	}
	if jsnCfg.Out_dataDB_host != nil {
		mg.OutDataDBHost = *jsnCfg.Out_dataDB_host
	}
	if jsnCfg.Out_dataDB_port != nil {
		mg.OutDataDBPort = *jsnCfg.Out_dataDB_port
	}
	if jsnCfg.Out_dataDB_name != nil {
		mg.OutDataDBName = *jsnCfg.Out_dataDB_name
	}
	if jsnCfg.Out_dataDB_user != nil {
		mg.OutDataDBUser = *jsnCfg.Out_dataDB_user
	}
	if jsnCfg.Out_dataDB_password != nil {
		mg.OutDataDBPassword = *jsnCfg.Out_dataDB_password
	}
	if jsnCfg.Out_dataDB_encoding != nil {
		mg.OutDataDBEncoding = strings.TrimPrefix(*jsnCfg.Out_dataDB_encoding, "*")
	}
	if jsnCfg.Out_storDB_type != nil {
		mg.OutStorDBType = *jsnCfg.Out_storDB_type
	}
	if jsnCfg.Out_storDB_host != nil {
		mg.OutStorDBHost = *jsnCfg.Out_storDB_host
	}
	if jsnCfg.Out_storDB_port != nil {
		mg.OutStorDBPort = *jsnCfg.Out_storDB_port
	}
	if jsnCfg.Out_storDB_name != nil {
		mg.OutStorDBName = *jsnCfg.Out_storDB_name
	}
	if jsnCfg.Out_storDB_user != nil {
		mg.OutStorDBUser = *jsnCfg.Out_storDB_user
	}
	if jsnCfg.Out_storDB_password != nil {
		mg.OutStorDBPassword = *jsnCfg.Out_storDB_password
	}
	if jsnCfg.Users_filters != nil && len(*jsnCfg.Users_filters) != 0 {
		mg.UsersFilters = make([]string, len(*jsnCfg.Users_filters))
		for i, v := range *jsnCfg.Users_filters {
			mg.UsersFilters[i] = v
		}
	}
	if jsnCfg.Out_dataDB_opts != nil {
		err = mg.OutDataDBOpts.loadFromJSONCfg(jsnCfg.Out_dataDB_opts)
	}
	if jsnCfg.Out_storDB_opts != nil {
		err = mg.OutStorDBOpts.loadFromJSONCfg(jsnCfg.Out_storDB_opts)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (mg *MigratorCgrCfg) AsMapInterface() (initialMP map[string]interface{}) {
	outDataDBOpts := map[string]interface{}{
		utils.RedisMaxConnsCfg:           mg.OutDataDBOpts.RedisMaxConns,
		utils.RedisConnectAttemptsCfg:    mg.OutDataDBOpts.RedisConnectAttempts,
		utils.RedisSentinelNameCfg:       mg.OutDataDBOpts.RedisSentinel,
		utils.RedisClusterCfg:            mg.OutDataDBOpts.RedisCluster,
		utils.RedisClusterSyncCfg:        mg.OutDataDBOpts.RedisClusterSync.String(),
		utils.RedisClusterOnDownDelayCfg: mg.OutDataDBOpts.RedisClusterOndownDelay.String(),
		utils.RedisConnectTimeoutCfg:     mg.OutDataDBOpts.RedisConnectTimeout.String(),
		utils.RedisReadTimeoutCfg:        mg.OutDataDBOpts.RedisReadTimeout.String(),
		utils.RedisWriteTimeoutCfg:       mg.OutDataDBOpts.RedisWriteTimeout.String(),
		utils.MongoQueryTimeoutCfg:       mg.OutDataDBOpts.MongoQueryTimeout.String(),
		utils.RedisTLS:                   mg.OutDataDBOpts.RedisTLS,
		utils.RedisClientCertificate:     mg.OutDataDBOpts.RedisClientCertificate,
		utils.RedisClientKey:             mg.OutDataDBOpts.RedisClientKey,
		utils.RedisCACertificate:         mg.OutDataDBOpts.RedisCACertificate,
	}
	outStorDBOpts := map[string]interface{}{
		utils.SQLMaxOpenConnsCfg:   mg.OutStorDBOpts.SQLMaxOpenConns,
		utils.SQLMaxIdleConnsCfg:   mg.OutStorDBOpts.SQLMaxIdleConns,
		utils.SQLConnMaxLifetime:   mg.OutStorDBOpts.SQLConnMaxLifetime.String(),
		utils.MongoQueryTimeoutCfg: mg.OutStorDBOpts.MongoQueryTimeout.String(),
		utils.PgSSLModeCfg:         mg.OutStorDBOpts.PgSSLMode,
		utils.MysqlLocation:        mg.OutStorDBOpts.MySQLLocation,
		utils.MYSQLDSNParams:       mg.OutStorDBOpts.MySQLDSNParams,
	}
	return map[string]interface{}{
		utils.OutDataDBTypeCfg:     mg.OutDataDBType,
		utils.OutDataDBHostCfg:     mg.OutDataDBHost,
		utils.OutDataDBPortCfg:     mg.OutDataDBPort,
		utils.OutDataDBNameCfg:     mg.OutDataDBName,
		utils.OutDataDBUserCfg:     mg.OutDataDBUser,
		utils.OutDataDBPasswordCfg: mg.OutDataDBPassword,
		utils.OutDataDBEncodingCfg: mg.OutDataDBEncoding,
		utils.OutStorDBTypeCfg:     mg.OutStorDBType,
		utils.OutStorDBHostCfg:     mg.OutStorDBHost,
		utils.OutStorDBPortCfg:     mg.OutStorDBPort,
		utils.OutStorDBNameCfg:     mg.OutStorDBName,
		utils.OutStorDBUserCfg:     mg.OutStorDBUser,
		utils.OutStorDBPasswordCfg: mg.OutStorDBPassword,
		utils.OutDataDBOptsCfg:     outDataDBOpts,
		utils.OutStorDBOptsCfg:     outStorDBOpts,
		utils.UsersFiltersCfg:      utils.CloneStringSlice(mg.UsersFilters),
	}
}

// Clone returns a deep copy of MigratorCgrCfg
func (mg MigratorCgrCfg) Clone() (cln *MigratorCgrCfg) {
	cln = &MigratorCgrCfg{
		OutDataDBType:     mg.OutDataDBType,
		OutDataDBHost:     mg.OutDataDBHost,
		OutDataDBPort:     mg.OutDataDBPort,
		OutDataDBName:     mg.OutDataDBName,
		OutDataDBUser:     mg.OutDataDBUser,
		OutDataDBPassword: mg.OutDataDBPassword,
		OutDataDBEncoding: mg.OutDataDBEncoding,
		OutStorDBType:     mg.OutStorDBType,
		OutStorDBHost:     mg.OutStorDBHost,
		OutStorDBPort:     mg.OutStorDBPort,
		OutStorDBName:     mg.OutStorDBName,
		OutStorDBUser:     mg.OutStorDBUser,
		OutStorDBPassword: mg.OutStorDBPassword,
		OutDataDBOpts:     mg.OutDataDBOpts.Clone(),
		OutStorDBOpts:     mg.OutStorDBOpts.Clone(),
	}
	if mg.UsersFilters != nil {
		cln.UsersFilters = utils.CloneStringSlice(mg.UsersFilters)
	}
	return
}
