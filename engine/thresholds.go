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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type ThresholdProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Time when this limit becomes active and expires
	MaxHits            int
	MinHits            int
	MinSleep           time.Duration
	Blocker            bool    // blocker flag to stop processing on filters matched
	Weight             float64 // Weight to sort the thresholds
	ActionIDs          []string
	Async              bool
}

func (tp *ThresholdProfile) TenantID() string {
	return utils.ConcatenatedKey(tp.Tenant, tp.ID)
}

// Threshold is the unit matched by filters
type Threshold struct {
	Tenant string
	ID     string
	Hits   int       // number of hits for this threshold
	Snooze time.Time // prevent threshold to run too early

	tPrfl *ThresholdProfile
	dirty *bool // needs save
}

func (t *Threshold) TenantID() string {
	return utils.ConcatenatedKey(t.Tenant, t.ID)
}

// ProcessEvent processes an ThresholdEvent
// concurrentActions limits the number of simultaneous action sets executed
func (t *Threshold) ProcessEvent(args *ArgsProcessEvent, dm *DataManager) (err error) {
	if t.Snooze.After(time.Now()) { // snoozed, not executing actions
		return
	}
	if t.Hits < t.tPrfl.MinHits { // number of hits was not met, will not execute actions
		return
	}
	if t.tPrfl.MaxHits != -1 && t.Hits > t.tPrfl.MaxHits {
		return
	}
	acnt, _ := args.FieldAsString(utils.Account)
	var acntID string
	if acnt != "" {
		acntID = utils.ConcatenatedKey(args.Tenant, acnt)
	}
	for _, actionSetID := range t.tPrfl.ActionIDs {
		at := &ActionTiming{
			Uuid:      utils.GenUUID(),
			ActionsID: actionSetID,
			ExtraData: args.CGREvent,
		}
		if acntID != "" {
			at.accountIDs = utils.NewStringMap(acntID)
		}
		if t.tPrfl.Async {

			go func() {
				if errExec := at.Execute(nil, nil); errExec != nil {
					utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed executing actions: %s, error: %s", actionSetID, errExec.Error()))
				}
			}()

		} else {
			if errExec := at.Execute(nil, nil); errExec != nil {
				utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed executing actions: %s, error: %s", actionSetID, errExec.Error()))
				err = utils.ErrPartiallyExecuted
			}
		}
	}
	return
}

// Thresholds is a sortable slice of Threshold
type Thresholds []*Threshold

// sort based on Weight
func (ts Thresholds) Sort() {
	sort.Slice(ts, func(i, j int) bool { return ts[i].tPrfl.Weight > ts[j].tPrfl.Weight })
}

func NewThresholdService(dm *DataManager, stringIndexedFields, prefixIndexedFields *[]string, storeInterval time.Duration,
	filterS *FilterS) (tS *ThresholdService, err error) {
	return &ThresholdService{dm: dm,
		stringIndexedFields: stringIndexedFields,
		prefixIndexedFields: prefixIndexedFields,
		storeInterval:       storeInterval,
		filterS:             filterS,
		stopBackup:          make(chan struct{}),
		storedTdIDs:         make(utils.StringMap)}, nil
}

// ThresholdService manages Threshold execution and storing them to dataDB
type ThresholdService struct {
	dm                  *DataManager
	stringIndexedFields *[]string // fields considered when searching for matching thresholds
	prefixIndexedFields *[]string
	storeInterval       time.Duration
	filterS             *FilterS
	stopBackup          chan struct{}
	storedTdIDs         utils.StringMap // keep a record of stats which need saving, map[statsTenantID]bool
	stMux               sync.RWMutex    // protects storedTdIDs
}

