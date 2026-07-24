// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/utils"
)

// Creates a new timing within a tariff plan
func (self *APIerSv1) SetTPTiming(attrs utils.ApierTPTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID", "Years", "Months", "MonthDays", "WeekDays", "Time"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPTimings([]*utils.ApierTPTiming{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPTiming struct {
	TPid string // Tariff plan id
	ID   string // Timing id
}

// Queries specific Timing on Tariff plan
func (self *APIerSv1) GetTPTiming(attrs AttrGetTPTiming, reply *utils.ApierTPTiming) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if tms, err := self.StorDb.GetTPTimings(attrs.TPid, attrs.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *tms[0]
	}
	return nil
}

type AttrGetTPTimingIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// Queries timing identities on specific tariff plan.
func (self *APIerSv1) GetTPTimingIds(attrs AttrGetTPTimingIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPTimings,
		utils.TPDistinctIds{"tag"}, nil, &attrs.PaginatorWithSearch); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Timing on Tariff plan
func (self *APIerSv1) RemoveTPTiming(attrs AttrGetTPTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPTimings, attrs.TPid, map[string]string{"tag": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
