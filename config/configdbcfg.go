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
	"strconv"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// ConfigDBCfg Database config
type ConfigDBCfg struct {
	Type     string
	Host     string // The host to connect to. Values that start with / are for UNIX domain sockets.
	Port     string // The port to bind to.
	Name     string // The name of the database to connect to.
	User     string // The user to sign in as.
	Password string // The user's password.
	Opts     map[string]interface{}
}

// loadConfigDBCfg loads the DataDB section of the configuration
func (dbcfg *ConfigDBCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnConfigDBCfg := new(DbJsonCfg)
	if err = jsnCfg.GetSection(ctx, ConfigDBJSON, jsnConfigDBCfg); err != nil {
		return
	}
	if err = dbcfg.loadFromJSONCfg(jsnConfigDBCfg); err != nil {
		return
	}
	return
}

// loadFromJSONCfg loads Database config from JsonCfg
func (dbcfg *ConfigDBCfg) loadFromJSONCfg(jsnDbCfg *DbJsonCfg) (err error) {
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

	if jsnDbCfg.Opts != nil {
		for k, v := range jsnDbCfg.Opts {
			dbcfg.Opts[k] = v
		}
	}
	return
}

func (ConfigDBCfg) SName() string               { return ConfigDBJSON }
func (dbcfg ConfigDBCfg) CloneSection() Section { return dbcfg.Clone() }

// Clone returns the cloned object
func (dbcfg ConfigDBCfg) Clone() (cln *ConfigDBCfg) {
	cln = &ConfigDBCfg{
		Type:     dbcfg.Type,
		Host:     dbcfg.Host,
		Port:     dbcfg.Port,
		Name:     dbcfg.Name,
		User:     dbcfg.User,
		Password: dbcfg.Password,
		Opts:     make(map[string]interface{}),
	}
	for k, v := range dbcfg.Opts {
		cln.Opts[k] = v
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (dbcfg ConfigDBCfg) AsMapInterface(string) interface{} {
	mp := map[string]interface{}{
		utils.DataDbTypeCfg: utils.Meta + dbcfg.Type,
		utils.DataDbHostCfg: dbcfg.Host,
		utils.DataDbNameCfg: dbcfg.Name,
		utils.DataDbUserCfg: dbcfg.User,
		utils.DataDbPassCfg: dbcfg.Password,
	}
	opts := make(map[string]interface{})
	for k, v := range dbcfg.Opts {
		opts[k] = v
	}
	mp[utils.OptsCfg] = opts
	if dbcfg.Port != "" {
		mp[utils.DataDbPortCfg], _ = strconv.Atoi(dbcfg.Port)
	}
	return mp
}
