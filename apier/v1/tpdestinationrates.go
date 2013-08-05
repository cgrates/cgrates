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

// This file deals with tp_destination_rates management over APIs

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new DestinationRate profile within a tariff plan
func (self *ApierV1) SetTPDestinationRate(attrs utils.TPDestinationRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationRateId", "DestinationRates"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if exists, err := self.StorDb.ExistsTPDestinationRate(attrs.TPid, attrs.DestinationRateId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if exists {
		return errors.New(utils.ERR_DUPLICATE)
	}
	drs := make([]*engine.DestinationRate, len(attrs.DestinationRates))
	for idx, dr := range attrs.DestinationRates {
		drs[idx] = &engine.DestinationRate{attrs.DestinationRateId, dr.DestinationId, dr.RateId, nil}
	}
	if err := self.StorDb.SetTPDestinationRates(attrs.TPid, map[string][]*engine.DestinationRate{attrs.DestinationRateId: drs}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPDestinationRate struct {
	TPid              string // Tariff plan id
	DestinationRateId string // Rate id
}

// Queries specific DestinationRate profile on tariff plan
func (self *ApierV1) GetTPDestinationRate(attrs AttrGetTPDestinationRate, reply *utils.TPDestinationRate) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationRateId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if dr, err := self.StorDb.GetTPDestinationRate(attrs.TPid, attrs.DestinationRateId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if dr == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *dr
	}
	return nil
}

type AttrTPDestinationRateIds struct {
	TPid string // Tariff plan id
}

// Queries DestinationRate identities on specific tariff plan.
func (self *ApierV1) GetTPDestinationRateIds(attrs AttrGetTPRateIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPDestinationRateIds(attrs.TPid); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}
