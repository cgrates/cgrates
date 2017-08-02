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
package stats

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewStatService initializes a StatService
func NewStatService(dataDB engine.DataDB, ms engine.Marshaler, storeInterval time.Duration) (ss *StatService, err error) {
	ss = &StatService{dataDB: dataDB, ms: ms, storeInterval: storeInterval,
		stopStoring: make(chan struct{}), evCache: NewStatsEventCache()}
	sqPrfxs, err := dataDB.GetKeysForPrefix(utils.StatsQueuePrefix)
	if err != nil {
		return nil, err
	}
	ss.stInsts = make(StatsInstances, len(sqPrfxs))
	for i, prfx := range sqPrfxs {
		sq, err := dataDB.GetStatsQueue(prfx[len(utils.StatsQueuePrefix):], false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		var sqSM *engine.SQStoredMetrics
		if sq.Store {
			if sqSM, err = dataDB.GetSQStoredMetrics(sq.ID); err != nil && err != utils.ErrNotFound {
				return nil, err
			}
		}
		if ss.stInsts[i], err = NewStatsInstance(ss.evCache, ss.ms, sq, sqSM); err != nil {
			return nil, err
		}
	}
	ss.stInsts.Sort()
	go ss.dumpStoredMetrics()
	return
}

// StatService builds stats for events
type StatService struct {
	sync.RWMutex
	dataDB        engine.DataDB
	ms            engine.Marshaler
	storeInterval time.Duration
	stopStoring   chan struct{}
	evCache       *StatsEventCache // so we can pass it to queues
	stInsts       StatsInstances   // ordered list of StatsQueues
}

// ListenAndServe loops keeps the service alive
func (ss *StatService) ListenAndServe(exitChan chan bool) error {
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Called to shutdown the service
// ToDo: improve with context, ie, following http implementation
func (ss *StatService) Shutdown() error {
	utils.Logger.Info("<StatS> service shutdown initialized")
	close(ss.stopStoring)
	ss.storeMetrics()
	utils.Logger.Info("<StatS> service shutdown complete")
	return nil
}

// store stores the necessary storedMetrics to dataDB
func (ss *StatService) storeMetrics() {
	for _, si := range ss.stInsts {
		if !si.cfg.Store || !si.dirty { // no need to save
			continue
		}
		if siSM := si.GetStoredMetrics(); siSM != nil {
			if err := ss.dataDB.SetSQStoredMetrics(siSM); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<StatService> failed saving StoredMetrics for QueueID: %s, error: %s",
						si.cfg.ID, err.Error()))
			}
		}
		// randomize the CPU load and give up thread control
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Nanosecond)
	}
	return
}

// dumpStoredMetrics regularly dumps metrics to dataDB
func (ss *StatService) dumpStoredMetrics() {
	for {
		select {
		case <-ss.stopStoring:
			return
		}
		ss.storeMetrics()
		time.Sleep(ss.storeInterval)
	}
}

// processEvent processes a StatsEvent through the queues and caches it when needed
func (ss *StatService) processEvent(ev engine.StatsEvent) (err error) {
	evStatsID := ev.ID()
	if evStatsID == "" { // ID is mandatory
		return errors.New("missing ID field")
	}
	for _, stInst := range ss.stInsts {
		if err := stInst.ProcessEvent(ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<StatService> QueueID: %s, ignoring event with ID: %s, error: %s",
					stInst.cfg.ID, evStatsID, err.Error()))
		}
	}
	return
}

// V1ProcessEvent implements StatV1 method for processing an Event
func (ss *StatService) V1ProcessEvent(ev engine.StatsEvent, reply *string) (err error) {
	if err = ss.processEvent(ev); err == nil {
		*reply = utils.OK
	}
	return
}
