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
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// CacheParamCfg represents the config of a single cache partition
type CacheParamCfg struct {
	Limit     int
	TTL       time.Duration
	StaticTTL bool
	Precache  bool
	Remote    bool
	Replicate bool
}

func (cParam *CacheParamCfg) loadFromJSONCfg(jsnCfg *CacheParamJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Limit != nil {
		cParam.Limit = *jsnCfg.Limit
	}
	if jsnCfg.Static_ttl != nil {
		cParam.StaticTTL = *jsnCfg.Static_ttl
	}
	if jsnCfg.Precache != nil {
		cParam.Precache = *jsnCfg.Precache
	}
	if jsnCfg.Remote != nil {
		cParam.Remote = *jsnCfg.Remote
	}
	if jsnCfg.Replicate != nil {
		cParam.Replicate = *jsnCfg.Replicate
	}
	if jsnCfg.Ttl != nil {
		cParam.TTL, err = utils.ParseDurationWithNanosecs(*jsnCfg.Ttl)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (cParam *CacheParamCfg) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.LimitCfg:     cParam.Limit,
		utils.StaticTTLCfg: cParam.StaticTTL,
		utils.PrecacheCfg:  cParam.Precache,
		utils.RemoteCfg:    cParam.Remote,
		utils.ReplicateCfg: cParam.Replicate,
	}
	if cParam.TTL != 0 {
		initialMP[utils.TTLCfg] = cParam.TTL.String()
	}
	return
}

// Clone returns a deep copy of CacheParamCfg
func (cParam CacheParamCfg) Clone() (cln *CacheParamCfg) {
	return &CacheParamCfg{
		Limit:     cParam.Limit,
		TTL:       cParam.TTL,
		StaticTTL: cParam.StaticTTL,
		Precache:  cParam.Precache,
		Remote:    cParam.Remote,
		Replicate: cParam.Replicate,
	}
}

// CacheCfg used to store the cache config
type CacheCfg struct {
	Partitions       map[string]*CacheParamCfg
	ReplicationConns []string
	RemoteConns      []string
}

// loadCacheCfg loads the Cache section of the configuration
func (cCfg *CacheCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnCacheCfg := new(CacheJsonCfg)
	if err = jsnCfg.GetSection(ctx, CacheJSON, jsnCacheCfg); err != nil {
		return
	}
	return cCfg.loadFromJSONCfg(jsnCacheCfg)
}

func (cCfg *CacheCfg) loadFromJSONCfg(jsnCfg *CacheJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	for kJsn, vJsn := range jsnCfg.Partitions {
		val := new(CacheParamCfg)
		if err := val.loadFromJSONCfg(vJsn); err != nil {
			return err
		}
		cCfg.Partitions[kJsn] = val
	}
	if jsnCfg.Replication_conns != nil {
		cCfg.ReplicationConns = make([]string, len(*jsnCfg.Replication_conns))
		for idx, connID := range *jsnCfg.Replication_conns {
			if connID == utils.MetaInternal {
				return fmt.Errorf("replication connection ID needs to be different than *internal")
			}
			cCfg.ReplicationConns[idx] = connID
		}
	}
	if jsnCfg.Remote_conns != nil {
		cCfg.RemoteConns = make([]string, len(*jsnCfg.Remote_conns))
		for idx, connID := range *jsnCfg.Remote_conns {
			cCfg.RemoteConns[idx] = connID
		}
	}
	return nil
}

// AsTransCacheConfig transforms the cache config in ltcache config
func (cCfg CacheCfg) AsTransCacheConfig() (tcCfg map[string]*ltcache.CacheConfig) {
	tcCfg = make(map[string]*ltcache.CacheConfig, len(cCfg.Partitions))
	for k, cPcfg := range cCfg.Partitions {
		tcCfg[k] = &ltcache.CacheConfig{
			MaxItems:  cPcfg.Limit,
			TTL:       cPcfg.TTL,
			StaticTTL: cPcfg.StaticTTL,
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (cCfg CacheCfg) AsMapInterface() any {
	partitions := make(map[string]any, len(cCfg.Partitions))
	for key, value := range cCfg.Partitions {
		partitions[key] = value.AsMapInterface()
	}
	mp := map[string]any{utils.PartitionsCfg: partitions}
	if cCfg.ReplicationConns != nil {
		mp[utils.ReplicationConnsCfg] = cCfg.ReplicationConns
	}
	if cCfg.RemoteConns != nil {
		mp[utils.RemoteConnsCfg] = cCfg.RemoteConns
	}
	return mp
}

func (CacheCfg) SName() string              { return CacheJSON }
func (cCfg CacheCfg) CloneSection() Section { return cCfg.Clone() }

// Clone returns a deep copy of CacheCfg
func (cCfg CacheCfg) Clone() (cln *CacheCfg) {
	cln = &CacheCfg{
		Partitions: make(map[string]*CacheParamCfg),
	}
	for key, par := range cCfg.Partitions {
		cln.Partitions[key] = par.Clone()
	}
	if cCfg.ReplicationConns != nil {
		cln.ReplicationConns = slices.Clone(cCfg.ReplicationConns)
	}
	if cCfg.RemoteConns != nil {
		cln.RemoteConns = slices.Clone(cCfg.RemoteConns)
	}
	return
}

type CacheParamJsonCfg struct {
	Limit      *int
	Ttl        *string
	Static_ttl *bool
	Precache   *bool
	Remote     *bool
	Replicate  *bool
}

func diffCacheParamsJsonCfg(d map[string]*CacheParamJsonCfg, v2 map[string]*CacheParamCfg) map[string]*CacheParamJsonCfg {
	if d == nil {
		d = make(map[string]*CacheParamJsonCfg)
	}
	for k, val2 := range v2 {
		d[k] = &CacheParamJsonCfg{
			Limit:      utils.IntPointer(val2.Limit),
			Ttl:        utils.StringPointer(val2.TTL.String()),
			Static_ttl: utils.BoolPointer(val2.StaticTTL),
			Precache:   utils.BoolPointer(val2.Precache),
			Remote:     utils.BoolPointer(val2.Remote),
			Replicate:  utils.BoolPointer(val2.Replicate),
		}
	}
	return d
}

type CacheJsonCfg struct {
	Partitions        map[string]*CacheParamJsonCfg
	Replication_conns *[]string
	Remote_conns      *[]string
}

func diffCacheJsonCfg(d *CacheJsonCfg, v1, v2 *CacheCfg) *CacheJsonCfg {
	if d == nil {
		d = new(CacheJsonCfg)
	}
	d.Partitions = diffCacheParamsJsonCfg(d.Partitions, v2.Partitions)
	if !slices.Equal(v1.ReplicationConns, v2.ReplicationConns) {
		d.Replication_conns = &v2.ReplicationConns
	}
	if !slices.Equal(v1.RemoteConns, v2.RemoteConns) {
		d.Remote_conns = &v2.RemoteConns
	}
	return d
}
