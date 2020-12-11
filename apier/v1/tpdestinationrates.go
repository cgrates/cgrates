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

// This file deals with tp_destination_rates management over APIs

import (
	"github.com/cgrates/cgrates/utils"
)

// SetTPDestinationRate creates a new DestinationRate profile within a tariff plan
func (apierSv1 *APIerSv1) SetTPDestinationRate(attrs *utils.TPDestinationRate, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID, utils.DestinationRates}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPDestinationRates([]*utils.TPDestinationRate{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPDestinationRate struct {
	TPid string // Tariff plan id
	ID   string // Rate id
	utils.Paginator
}

// GetTPDestinationRate queries specific DestinationRate profile on tariff plan
func (apierSv1 *APIerSv1) GetTPDestinationRate(attrs *AttrGetTPDestinationRate, reply *utils.TPDestinationRate) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	drs, err := apierSv1.StorDb.GetTPDestinationRates(attrs.TPid, attrs.ID, &attrs.Paginator)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *drs[0]
	return nil
}

type AttrTPDestinationRateIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPDestinationRateIds queries DestinationRate identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPDestinationRateIds(attrs *AttrGetTPRateIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDestinationRates,
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

// RemoveTPDestinationRate removes specific DestinationRate on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPDestinationRate(attrs *AttrGetTPDestinationRate, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPDestinationRates, attrs.TPid, map[string]string{utils.TagCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
