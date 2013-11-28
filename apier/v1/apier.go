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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"

	"path"
)

const (
	OK = "OK"
)

type ApierV1 struct {
	StorDb engine.LoadStorage
	DataDb engine.DataStorage
	CdrDb  engine.CdrStorage
	Sched  *scheduler.Scheduler
	Config *config.CGRConfig
}

type AttrDestination struct {
	Id       string
	Prefixes []string
}

func (self *ApierV1) GetDestination(attr *AttrDestination, reply *AttrDestination) error {
	if dst, err := self.DataDb.GetDestination(attr.Id, false); err != nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		reply.Id = dst.Id
		reply.Prefixes = dst.Prefixes
	}
	return nil
}

type AttrGetBalance struct {
	Tenant    string
	Account   string
	BalanceId string
	Direction string
}

// Get balance
func (self *ApierV1) GetBalance(attr *AttrGetBalance, reply *float64) error {
	tag := fmt.Sprintf("%s:%s:%s", attr.Direction, attr.Tenant, attr.Account)
	userBalance, err := self.DataDb.GetUserBalance(tag)
	if err != nil {
		return err
	}

	if attr.Direction == "" {
		attr.Direction = engine.OUTBOUND
	}

	if balance, balExists := userBalance.BalanceMap[attr.BalanceId+attr.Direction]; !balExists {
		*reply = 0.0
	} else {
		*reply = balance.GetTotalValue()
	}
	return nil
}

type AttrAddBalance struct {
	Tenant    string
	Account   string
	BalanceId string
	Direction string
	Value     float64
}

