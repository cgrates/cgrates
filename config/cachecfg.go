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

func (cCfg *CacheCfg) loadFromJSONCfg(jsnCfg *CacheJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Partitions != nil {
		for kJsn, vJsn := range *jsnCfg.Partitions {
			val := new(CacheParamCfg)
			if err := val.loadFromJSONCfg(vJsn); err != nil {
				return err
			}
			cCfg.Partitions[kJsn] = val
		}
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
		copy(cCfg.RemoteConns, *jsnCfg.Remote_conns)
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

// AddTmpCaches adds all the temporary caches configuration needed
func (cCfg *CacheCfg) AddTmpCaches() {
	cCfg.Partitions[utils.CacheRatingProfilesTmp] = &CacheParamCfg{
		Limit: -1,
		TTL:   time.Minute,
	}
}

// AsMapInterface returns the config as a map[string]any
func (cCfg *CacheCfg) AsMapInterface() (mp map[string]any) {
	mp = make(map[string]any)
	partitions := make(map[string]any, len(cCfg.Partitions))
	for key, value := range cCfg.Partitions {
		partitions[key] = value.AsMapInterface()
	}
	mp[utils.PartitionsCfg] = partitions
	if cCfg.ReplicationConns != nil {
		mp[utils.ReplicationConnsCfg] = cCfg.ReplicationConns
	}
	if cCfg.RemoteConns != nil {
		mp[utils.RemoteConnsCfg] = cCfg.RemoteConns
	}
	return
}

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
