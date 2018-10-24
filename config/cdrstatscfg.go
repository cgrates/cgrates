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

package config

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Cdrstats config section
type CdrStatsCfg struct {
	CDRStatsEnabled      bool          // Enable CDR Stats service
	CDRStatsSaveInterval time.Duration // Save interval duration
}

//loadFromJsonCfg loads CdrStatsconfig from JsonCfg
func (cdrstsCfg *CdrStatsCfg) loadFromJsonCfg(jsnCdrstatsCfg *CdrStatsJsonCfg) (err error) {
	if jsnCdrstatsCfg == nil {
		return nil
	}
	if jsnCdrstatsCfg.Enabled != nil {
		cdrstsCfg.CDRStatsEnabled = *jsnCdrstatsCfg.Enabled

	}
	if jsnCdrstatsCfg.Save_Interval != nil {
		if cdrstsCfg.CDRStatsSaveInterval, err = utils.ParseDurationWithNanosecs(*jsnCdrstatsCfg.Save_Interval); err != nil {
			return err
		}
	}
	return nil
}
