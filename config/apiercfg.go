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

// ApierCfg is the configuration of Apier service
type ApierCfg struct {
	CachesConns     []*RemoteHost // connections towards Cache
	SchedulerConns  []*RemoteHost // connections towards Scheduler
	AttributeSConns []*RemoteHost // connections towards AttributeS
}

func (aCfg *ApierCfg) loadFromJsonCfg(jsnCfg *ApierJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Caches_conns != nil {
		aCfg.CachesConns = make([]*RemoteHost, len(*jsnCfg.Caches_conns))
		for idx, jsnHaCfg := range *jsnCfg.Caches_conns {
			aCfg.CachesConns[idx] = NewDfltRemoteHost()
			aCfg.CachesConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Scheduler_conns != nil {
		aCfg.SchedulerConns = make([]*RemoteHost, len(*jsnCfg.Scheduler_conns))
		for idx, jsnHaCfg := range *jsnCfg.Scheduler_conns {
			aCfg.SchedulerConns[idx] = NewDfltRemoteHost()
			aCfg.SchedulerConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Attributes_conns != nil {
		aCfg.AttributeSConns = make([]*RemoteHost, len(*jsnCfg.Attributes_conns))
		for idx, jsnHaCfg := range *jsnCfg.Attributes_conns {
			aCfg.AttributeSConns[idx] = NewDfltRemoteHost()
			aCfg.AttributeSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}

	return nil
}
