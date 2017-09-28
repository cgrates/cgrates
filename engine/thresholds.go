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
	//"fmt"
	//"math/rand"
	"sync"
	"time"

	//"github.com/cgrates/cgrates/cache"
	//"github.com/cgrates/cgrates/config"
	//"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

type ThresholdProfile struct {
	Tenant             string
	ID                 string
	Filters            []*RequestFilter          // Filters for the request
	ActivationInterval *utils.ActivationInterval // Time when this limit becomes active and expires
	MinItems           int                       // number of items agregated for the threshold to match
	Recurrent          bool
	MinSleep           time.Duration
	Blocker            bool    // blocker flag to stop processing on filters matched
	Weight             float64 // Weight to sort the thresholds
	ActionIDs          []string
}

func (tp *ThresholdProfile) TenantID() string {
	return utils.ConcatenatedKey(tp.Tenant, tp.ID)
}

// ThresholdEvent is an event processed by ThresholdService
type ThresholdEvent struct {
	Tenant string
	ID     string
	Fields map[string]interface{}
}

// Threshold is the unit matched by filters
// It's WakeupTime is stored on demand
type Threshold struct {
	Tenant       string
	ID           string
	LastExecuted time.Time
	WakeupTime   time.Time // prevent threshold to run too early

	tPrfl *ThresholdProfile
	dirty *bool // needs save
}

func (t *Threshold) TenantID() string {
	return utils.ConcatenatedKey(t.Tenant, t.ID)
}

func NewThresholdService(dm *DataManager, storeInterval time.Duration) (tS *ThresholdService, err error) {
	return &ThresholdService{dm: dm, storeInterval: storeInterval,
		stopBackup: make(chan struct{})}, nil
}

// ThresholdService manages Threshold execution and storing them to dataDB
type ThresholdService struct {
	dm            *DataManager
	storeInterval time.Duration
	stopBackup    chan struct{}
	storedTdIDs   utils.StringMap // keep a record of stats which need saving, map[statsTenantID]bool
	stMux         sync.RWMutex    // protects storedTdIDs
}

// Called to start the service
func (tS *ThresholdService) ListenAndServe(exitChan chan bool) error {
	//go tS.runBackup() // start backup loop
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

/*
// backup will regularly store resources changed to dataDB
func (tS *ThresholdService) runBackup() {
	if tS.storeInterval <= 0 {
		return
	}
	for {
		select {
		case <-tS.stopBackup:
			return
		default:
		}
		tS.storeThresholds()
		time.Sleep(tS.storeInterval)
	}
}

// storeThresholds represents one task of complete backup
func (tS *ThresholdService) storeThresholds() {
	var failedTdIDs []string
	for { // don't stop untill we store all dirty resources
		tS.stMux.Lock()
		tID := tS.storedTdIDs.GetOne()
		if tID != "" {
			delete(tS.storedTdIDs, tID)
		}
		tS.stMux.Unlock()
		if tID == "" {
			break // no more keys, backup completed
		}
		if tIf, ok := cache.Get(utils.ThresholdsPrefix + tID); !ok || tIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed retrieving from cache resource with ID: %s", tID))
		} else if err := tS.StoreThreshold(tIf.(*Threshold)); err != nil {
			failedTdIDs = append(failedTdIDs, tID) // record failure so we can schedule it for next backup
		}
		// randomize the CPU load and give up thread control
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Nanosecond)
	}
	if len(failedTdIDs) != 0 { // there were errors on save, schedule the keys for next backup
		tS.stMux.Lock()
		for _, tID := range failedTdIDs {
			tS.storedTdIDs[tID] = true
		}
		tS.stMux.Unlock()
	}
}

// StoreThreshold stores the threshold in DB and corrects dirty flag
func (tS *ThresholdService) StoreThreshold(t *Threshold) (err error) {
	if t.dirty == nil || !*t.dirty {
		return
	}
	guardian.Guardian.GuardIDs(config.CgrConfig().LockingTimeout, utils.ThresholdsPrefix+t.TenantID())
	defer guardian.Guardian.UnguardIDs(utils.ThresholdsPrefix + t.TenantID())
	if err = tS.dm.DataDB().SetThreshold(t); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<ThresholdS> failed saving Threshold with tenant: %s and ID: %s, error: %s",
				t.Tenant, t.ID, err.Error()))
		return
	} else {
		*t.dirty = false
	}
	return
}
*/
