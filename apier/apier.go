/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package apier

import (
	"errors"
	"fmt"
	"path"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
)

const (
	OK = "OK"
)

type ApierV1 struct {
	StorDb    engine.LoadStorage
	RatingDb  engine.RatingStorage
	AccountDb engine.AccountingStorage
	CdrDb     engine.CdrStorage
	Sched     *scheduler.Scheduler
	Config    *config.CGRConfig
}

func (self *ApierV1) GetDestination(dstId string, reply *engine.Destination) error {
	if dst, err := self.RatingDb.GetDestination(dstId); err != nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *dst
	}
	return nil
}

func (self *ApierV1) GetRatingPlan(rplnId string, reply *engine.RatingPlan) error {
	if rpln, err := self.RatingDb.GetRatingPlan(rplnId, false); err != nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *rpln
	}
	return nil
}

type AttrGetUserBalance struct {
	Tenant    string
	Account   string
	BalanceId string
	Direction string
}

// Get balance
func (self *ApierV1) GetUserBalance(attr *AttrGetUserBalance, reply *engine.UserBalance) error {
	tag := fmt.Sprintf("%s:%s:%s", attr.Direction, attr.Tenant, attr.Account)
	userBalance, err := self.AccountDb.GetUserBalance(tag)
	if err != nil {
		return err
	}

	*reply = *userBalance
	return nil
}

type AttrAddBalance struct {
	Tenant    string
	Account   string
	BalanceId string
	Direction string
	Value     float64
	//ExpirationDate time.Time
	//RateSubject    string
	DestinationId string
	Weight        float64
	Overwrite     bool // When true it will reset if the balance is already there
}