// Called to start the service
func (tS *ThresholdService) ListenAndServe(exitChan chan bool) error {
	go tS.runBackup() // start backup loop
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (tS *ThresholdService) Shutdown() error {
	utils.Logger.Info("<ThresholdS> shutdown initialized")
	close(tS.stopBackup)
	tS.storeThresholds()
	utils.Logger.Info("<ThresholdS> shutdown complete")
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
		if tIf, ok := Cache.Get(utils.CacheThresholds, tID); !ok || tIf == nil {
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
	if err = tS.dm.SetThreshold(t); err != nil {
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
func (tS *ThresholdService) matchingThresholdsForEvent(args *ArgsProcessEvent) (ts Thresholds, err error) {
	matchingTs := make(map[string]*Threshold)
	var tIDs []string
	if len(args.ThresholdIDs) != 0 {
		tIDs = args.ThresholdIDs
	} else {
		tIDsMap, err := matchingItemIDsForEvent(args.Event, tS.stringIndexedFields,
			tS.prefixIndexedFields, tS.dm, utils.CacheThresholdFilterIndexes,
			args.Tenant, tS.filterS.cfg.FilterSCfg().IndexedSelects)
		if err != nil {
			return nil, err
		}
		tIDs = tIDsMap.Slice()
	}
	for _, tID := range tIDs {
		tPrfl, err := tS.dm.GetThresholdProfile(args.Tenant, tID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if tPrfl.ActivationInterval != nil && args.Time != nil &&
			!tPrfl.ActivationInterval.IsActiveAtTime(*args.Time) { // not active
			continue
		}
		if pass, err := tS.filterS.Pass(args.Tenant, tPrfl.FilterIDs,
			config.NewNavigableMap(args.Event)); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		t, err := tS.dm.GetThreshold(tPrfl.Tenant, tPrfl.ID, true, true, "")
		if err != nil {
			return nil, err
		}
		if t.dirty == nil || tPrfl.MaxHits == -1 || t.Hits < tPrfl.MaxHits {
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

type ArgsProcessEvent struct {
	ThresholdIDs []string
	utils.CGREvent
}

// processEvent processes a new event, dispatching to matching thresholds
func (tS *ThresholdService) processEvent(args *ArgsProcessEvent) (thresholdsIDs []string, err error) {
	matchTs, err := tS.matchingThresholdsForEvent(args)
	if err != nil {
		return nil, err
	}
	var withErrors bool
	var tIDs []string
	for _, t := range matchTs {
		tIDs = append(tIDs, t.ID)
		t.Hits += 1
		err = t.ProcessEvent(args, tS.dm)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<ThresholdService> threshold: %s, ignoring event: %s, error: %s",
					t.TenantID(), args.CGREvent.TenantID(), err.Error()))
			withErrors = true
			continue
		}
		if t.dirty == nil || t.Hits == t.tPrfl.MaxHits { // one time threshold
			if err = tS.dm.RemoveThreshold(t.Tenant, t.ID, utils.NonTransactional); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<ThresholdService> failed removing non-recurrent threshold: %s, error: %s",
						t.TenantID(), err.Error()))
				withErrors = true
			}
			continue
		}
		t.Snooze = time.Now().Add(t.tPrfl.MinSleep)
		// recurrent threshold
		if tS.storeInterval == -1 {
			tS.StoreThreshold(t)
		} else {
			*t.dirty = true // mark it to be saved
			tS.stMux.Lock()
			tS.storedTdIDs[t.TenantID()] = true
			tS.stMux.Unlock()
		}
	}
	if len(tIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	thresholdsIDs = append(thresholdsIDs, tIDs...)
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1ProcessEvent implements ThresholdService method for processing an Event
func (tS *ThresholdService) V1ProcessEvent(args *ArgsProcessEvent, reply *[]string) (err error) {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	if ids, err := tS.processEvent(args); err != nil {
		return err
	} else {
		*reply = ids
	}
	return
}

// V1GetThresholdsForEvent queries thresholds matching an Event
func (tS *ThresholdService) V1GetThresholdsForEvent(args *ArgsProcessEvent, reply *Thresholds) (err error) {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	var ts Thresholds
	if ts, err = tS.matchingThresholdsForEvent(args); err == nil {
		*reply = ts
	}
	return
}

// V1GetQueueIDs returns list of queueIDs registered for a tenant
func (tS *ThresholdService) V1GetThresholdIDs(tenant string, tIDs *[]string) (err error) {
	prfx := utils.ThresholdPrefix + tenant + ":"
	keys, err := tS.dm.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*tIDs = retIDs
	return
}

// V1GetThreshold retrieves a Threshold
func (tS *ThresholdService) V1GetThreshold(tntID *utils.TenantID, t *Threshold) (err error) {
	if thd, err := tS.dm.GetThreshold(tntID.Tenant, tntID.ID, true, true, ""); err != nil {
		return err
	} else {
		*t = *thd
	}
	return
}
