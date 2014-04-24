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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Transparently handles merging between storage data and configuration, useful as local handler
func HandleGetDerivedChargers(acntStorage AccountingStorage, cfg *config.CGRConfig, attrs utils.AttrDerivedChargers) (utils.DerivedChargers, error) {
	var dcs utils.DerivedChargers
	var err error
	strictKey := utils.DerivedChargersKey(attrs.Tenant, attrs.Tor, attrs.Direction, attrs.Account, attrs.Subject)
	anySubjKey := utils.DerivedChargersKey(attrs.Tenant, attrs.Tor, attrs.Direction, attrs.Account, utils.ANY)
	for _, dcKey := range []string{strictKey, anySubjKey} {
		if dcsDb, err := acntStorage.GetDerivedChargers(dcKey, false); err != nil && err.Error() != utils.ERR_NOT_FOUND {
			return nil, err
		} else if dcsDb != nil {
			dcs = dcsDb
			break
		}
	}
	if dcs == nil {
		dcs = cfg.DerivedChargers
		return dcs, nil
	}
	if cfg.CombinedDerivedChargers {
		for _, cfgDc := range cfg.DerivedChargers {
			if dcs, err = dcs.Append(cfgDc); err != nil {
				return nil, err
			}
		}
	}
	return dcs, nil
}