func (self *ApierV1) AddBalance(attr *AttrAddBalance, reply *string) error {
	tag := fmt.Sprintf("%s:%s:%s", attr.Direction, attr.Tenant, attr.Account)
	if _, err := self.AccountDb.GetUserBalance(tag); err != nil {
		// create user balance if not exists
		ub := &engine.UserBalance{
			Id: tag,
		}
		if err := self.AccountDb.SetUserBalance(ub); err != nil {
			*reply = err.Error()
			return err
		}
	}
	at := &engine.ActionTiming{
		UserBalanceIds: []string{tag},
	}

	if attr.Direction == "" {
		attr.Direction = engine.OUTBOUND
	}
	aType := engine.TOPUP
	if attr.Overwrite {
		aType = engine.TOPUP_RESET
	}
	at.SetActions(engine.Actions{
		&engine.Action{
			ActionType: aType,
			BalanceId:  attr.BalanceId,
			Direction:  attr.Direction,
			Balance: &engine.Balance{
				Value: attr.Value,
				//ExpirationDate: attr.ExpirationDate,
				//RateSubject:    attr.RateSubject,
				DestinationId: attr.DestinationId,
				Weight:        attr.Weight,
			},
		},
	})
	if err := at.Execute(); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

type AttrExecuteAction struct {
	Direction string
	Tenant    string
	Account   string
	ActionsId string
}

func (self *ApierV1) ExecuteAction(attr *AttrExecuteAction, reply *string) error {
	tag := fmt.Sprintf("%s:%s:%s", attr.Direction, attr.Tenant, attr.Account)
	at := &engine.ActionTiming{
		UserBalanceIds: []string{tag},
		ActionsId:      attr.ActionsId,
	}

	if err := at.Execute(); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

type AttrLoadRatingPlan struct {
	TPid         string
	RatingPlanId string
}

// Process dependencies and load a specific rating plan from storDb into dataDb.
func (self *ApierV1) LoadRatingPlan(attrs AttrLoadRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)
	if loaded, err := dbReader.LoadRatingPlanByTag(attrs.RatingPlanId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if !loaded {
		return errors.New("NOT_FOUND")
	}
	*reply = OK
	return nil
}

// Process dependencies and load a specific rating profile from storDb into dataDb.
func (self *ApierV1) LoadRatingProfile(attrs utils.TPRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LoadId", "Tenant", "TOR", "Direction", "Subject"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)
	if err := dbReader.LoadRatingProfileFiltered(&attrs); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = OK
	return nil
}

type AttrSetRatingProfile struct {
	Tenant                string                      // Tenant's Id
	TOR                   string                      // TypeOfRecord
	Direction             string                      // Traffic direction, OUT is the only one supported for now
	Subject               string                      // Rating subject, usually the same as account
	Overwrite             bool                        // Overwrite if exists
	RatingPlanActivations []*utils.TPRatingActivation // Activate rate profiles at specific time
}

// Sets a specific rating profile working with data directly in the RatingDb without involving storDb
func (self *ApierV1) SetRatingProfile(attrs AttrSetRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "TOR", "Direction", "Subject", "RatingPlanActivations"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	for _, rpa := range attrs.RatingPlanActivations {
		if missing := utils.MissingStructFields(rpa, []string{"ActivationTime", "RatingPlanId"}); len(missing) != 0 {
			return fmt.Errorf("%s:RatingPlanActivation:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
		}
	}
	tpRpf := utils.TPRatingProfile{Tenant: attrs.Tenant, TOR: attrs.TOR, Direction: attrs.Direction, Subject: attrs.Subject}
	keyId := tpRpf.KeyId()
	if !attrs.Overwrite {
		if exists, err := self.RatingDb.HasData(engine.RATING_PROFILE_PREFIX, keyId); err != nil {
			return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
		} else if exists {
			return errors.New(utils.ERR_EXISTS)
		}
	}
	rpfl := &engine.RatingProfile{Id: keyId, RatingPlanActivations: make(engine.RatingPlanActivations, len(attrs.RatingPlanActivations))}
	for idx, ra := range attrs.RatingPlanActivations {
		at, err := utils.ParseDate(ra.ActivationTime)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("%s:Cannot parse activation time from %v", utils.ERR_SERVER_ERROR, ra.ActivationTime))
		}
		if exists, err := self.RatingDb.HasData(engine.RATING_PLAN_PREFIX, ra.RatingPlanId); err != nil {
			return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
		} else if !exists {
			return fmt.Errorf(fmt.Sprintf("%s:RatingPlanId:%s", utils.ERR_NOT_FOUND, ra.RatingPlanId))
		}
		rpfl.RatingPlanActivations[idx] = &engine.RatingPlanActivation{ActivationTime: at, RatingPlanId: ra.RatingPlanId,
			FallbackKeys: utils.FallbackSubjKeys(tpRpf.Direction, tpRpf.Tenant, tpRpf.TOR, ra.FallbackSubjects)}
	}
	if err := self.RatingDb.SetRatingProfile(rpfl); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = OK
	return nil
}

type AttrSetActions struct {
	ActionsId string            // Actions id
	Overwrite bool              // If previously defined, will be overwritten
	Actions   []*utils.TPAction // Set of actions this Actions profile will perform
}

func (self *ApierV1) SetActions(attrs AttrSetActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"ActionsId", "Actions"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	for _, action := range attrs.Actions {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ERR_MANDATORY_IE_MISSING, action.Identifier, missing)
		}
	}
	if !attrs.Overwrite {
		if exists, err := self.AccountDb.HasData(engine.ACTION_PREFIX, attrs.ActionsId); err != nil {
			return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
		} else if exists {
			return errors.New(utils.ERR_EXISTS)
		}
	}
	storeActions := make(engine.Actions, len(attrs.Actions))
	for idx, apiAct := range attrs.Actions {
		a := &engine.Action{
			Id:               utils.GenUUID(),
			ActionType:       apiAct.Identifier,
			BalanceId:        apiAct.BalanceType,
			Direction:        apiAct.Direction,
			Weight:           apiAct.Weight,
			ExpirationString: apiAct.ExpiryTime,
			ExtraParameters:  apiAct.ExtraParameters,
			Balance: &engine.Balance{
				Uuid:          utils.GenUUID(),
				Value:         apiAct.Units,
				Weight:        apiAct.BalanceWeight,
				DestinationId: apiAct.DestinationId,
				RateSubject:   apiAct.RatingSubject,
			},
		}
		storeActions[idx] = a
	}
	if err := self.AccountDb.SetActions(attrs.ActionsId, storeActions); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = OK
	return nil
}

