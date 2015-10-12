/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new DerivedCharges profile within a tariff plan
func (self *ApierV1) SetTPDerivedChargers(attrs utils.TPDerivedChargers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Direction", "Tenant", "Category", "Account", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	/*for _, action := range attrs.DerivedCharges {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:DerivedCharge:%s:%v", utils.ERR_MANDATORY_IE_MISSING, action.Identifier, missing)
		}
	}*/
	dc := engine.APItoModelDerivedCharger(&attrs)
	if err := self.StorDb.SetTpDerivedChargers(dc); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPDerivedChargers struct {
	TPid              string // Tariff plan id
	DerivedChargersId string // DerivedCharge id
}

// Queries specific DerivedCharge on tariff plan
func (self *ApierV1) GetTPDerivedChargers(attrs AttrGetTPDerivedChargers, reply *utils.TPDerivedChargers) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DerivedChargersId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tmpDc := &utils.TPDerivedChargers{TPid: attrs.TPid}
	if err := tmpDc.SetDerivedChargersId(attrs.DerivedChargersId); err != nil {
		return err
	}
	dcs := engine.APItoModelDerivedCharger(tmpDc)
	if sgs, err := self.StorDb.GetTpDerivedChargers(&dcs[0]); err != nil {
		return utils.NewErrServerError(err)
	} else if len(sgs) == 0 {
		return utils.ErrNotFound
	} else {
		dcsMap, err := engine.TpDerivedChargers(dcs).GetDerivedChargers()
		if err != nil {
			return err
		}
		*reply = *dcsMap[attrs.DerivedChargersId]
	}
	return nil
}

type AttrGetTPDerivedChargeIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries DerivedCharges identities on specific tariff plan.
func (self *ApierV1) GetTPDerivedChargerIds(attrs AttrGetTPDerivedChargeIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_DERIVED_CHARGERS, utils.TPDistinctIds{"loadid", "direction", "tenant", "category", "account", "subject"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific DerivedCharges on Tariff plan
func (self *ApierV1) RemTPDerivedChargers(attrs AttrGetTPDerivedChargers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DerivedChargesId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tmpDc := engine.TpDerivedCharger{}
	if err := tmpDc.SetDerivedChargersId(attrs.DerivedChargersId); err != nil {
		return err
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_DERIVED_CHARGERS, attrs.TPid, map[string]string{"loadid": tmpDc.Loadid, "direction": tmpDc.Direction, "tenant": tmpDc.Tenant, "category": tmpDc.Category, "account": tmpDc.Account, "subject": tmpDc.Subject}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
