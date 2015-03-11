/*
Real-Time Charging System for Telecom Environments
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

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var rds *RedisStorage
var err error

func TestConnectRedis(t *testing.T) {
	if !*testLocal {
		return
	}
	cfg, _ = config.NewDefaultCGRConfig()
	rds, err = NewRedisStorage(fmt.Sprintf("%s:%s", cfg.RatingDBHost, cfg.RatingDBPort), 4, cfg.RatingDBPass, cfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
}

func TestFlush(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := rds.Flush(""); err != nil {
		t.Error("Failed to Flush redis database", err.Error())
	}
	rds.CacheAccounting(nil, nil, nil, nil)
}

func TestSetGetDerivedCharges(t *testing.T) {
	if !*testLocal {
		return
	}
	keyCharger1 := utils.ConcatenatedKey("*out", "cgrates.org", "call", "dan", "dan")
	charger1 := utils.DerivedChargers{
		&utils.DerivedCharger{RunId: "extra1", ReqTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunId: "extra2", ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}
	if err := rds.SetDerivedChargers(keyCharger1, charger1); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	// Try retrieving from cache, should not be in yet
	if _, err := rds.GetDerivedChargers(keyCharger1, false); err == nil {
		t.Error("DerivedCharger should not be in the cache")
	}
	// Retrieve from db
	if rcvCharger, err := rds.GetDerivedChargers(keyCharger1, true); err != nil {
		t.Error("Error when retrieving DerivedCHarger", err.Error())
	} else if !reflect.DeepEqual(rcvCharger, charger1) {
		t.Errorf("Expecting %v, received: %v", charger1, rcvCharger)
	}
	// Retrieve from cache
	if rcvCharger, err := rds.GetDerivedChargers(keyCharger1, false); err != nil {
		t.Error("Error when retrieving DerivedCHarger", err.Error())
	} else if !reflect.DeepEqual(rcvCharger, charger1) {
		t.Errorf("Expecting %v, received: %v", charger1, rcvCharger)
	}
}
