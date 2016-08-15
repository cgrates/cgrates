package config

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

type CacheParamConfig struct {
	Limit    int
	TTL      time.Duration
	Precache bool
}

func (self *CacheParamConfig) loadFromJsonCfg(jsnCfg *CacheParamJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Limit != nil {
		self.Limit = *jsnCfg.Limit
	}
	if jsnCfg.Ttl != nil {
		if self.TTL, err = utils.ParseDurationWithSecs(*jsnCfg.Ttl); err != nil {
			return err
		}
	}
	if jsnCfg.Precache != nil {
		self.Precache = *jsnCfg.Precache
	}
	return nil
}

type CacheConfig struct {
	Destinations        *CacheParamConfig
	ReverseDestinations *CacheParamConfig
	RatingPlans         *CacheParamConfig
	RatingProfiles      *CacheParamConfig
	Lcr                 *CacheParamConfig
	CdrStats            *CacheParamConfig
	Actions             *CacheParamConfig
	ActionPlans         *CacheParamConfig
	ActionTriggers      *CacheParamConfig
	SharedGroups        *CacheParamConfig
	Aliases             *CacheParamConfig
	ReverseAliases      *CacheParamConfig
}

func (self *CacheConfig) loadFromJsonCfg(jsnCfg *CacheJsonCfg) error {
	if jsnCfg.Destinations != nil {
		self.Destinations = &CacheParamConfig{}
		self.Destinations.loadFromJsonCfg(jsnCfg.Destinations)
	}
	return nil
}
