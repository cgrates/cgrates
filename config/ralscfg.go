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
	"time"

	"github.com/cgrates/cgrates/utils"
)

// RalsCfg is rater config section
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
}

// loadFromJSONCfg loads Rals config from JsonCfg
func (ralsCfg *RalsCfg) loadFromJSONCfg(jsnRALsCfg *RalsJsonCfg) (err error) {
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
			ralsCfg.ThresholdSConns[idx] = conn
			if conn == utils.MetaInternal {
				ralsCfg.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			}
		}
	}
	if jsnRALsCfg.Stats_conns != nil {
		ralsCfg.StatSConns = make([]string, len(*jsnRALsCfg.Stats_conns))
		for idx, conn := range *jsnRALsCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			ralsCfg.StatSConns[idx] = conn
			if conn == utils.MetaInternal {
				ralsCfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
	}
	if jsnRALsCfg.CacheS_conns != nil {
		ralsCfg.CacheSConns = make([]string, len(*jsnRALsCfg.CacheS_conns))
		for idx, conn := range *jsnRALsCfg.CacheS_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			ralsCfg.CacheSConns[idx] = conn
			if conn == utils.MetaInternal {
				ralsCfg.CacheSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
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

	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (ralsCfg *RalsCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:                 ralsCfg.Enabled,
		utils.RpSubjectPrefixMatchingCfg: ralsCfg.RpSubjectPrefixMatching,
		utils.RemoveExpiredCfg:           ralsCfg.RemoveExpired,
		utils.MaxIncrementsCfg:           ralsCfg.MaxIncrements,
	}
	if ralsCfg.ThresholdSConns != nil {
		threSholds := make([]string, len(ralsCfg.ThresholdSConns))
		for i, item := range ralsCfg.ThresholdSConns {
			threSholds[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				threSholds[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = threSholds
	}
	if ralsCfg.StatSConns != nil {
		statS := make([]string, len(ralsCfg.StatSConns))
		for i, item := range ralsCfg.StatSConns {
			statS[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				statS[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statS
	}
	if ralsCfg.CacheSConns != nil {
		cacheSConns := make([]string, len(ralsCfg.CacheSConns))
		for i, item := range ralsCfg.CacheSConns {
			cacheSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches) {
				cacheSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.CachesConnsCfg] = cacheSConns
	}
	maxComputed := make(map[string]interface{})
	for key, item := range ralsCfg.MaxComputedUsage {
		if key == utils.MetaAny || key == utils.MetaVoice {
			maxComputed[key] = item.String()
		} else {
			maxComputed[key] = strconv.Itoa(int(item))
		}
	}
	initialMP[utils.MaxComputedUsageCfg] = maxComputed
	balanceRatSubj := make(map[string]string)
	for k, v := range ralsCfg.BalanceRatingSubject {
		balanceRatSubj[k] = v
	}
	initialMP[utils.BalanceRatingSubjectCfg] = balanceRatSubj
	return
}

// Clone returns a deep copy of RalsCfg
func (ralsCfg RalsCfg) Clone() (cln *RalsCfg) {
	cln = &RalsCfg{
		Enabled:                 ralsCfg.Enabled,
		RpSubjectPrefixMatching: ralsCfg.RpSubjectPrefixMatching,
		RemoveExpired:           ralsCfg.RemoveExpired,
		MaxIncrements:           ralsCfg.MaxIncrements,

		MaxComputedUsage:     make(map[string]time.Duration),
		BalanceRatingSubject: make(map[string]string),
	}
	if ralsCfg.ThresholdSConns != nil {
		cln.ThresholdSConns = make([]string, len(ralsCfg.ThresholdSConns))
		for i, con := range ralsCfg.ThresholdSConns {
			cln.ThresholdSConns[i] = con
		}
	}
	if ralsCfg.StatSConns != nil {
		cln.StatSConns = make([]string, len(ralsCfg.StatSConns))
		for i, con := range ralsCfg.StatSConns {
			cln.StatSConns[i] = con
		}
	}
	if ralsCfg.CacheSConns != nil {
		cln.CacheSConns = make([]string, len(ralsCfg.CacheSConns))
		for i, con := range ralsCfg.CacheSConns {
			cln.CacheSConns[i] = con
		}
	}

	for k, u := range ralsCfg.MaxComputedUsage {
		cln.MaxComputedUsage[k] = u
	}
	for k, r := range ralsCfg.BalanceRatingSubject {
		cln.BalanceRatingSubject[k] = r
	}
	return
}