func (self *ApierV1) AddBalance(attr *AttrAddBalance, reply *string) error {
	// what storage instance do we use?
	tag := fmt.Sprintf("%s:%s:%s", attr.Direction, attr.Tenant, attr.Account)

	if _, err := self.DataDb.GetUserBalance(tag); err != nil {
		// create user balance if not exists
		ub := &engine.UserBalance{
			Id: tag,
		}
		if err := self.DataDb.SetUserBalance(ub); err != nil {
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

	at.SetActions(engine.Actions{&engine.Action{ActionType: engine.TOPUP, BalanceId: attr.BalanceId, Direction: attr.Direction, Balance: &engine.Balance{Value: attr.Value}}})

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

type AttrSetRatingPlan struct {
	TPid         string
	RatingPlanId string
}

// Process dependencies and load a specific rating plan from storDb into dataDb.
func (self *ApierV1) SetRatingPlan(attrs AttrSetRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	dbReader := engine.NewDbReader(self.StorDb, self.DataDb, attrs.TPid)
	if loaded, err := dbReader.LoadRatingPlanByTag(attrs.RatingPlanId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if !loaded {
		return errors.New("NOT_FOUND")
	}
	*reply = OK
	return nil
}

// Process dependencies and load a specific rating profile from storDb into dataDb.
func (self *ApierV1) SetRatingProfile(attrs utils.TPRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LoadId", "Tenant", "TOR", "Direction", "Subject"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	dbReader := engine.NewDbReader(self.StorDb, self.DataDb, attrs.TPid)

	if err := dbReader.LoadRatingProfileFiltered(&attrs); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}

	*reply = OK
	return nil
}

type AttrAddActionTrigger struct {
	Tenant         string
	Account        string
	Direction      string
	BalanceId      string
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
		ThresholdValue: attr.ThresholdValue,
		DestinationId:  attr.DestinationId,
		Weight:         attr.Weight,
		ActionsId:      attr.ActionsId,
		Executed:       false,
	}

	tag := fmt.Sprintf("%s:%s:%s", attr.Direction, attr.Tenant, attr.Account)
	_, err := engine.AccLock.Guard(tag, func() (float64, error) {
		userBalance, err := self.DataDb.GetUserBalance(tag)
		if err != nil {
			return 0, err
		}

		userBalance.ActionTriggers = append(userBalance.ActionTriggers, at)

		if err = self.DataDb.SetUserBalance(userBalance); err != nil {
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

type AttrAddAccount struct {
	Tenant          string
	Direction       string
	Account         string
	Type            string // prepaid-postpaid
	ActionTimingsId string
}

// Ads a new account into dataDb. If already defined, returns success.
func (self *ApierV1) AddAccount(attr AttrAddAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Direction", "Account", "Type", "ActionTimingsId"}); len(missing) != 0 {
		*reply = fmt.Sprintf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	tag := fmt.Sprintf("%s:%s:%s", attr.Direction, attr.Tenant, attr.Account)
	ub := &engine.UserBalance{
		Id:   tag,
		Type: attr.Type,
	}

	if attr.ActionTimingsId != "" {
		if ats, err := self.DataDb.GetActionTimings(attr.ActionTimingsId); err == nil {
			for _, at := range ats {
				at.UserBalanceIds = append(at.UserBalanceIds, tag)
			}
			err = self.DataDb.SetActionTimings(attr.ActionTimingsId, ats)
			if err != nil {
				if self.Sched != nil {
					self.Sched.LoadActionTimings(self.DataDb)
					self.Sched.Restart()
				}
			}
			if err := self.DataDb.SetUserBalance(ub); err != nil {
				*reply = fmt.Sprintf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
				return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
			}
		} else {
			*reply = fmt.Sprintf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
			return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
		}
	}
	*reply = OK
	return nil
}

// Process dependencies and load a specific AccountActions profile from storDb into dataDb.
func (self *ApierV1) SetAccountActions(attrs utils.TPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LoadId", "Tenant", "Account", "Direction"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	dbReader := engine.NewDbReader(self.StorDb, self.DataDb, attrs.TPid)

	if _, err := engine.AccLock.Guard(attrs.KeyId(), func() (float64, error) {
		if err := dbReader.LoadAccountActionsFiltered(&attrs); err != nil {
			return 0, err
		}
		return 0, nil
	}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if self.Sched != nil {
		self.Sched.LoadActionTimings(self.DataDb)
		self.Sched.Restart()
	}
	*reply = OK
	return nil
}

func (self *ApierV1) ReloadScheduler(input string, reply *string) error {
	if self.Sched != nil {
		self.Sched.LoadActionTimings(self.DataDb)
		self.Sched.Restart()
		*reply = OK
		return nil
	}
	*reply = utils.ERR_NOT_FOUND
	return errors.New(utils.ERR_NOT_FOUND)
}

func (self *ApierV1) ReloadCache(attrs utils.ApiReloadCache, reply *string) error {
	var dstKeys, rpKeys []string
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
	if err := self.DataDb.PreCache(dstKeys, rpKeys); err != nil {
		return err
	}
	*reply = "OK"
	return nil
}

type AttrLoadTPFromFolder struct {
	FolderPath string // Take files from folder absolute path
	DryRun     bool   // Do not write to database but parse only
	FlushDb    bool   // Flush previous data before loading new one
}

func (self *ApierV1) LoadTariffPlanFromFolder(attrs AttrLoadTPFromFolder, reply *string) error {
	loader := engine.NewFileCSVReader(self.DataDb, utils.CSV_SEP,
		path.Join(attrs.FolderPath, utils.DESTINATIONS_CSV),
		path.Join(attrs.FolderPath, utils.TIMINGS_CSV),
		path.Join(attrs.FolderPath, utils.RATES_CSV),
		path.Join(attrs.FolderPath, utils.DESTINATION_RATES_CSV),
		path.Join(attrs.FolderPath, utils.RATING_PLANS_CSV),
		path.Join(attrs.FolderPath, utils.RATING_PROFILES_CSV),
		path.Join(attrs.FolderPath, utils.ACTIONS_CSV),
		path.Join(attrs.FolderPath, utils.ACTION_TIMINGS_CSV),
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
	*reply = "OK"
	return nil
}
