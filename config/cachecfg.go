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
		cln.ReplicationConns = utils.CloneStringSlice(cCfg.ReplicationConns)
	}
	return
}

type CacheParamJsonCfg struct {
	Limit      *int
	Ttl        *string
	Static_ttl *bool
	Precache   *bool
	Replicate  *bool
}

func diffCacheParamJsonCfg(d *CacheParamJsonCfg, v1, v2 *CacheParamCfg) *CacheParamJsonCfg {
	if d == nil {
		d = new(CacheParamJsonCfg)
	}
	if v1.Limit != v2.Limit {
		d.Limit = utils.IntPointer(v2.Limit)
	}
	if v1.TTL != v2.TTL {
		d.Ttl = utils.StringPointer(v2.TTL.String())
	}
	if v1.StaticTTL != v2.StaticTTL {
		d.Static_ttl = utils.BoolPointer(v2.StaticTTL)
	}
	if v1.Precache != v2.Precache {
		d.Precache = utils.BoolPointer(v2.Precache)
	}
	if v1.Replicate != v2.Replicate {
		d.Replicate = utils.BoolPointer(v2.Replicate)
	}
	return d
}

func diffCacheParamsJsonCfg(d map[string]*CacheParamJsonCfg, v1, v2 map[string]*CacheParamCfg) map[string]*CacheParamJsonCfg {
	if d == nil {
		d = make(map[string]*CacheParamJsonCfg)
	}
	for k, val2 := range v2 {
		val1, has := v1[k]
		if !has || val1 == nil {
			val1 = new(CacheParamCfg)
		}
		d[k] = diffCacheParamJsonCfg(d[k], val1, val2)
	}
	return d
}

type CacheJsonCfg struct {
	Partitions        map[string]*CacheParamJsonCfg
	Replication_conns *[]string
}

func diffCacheJsonCfg(d *CacheJsonCfg, v1, v2 *CacheCfg) *CacheJsonCfg {
	if d == nil {
		d = new(CacheJsonCfg)
	}
	d.Partitions = diffCacheParamsJsonCfg(d.Partitions, v1.Partitions, v2.Partitions)
	if !utils.SliceStringEqual(v1.ReplicationConns, v2.ReplicationConns) {
		d.Replication_conns = &v2.ReplicationConns
	}
	return d
}
