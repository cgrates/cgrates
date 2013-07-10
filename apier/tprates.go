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

// This file deals with tp_rates management over APIs

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/rater"
	"github.com/cgrates/cgrates/utils"
	"strings"
)

type ApierTPRate struct {
	TPid           string // Tariff plan id
	RateId         string // Rates id
	ConnectFee     string // ConnectFee applied once the call is answered
	Rate           string // Rate applied
	RatedUnits     string //  Number of billing units this rate applies to
	RateIncrements string // This rate will apply in increments of
	Weight         string // Rate's priority when dealing with grouped rates
}

// Creates a new rate within a tariff plan
func (self *Apier) SetTPRate(attrs ApierTPRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RateId", "ConnectFee", "Rate", "RatedUnits", "RateIncrements", "Weight"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	rt, errRt := rater.NewRate(attrs.RateId, attrs.ConnectFee, attrs.Rate, attrs.RatedUnits, attrs.RateIncrements, attrs.Weight)
	if errRt != nil {
		return fmt.Errorf("%s:%v", utils.ERR_SERVER_ERROR, errRt.Error())
	}
	err := self.StorDb.SetTPRate(attrs.TPid, rt)
	switch {
	case strings.HasPrefix(err.Error(), "Error 1062"): //MySQL way of saying duplicate
		return errors.New(utils.ERR_DUPLICATE)
	case err != nil:
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}
