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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
)

const (
	OK = "OK"
)

type ApierV1 struct {
	StorDb engine.DataStorage
	DataDb engine.DataStorage
	Sched  *scheduler.Scheduler
}

type AttrDestination struct {
	Id       string
	Prefixes []string
}

func (self *ApierV1) GetDestination(attr *AttrDestination, reply *AttrDestination) error {
	if dst, err := self.DataDb.GetDestination(attr.Id); err != nil {
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

	at.SetActions(engine.Actions{&engine.Action{ActionType: engine.TOPUP, BalanceId: attr.BalanceId, Direction: attr.Direction, Units: attr.Value}})

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

type AttrSetRatingProfile struct {
	TPid          string
	RateProfileId string
}

// Process dependencies and load a specific rating profile from storDb into dataDb.
func (self *ApierV1) SetRatingProfile(attrs AttrSetRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RateProfileId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	dbReader := engine.NewDbReader(self.StorDb, self.DataDb, attrs.TPid)

	if err := dbReader.LoadRatingProfileByTag(attrs.RateProfileId); err != nil {
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
func (self *ApierV1) AddAccount(attr *AttrAddAccount, reply *string) error {
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

type AttrSetAccountActions struct {
	TPid             string
	AccountActionsId string
}

// Process dependencies and load a specific AccountActions profile from storDb into dataDb.
func (self *ApierV1) SetAccountActions(attrs AttrSetAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "AccountActionsId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	dbReader := engine.NewDbReader(self.StorDb, self.DataDb, attrs.TPid)

	if _, err := engine.AccLock.Guard(attrs.AccountActionsId, func() (float64, error) {
		if err := dbReader.LoadAccountActionsByTag(attrs.AccountActionsId); err != nil {
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
