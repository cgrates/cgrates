// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

// This file deals with tp_rates management over APIs

import (
	"github.com/cgrates/cgrates/utils"
)

// Creates a new rate within a tariff plan
func (self *APIerSv1) SetTPRate(attrs utils.TPRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID", "RateSlots"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPRates([]*utils.TPRate{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPRate struct {
	TPid string // Tariff plan id
	ID   string // Rate id
}

// Queries specific Rate on tariff plan
func (self *APIerSv1) GetTPRate(attrs AttrGetTPRate, reply *utils.TPRate) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rs, err := self.StorDb.GetTPRates(attrs.TPid, attrs.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *rs[0]
	}
	return nil
}

type AttrGetTPRateIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// Queries rate identities on specific tariff plan.
func (self *APIerSv1) GetTPRateIds(attrs AttrGetTPRateIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRates,
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

// Removes specific Rate on Tariff plan
func (self *APIerSv1) RemoveTPRate(attrs AttrGetTPRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPRates, attrs.TPid, map[string]string{"tag": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