type AttrSetActionPlan struct {
	Id              string             // Profile id
	ActionPlan      []*ApiActionTiming // Set of actions this Actions profile will perform
	Overwrite       bool               // If previously defined, will be overwritten
	ReloadScheduler bool               // Enables automatic reload of the scheduler (eg: useful when adding a single action timing)
}

type ApiActionTiming struct {
	ActionsId string  // Actions id
	Years     string  // semicolon separated list of years this timing is valid on, *any or empty supported
	Months    string  // semicolon separated list of months this timing is valid on, *any or empty supported
	MonthDays string  // semicolon separated list of month's days this timing is valid on, *any or empty supported
	WeekDays  string  // semicolon separated list of week day names this timing is valid on *any or empty supported
	Time      string  // String representing the time this timing starts on, *asap supported
	Weight    float64 // Binding's weight
}

func (self *ApierV1) SetActionPlan(attrs AttrSetActionPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Id", "ActionPlan"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	for _, at := range attrs.ActionPlan {
		requiredFields := []string{"ActionsId", "Time", "Weight"}
		if missing := utils.MissingStructFields(at, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ERR_MANDATORY_IE_MISSING, at.ActionsId, missing)
		}
	}
	if !attrs.Overwrite {
		if exists, err := self.AccountDb.HasData(engine.ACTION_TIMING_PREFIX, attrs.Id); err != nil {
			return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
		} else if exists {
			return errors.New(utils.ERR_EXISTS)
		}
	}
	storeAtms := make(engine.ActionPlan, len(attrs.ActionPlan))
	for idx, apiAtm := range attrs.ActionPlan {
		if exists, err := self.AccountDb.HasData(engine.ACTION_PREFIX, apiAtm.ActionsId); err != nil {
			return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
		} else if !exists {
			return fmt.Errorf("%s:%s", utils.ERR_BROKEN_REFERENCE, err.Error())
		}
		timing := new(engine.RITiming)
		timing.Years.Parse(apiAtm.Years, ";")
		timing.Months.Parse(apiAtm.Months, ";")
		timing.MonthDays.Parse(apiAtm.MonthDays, ";")
		timing.WeekDays.Parse(apiAtm.WeekDays, ";")
		timing.StartTime = apiAtm.Time
		at := &engine.ActionTiming{
			Id:        utils.GenUUID(),
			Tag:       attrs.Id,
			Weight:    apiAtm.Weight,
			Timing:    &engine.RateInterval{Timing: timing},
			ActionsId: apiAtm.ActionsId,
		}
		storeAtms[idx] = at
	}
	if err := self.AccountDb.SetActionTimings(attrs.Id, storeAtms); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if attrs.ReloadScheduler {
		if self.Sched == nil {
			return errors.New("SCHEDULER_NOT_ENABLED")
		}
		self.Sched.LoadActionTimings(self.AccountDb)
		self.Sched.Restart()
	}
	*reply = OK
	return nil
}

type AttrAddActionTrigger struct {
	Tenant         string
	Account        string
	Direction      string
	BalanceId      string
	ThresholdType  string
	ThresholdValue float64
	DestinationId  string
	Weight         float64
	ActionsId      string
}

