/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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
	"github.com/cgrates/cgrates/utils"
)

// Creates a new CdrStats profile within a tariff plan
func (self *ApierV1) SetTPCdrStats(attrs utils.TPCdrStats, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID", "CdrStats"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPCdrStats([]*utils.TPCdrStats{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPCdrStats struct {
	TPid string // Tariff plan id
	ID   string // CdrStat id
}

// Queries specific CdrStat on tariff plan
func (self *ApierV1) GetTPCdrStats(attrs AttrGetTPCdrStats, reply *utils.TPCdrStats) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if css, err := self.StorDb.GetTPCdrStats(attrs.TPid, attrs.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *css[0]
	}
	return nil
}

type AttrGetTPCdrStatIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries CdrStats identities on specific tariff plan.
func (self *ApierV1) GetTPCdrStatsIds(attrs AttrGetTPCdrStatIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPCdrStats, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific CdrStats on Tariff plan
func (self *ApierV1) RemTPCdrStats(attrs AttrGetTPCdrStats, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPCdrStats, attrs.TPid, map[string]string{"tag": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
