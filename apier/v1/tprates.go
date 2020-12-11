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

// This file deals with tp_rates management over APIs

import (
	"github.com/cgrates/cgrates/utils"
)

// SetTPRate creates a new rate within a tariff plan
func (apierSv1 *APIerSv1) SetTPRate(attrs *utils.TPRateRALs, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID, utils.RateSlots}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPRates([]*utils.TPRateRALs{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPRate struct {
	TPid string // Tariff plan id
	ID   string // Rate id
}

// GetTPRate queries specific Rate on tariff plan
func (apierSv1 *APIerSv1) GetTPRate(attrs *AttrGetTPRate, reply *utils.TPRateRALs) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rs, err := apierSv1.StorDb.GetTPRates(attrs.TPid, attrs.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rs[0]
	return nil
}

type AttrGetTPRateIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPRateIds queries rate identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPRateIds(attrs *AttrGetTPRateIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRates,
		utils.TPDistinctIds{utils.TagCfg}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPRate removes specific Rate on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPRate(attrs *AttrGetTPRate, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPRates, attrs.TPid, map[string]string{utils.TagCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
