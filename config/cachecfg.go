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
	Replicate bool
}

func (cParam *CacheParamCfg) loadFromJSONCfg(jsnCfg *CacheParamJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Limit != nil {
		cParam.Limit = *jsnCfg.Limit
	}
	if jsnCfg.Ttl != nil {
		if cParam.TTL, err = utils.ParseDurationWithNanosecs(*jsnCfg.Ttl); err != nil {
			return err
		}
	}
	if jsnCfg.Static_ttl != nil {
		cParam.StaticTTL = *jsnCfg.Static_ttl
	}
	if jsnCfg.Precache != nil {
		cParam.Precache = *jsnCfg.Precache
	}
	if jsnCfg.Replicate != nil {
		cParam.Replicate = *jsnCfg.Replicate
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (cParam *CacheParamCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.LimitCfg:     cParam.Limit,
		utils.StaticTTLCfg: cParam.StaticTTL,
		utils.PrecacheCfg:  cParam.Precache,
		utils.ReplicateCfg: cParam.Replicate,
		utils.TTLCfg:       utils.EmptyString,
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
		Replicate: cParam.Replicate,
	}
}

// CacheCfg used to store the cache config
type CacheCfg struct {
	Partitions       map[string]*CacheParamCfg
	ReplicationConns []string
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

// AsMapInterface returns the config as a map[string]interface{}
func (cCfg *CacheCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = make(map[string]interface{})
	partitions := make(map[string]interface{}, len(cCfg.Partitions))
	for key, value := range cCfg.Partitions {
		partitions[key] = value.AsMapInterface()
	}
	initialMP[utils.PartitionsCfg] = partitions
	if cCfg.ReplicationConns != nil {
		initialMP[utils.ReplicationConnsCfg] = cCfg.ReplicationConns
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
		cln.ReplicationConns = make([]string, len(cCfg.ReplicationConns))
		for i, c := range cCfg.ReplicationConns {
			cln.ReplicationConns[i] = c
		}
	}
	return
}
