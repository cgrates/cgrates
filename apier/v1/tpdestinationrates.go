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

package v1

// This file deals with tp_destination_rates management over APIs

import (
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

// Creates a new DestinationRate profile within a tariff plan
func (self *ApierV1) SetTPDestinationRate(attrs utils.TPDestinationRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationRateId", "DestinationRates"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.SetTPDestinationRates(attrs.TPid, map[string][]*utils.DestinationRate{attrs.DestinationRateId: attrs.DestinationRates}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPDestinationRate struct {
	TPid              string // Tariff plan id
	DestinationRateId string // Rate id
	utils.Paginator
}

// Queries specific DestinationRate profile on tariff plan
func (self *ApierV1) GetTPDestinationRate(attrs AttrGetTPDestinationRate, reply *utils.TPDestinationRate) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationRateId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if drs, err := self.StorDb.GetTpDestinationRates(attrs.TPid, attrs.DestinationRateId, &attrs.Paginator); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(drs) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *drs[attrs.DestinationRateId]
	}
	return nil
}

type AttrTPDestinationRateIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries DestinationRate identities on specific tariff plan.
func (self *ApierV1) GetTPDestinationRateIds(attrs AttrGetTPRateIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_DESTINATION_RATES, utils.TPDistinctIds{"id"}, nil, &attrs.Paginator); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific DestinationRate on Tariff plan
func (self *ApierV1) RemTPDestinationRate(attrs AttrGetTPDestinationRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationRateId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_DESTINATION_RATES, attrs.TPid, attrs.DestinationRateId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
