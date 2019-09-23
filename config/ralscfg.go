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
)

// Rater config section
type RalsCfg struct {
	Enabled                 bool          // start standalone server (no balancer)
	ThresholdSConns         []*RemoteHost // address where to reach ThresholdS config
	StatSConns              []*RemoteHost
	RpSubjectPrefixMatching bool // enables prefix matching for the rating profile subject
	RemoveExpired           bool
	MaxComputedUsage        map[string]time.Duration
	BalanceRatingSubject    map[string]string
	RALsMaxIncrements        int
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
		ralsCfg.ThresholdSConns = make([]*RemoteHost, len(*jsnRALsCfg.Thresholds_conns))
		for idx, jsnHaCfg := range *jsnRALsCfg.Thresholds_conns {
			ralsCfg.ThresholdSConns[idx] = NewDfltRemoteHost()
			ralsCfg.ThresholdSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnRALsCfg.Stats_conns != nil {
		ralsCfg.StatSConns = make([]*RemoteHost, len(*jsnRALsCfg.Stats_conns))
		for idx, jsnHaCfg := range *jsnRALsCfg.Stats_conns {
			ralsCfg.StatSConns[idx] = NewDfltRemoteHost()
			ralsCfg.StatSConns[idx].loadFromJsonCfg(jsnHaCfg)
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
		ralsCfg.RALsMaxIncrements = *jsnRALsCfg.Max_increments
	}
	if jsnRALsCfg.Balance_rating_subject != nil {
		for k, v := range *jsnRALsCfg.Balance_rating_subject {
			ralsCfg.BalanceRatingSubject[k] = v
		}
	}

	return nil
}
