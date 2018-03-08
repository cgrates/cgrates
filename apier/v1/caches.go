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
	"time"

	"github.com/cgrates/cgrates/engine"
)

func NewCacheSv1(cacheS *engine.CacheS) *CacheSv1 {
	return &CacheSv1{cacheS: cacheS}
}

// Exports RPC from CacheS
type CacheSv1 struct {
	cacheS *engine.CacheS
}

// GetItemExpiryTime re
func (chSv1 *CacheSv1) GetItemExpiryTime(args *engine.ArgsGetCacheItem,
	reply *time.Time) error {
	return chSv1.cacheS.V1GetItemExpiryTime(args, reply)
}
