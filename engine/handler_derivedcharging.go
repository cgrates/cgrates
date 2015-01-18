/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"github.com/cgrates/cgrates/utils"
)

// Handles retrieving of DerivedChargers profile based on longest match from AccountingDb
func HandleGetDerivedChargers(acntStorage AccountingStorage, attrs utils.AttrDerivedChargers) (utils.DerivedChargers, error) {
	var dcs utils.DerivedChargers
	strictKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject)
	anySubjKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, utils.ANY)
	anyAcntKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, utils.ANY, utils.ANY)
	anyCategKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, utils.ANY, utils.ANY, utils.ANY)
	anyTenantKey := utils.DerivedChargersKey(attrs.Direction, utils.ANY, utils.ANY, utils.ANY, utils.ANY)
	for _, dcKey := range []string{strictKey, anySubjKey, anyAcntKey, anyCategKey, anyTenantKey} {
		if dcsDb, err := acntStorage.GetDerivedChargers(dcKey, false); err != nil && err.Error() != utils.ERR_NOT_FOUND {
			return nil, err
		} else if dcsDb != nil {
			dcs = dcsDb
			break
		}
	}
	return dcs, nil
}
