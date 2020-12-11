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

// SetTPActionTriggers creates a new ActionTriggers profile within a tariff plan
func (apierSv1 *APIerSv1) SetTPActionTriggers(attrs *utils.TPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPActionTriggers([]*utils.TPActionTriggers{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// AttrGetTPActionTriggers used as argument for GetTPActionTriggers
type AttrGetTPActionTriggers struct {
	TPid string // Tariff plan id
	ID   string // ActionTrigger id
}

// GetTPActionTriggers queries specific ActionTriggers profile on tariff plan
func (apierSv1 *APIerSv1) GetTPActionTriggers(attrs *AttrGetTPActionTriggers, reply *utils.TPActionTriggers) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ats, err := apierSv1.StorDb.GetTPActionTriggers(attrs.TPid, attrs.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *ats[0]
	return nil
}

// AttrGetTPActionTriggerIds used as argument for GetTPActionTriggerIds and RemoveTPActionTriggers
type AttrGetTPActionTriggerIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPActionTriggerIds queries ActionTriggers identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPActionTriggerIds(attrs *AttrGetTPActionTriggerIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPActionTriggers,
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

// RemoveTPActionTriggers removes specific ActionTriggers on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPActionTriggers(attrs *AttrGetTPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	err := apierSv1.StorDb.RemTpData(utils.TBLTPActionTriggers,
		attrs.TPid, map[string]string{utils.TagCfg: attrs.ID})
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
