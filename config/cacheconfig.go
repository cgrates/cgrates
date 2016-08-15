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
		if err := self.Destinations.loadFromJsonCfg(jsnCfg.Destinations); err != nil {
			return err
		}
	}
	if jsnCfg.Reverse_destinations != nil {
		self.ReverseDestinations = &CacheParamConfig{}
		if err := self.ReverseDestinations.loadFromJsonCfg(jsnCfg.Reverse_destinations); err != nil {
			return err
		}
	}
	if jsnCfg.Rating_plans != nil {
		self.RatingPlans = &CacheParamConfig{}
		if err := self.RatingPlans.loadFromJsonCfg(jsnCfg.Rating_plans); err != nil {
			return err
		}
	}
	if jsnCfg.Rating_profiles != nil {
		self.RatingProfiles = &CacheParamConfig{}
		if err := self.RatingProfiles.loadFromJsonCfg(jsnCfg.Rating_profiles); err != nil {
			return err
		}
	}
	if jsnCfg.Lcr != nil {
		self.Lcr = &CacheParamConfig{}
		if err := self.Lcr.loadFromJsonCfg(jsnCfg.Lcr); err != nil {
			return err
		}
	}
	if jsnCfg.Cdr_stats != nil {
		self.CdrStats = &CacheParamConfig{}
		if err := self.CdrStats.loadFromJsonCfg(jsnCfg.Cdr_stats); err != nil {
			return err
		}
	}
	if jsnCfg.Actions != nil {
		self.Actions = &CacheParamConfig{}
		if err := self.Actions.loadFromJsonCfg(jsnCfg.Actions); err != nil {
			return err
		}
	}
	if jsnCfg.Action_plans != nil {
		self.ActionPlans = &CacheParamConfig{}
		if err := self.ActionPlans.loadFromJsonCfg(jsnCfg.Action_plans); err != nil {
			return err
		}
	}
	if jsnCfg.Action_triggers != nil {
		self.ActionTriggers = &CacheParamConfig{}
		if err := self.ActionTriggers.loadFromJsonCfg(jsnCfg.Action_triggers); err != nil {
			return err
		}
	}
	if jsnCfg.Shared_groups != nil {
		self.SharedGroups = &CacheParamConfig{}
		if err := self.SharedGroups.loadFromJsonCfg(jsnCfg.Shared_groups); err != nil {
			return err
		}
	}
	if jsnCfg.Aliases != nil {
		self.Aliases = &CacheParamConfig{}
		if err := self.Aliases.loadFromJsonCfg(jsnCfg.Aliases); err != nil {
			return err
		}
	}
	if jsnCfg.Reverse_aliases != nil {
		self.ReverseAliases = &CacheParamConfig{}
		if err := self.ReverseAliases.loadFromJsonCfg(jsnCfg.Reverse_aliases); err != nil {
			return err
		}
	}
	return nil
}
