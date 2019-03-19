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

// getCacheOpt receive the apiOpt and compare with default value
// overwrite the default if it's present
func (v1 *ApierV1) getCacheOpt(apiOpt *string) string {
	cacheOpt := v1.Config.ApierCfg().DefaultCache
	if apiOpt != nil && *apiOpt != utils.EmptyString {
		cacheOpt = *apiOpt
	}
	return cacheOpt
}

func composeArgsCache(engine.ArgsGetCacheItem) {

}
