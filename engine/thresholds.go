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
	"reflect"
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
	Async              bool
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

func (te *ThresholdEvent) TenantID() string {
	return utils.ConcatenatedKey(te.Tenant, te.ID)
}

func (te *ThresholdEvent) Account() (acnt string, err error) {
	acntIf, has := te.Fields[utils.ACCOUNT]
	if !has {
		return "", utils.ErrNotFound
	}
	var canCast bool
	if acnt, canCast = acntIf.(string); !canCast {
		return "", fmt.Errorf("field %s is not string", utils.ACCOUNT)
	}
	return
}

func (te *ThresholdEvent) FilterableEvent(fltredFields []string) (fEv map[string]interface{}) {
	fEv = make(map[string]interface{})
	if len(fltredFields) == 0 {
		i := 0
		fltredFields = make([]string, len(te.Fields))
		for k := range te.Fields {
			fltredFields[i] = k
			i++
		}
	}
	for _, fltrFld := range fltredFields {
		fldVal, has := te.Fields[fltrFld]
		if !has {
			continue // the field does not exist in map, ignore it
		}
		valOf := reflect.ValueOf(fldVal)
		if valOf.Kind() == reflect.String {
			fEv[fltrFld] = utils.StringToInterface(valOf.String()) // attempt converting from string to comparable interface
		} else {
			fEv[fltrFld] = fldVal
		}
	}
	return
}

// Threshold is the unit matched by filters
type Threshold struct {
	Tenant string
	ID     string
	Snooze time.Time // prevent threshold to run too early

	tPrfl *ThresholdProfile
	dirty *bool // needs save
}

func (t *Threshold) TenantID() string {
	return utils.ConcatenatedKey(t.Tenant, t.ID)
}

// ProcessEvent processes an ThresholdEvent
// concurrentActions limits the number of simultaneous action sets executed
func (t *Threshold) ProcessEvent(ev *ThresholdEvent, dm *DataManager) (err error) {
	if t.Snooze.After(time.Now()) { // ignore the event
		return
	}
	acnt, _ := ev.Account()
	var acntID string
	if acnt != "" {
		acntID = utils.ConcatenatedKey(ev.Tenant, acnt)
	}
	for _, actionSetID := range t.tPrfl.ActionIDs {
		at := &ActionTiming{
			Uuid:      utils.GenUUID(),
			ActionsID: actionSetID,
		}
		if acntID != "" {
			at.accountIDs = utils.NewStringMap(acntID)
		}
		if errExec := at.Execute(nil, nil); errExec != nil {
			utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed executing actions: %s, error: %s", actionSetID, errExec.Error()))
			err = utils.ErrPartiallyExecuted
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

func NewThresholdService(dm *DataManager, filteredFields []string, storeInterval time.Duration,
	statS *rpcclient.RpcClientPool) (tS *ThresholdService, err error) {
	return &ThresholdService{dm: dm,
		filteredFields: filteredFields,
		storeInterval:  storeInterval,
		statS:          statS,
		stopBackup:     make(chan struct{}),
		storedTdIDs:    make(utils.StringMap)}, nil
}

// ThresholdService manages Threshold execution and storing them to dataDB
type ThresholdService struct {
	dm             *DataManager
	filteredFields []string // fields considered when searching for matching thresholds
	storeInterval  time.Duration
	statS          *rpcclient.RpcClientPool // allows applying filters based on stats
	stopBackup     chan struct{}
	storedTdIDs    utils.StringMap // keep a record of stats which need saving, map[statsTenantID]bool
	stMux          sync.RWMutex    // protects storedTdIDs
}

// Called to start the service
func (tS *ThresholdService) ListenAndServe(exitChan chan bool) error {
	//go tS.runBackup() // start backup loop
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
		if tIf, ok := cache.Get(utils.ThresholdPrefix + tID); !ok || tIf == nil {
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
	guardian.Guardian.GuardIDs(config.CgrConfig().LockingTimeout, utils.ThresholdPrefix+t.TenantID())
	defer guardian.Guardian.UnguardIDs(utils.ThresholdPrefix + t.TenantID())
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
	tIDs, err := matchingItemIDsForEvent(ev.Fields, tS.dm, utils.ThresholdsIndex+ev.Tenant)
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
			if pass, err := fltr.Pass(ev.FilterableEvent(nil), "", tS.statS); err != nil {
				return nil, err
			} else if !pass {
				passAllFilters = false
				continue
			}
		}
		if !passAllFilters {
			continue
		}
		lockThreshold := utils.ThresholdPrefix + tPrfl.TenantID()
		guardian.Guardian.GuardIDs(config.CgrConfig().LockingTimeout, lockThreshold)
		t, err := tS.dm.DataDB().GetThreshold(tPrfl.Tenant, tPrfl.ID, false, "")
		if err != nil {
			guardian.Guardian.UnguardIDs(lockThreshold)
			return nil, err
		}
		if tPrfl.Recurrent && t.dirty == nil {
			t.dirty = utils.BoolPointer(false)
		}
		t.tPrfl = tPrfl
		guardian.Guardian.UnguardIDs(lockThreshold)
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

// processEvent processes a new event, dispatching to matching thresholds
func (tS *ThresholdService) processEvent(ev *ThresholdEvent) (hits int, err error) {
	matchTs, err := tS.matchingThresholdsForEvent(ev)
	if err != nil {
		return 0, err
	}
	hits = len(matchTs)
	var withErrors bool
	for _, t := range matchTs {
		err = t.ProcessEvent(ev, tS.dm)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<ThresholdService> threshold: %s, ignoring event: %s, error: %s",
					t.TenantID(), ev.TenantID(), err.Error()))
			withErrors = true
			continue
		}
		if t.dirty == nil { // one time threshold
			lockThreshold := utils.ThresholdPrefix + t.TenantID()
			guardian.Guardian.GuardIDs(config.CgrConfig().LockingTimeout, lockThreshold)
			if err = tS.dm.DataDB().RemoveThreshold(t.Tenant, t.ID, utils.NonTransactional); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<ThresholdService> failed removing non-recurrent threshold: %s, error: %s",
						t.TenantID(), err.Error()))
				withErrors = true

			}
			guardian.Guardian.UnguardIDs(lockThreshold)
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
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1ProcessEvent implements ThresholdService method for processing an Event
func (tS *ThresholdService) V1ProcessEvent(ev *ThresholdEvent, reply *int) (err error) {
	if missing := utils.MissingStructFields(ev, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if hits, err := tS.processEvent(ev); err != nil {
		return err
	} else {
		*reply = hits
	}
	return
}

// V1GetThresholdsForEvent queries thresholds matching an Event
func (tS *ThresholdService) V1GetThresholdsForEvent(ev *ThresholdEvent, reply *Thresholds) (err error) {
	if missing := utils.MissingStructFields(ev, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var ts Thresholds
	if ts, err = tS.matchingThresholdsForEvent(ev); err == nil {
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
	if thd, err := tS.dm.DataDB().GetThreshold(tntID.Tenant, tntID.ID, false, ""); err != nil {
		return err
	} else {
		*t = *thd
	}
	return
}