func (self *ApierV1) AddTriggeredAction(attr AttrAddActionTrigger, reply *string) error {
	if attr.Direction == "" {
		attr.Direction = engine.OUTBOUND
	}

	at := &engine.ActionTrigger{
		Id:             utils.GenUUID(),
		BalanceId:      attr.BalanceId,
		Direction:      attr.Direction,
		ThresholdType:  attr.ThresholdType,
		ThresholdValue: attr.ThresholdValue,
		DestinationId:  attr.DestinationId,
		Weight:         attr.Weight,
		ActionsId:      attr.ActionsId,
		Executed:       false,
	}

	tag := utils.BalanceKey(attr.Tenant, attr.Account, attr.Direction)
	_, err := engine.AccLock.Guard(tag, func() (float64, error) {
		userBalance, err := self.AccountDb.GetUserBalance(tag)
		if err != nil {
			return 0, err
		}

		userBalance.ActionTriggers = append(userBalance.ActionTriggers, at)

		if err = self.AccountDb.SetUserBalance(userBalance); err != nil {
			return 0, err
		}
		return 0, nil
	})
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

// Process dependencies and load a specific AccountActions profile from storDb into dataDb.
func (self *ApierV1) LoadAccountActions(attrs utils.TPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LoadId", "Tenant", "Account", "Direction"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)

	if _, err := engine.AccLock.Guard(attrs.KeyId(), func() (float64, error) {
		if err := dbReader.LoadAccountActionsFiltered(&attrs); err != nil {
			return 0, err
		}
		return 0, nil
	}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	// ToDo: Get the action keys loaded by dbReader so we reload only these in cache
	// Need to do it before scheduler otherwise actions to run will be unknown
	if err := self.AccountDb.CacheAccounting(nil, nil); err != nil {
		return err
	}
	if self.Sched != nil {
		self.Sched.LoadActionTimings(self.AccountDb)
		self.Sched.Restart()
	}
	*reply = OK
	return nil
}

func (self *ApierV1) ReloadScheduler(input string, reply *string) error {
	if self.Sched == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	}
	self.Sched.LoadActionTimings(self.AccountDb)
	self.Sched.Restart()
	*reply = OK
	return nil

}

func (self *ApierV1) ReloadCache(attrs utils.ApiReloadCache, reply *string) error {
	var dstKeys, rpKeys, rpfKeys, actKeys, shgKeys []string
	if len(attrs.DestinationIds) > 0 {
		dstKeys = make([]string, len(attrs.DestinationIds))
		for idx, dId := range attrs.DestinationIds {
			dstKeys[idx] = engine.DESTINATION_PREFIX + dId // Cache expects them as redis keys
		}
	}
	if len(attrs.RatingPlanIds) > 0 {
		rpKeys = make([]string, len(attrs.RatingPlanIds))
		for idx, rpId := range attrs.RatingPlanIds {
			rpKeys[idx] = engine.RATING_PLAN_PREFIX + rpId
		}
	}
	if len(attrs.RatingProfileIds) > 0 {
		rpfKeys = make([]string, len(attrs.RatingProfileIds))
		for idx, rpfId := range attrs.RatingProfileIds {
			rpfKeys[idx] = engine.RATING_PROFILE_PREFIX + rpfId
		}
	}
	if len(attrs.ActionIds) > 0 {
		actKeys = make([]string, len(attrs.ActionIds))
		for idx, actId := range attrs.ActionIds {
			actKeys[idx] = engine.ACTION_PREFIX + actId
		}
	}
	if len(attrs.SharedGroupIds) > 0 {
		shgKeys = make([]string, len(attrs.SharedGroupIds))
		for idx, shgId := range attrs.SharedGroupIds {
			shgKeys[idx] = engine.SHARED_GROUP_PREFIX + shgId
		}
	}
	if err := self.RatingDb.CacheRating(dstKeys, rpKeys, rpfKeys); err != nil {
		return err
	}
	if err := self.AccountDb.CacheAccounting(actKeys, shgKeys); err != nil {
		return err
	}
	*reply = "OK"
	return nil
}

func (self *ApierV1) GetCacheStats(attrs utils.AttrCacheStats, reply *utils.CacheStats) error {
	cs := new(utils.CacheStats)
	cs.Destinations = cache2go.CountEntries(engine.DESTINATION_PREFIX)
	cs.RatingPlans = cache2go.CountEntries(engine.RATING_PLAN_PREFIX)
	cs.RatingProfiles = cache2go.CountEntries(engine.RATING_PROFILE_PREFIX)
	cs.Actions = cache2go.CountEntries(engine.ACTION_PREFIX)
	*reply = *cs
	return nil
}

func (self *ApierV1) GetCachedItemAge(itemId string, reply *utils.CachedItemAge) error {
	if len(itemId) == 0 {
		return fmt.Errorf("%s:ItemId", utils.ERR_MANDATORY_IE_MISSING)
	}
	cachedItemAge := new(utils.CachedItemAge)
	var found bool
	for idx, cacheKey := range []string{engine.DESTINATION_PREFIX + itemId, engine.RATING_PLAN_PREFIX + itemId, engine.RATING_PROFILE_PREFIX + itemId,
		engine.ACTION_PREFIX + itemId} {
		if age, err := cache2go.GetKeyAge(cacheKey); err == nil {
			found = true
			switch idx {
			case 0:
				cachedItemAge.Destination = age
			case 1:
				cachedItemAge.RatingPlan = age
			case 2:
				cachedItemAge.RatingProfile = age
			case 3:
				cachedItemAge.Action = age
			}
		}
	}
	if !found {
		return errors.New(utils.ERR_NOT_FOUND)
	}
	*reply = *cachedItemAge
	return nil
}

type AttrLoadTPFromFolder struct {
	FolderPath string // Take files from folder absolute path
	DryRun     bool   // Do not write to database but parse only
	FlushDb    bool   // Flush previous data before loading new one
}

func (self *ApierV1) LoadTariffPlanFromFolder(attrs AttrLoadTPFromFolder, reply *string) error {
	loader := engine.NewFileCSVReader(self.RatingDb, self.AccountDb, utils.CSV_SEP,
		path.Join(attrs.FolderPath, utils.DESTINATIONS_CSV),
		path.Join(attrs.FolderPath, utils.TIMINGS_CSV),
		path.Join(attrs.FolderPath, utils.RATES_CSV),
		path.Join(attrs.FolderPath, utils.DESTINATION_RATES_CSV),
		path.Join(attrs.FolderPath, utils.RATING_PLANS_CSV),
		path.Join(attrs.FolderPath, utils.RATING_PROFILES_CSV),
		path.Join(attrs.FolderPath, utils.SHARED_GROUPS_CSV),
		path.Join(attrs.FolderPath, utils.ACTIONS_CSV),
		path.Join(attrs.FolderPath, utils.ACTION_PLANS_CSV),
		path.Join(attrs.FolderPath, utils.ACTION_TRIGGERS_CSV),
		path.Join(attrs.FolderPath, utils.ACCOUNT_ACTIONS_CSV))
	if err := loader.LoadAll(); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if attrs.DryRun {
		*reply = "OK"
		return nil // Mission complete, no errors
	}
	if err := loader.WriteToDatabase(attrs.FlushDb, false); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	// Make sure the items are in the cache
	dstIds, _ := loader.GetLoadedIds(engine.DESTINATION_PREFIX)
	dstKeys := make([]string, len(dstIds))
	for idx, dId := range dstIds {
		dstKeys[idx] = engine.DESTINATION_PREFIX + dId // Cache expects them as redis keys
	}
	rplIds, _ := loader.GetLoadedIds(engine.RATING_PLAN_PREFIX)
	rpKeys := make([]string, len(rplIds))
	for idx, rpId := range rplIds {
		rpKeys[idx] = engine.RATING_PLAN_PREFIX + rpId
	}
	rpfIds, _ := loader.GetLoadedIds(engine.RATING_PROFILE_PREFIX)
	rpfKeys := make([]string, len(rpfIds))
	for idx, rpfId := range rpfIds {
		rpfKeys[idx] = engine.RATING_PROFILE_PREFIX + rpfId
	}
	actIds, _ := loader.GetLoadedIds(engine.ACTION_PREFIX)
	actKeys := make([]string, len(actIds))
	for idx, actId := range actIds {
		actKeys[idx] = engine.ACTION_PREFIX + actId
	}
	shgIds, _ := loader.GetLoadedIds(engine.SHARED_GROUP_PREFIX)
	shgKeys := make([]string, len(shgIds))
	for idx, shgId := range shgIds {
		shgKeys[idx] = engine.SHARED_GROUP_PREFIX + shgId
	}
	if err := self.RatingDb.CacheRating(dstKeys, rpKeys, rpfKeys); err != nil {
		return err
	}
	if err := self.AccountDb.CacheAccounting(actKeys, shgKeys); err != nil {
		return err
	}
	if self.Sched != nil {
		self.Sched.LoadActionTimings(self.AccountDb)
		self.Sched.Restart()
	}
	*reply = "OK"
	return nil
}
