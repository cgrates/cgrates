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
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestPopulateCostForRoutesConnRefused(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMgr := NewConnManager(cfg)
	fltrS := NewFilterS(cfg, connMgr, nil)
	routes := map[string]*RouteWithWeight{
		"RW": {
			Route: &Route{
				ID:             "local",
				RateProfileIDs: []string{"RP_LOCAL"},
			},
			Weight: 10,
		},
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
		APIOpts: map[string]interface{}{
			utils.OptsRatesProfileIDs: []string{},
		},
	}
	extraOpts := &optsGetRoutes{}
	cfg.RouteSCfg().RateSConns = []string{"*localhost"}
	_, err := populateCostForRoutes(context.Background(), cfg, connMgr, fltrS, routes, ev, extraOpts)
	errExpect := "RATES_ERROR:dial tcp 127.0.0.1:2012: connect: connection refused"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}

func TestLeastCostSorterSortRoutesErr(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cM := NewConnManager(cfg)
	fltrS := NewFilterS(cfg, cM, nil)
	lcs := NewLeastCostSorter(cfg, cM, fltrS)

	expErr := "MANDATORY_IE_MISSING: [connIDs]"
	if _, err := lcs.SortRoutes(context.Background(), "", map[string]*RouteWithWeight{}, &utils.CGREvent{}, &optsGetRoutes{}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%+v>, received <%+v>", expErr, err)
	}

}
