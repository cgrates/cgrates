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

/*
// Creates a new DerivedCharges profile within a tariff plan
func (self *ApierV1) SetTPDerivedCharges(attrs utils.TPDerivedCharges, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DerivedChargesId", "DerivedCharges"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	for _, action := range attrs.DerivedCharges {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:DerivedCharge:%s:%v", utils.ERR_MANDATORY_IE_MISSING, action.Identifier, missing)
		}
	}
	if err := self.StorDb.SetTPDerivedCharges(attrs.TPid, map[string][]*utils.TPDerivedCharge{attrs.DerivedChargesId: attrs.DerivedCharges}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPDerivedCharges struct {
	TPid             string // Tariff plan id
	DerivedChargesId string // DerivedCharge id
}

// Queries specific DerivedCharge on tariff plan
func (self *ApierV1) GetTPDerivedCharges(attrs AttrGetTPDerivedCharges, reply *utils.TPDerivedCharges) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DerivedChargesId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if sgs, err := self.StorDb.GetTpDerivedCharges(attrs.TPid, attrs.DerivedChargesId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(sgs) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = utils.TPDerivedCharges{TPid: attrs.TPid, DerivedChargesId: attrs.DerivedChargesId, DerivedCharges: sgs[attrs.DerivedChargesId]}
	}
	return nil
}

type AttrGetTPDerivedChargeIds struct {
	TPid string // Tariff plan id
}

// Queries DerivedCharges identities on specific tariff plan.
func (self *ApierV1) GetTPDerivedChargeIds(attrs AttrGetTPDerivedChargeIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_DERIVED_CHARGES, "id", nil); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific DerivedCharges on Tariff plan
func (self *ApierV1) RemTPDerivedCharges(attrs AttrGetTPDerivedCharges, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DerivedChargesId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_DERIVED_CHARGES, attrs.TPid, attrs.DerivedChargesId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
*/
