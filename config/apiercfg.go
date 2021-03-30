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
	"github.com/cgrates/cgrates/utils"
)

// ApierCfg is the configuration of Apier service
type ApierCfg struct {
	Enabled         bool
	CachesConns     []string // connections towards Cache
	ActionsConns    []string // connections towards Scheduler
	AttributeSConns []string // connections towards AttributeS
	EEsConns        []string // connections towards EEs
}

func (aCfg *ApierCfg) loadFromJSONCfg(jsnCfg *ApierJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		aCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Caches_conns != nil {
		aCfg.CachesConns = make([]string, len(*jsnCfg.Caches_conns))
		for idx, conn := range *jsnCfg.Caches_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			aCfg.CachesConns[idx] = conn
			if conn == utils.MetaInternal {
				aCfg.CachesConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
			}
		}
	}
	if jsnCfg.Actions_conns != nil {
		aCfg.ActionsConns = make([]string, len(*jsnCfg.Actions_conns))
		for idx, conn := range *jsnCfg.Actions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				aCfg.ActionsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions)
			}
		}
	}
	if jsnCfg.Attributes_conns != nil {
		aCfg.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, conn := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			aCfg.AttributeSConns[idx] = conn
			if conn == utils.MetaInternal {
				aCfg.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			}
		}
	}
	if jsnCfg.Ees_conns != nil {
		aCfg.EEsConns = make([]string, len(*jsnCfg.Ees_conns))
		for idx, connID := range *jsnCfg.Ees_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			aCfg.EEsConns[idx] = connID
			if connID == utils.MetaInternal {
				aCfg.EEsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)
			}
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (aCfg *ApierCfg) AsMapInterface() (initialMap map[string]interface{}) {
	initialMap = map[string]interface{}{
		utils.EnabledCfg: aCfg.Enabled,
	}
	if aCfg.CachesConns != nil {
		cachesConns := make([]string, len(aCfg.CachesConns))
		for i, item := range aCfg.CachesConns {
			cachesConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches) {
				cachesConns[i] = utils.MetaInternal
			}
		}
		initialMap[utils.CachesConnsCfg] = cachesConns
	}
	if aCfg.ActionsConns != nil {
		actionsConns := make([]string, len(aCfg.ActionsConns))
		for i, item := range aCfg.ActionsConns {
			actionsConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions) {
				actionsConns[i] = utils.MetaInternal
			}
		}
		initialMap[utils.ActionSConnsCfg] = actionsConns
	}
	if aCfg.AttributeSConns != nil {
		attributeSConns := make([]string, len(aCfg.AttributeSConns))
		for i, item := range aCfg.AttributeSConns {
			attributeSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = utils.MetaInternal
			}
		}
		initialMap[utils.AttributeSConnsCfg] = attributeSConns
	}
	if aCfg.EEsConns != nil {
		eesConns := make([]string, len(aCfg.EEsConns))
		for i, item := range aCfg.EEsConns {
			eesConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs) {
				eesConns[i] = utils.MetaInternal
			}
		}
		initialMap[utils.EEsConnsCfg] = eesConns
	}
	return
}

// Clone returns a deep copy of ApierCfg
func (aCfg ApierCfg) Clone() (cln *ApierCfg) {
	cln = &ApierCfg{
		Enabled: aCfg.Enabled,
	}
	if aCfg.CachesConns != nil {
		cln.CachesConns = make([]string, len(aCfg.CachesConns))
		for i, k := range aCfg.CachesConns {
			cln.CachesConns[i] = k
		}
	}
	if aCfg.ActionsConns != nil {
		cln.ActionsConns = make([]string, len(aCfg.ActionsConns))
		for i, k := range aCfg.ActionsConns {
			cln.ActionsConns[i] = k
		}
	}
	if aCfg.AttributeSConns != nil {
		cln.AttributeSConns = make([]string, len(aCfg.AttributeSConns))
		for i, k := range aCfg.AttributeSConns {
			cln.AttributeSConns[i] = k
		}
	}
	if aCfg.EEsConns != nil {
		cln.EEsConns = make([]string, len(aCfg.EEsConns))
		for i, k := range aCfg.EEsConns {
			cln.EEsConns[i] = k
		}
	}
	return
}
