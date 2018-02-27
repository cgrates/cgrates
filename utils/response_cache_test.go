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
package utils

import (
	"testing"
	"time"
)

func TestRCacheSetGet(t *testing.T) {
	rc := NewResponseCache(5 * time.Second)
	rc.Cache("test", &ResponseCacheItem{Value: "best"})
	v, err := rc.Get("test")
	if err != nil || v.Value.(string) != "best" {
		t.Error("Error retriving response cache: ", v, err)
	}
}

/*
func TestRCacheExpire(t *testing.T) {
	rc := NewResponseCache(1 * time.Microsecond)
	rc.Cache("test", &CacheItem{Value: "best"})
	time.Sleep(3 * time.Millisecond)
	o, err := rc.Get("test")
	if err == nil {
		t.Error("Error expiring response cache: ", o)
	}
}
*/
