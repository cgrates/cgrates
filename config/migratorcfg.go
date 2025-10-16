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
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// MigratorCgrCfg the migrator config section
type MigratorCgrCfg struct {
	UsersFilters []string
	FromItems    map[string]*MigratorFromItem // contains the in items as the keys of the map, and the DataDB ids of each item in MigratorFromItems
	OutDBOpts    *DBOpts
}

// MigratorFromItem contains the DataDB id of the item
type MigratorFromItem struct {
	DBConn string // ID of the DB connection that this item belongs to
}

// loadFromJSONCfg loads Database config from JsonCfg
func (mfi *MigratorFromItem) loadFromJSONCfg(jsonII *FromItemJson) (err error) {
	if jsonII == nil {
		return
	}
	if jsonII.DbConn != nil {
		mfi.DBConn = *jsonII.DbConn
	}
	return
}

// Clone returns the cloned object
func (mfi *MigratorFromItem) Clone() *MigratorFromItem {
	if mfi == nil {
		return nil
	}
	return &MigratorFromItem{
		DBConn: mfi.DBConn,
	}
}

// AsMapInterface returns the config as a map[string]any
func (mfi *MigratorFromItem) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.DBConnCfg: mfi.DBConn,
	}
	return
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
	if jsnCfg.Users_filters != nil && len(*jsnCfg.Users_filters) != 0 {
		mg.UsersFilters = slices.Clone(*jsnCfg.Users_filters)
	}
	if jsnCfg.FromItems != nil {
		for kJsn, vJsn := range jsnCfg.FromItems {
			val, has := mg.FromItems[kJsn]
			if val == nil || !has {
				val = new(MigratorFromItem)
			}
			if err = val.loadFromJSONCfg(vJsn); err != nil {
				return
			}
			mg.FromItems[kJsn] = val
		}
	}
	if jsnCfg.Out_db_opts != nil {
		err = mg.OutDBOpts.loadFromJSONCfg(jsnCfg.Out_db_opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (mg MigratorCgrCfg) AsMapInterface() any {
	outDBOpts := map[string]any{
		utils.RedisMaxConnsCfg:           mg.OutDBOpts.RedisMaxConns,
		utils.RedisConnectAttemptsCfg:    mg.OutDBOpts.RedisConnectAttempts,
		utils.RedisSentinelNameCfg:       mg.OutDBOpts.RedisSentinel,
		utils.RedisClusterCfg:            mg.OutDBOpts.RedisCluster,
		utils.RedisClusterSyncCfg:        mg.OutDBOpts.RedisClusterSync.String(),
		utils.RedisClusterOnDownDelayCfg: mg.OutDBOpts.RedisClusterOndownDelay.String(),
		utils.RedisConnectTimeoutCfg:     mg.OutDBOpts.RedisConnectTimeout.String(),
		utils.RedisReadTimeoutCfg:        mg.OutDBOpts.RedisReadTimeout.String(),
		utils.RedisWriteTimeoutCfg:       mg.OutDBOpts.RedisWriteTimeout.String(),
		utils.RedisPoolPipelineWindowCfg: mg.OutDBOpts.RedisPoolPipelineWindow.String(),
		utils.RedisPoolPipelineLimitCfg:  mg.OutDBOpts.RedisPoolPipelineLimit,
		utils.RedisTLSCfg:                mg.OutDBOpts.RedisTLS,
		utils.RedisClientCertificateCfg:  mg.OutDBOpts.RedisClientCertificate,
		utils.RedisClientKeyCfg:          mg.OutDBOpts.RedisClientKey,
		utils.RedisCACertificateCfg:      mg.OutDBOpts.RedisCACertificate,
		utils.MongoQueryTimeoutCfg:       mg.OutDBOpts.MongoQueryTimeout.String(),
		utils.MongoConnSchemeCfg:         mg.OutDBOpts.MongoConnScheme,
	}
	var items map[string]any
	if mg.FromItems != nil {
		items = make(map[string]any)
		for itemID, item := range mg.FromItems {
			items[itemID] = item.AsMapInterface()
		}
	}
	return map[string]any{
		utils.FromItemsCfg:    items,
		utils.OutDBOptsCfg:    outDBOpts,
		utils.UsersFiltersCfg: slices.Clone(mg.UsersFilters),
	}
}

func (MigratorCgrCfg) SName() string            { return MigratorJSON }
func (mg MigratorCgrCfg) CloneSection() Section { return mg.Clone() }

// Clone returns a deep copy of MigratorCgrCfg
func (mg MigratorCgrCfg) Clone() (cln *MigratorCgrCfg) {
	cln = &MigratorCgrCfg{
		FromItems: make(map[string]*MigratorFromItem),
		OutDBOpts: mg.OutDBOpts.Clone(),
	}
	for k, v := range mg.FromItems {
		cln.FromItems[k] = v.Clone()
	}
	if mg.UsersFilters != nil {
		cln.UsersFilters = slices.Clone(mg.UsersFilters)
	}
	return
}

type MigratorCfgJson struct {
	Users_filters *[]string
	FromItems     map[string]*FromItemJson
	Out_db_opts   *DBOptsJson
}

type FromItemJson struct {
	DbConn *string
}

func (mfi *MigratorFromItem) Equals(itm2 *MigratorFromItem) bool {
	return mfi == nil && itm2 == nil ||
		mfi != nil && itm2 != nil && mfi.DBConn == itm2.DBConn
}

func diffFromItemJson(d *FromItemJson, v1, v2 *MigratorFromItem) *FromItemJson {
	if d == nil {
		d = new(FromItemJson)
	}
	if v2.DBConn != v1.DBConn {
		d.DbConn = utils.StringPointer(v2.DBConn)
	}
	return d
}

func diffMapFromItemJson(d map[string]*FromItemJson, v1 map[string]*MigratorFromItem,
	v2 map[string]*MigratorFromItem) map[string]*FromItemJson {
	if d == nil {
		d = make(map[string]*FromItemJson)
	}
	for k, val2 := range v2 {
		if val1, has := v1[k]; !has {
			d[k] = diffFromItemJson(d[k], new(MigratorFromItem), val2)
		} else if !val1.Equals(val2) {
			d[k] = diffFromItemJson(d[k], val1, val2)
		}
	}
	return d
}

func diffMigratorCfgJson(d *MigratorCfgJson, v1, v2 *MigratorCgrCfg) *MigratorCfgJson {
	if d == nil {
		d = new(MigratorCfgJson)
	}

	if !slices.Equal(v1.UsersFilters, v2.UsersFilters) {
		d.Users_filters = utils.SliceStringPointer(slices.Clone(v2.UsersFilters))
	}
	d.FromItems = diffMapFromItemJson(d.FromItems, v1.FromItems, v2.FromItems)
	d.Out_db_opts = diffDataDBOptsJsonCfg(d.Out_db_opts, v1.OutDBOpts, v2.OutDBOpts)
	return d
}
