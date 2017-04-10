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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new DerivedCharges profile within a tariff plan
func (self *ApierV1) SetTPDerivedChargers(attrs utils.TPDerivedChargers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Direction", "Tenant", "Category", "Account", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPDerivedChargers([]*utils.TPDerivedChargers{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
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
	filter := &utils.TPDerivedChargers{TPid: attrs.TPid}
	filter.SetDerivedChargersId(attrs.DerivedChargersId)
	if dcs, err := self.StorDb.GetTPDerivedChargers(filter); err != nil {
		return utils.NewErrServerError(err)
	} else if len(dcs) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = *dcs[0]
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
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDerivedChargers, utils.TPDistinctIds{"loadid", "direction", "tenant", "category", "account", "subject"}, nil, &attrs.Paginator); err != nil {
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
	if err := self.StorDb.RemTpData(utils.TBLTPDerivedChargers, attrs.TPid, map[string]string{"loadid": tmpDc.Loadid, "direction": tmpDc.Direction, "tenant": tmpDc.Tenant, "category": tmpDc.Category, "account": tmpDc.Account, "subject": tmpDc.Subject}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
