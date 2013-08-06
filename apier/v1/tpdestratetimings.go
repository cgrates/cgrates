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

// This file deals with tp_destrates_timing management over APIs

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new DestinationRateTiming profile within a tariff plan
func (self *ApierV1) SetTPDestRateTiming(attrs utils.TPDestRateTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestRateTimingId", "DestRateTimings"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if exists, err := self.StorDb.ExistsTPDestRateTiming(attrs.TPid, attrs.DestRateTimingId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if exists {
		return errors.New(utils.ERR_DUPLICATE)
	}
	drts := make([]*engine.DestinationRateTiming, len(attrs.DestRateTimings))
	for idx, drt := range attrs.DestRateTimings {
		drts[idx] = &engine.DestinationRateTiming{Tag: attrs.DestRateTimingId,
			DestinationRatesTag: drt.DestRatesId,
			Weight:              drt.Weight,
			TimingsTag:          drt.TimingId,
		}
	}
	if err := self.StorDb.SetTPDestRateTimings(attrs.TPid, map[string][]*engine.DestinationRateTiming{attrs.DestRateTimingId: drts}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPDestRateTiming struct {
	TPid             string // Tariff plan id
	DestRateTimingId string // Rate id
}

// Queries specific DestRateTiming profile on tariff plan
func (self *ApierV1) GetTPDestRateTiming(attrs AttrGetTPDestRateTiming, reply *utils.TPDestRateTiming) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestRateTimingId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if dr, err := self.StorDb.GetTPDestRateTiming(attrs.TPid, attrs.DestRateTimingId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if dr == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *dr
	}
	return nil
}

type AttrTPDestRateTimingIds struct {
	TPid string // Tariff plan id
}

// Queries DestRateTiming identities on specific tariff plan.
func (self *ApierV1) GetTPDestRateTimingIds(attrs AttrGetTPRateIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPDestRateTimingIds(attrs.TPid); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}
