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

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestGetFltrIdxHealthForRateRates(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	rt := &utils.RateProfile{
		Tenant:          utils.CGRateSorg,
		ID:              "TEST_RATE_TEST",
		FilterIDs:       []string{"*string:~*req.Account:dan"},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rt, false, true); err != nil {
		t.Error(err)
	}
	rply, err := GetFltrIdxHealthForRateRates(context.Background(), dm, ltcache.NewCache(50, 60*time.Second, true, nil),
		ltcache.NewCache(40, 30*time.Second, false, nil),
		ltcache.NewCache(20, 20*time.Second, true, nil))
	if err != nil {
		t.Error(err)
	}
	exp := &FilterIHReply{
		MissingObjects: nil,
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  make(map[string][]string),
		MissingFilters: make(map[string][]string),
	}
	if !reflect.DeepEqual(rply, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rply)
	}
}
