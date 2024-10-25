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
	"slices"
	"strings"

	"github.com/cgrates/birpc/context"
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
	UsersFilters      []string
	OutDataDBOpts     *DataDBOpts
}

// loadMigratorCgrCfg loads the Migrator section of the configuration
func (mg *MigratorCgrCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnMigratorCgrCfg := new(MigratorCfgJson)
	if err = jsnCfg.GetSection(ctx, MigratorJSON, jsnMigratorCgrCfg); err != nil {
		return
	}
	return mg.loadFromJSONCfg(jsnMigratorCgrCfg)
}

func (mg *MigratorCgrCfg) loadFromJSONCfg(jsnCfg *MigratorCfgJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Out_dataDB_type != nil {

		if !strings.HasPrefix(*jsnCfg.Out_dataDB_type, "*") {
			mg.OutDataDBType = fmt.Sprintf("*%v", *jsnCfg.Out_dataDB_type)
		} else {
			mg.OutDataDBType = *jsnCfg.Out_dataDB_type
		}
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
	if jsnCfg.Users_filters != nil && len(*jsnCfg.Users_filters) != 0 {
		mg.UsersFilters = slices.Clone(*jsnCfg.Users_filters)
	}
	if jsnCfg.Out_dataDB_opts != nil {
		err = mg.OutDataDBOpts.loadFromJSONCfg(jsnCfg.Out_dataDB_opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (mg MigratorCgrCfg) AsMapInterface(string) any {
	outDataDBOpts := map[string]any{
		utils.RedisMaxConnsCfg:           mg.OutDataDBOpts.RedisMaxConns,
		utils.RedisConnectAttemptsCfg:    mg.OutDataDBOpts.RedisConnectAttempts,
		utils.RedisSentinelNameCfg:       mg.OutDataDBOpts.RedisSentinel,
		utils.RedisClusterCfg:            mg.OutDataDBOpts.RedisCluster,
		utils.RedisClusterSyncCfg:        mg.OutDataDBOpts.RedisClusterSync.String(),
		utils.RedisClusterOnDownDelayCfg: mg.OutDataDBOpts.RedisClusterOndownDelay.String(),
		utils.RedisConnectTimeoutCfg:     mg.OutDataDBOpts.RedisConnectTimeout.String(),
		utils.RedisReadTimeoutCfg:        mg.OutDataDBOpts.RedisReadTimeout.String(),
		utils.RedisWriteTimeoutCfg:       mg.OutDataDBOpts.RedisWriteTimeout.String(),
		utils.RedisPoolPipelineWindowCfg: mg.OutDataDBOpts.RedisPoolPipelineWindow.String(),
		utils.RedisPoolPipelineLimitCfg:  mg.OutDataDBOpts.RedisPoolPipelineLimit,
		utils.RedisTLSCfg:                mg.OutDataDBOpts.RedisTLS,
		utils.RedisClientCertificateCfg:  mg.OutDataDBOpts.RedisClientCertificate,
		utils.RedisClientKeyCfg:          mg.OutDataDBOpts.RedisClientKey,
		utils.RedisCACertificateCfg:      mg.OutDataDBOpts.RedisCACertificate,
		utils.MongoQueryTimeoutCfg:       mg.OutDataDBOpts.MongoQueryTimeout.String(),
		utils.MongoConnSchemeCfg:         mg.OutDataDBOpts.MongoConnScheme,
	}
	return map[string]any{
		utils.OutDataDBTypeCfg:     mg.OutDataDBType,
		utils.OutDataDBHostCfg:     mg.OutDataDBHost,
		utils.OutDataDBPortCfg:     mg.OutDataDBPort,
		utils.OutDataDBNameCfg:     mg.OutDataDBName,
		utils.OutDataDBUserCfg:     mg.OutDataDBUser,
		utils.OutDataDBPasswordCfg: mg.OutDataDBPassword,
		utils.OutDataDBEncodingCfg: mg.OutDataDBEncoding,
		utils.OutDataDBOptsCfg:     outDataDBOpts,
		utils.UsersFiltersCfg:      slices.Clone(mg.UsersFilters),
	}
}

func (MigratorCgrCfg) SName() string            { return MigratorJSON }
func (mg MigratorCgrCfg) CloneSection() Section { return mg.Clone() }

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

		OutDataDBOpts: mg.OutDataDBOpts.Clone(),
	}
	if mg.UsersFilters != nil {
		cln.UsersFilters = slices.Clone(mg.UsersFilters)
	}
	return
}

type MigratorCfgJson struct {
	Out_dataDB_type     *string
	Out_dataDB_host     *string
	Out_dataDB_port     *string
	Out_dataDB_name     *string
	Out_dataDB_user     *string
	Out_dataDB_password *string
	Out_dataDB_encoding *string
	Users_filters       *[]string
	Out_dataDB_opts     *DBOptsJson
}

func diffMigratorCfgJson(d *MigratorCfgJson, v1, v2 *MigratorCgrCfg) *MigratorCfgJson {
	if d == nil {
		d = new(MigratorCfgJson)
	}
	if v1.OutDataDBType != v2.OutDataDBType {
		d.Out_dataDB_type = utils.StringPointer(v2.OutDataDBType)
	}
	if v1.OutDataDBHost != v2.OutDataDBHost {
		d.Out_dataDB_host = utils.StringPointer(v2.OutDataDBHost)
	}
	if v1.OutDataDBPort != v2.OutDataDBPort {
		d.Out_dataDB_port = utils.StringPointer(v2.OutDataDBPort)
	}
	if v1.OutDataDBName != v2.OutDataDBName {
		d.Out_dataDB_name = utils.StringPointer(v2.OutDataDBName)
	}
	if v1.OutDataDBUser != v2.OutDataDBUser {
		d.Out_dataDB_user = utils.StringPointer(v2.OutDataDBUser)
	}
	if v1.OutDataDBPassword != v2.OutDataDBPassword {
		d.Out_dataDB_password = utils.StringPointer(v2.OutDataDBPassword)
	}
	if v1.OutDataDBEncoding != v2.OutDataDBEncoding {
		d.Out_dataDB_encoding = utils.StringPointer(v2.OutDataDBEncoding)
	}

	if !slices.Equal(v1.UsersFilters, v2.UsersFilters) {
		d.Users_filters = utils.SliceStringPointer(slices.Clone(v2.UsersFilters))
	}
	d.Out_dataDB_opts = diffDataDBOptsJsonCfg(d.Out_dataDB_opts, v1.OutDataDBOpts, v2.OutDataDBOpts)
	return d
}
