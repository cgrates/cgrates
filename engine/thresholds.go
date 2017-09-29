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
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type ThresholdProfile struct {
	Tenant             string
	ID                 string
	Filters            []*RequestFilter          // Filters for the request
	ActivationInterval *utils.ActivationInterval // Time when this limit becomes active and expires
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

// Thresholds is a sortable slice of Threshold
type Thresholds []*Threshold

// sort based on Weight
func (ts Thresholds) Sort() {
	sort.Slice(ts, func(i, j int) bool { return ts[i].tPrfl.Weight > ts[j].tPrfl.Weight })
}

func NewThresholdService(dm *DataManager, storeInterval time.Duration,
	statS rpcclient.RpcClientConnection) (tS *ThresholdService, err error) {
	return &ThresholdService{dm: dm, storeInterval: storeInterval,
		statS:      statS,
		stopBackup: make(chan struct{})}, nil
}

// ThresholdService manages Threshold execution and storing them to dataDB
type ThresholdService struct {
	dm            *DataManager
	storeInterval time.Duration
	statS         rpcclient.RpcClientConnection // allows applying filters based on stats
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

// matchingThresholdsForEvent returns ordered list of matching thresholds which are active for an Event
func (tS *ThresholdService) matchingThresholdsForEvent(ev *ThresholdEvent) (ts Thresholds, err error) {
	matchingTs := make(map[string]*Threshold)
	tIDs, err := matchingItemIDsForEvent(ev.Fields, tS.dm.DataDB(), utils.ThresholdsIndex+ev.Tenant)
	if err != nil {
		return nil, err
	}
	lockIDs := utils.PrefixSliceItems(tIDs.Slice(), utils.ThresholdsIndex)
	guardian.Guardian.GuardIDs(config.CgrConfig().LockingTimeout, lockIDs...)
	defer guardian.Guardian.UnguardIDs(lockIDs...)
	for tID := range tIDs {
		tPrfl, err := tS.dm.DataDB().GetThresholdProfile(ev.Tenant, tID, false, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if tPrfl.ActivationInterval != nil &&
			!tPrfl.ActivationInterval.IsActiveAtTime(time.Now()) { // not active
			continue
		}
		passAllFilters := true
		for _, fltr := range tPrfl.Filters {
			if pass, err := fltr.Pass(ev.Fields, "", tS.statS); err != nil {
				return nil, err
			} else if !pass {
				passAllFilters = false
				continue
			}
		}
		if !passAllFilters {
			continue
		}
		t, err := tS.dm.DataDB().GetThreshold(tPrfl.Tenant, tPrfl.ID, false, "")
		if err != nil {
			return nil, err
		}
		if tPrfl.Recurrent && t.dirty == nil {
			t.dirty = utils.BoolPointer(false)
		}
		t.tPrfl = tPrfl
		matchingTs[tPrfl.ID] = t
	}
	// All good, convert from Map to Slice so we can sort
	ts = make(Thresholds, len(matchingTs))
	i := 0
	for _, t := range matchingTs {
		ts[i] = t
		i++
	}
	ts.Sort()
	for i, t := range ts {
		if t.tPrfl.Blocker { // blocker will stop processing
			ts = ts[:i+1]
			break
		}
	}
	return
}
