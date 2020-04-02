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

// SetTPTiming creates a new timing within a tariff plan
func (api *APIerSv1) SetTPTiming(attrs utils.ApierTPTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID", "Years", "Months", "MonthDays", "WeekDays", "Time"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.SetTPTimings([]*utils.ApierTPTiming{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPTiming struct {
	TPid string // Tariff plan id
	ID   string // Timing id
}

// GetTPTiming queries specific Timing on Tariff plan
func (api *APIerSv1) GetTPTiming(attrs AttrGetTPTiming, reply *utils.ApierTPTiming) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tms, err := api.StorDb.GetTPTimings(attrs.TPid, attrs.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *tms[0]
	return nil
}

type AttrGetTPTimingIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPTimingIds queries timing identities on specific tariff plan.
func (api *APIerSv1) GetTPTimingIds(attrs AttrGetTPTimingIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPTimings,
		utils.TPDistinctIds{"tag"}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPTiming removes specific Timing on Tariff plan
func (api *APIerSv1) RemoveTPTiming(attrs AttrGetTPTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.RemTpData(utils.TBLTPTimings, attrs.TPid, map[string]string{"tag": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
