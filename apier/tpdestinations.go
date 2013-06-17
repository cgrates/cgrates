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
	"github.com/cgrates/cgrates/rater"
	"github.com/cgrates/cgrates/utils"
)

type AttrGetTPDestinations struct {
	TPid            string
	DestinationsTag string
}

// Return destinations profile for a destination tag received as parameter
func (self *Apier) GetTPDestinations(attrs AttrGetTPDestinations, reply *rater.Destination) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationsTag"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if dst, err := self.StorDb.GetTPDestination(attrs.TPid, attrs.DestinationsTag); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if dst == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *dst
	}
	return nil
}

func (self *Apier) GetDestination(tag string, reply *rater.Destination) error {
	if dst, err := self.StorDb.GetDestination(tag); err != nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *dst
	}
	return nil
}

func (self *Apier) SetDestination(dest *rater.Destination, reply *rater.Destination) error {
	if err := self.StorDb.SetDestination(dest); err != nil {
		return err
	}
	return nil
}

type AttrGetBalance struct {
	Account   string
	BalanceId string
	Direction string
}

// Get balance
func (self *Apier) GetBalance(attr *AttrGetBalance, reply *float64) error {
	userBalance, err := self.StorDb.GetUserBalance(attr.Account)
	if err != nil {
		return err
	}

	if attr.Direction == "" {
		attr.Direction = rater.OUTBOUND
	}

	if balance, balExists := userBalance.BalanceMap[attr.BalanceId+attr.Direction]; !balExists {
		// No match, balanceId not found
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = balance
	}
	return nil
}

type AttrAddBalance struct {
	Account   string
	BalanceId string
	Direction string
	Value     float64
}

func (self *Apier) AddBalance(attr *AttrAddBalance, reply *float64) error {
	// what storage instance do we use?

	at := &rater.ActionTiming{
		UserBalanceIds: []string{attr.Account},
	}

	if attr.Direction == "" {
		attr.Direction = rater.OUTBOUND
	}

	at.SetActions(rater.Actions{&rater.Action{BalanceId: attr.BalanceId, Direction: attr.Direction, Units: attr.Value}})

	if err := at.Execute(); err != nil {
		return err
	}
	// what to put in replay?
	return nil
}

type AttrExecuteAction struct {
	Account   string
	BalanceId string
	ActionsId string
}

func (self *Apier) ExecuteAction(attr *AttrExecuteAction, reply *float64) error {
	at := &rater.ActionTiming{
		UserBalanceIds: []string{attr.Account},
		ActionsId:      attr.ActionsId,
	}

	if err := at.Execute(); err != nil {
		return err
	}
	// what to put in replay
	return nil
}

type AttrSetRatingProfile struct {
	Subject         string
	RatingProfileId string
}

func (self *Apier) SetRatingProfile(attr *AttrSetRatingProfile, reply *float64) error {
	subject, err := self.StorDb.GetRatingProfile(attr.Subject)
	if err != nil {
		return err
	}
	rp, err := self.StorDb.GetRatingProfile(attr.RatingProfileId)
	if err != nil {
		return err
	}
	subject.DestinationMap = rp.DestinationMap
	err = self.StorDb.SetRatingProfile(subject)
	return err
}
