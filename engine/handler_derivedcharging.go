/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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

import "github.com/cgrates/cgrates/utils"

// Handles retrieving of DerivedChargers profile based on longest match from AccountingDb
func HandleGetDerivedChargers(ratingStorage RatingStorage, attrs *utils.AttrDerivedChargers) (*utils.DerivedChargers, error) {
	dcs := &utils.DerivedChargers{}
	strictKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject)
	anySubjKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, utils.ANY)
	anyAcntKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, utils.ANY, utils.ANY)
	anyCategKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, utils.ANY, utils.ANY, utils.ANY)
	anyTenantKey := utils.DerivedChargersKey(attrs.Direction, utils.ANY, utils.ANY, utils.ANY, utils.ANY)
	for _, dcKey := range []string{strictKey, anySubjKey, anyAcntKey, anyCategKey, anyTenantKey} {
		if dcsDb, err := ratingStorage.GetDerivedChargers(dcKey, false); err != nil && err != utils.ErrNotFound {
			return nil, err
		} else if dcsDb != nil && DerivedChargersMatchesDest(dcsDb, attrs.Destination) {
			dcs = dcsDb
			break
		}
	}
	return dcs, nil
}

func DerivedChargersMatchesDest(dcs *utils.DerivedChargers, dest string) bool {
	if len(dcs.DestinationIDs) == 0 || dcs.DestinationIDs[utils.ANY] {
		return true
	}
	// check destination ids
	for _, p := range utils.SplitPrefix(dest, MIN_PREFIX_MATCH) {
		if destIDs, err := ratingStorage.GetReverseDestination(p, false); err == nil {
			for _, dId := range destIDs {
				includeDest, found := dcs.DestinationIDs[dId]
				if found {
					return includeDest
				}
			}
		}
	}
	return false
}
