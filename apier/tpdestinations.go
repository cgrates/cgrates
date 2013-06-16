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

type AttrBalance struct {
	Account   string
	Direction string
}

func (self *Apier) GetMoneyBalance(attr *AttrBalance, reply *float64) (err error) {
	err = self.getBalance(attr, rater.CREDIT, reply)
	return err
}

func (self *Apier) GetSMSBalance(attr *AttrBalance, reply *float64) (err error) {
	err = self.getBalance(attr, rater.SMS, reply)
	return err
}

func (self *Apier) GetInternetBalance(attr *AttrBalance, reply *float64) (err error) {
	err = self.getBalance(attr, rater.TRAFFIC, reply)
	return err
}

func (self *Apier) GetInternetTimeBalance(attr *AttrBalance, reply *float64) (err error) {
	err = self.getBalance(attr, rater.TRAFFIC_TIME, reply)
	return err
}

func (self *Apier) GetMinutesBalance(attr *AttrBalance, reply *float64) (err error) {
	err = self.getBalance(attr, rater.MINUTES, reply)
	return err
}

// Get balance
func (self *Apier) getBalance(attr *AttrBalance, balanceId string, reply *float64) (err error) {
	userBalance, err := self.StorDb.GetUserBalance(attr.Account)
	if err != nil {
		return err
	}
	if balance, balExists := userBalance.BalanceMap[balanceId+attr.Direction]; !balExists {
		// No match, balanceId not found
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = balance
	}
	return nil
}

func (self *Apier) AddTimedAction(attr *AttrBalance, reply *float64) (err error) {
	err = self.getBalance(attr, rater.MINUTES, reply)
	return err
}
