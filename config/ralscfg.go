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
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Rater config section
type RalsCfg struct {
	Enabled                 bool     // start standalone server (no balancer)
	ThresholdSConns         []string // address where to reach ThresholdS config
	StatSConns              []string
	CacheSConns             []string
	RpSubjectPrefixMatching bool // enables prefix matching for the rating profile subject
	RemoveExpired           bool
	MaxComputedUsage        map[string]time.Duration
	BalanceRatingSubject    map[string]string
	MaxIncrements           int
	DynaprepaidActionPlans  []string
}

//loadFromJsonCfg loads Rals config from JsonCfg
func (ralsCfg *RalsCfg) loadFromJsonCfg(jsnRALsCfg *RalsJsonCfg) (err error) {
	if jsnRALsCfg == nil {
		return nil
	}
	if jsnRALsCfg.Enabled != nil {
		ralsCfg.Enabled = *jsnRALsCfg.Enabled
	}
	if jsnRALsCfg.Thresholds_conns != nil {
		ralsCfg.ThresholdSConns = make([]string, len(*jsnRALsCfg.Thresholds_conns))
		for idx, conn := range *jsnRALsCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				ralsCfg.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			} else {
				ralsCfg.ThresholdSConns[idx] = conn
			}
		}
	}
	if jsnRALsCfg.Stats_conns != nil {
		ralsCfg.StatSConns = make([]string, len(*jsnRALsCfg.Stats_conns))
		for idx, conn := range *jsnRALsCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				ralsCfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)
			} else {
				ralsCfg.StatSConns[idx] = conn
			}
		}
	}
	if jsnRALsCfg.CacheS_conns != nil {
		ralsCfg.CacheSConns = make([]string, len(*jsnRALsCfg.CacheS_conns))
		for idx, conn := range *jsnRALsCfg.CacheS_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				ralsCfg.CacheSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
			} else {
				ralsCfg.CacheSConns[idx] = conn
			}
		}
	}
	if jsnRALsCfg.Rp_subject_prefix_matching != nil {
		ralsCfg.RpSubjectPrefixMatching = *jsnRALsCfg.Rp_subject_prefix_matching
	}
	if jsnRALsCfg.Remove_expired != nil {
		ralsCfg.RemoveExpired = *jsnRALsCfg.Remove_expired
	}
	if jsnRALsCfg.Max_computed_usage != nil {
		for k, v := range *jsnRALsCfg.Max_computed_usage {
			if ralsCfg.MaxComputedUsage[k], err = utils.ParseDurationWithNanosecs(v); err != nil {
				return
			}
		}
	}
	if jsnRALsCfg.Max_increments != nil {
		ralsCfg.MaxIncrements = *jsnRALsCfg.Max_increments
	}
	if jsnRALsCfg.Balance_rating_subject != nil {
		for k, v := range *jsnRALsCfg.Balance_rating_subject {
			ralsCfg.BalanceRatingSubject[k] = v
		}
	}
	if jsnRALsCfg.Dynaprepaid_actionplans != nil {
		ralsCfg.DynaprepaidActionPlans = make([]string, len(*jsnRALsCfg.Dynaprepaid_actionplans))
		for i, val := range *jsnRALsCfg.Dynaprepaid_actionplans {
			ralsCfg.DynaprepaidActionPlans[i] = val
		}
	}

	return nil
}

func (ralsCfg *RalsCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:                 ralsCfg.Enabled,
		utils.RpSubjectPrefixMatchingCfg: ralsCfg.RpSubjectPrefixMatching,
		utils.RemoveExpiredCfg:           ralsCfg.RemoveExpired,
		utils.MaxIncrementsCfg:           ralsCfg.MaxIncrements,
		utils.DynaprepaidActionplansCfg:  ralsCfg.DynaprepaidActionPlans,
	}
	if ralsCfg.ThresholdSConns != nil {
		threSholds := make([]string, len(ralsCfg.ThresholdSConns))
		for i, item := range ralsCfg.ThresholdSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				threSholds[i] = strings.ReplaceAll(item, ":*thresholds", utils.EmptyString)
			} else {
				threSholds[i] = item
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = threSholds
	}
	if ralsCfg.StatSConns != nil {
		statS := make([]string, len(ralsCfg.StatSConns))
		for i, item := range ralsCfg.StatSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS) {
				statS[i] = strings.ReplaceAll(item, ":*stats", utils.EmptyString)
			} else {
				statS[i] = item
			}
		}
		initialMP[utils.StatSConnsCfg] = statS
	}
	if ralsCfg.MaxComputedUsage != nil {
		maxComputed := make(map[string]interface{})
		for key, item := range ralsCfg.MaxComputedUsage {
			if key == utils.ANY || key == utils.VOICE {
				maxComputed[key] = item.String()
			} else {
				maxComputed[key] = strconv.Itoa(int(item))
			}
		}
		initialMP[utils.MaxComputedUsageCfg] = maxComputed
	}
	if ralsCfg.CacheSConns != nil {
		cacheSConns := make([]string, len(ralsCfg.CacheSConns))
		for i, item := range ralsCfg.CacheSConns {
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches) {
				cacheSConns[i] = strings.ReplaceAll(item, ":*caches", utils.EmptyString)
			} else {
				cacheSConns[i] = item
			}
		}
		initialMP[utils.CachesConnsCfg] = cacheSConns
	}
	if ralsCfg.BalanceRatingSubject != nil {
		balanceRating := make(map[string]interface{})
		for key, item := range ralsCfg.BalanceRatingSubject {
			balanceRating[key] = item
		}
		initialMP[utils.BalanceRatingSubjectCfg] = balanceRating
	}
	return
}
