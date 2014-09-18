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

package v1

/*
// Creates a new LcrRules profile within a tariff plan
func (self *ApierV1) SetTPLcrRules(attrs utils.TPLcrRules, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LcrRulesId", "LcrRules"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	for _, action := range attrs.LcrRules {
		requiredFields := []string{"Identifier", "Weight"}

		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:LcrAction:%s:%v", utils.ERR_MANDATORY_IE_MISSING, action.Identifier, missing)
		}
	}
	if err := self.StorDb.SetTPLcrRules(attrs.TPid, map[string][]*utils.TPLcrRule{attrs.LcrRulesId: attrs.LcrRules}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPLcrRules struct {
	TPid  string // Tariff plan id
	LcrId string // Lcr id
}

// Queries specific LcrRules profile on tariff plan
func (self *ApierV1) GetTPLcrRules(attrs AttrGetTPLcrRules, reply *utils.TPLcrRules) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LcrId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if lcrs, err := self.StorDb.GetTpLCRs(attrs.TPid, attrs.LcrId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(acts) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = utils.TPLcrRules{TPid: attrs.TPid, LcrRulesId: attrs.LcrRulesId, LcrRules: lcrs[attrs.LcrRulesId]}
	}
	return nil
}

type AttrGetTPLcrActionIds struct {
	TPid string // Tariff plan id
}

// Queries LcrRules identities on specific tariff plan.
func (self *ApierV1) GetTPLcrActionIds(attrs AttrGetTPLcrActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_LCRS, "id", nil); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific LcrRules on Tariff plan
func (self *ApierV1) RemTPLcrRules(attrs AttrGetTPLcrRules, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LcrRulesId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_LCRS, attrs.TPid, attrs.LcrRulesId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
*/
