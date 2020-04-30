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
}

func (cParam *CacheParamCfg) loadFromJsonCfg(jsnCfg *CacheParamJsonCfg) error {
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
	return nil
}

func (cParam *CacheParamCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.Limit:     cParam.Limit,
		utils.TTL:       cParam.TTL,
		utils.StaticTTL: cParam.StaticTTL,
		utils.Precache:  cParam.Precache,
	}
}

// CacheCfg used to store the cache config
type CacheCfg map[string]*CacheParamCfg

func (cCfg CacheCfg) loadFromJsonCfg(jsnCfg *CacheJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	for kJsn, vJsn := range *jsnCfg {
		val := new(CacheParamCfg)
		if err := val.loadFromJsonCfg(vJsn); err != nil {
			return err
		}
		cCfg[kJsn] = val
	}
	return nil
}

// AsTransCacheConfig transforms the cache config in ltcache config
func (cCfg CacheCfg) AsTransCacheConfig() (tcCfg map[string]*ltcache.CacheConfig) {
	tcCfg = make(map[string]*ltcache.CacheConfig, len(cCfg))
	for k, cPcfg := range cCfg {
		tcCfg[k] = &ltcache.CacheConfig{
			MaxItems:  cPcfg.Limit,
			TTL:       cPcfg.TTL,
			StaticTTL: cPcfg.StaticTTL,
		}
	}
	return
}

// AddTmpCaches adds all the temotrary caches configuration needed
func (cCfg CacheCfg) AddTmpCaches() {
	cCfg[utils.CacheRatingProfilesTmp] = &CacheParamCfg{
		Limit: -1,
		TTL:   time.Minute,
	}
}

func (cCfg *CacheCfg) AsMapInterface() map[string]interface{} {
	partitions := make(map[string]interface{}, len(cCfg.Partitions))
	for key, value := range cCfg.Partitions {
		partitions[key] = value.AsMapInterface()
	}

	return map[string]interface{}{
		utils.PartitionsCfg: partitions,
		utils.RplConnsCfg:   cCfg.ReplicationConns,
	}

}
