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

// DispatcherSCfg is the configuration of dispatcher service
type DispatcherSCfg struct {
	Enabled             bool
	RALsConns           []*HaPoolConfig
	ResSConns           []*HaPoolConfig
	ThreshSConns        []*HaPoolConfig
	StatSConns          []*HaPoolConfig
	SupplSConns         []*HaPoolConfig
	AttrSConns          []*HaPoolConfig
	SessionSConns       []*HaPoolConfig
	ChargerSConns       []*HaPoolConfig
	DispatchingStrategy string
}

func (dps *DispatcherSCfg) loadFromJsonCfg(jsnCfg *DispatcherSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		dps.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Rals_conns != nil {
		dps.RALsConns = make([]*HaPoolConfig, len(*jsnCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCfg.Rals_conns {
			dps.RALsConns[idx] = NewDfltHaPoolConfig()
			dps.RALsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Resources_conns != nil {
		dps.ResSConns = make([]*HaPoolConfig, len(*jsnCfg.Resources_conns))
		for idx, jsnHaCfg := range *jsnCfg.Resources_conns {
			dps.ResSConns[idx] = NewDfltHaPoolConfig()
			dps.ResSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Thresholds_conns != nil {
		dps.ThreshSConns = make([]*HaPoolConfig, len(*jsnCfg.Thresholds_conns))
		for idx, jsnHaCfg := range *jsnCfg.Thresholds_conns {
			dps.ThreshSConns[idx] = NewDfltHaPoolConfig()
			dps.ThreshSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Stats_conns != nil {
		dps.StatSConns = make([]*HaPoolConfig, len(*jsnCfg.Stats_conns))
		for idx, jsnHaCfg := range *jsnCfg.Stats_conns {
			dps.StatSConns[idx] = NewDfltHaPoolConfig()
			dps.StatSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Suppliers_conns != nil {
		dps.SupplSConns = make([]*HaPoolConfig, len(*jsnCfg.Suppliers_conns))
		for idx, jsnHaCfg := range *jsnCfg.Suppliers_conns {
			dps.SupplSConns[idx] = NewDfltHaPoolConfig()
			dps.SupplSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Attributes_conns != nil {
		dps.AttrSConns = make([]*HaPoolConfig, len(*jsnCfg.Attributes_conns))
		for idx, jsnHaCfg := range *jsnCfg.Attributes_conns {
			dps.AttrSConns[idx] = NewDfltHaPoolConfig()
			dps.AttrSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Sessions_conns != nil {
		dps.SessionSConns = make([]*HaPoolConfig, len(*jsnCfg.Sessions_conns))
		for idx, jsnHaCfg := range *jsnCfg.Sessions_conns {
			dps.SessionSConns[idx] = NewDfltHaPoolConfig()
			dps.SessionSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Chargers_conns != nil {
		dps.ChargerSConns = make([]*HaPoolConfig, len(*jsnCfg.Chargers_conns))
		for idx, jsnHaCfg := range *jsnCfg.Chargers_conns {
			dps.ChargerSConns[idx] = NewDfltHaPoolConfig()
			dps.ChargerSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Dispatching_strategy != nil {
		dps.DispatchingStrategy = *jsnCfg.Dispatching_strategy
	}
	return nil
}
