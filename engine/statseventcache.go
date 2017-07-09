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
package engine

import (
	"sync"

	"github.com/cgrates/cgrates/utils"
)

// NewStatsEventCache instantiates a StatsEventCache
func NewStatsEventCache() *StatsEventCache {
	return &StatsEventCache{
		evCacheIdx: make(map[string]utils.StringMap),
		evCache:    make(map[string]StatsEvent)}
}

// StatsEventCache keeps a cache of StatsEvents which are referenced by StatsQueues
type StatsEventCache struct {
	sync.RWMutex
	evCacheIdx map[string]utils.StringMap // index events used in queues, map[eventID]map[queueID]bool
	evCache    map[string]StatsEvent      // cache for the processed events
}

// Cache will cache an event and reference it in the index
func (sec *StatsEventCache) Cache(evID string, ev StatsEvent, queueID string) {
	if utils.IsSliceMember([]string{evID, queueID}, "") {
		return
	}
	sec.Lock()
	if _, hasIt := sec.evCache[evID]; !hasIt {
		sec.evCache[evID] = ev
	}
	sec.evCacheIdx[evID][queueID] = true
	sec.Unlock()
}

func (sec *StatsEventCache) UnCache(evID string, ev StatsEvent, queueID string) {
	sec.Lock()
	if _, hasIt := sec.evCache[evID]; !hasIt {
		return
	}
	delete(sec.evCacheIdx[evID], queueID)
	if len(sec.evCacheIdx[evID]) == 0 {
		delete(sec.evCacheIdx, evID)
		delete(sec.evCache, evID)
	}
	sec.Unlock()
}

// GetEvent returns the event based on ID
func (sec *StatsEventCache) GetEvent(evID string) StatsEvent {
	sec.RLock()
	defer sec.RUnlock()
	return sec.evCache[evID]
}
