/*
Real-time Charging System for Telecom & ISP environments
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

package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

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
	StorDb      engine.LoadStorage
	RatingDb    engine.RatingStorage
	AccountDb   engine.AccountingStorage
	CdrDb       engine.CdrStorage
	LogDb       engine.LogStorage
	Sched       *scheduler.Scheduler
	Config      *config.CGRConfig
	Responder   *engine.Responder
	CdrStatsSrv *engine.Stats
}

func (self *ApierV1) GetDestination(dstId string, reply *engine.Destination) error {
	if dst, err := self.RatingDb.GetDestination(dstId); err != nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *dst
	}
	return nil
}

func (apier *ApierV1) GetSharedGroup(sgId string, reply *engine.SharedGroup) error {
	if sg, err := apier.AccountDb.GetSharedGroup(sgId, false); err != nil && err.Error() != utils.ERR_NOT_FOUND { // Not found is not an error here
		return err
	} else {
		if sg != nil {
			*reply = *sg
		}
	}
	return nil
}

type AttrSetDestination struct { //ToDo
	Id        string
	Prefixes  []string
	Overwrite bool
}

func (self *ApierV1) GetRatingPlan(rplnId string, reply *engine.RatingPlan) error {
	if rpln, err := self.RatingDb.GetRatingPlan(rplnId, false); err != nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *rpln
	}
	return nil
}

// Get balance
func (self *ApierV1) GetAccount(attr *utils.AttrGetAccount, reply *engine.Account) error {
	tag := fmt.Sprintf("%s:%s:%s", attr.Direction, attr.Tenant, attr.Account)
	userBalance, err := self.AccountDb.GetAccount(tag)
	if err != nil {
		return err
	}

	*reply = *userBalance
	return nil
}

type AttrAddBalance struct {
	Tenant        string
	Account       string
	BalanceId     string
	BalanceType   string
	Direction     string
	Value         float64
	ExpiryTime    string
	RatingSubject string
	DestinationId string
	Weight        float64
	SharedGroup   string
	Overwrite     bool // When true it will reset if the balance is already there
}

func (self *ApierV1) AddBalance(attr *AttrAddBalance, reply *string) error {
	expTime, err := utils.ParseDate(attr.ExpiryTime)
	if err != nil {
		*reply = err.Error()
		return err
	}
	tag := fmt.Sprintf("%s:%s:%s", attr.Direction, attr.Tenant, attr.Account)
	if _, err := self.AccountDb.GetAccount(tag); err != nil {
		// create user balance if not exists
		account := &engine.Account{
			Id: tag,
		}
		if err := self.AccountDb.SetAccount(account); err != nil {
			*reply = err.Error()
			return err
		}
	}
	at := &engine.ActionTiming{
		AccountIds: []string{tag},
	}
	if attr.Direction == "" {
		attr.Direction = engine.OUTBOUND
	}
	aType := engine.DEBIT
	// reverse the sign as it is a debit
	attr.Value = -attr.Value

	if attr.Overwrite {
		aType = engine.DEBIT_RESET
	}
	at.SetActions(engine.Actions{
		&engine.Action{
			ActionType:  aType,
			BalanceType: attr.BalanceType,
			Direction:   attr.Direction,
			Balance: &engine.Balance{
				Id:             attr.BalanceId,
				Value:          attr.Value,
				ExpirationDate: expTime,
				RatingSubject:  attr.RatingSubject,
				DestinationId:  attr.DestinationId,
				Weight:         attr.Weight,
				SharedGroup:    attr.SharedGroup,
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
		AccountIds: []string{tag},
		ActionsId:  attr.ActionsId,
	}

	if err := at.Execute(); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

type AttrLoadDestination struct {
	TPid          string
	DestinationId string
}

// Load destinations from storDb into dataDb.
func (self *ApierV1) LoadDestination(attrs AttrLoadDestination, reply *string) error {
	if len(attrs.TPid) == 0 {
		return fmt.Errorf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, "TPid")
	}
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)
	if loaded, err := dbReader.LoadDestinationByTag(attrs.DestinationId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if !loaded {
		return errors.New(utils.ERR_NOT_FOUND)
	}
	//Automatic cache of the newly inserted rating plan
	didNotChange := []string{}
	destIds := []string{engine.DESTINATION_PREFIX + attrs.DestinationId}
	if len(attrs.DestinationId) == 0 {
		destIds = nil // Cache all destinations, temporary here until we add ApierV2.LoadDestinations
	}
	if err := self.RatingDb.CacheRating(destIds, didNotChange, didNotChange, didNotChange, didNotChange); err != nil {
		return err
	}
	*reply = OK
	return nil
}

// Load derived chargers from storDb into dataDb.
func (self *ApierV1) LoadDerivedChargers(attrs utils.TPDerivedChargers, reply *string) error {
	if len(attrs.TPid) == 0 {
		return fmt.Errorf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, "TPid")
	}
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)
	if err := dbReader.LoadDerivedChargersFiltered(&attrs); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	//Automatic cache of the newly inserted rating plan
	didNotChange := []string{}
	var derivedChargingKeys []string
	if len(attrs.Direction) != 0 && len(attrs.Tenant) != 0 && len(attrs.Category) != 0 && len(attrs.Account) != 0 && len(attrs.Subject) != 0 {
		derivedChargingKeys = []string{engine.DERIVEDCHARGERS_PREFIX + attrs.GetDerivedChargersKey()}
	}
	if err := self.AccountDb.CacheAccounting(didNotChange, didNotChange, didNotChange, derivedChargingKeys); err != nil {
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
	if len(attrs.TPid) == 0 {
		return fmt.Errorf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, "TPid")
	}
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)
	if loaded, err := dbReader.LoadRatingPlanByTag(attrs.RatingPlanId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if !loaded {
		return errors.New(utils.ERR_NOT_FOUND)
	}
	//Automatic cache of the newly inserted rating plan
	didNotChange := []string{}
	var changedRPlKeys []string
	if len(attrs.TPid) != 0 {
		changedRPlKeys = []string{engine.RATING_PLAN_PREFIX + attrs.RatingPlanId}
	}
	if err := self.RatingDb.CacheRating(nil, changedRPlKeys, didNotChange, didNotChange, didNotChange); err != nil {
		return err
	}
	*reply = OK
	return nil
}

// Process dependencies and load a specific rating profile from storDb into dataDb.
func (self *ApierV1) LoadRatingProfile(attrs utils.TPRatingProfile, reply *string) error {
	if len(attrs.TPid) == 0 {
		return fmt.Errorf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, "TPid")
	}
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)
	if err := dbReader.LoadRatingProfileFiltered(&attrs); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	//Automatic cache of the newly inserted rating profile
	didNotChange := []string{}
	var ratingProfile []string
	if attrs.KeyId() != ":::" { // if has some filters
		ratingProfile = []string{engine.RATING_PROFILE_PREFIX + attrs.KeyId()}
	}
	if err := self.RatingDb.CacheRating(didNotChange, didNotChange, ratingProfile, didNotChange, didNotChange); err != nil {
		return err
	}
	*reply = OK
	return nil
}

type AttrLoadSharedGroup struct {
	TPid          string
	SharedGroupId string
}

// Load destinations from storDb into dataDb.
func (self *ApierV1) LoadSharedGroup(attrs AttrLoadSharedGroup, reply *string) error {
	if len(attrs.TPid) == 0 {
		return fmt.Errorf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, "TPid")
	}
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)
	if err := dbReader.LoadSharedGroupByTag(attrs.SharedGroupId, true); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	//Automatic cache of the newly inserted rating plan
	didNotChange := []string{}
	var changedSharedGroup []string
	if len(attrs.SharedGroupId) != 0 {
		changedSharedGroup = []string{engine.SHARED_GROUP_PREFIX + attrs.SharedGroupId}
	}
	if err := self.AccountDb.CacheAccounting(didNotChange, changedSharedGroup, didNotChange, didNotChange); err != nil {
		return err
	}
	*reply = OK
	return nil
}

type AttrLoadCdrStats struct {
	TPid       string
	CdrStatsId string
}

// Load destinations from storDb into dataDb.
func (self *ApierV1) LoadCdrStats(attrs AttrLoadCdrStats, reply *string) error {
	if len(attrs.TPid) == 0 {
		return fmt.Errorf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, "TPid")
	}
	dbReader := engine.NewDbReader(self.StorDb, self.RatingDb, self.AccountDb, attrs.TPid)
	if err := dbReader.LoadCdrStatsByTag(attrs.CdrStatsId, true); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = OK
	return nil
}

type AttrSetRatingProfile struct {
	Tenant                string                      // Tenant's Id
	Category              string                      // TypeOfRecord
	Direction             string                      // Traffic direction, OUT is the only one supported for now
	Subject               string                      // Rating subject, usually the same as account
	Overwrite             bool                        // Overwrite if exists
	RatingPlanActivations []*utils.TPRatingActivation // Activate rating plans at specific time
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
	tpRpf := utils.TPRatingProfile{Tenant: attrs.Tenant, Category: attrs.Category, Direction: attrs.Direction, Subject: attrs.Subject}
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
			FallbackKeys: utils.FallbackSubjKeys(tpRpf.Direction, tpRpf.Tenant, tpRpf.Category, ra.FallbackSubjects)}
	}
	if err := self.RatingDb.SetRatingProfile(rpfl); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	//Automatic cache of the newly inserted rating profile
	didNotChange := []string{}
	if err := self.RatingDb.CacheRating(didNotChange, didNotChange, []string{engine.RATING_PROFILE_PREFIX + keyId}, didNotChange, didNotChange); err != nil {
		return err
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
			BalanceType:      apiAct.BalanceType,
			Direction:        apiAct.Direction,
			Weight:           apiAct.Weight,
			ExpirationString: apiAct.ExpiryTime,
			ExtraParameters:  apiAct.ExtraParameters,
			Balance: &engine.Balance{
				Uuid:          utils.GenUUID(),
				Id:            apiAct.BalanceId,
				Value:         apiAct.Units,
				Weight:        apiAct.BalanceWeight,
				DestinationId: apiAct.DestinationId,
				RatingSubject: apiAct.RatingSubject,
				SharedGroup:   apiAct.SharedGroup,
			},
		}
		storeActions[idx] = a
	}
	if err := self.AccountDb.SetActions(attrs.ActionsId, storeActions); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	didNotChange := []string{}
	self.AccountDb.CacheAccounting(nil, didNotChange, didNotChange, didNotChange)
	*reply = OK
	return nil
}

// Retrieves actions attached to specific ActionsId within cache
func (self *ApierV1) GetActions(actsId string, reply *[]*utils.TPAction) error {
	if len(actsId) == 0 {
		return fmt.Errorf("%s ActionsId: %s", utils.ERR_MANDATORY_IE_MISSING, actsId)
	}
	acts := make([]*utils.TPAction, 0)
	engActs, err := self.AccountDb.GetActions(actsId, false)
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	for _, engAct := range engActs {
		act := &utils.TPAction{Identifier: engAct.ActionType,
			BalanceType:     engAct.BalanceType,
			Direction:       engAct.Direction,
			ExpiryTime:      engAct.ExpirationString,
			ExtraParameters: engAct.ExtraParameters,
			Weight:          engAct.Weight,
		}
		if engAct.Balance != nil {
			act.Units = engAct.Balance.Value
			act.DestinationId = engAct.Balance.DestinationId
			act.RatingSubject = engAct.Balance.RatingSubject
			act.SharedGroup = engAct.Balance.SharedGroup
			act.BalanceWeight = engAct.Balance.Weight
		}
		acts = append(acts, act)
	}
	*reply = acts
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
			return fmt.Errorf("%s:%s", utils.ERR_BROKEN_REFERENCE, apiAtm.ActionsId)
		}
		timing := new(engine.RITiming)
		timing.Years.Parse(apiAtm.Years, ";")
		timing.Months.Parse(apiAtm.Months, ";")
		timing.MonthDays.Parse(apiAtm.MonthDays, ";")
		timing.WeekDays.Parse(apiAtm.WeekDays, ";")
		timing.StartTime = apiAtm.Time
		at := &engine.ActionTiming{
			Uuid:      utils.GenUUID(),
			Id:        attrs.Id,
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
	ActionTriggersId     string
	Tenant               string
	Account              string
	ThresholdType        string
	ThresholdValue       float64
	BalanceId            string
	BalanceType          string
	BalanceDirection     string
	BalanceDestinationId string
	BalanceRatingSubject string //ToDo
	BalanceWeight        float64
	BalanceExpiryTime    string
	BalanceSharedGroup   string //ToDo
	Weight               float64
	ActionsId            string
}

func (self *ApierV1) AddTriggeredAction(attr AttrAddActionTrigger, reply *string) error {
	if attr.BalanceDirection == "" {
		attr.BalanceDirection = engine.OUTBOUND
	}
	balExpiryTime, err := utils.ParseTimeDetectLayout(attr.BalanceExpiryTime)
	if err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	at := &engine.ActionTrigger{
		Id:                    attr.ActionTriggersId,
		ThresholdType:         attr.ThresholdType,
		ThresholdValue:        attr.ThresholdValue,
		BalanceId:             attr.BalanceId,
		BalanceType:           attr.BalanceType,
		BalanceDirection:      attr.BalanceDirection,
		BalanceDestinationId:  attr.BalanceDestinationId,
		BalanceWeight:         attr.BalanceWeight,
		BalanceExpirationDate: balExpiryTime,
		Weight:                attr.Weight,
		ActionsId:             attr.ActionsId,
		Executed:              false,
	}

	tag := utils.AccountKey(attr.Tenant, attr.Account, attr.BalanceDirection)
	_, err = engine.AccLock.Guard(tag, func() (float64, error) {
		userBalance, err := self.AccountDb.GetAccount(tag)
		if err != nil {
			return 0, err
		}

		userBalance.ActionTriggers = append(userBalance.ActionTriggers, at)

		if err = self.AccountDb.SetAccount(userBalance); err != nil {
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

type AttrResetTriggeredAction struct {
	Id                   string
	Tenant               string
	Account              string
	Direction            string
	BalanceType          string
	ThresholdType        string
	ThresholdValue       float64
	DestinationId        string
	BalanceWeight        float64
	BalanceRatingSubject string
	BalanceSharedGroup   string
}

func (self *ApierV1) ResetTriggeredActions(attr AttrResetTriggeredAction, reply *string) error {
	var a *engine.Action
	if attr.Id != "" {
		// we can identify the trigge by the id
		a = &engine.Action{Id: attr.Id}
	} else {
		if attr.Direction == "" {
			attr.Direction = engine.OUTBOUND
		}
		extraParameters, err := json.Marshal(struct {
			ThresholdType        string
			ThresholdValue       float64
			DestinationId        string
			BalanceWeight        float64
			BalanceRatingSubject string
			BalanceSharedGroup   string
		}{
			attr.ThresholdType,
			attr.ThresholdValue,
			attr.DestinationId,
			attr.BalanceWeight,
			attr.BalanceRatingSubject,
			attr.BalanceSharedGroup,
		})
		if err != nil {
			*reply = err.Error()
			return err
		}
		a = &engine.Action{
			BalanceType:     attr.BalanceType,
			Direction:       attr.Direction,
			ExtraParameters: string(extraParameters),
		}
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account, attr.Direction)
	_, err := engine.AccLock.Guard(accID, func() (float64, error) {
		acc, err := self.AccountDb.GetAccount(accID)
		if err != nil {
			return 0, err
		}

		acc.ResetActionTriggers(a)

		if err = self.AccountDb.SetAccount(acc); err != nil {
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
	if len(attrs.TPid) == 0 {
		return fmt.Errorf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, "TPid")
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
	if err := self.AccountDb.CacheAccounting(nil, nil, nil, []string{}); err != nil {
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
	var dstKeys, rpKeys, rpfKeys, actKeys, shgKeys, rpAlsKeys, accAlsKeys, lcrKeys, dcsKeys []string
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
	if len(attrs.RpAliases) > 0 {
		rpAlsKeys = make([]string, len(attrs.RpAliases))
		for idx, alias := range attrs.RpAliases {
			rpAlsKeys[idx] = engine.RP_ALIAS_PREFIX + alias
		}
	}
	if len(attrs.AccAliases) > 0 {
		accAlsKeys = make([]string, len(attrs.AccAliases))
		for idx, alias := range attrs.AccAliases {
			accAlsKeys[idx] = engine.ACC_ALIAS_PREFIX + alias
		}
	}
	if len(attrs.LCRIds) > 0 {
		lcrKeys = make([]string, len(attrs.LCRIds))
		for idx, lcrId := range attrs.LCRIds {
			lcrKeys[idx] = engine.LCR_PREFIX + lcrId
		}
	}

	if len(attrs.DerivedChargers) > 0 {
		dcsKeys = make([]string, len(attrs.DerivedChargers))
		for idx, dc := range attrs.DerivedChargers {
			dcsKeys[idx] = engine.DERIVEDCHARGERS_PREFIX + dc
		}
	}
	if err := self.RatingDb.CacheRating(dstKeys, rpKeys, rpfKeys, rpAlsKeys, lcrKeys); err != nil {
		return err
	}
	if err := self.AccountDb.CacheAccounting(actKeys, shgKeys, accAlsKeys, dcsKeys); err != nil {
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
	cs.SharedGroups = cache2go.CountEntries(engine.SHARED_GROUP_PREFIX)
	cs.RatingAliases = cache2go.CountEntries(engine.RP_ALIAS_PREFIX)
	cs.AccountAliases = cache2go.CountEntries(engine.ACC_ALIAS_PREFIX)
	cs.DerivedChargers = cache2go.CountEntries(engine.DERIVEDCHARGERS_PREFIX)
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
		engine.ACTION_PREFIX + itemId, engine.SHARED_GROUP_PREFIX + itemId, engine.RP_ALIAS_PREFIX + itemId, engine.ACC_ALIAS_PREFIX + itemId} {
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
			case 4:
				cachedItemAge.SharedGroup = age
			case 5:
				cachedItemAge.RatingAlias = age
			case 6:
				cachedItemAge.AccountAlias = age
			}
		}
	}
	if !found {
		return errors.New(utils.ERR_NOT_FOUND)
	}
	*reply = *cachedItemAge
	return nil
}

func (self *ApierV1) LoadTariffPlanFromFolder(attrs utils.AttrLoadTpFromFolder, reply *string) error {
	if len(attrs.FolderPath) == 0 {
		return fmt.Errorf("%s:%s", utils.ERR_MANDATORY_IE_MISSING, "FolderPath")
	}
	if fi, err := os.Stat(attrs.FolderPath); err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return errors.New(utils.ERR_INVALID_PATH)
		}
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if !fi.IsDir() {
		return errors.New(utils.ERR_INVALID_PATH)
	}
	loader := engine.NewFileCSVReader(self.RatingDb, self.AccountDb, utils.CSV_SEP,
		path.Join(attrs.FolderPath, utils.DESTINATIONS_CSV),
		path.Join(attrs.FolderPath, utils.TIMINGS_CSV),
		path.Join(attrs.FolderPath, utils.RATES_CSV),
		path.Join(attrs.FolderPath, utils.DESTINATION_RATES_CSV),
		path.Join(attrs.FolderPath, utils.RATING_PLANS_CSV),
		path.Join(attrs.FolderPath, utils.RATING_PROFILES_CSV),
		path.Join(attrs.FolderPath, utils.SHARED_GROUPS_CSV),
		path.Join(attrs.FolderPath, utils.LCRS_CSV),
		path.Join(attrs.FolderPath, utils.ACTIONS_CSV),
		path.Join(attrs.FolderPath, utils.ACTION_PLANS_CSV),
		path.Join(attrs.FolderPath, utils.ACTION_TRIGGERS_CSV),
		path.Join(attrs.FolderPath, utils.ACCOUNT_ACTIONS_CSV),
		path.Join(attrs.FolderPath, utils.DERIVED_CHARGERS_CSV),
		path.Join(attrs.FolderPath, utils.CDR_STATS_CSV))
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
	rpAliases, _ := loader.GetLoadedIds(engine.RP_ALIAS_PREFIX)
	rpAlsKeys := make([]string, len(rpAliases))
	for idx, alias := range rpAliases {
		rpAlsKeys[idx] = engine.RP_ALIAS_PREFIX + alias
	}
	accAliases, _ := loader.GetLoadedIds(engine.ACC_ALIAS_PREFIX)
	accAlsKeys := make([]string, len(accAliases))
	for idx, alias := range accAliases {
		accAlsKeys[idx] = engine.ACC_ALIAS_PREFIX + alias
	}
	lcrIds, _ := loader.GetLoadedIds(engine.LCR_PREFIX)
	lcrKeys := make([]string, len(lcrIds))
	for idx, lcrId := range lcrIds {
		lcrKeys[idx] = engine.LCR_PREFIX + lcrId
	}
	dcs, _ := loader.GetLoadedIds(engine.DERIVEDCHARGERS_PREFIX)
	dcsKeys := make([]string, len(dcs))
	for idx, dc := range dcs {
		dcsKeys[idx] = engine.DERIVEDCHARGERS_PREFIX + dc
	}
	if err := self.RatingDb.CacheRating(dstKeys, rpKeys, rpfKeys, rpAlsKeys, lcrKeys); err != nil {
		return err
	}
	if err := self.AccountDb.CacheAccounting(actKeys, shgKeys, accAlsKeys, dcsKeys); err != nil {
		return err
	}
	if self.Sched != nil {
		self.Sched.LoadActionTimings(self.AccountDb)
		self.Sched.Restart()
	}
	cstKeys, _ := loader.GetLoadedIds(engine.CDR_STATS_PREFIX)
	if len(cstKeys) != 0 && self.CdrStatsSrv != nil {
		if err := self.CdrStatsSrv.ReloadQueues(cstKeys, nil); err != nil {
			return err
		}
	}
	*reply = "OK"
	return nil
}
