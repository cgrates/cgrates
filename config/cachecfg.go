// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func (cParam *CacheParamCfg) AsMapInterface() map[string]any {
	var TTL string = ""
	if cParam.TTL != 0 {
		TTL = cParam.TTL.String()
	}

	return map[string]any{
		utils.LimitCfg:     cParam.Limit,
		utils.TTLCfg:       TTL,
		utils.StaticTTLCfg: cParam.StaticTTL,
		utils.PrecacheCfg:  cParam.Precache,
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

func (cCfg *CacheCfg) AsMapInterface() map[string]any {
	mp := make(map[string]any, len(*cCfg))
	for key, value := range *cCfg {
		mp[key] = value.AsMapInterface()
	}
	return mp

}
