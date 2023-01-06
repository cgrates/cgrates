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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestFilterHelpersWeightFromDynamics(t *testing.T) {
	var expected float64 = 64
	ctx := context.Background()
	dWs := []*utils.DynamicWeight{
		{
			Weight: 64,
		},
	}
	fltrs := &FilterS{}
	tnt := utils.CGRateSorg
	ev := utils.MapStorage{}
	result, err := WeightFromDynamics(ctx, dWs, fltrs, tnt, ev)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestFilterHelpersWeightFromDynamicsErr(t *testing.T) {

	ctx := context.Background()
	dWs := []*utils.DynamicWeight{
		{
			FilterIDs: []string{"*stirng:~*req.Account:1001"},
			Weight:    64,
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	cM := NewConnManager(cfg)
	fltrs := NewFilterS(cfg, cM, dm)
	tnt := utils.CGRateSorg
	ev := utils.MapStorage{}

	expErr := "NOT_IMPLEMENTED:*stirng"
	_, err := WeightFromDynamics(ctx, dWs, fltrs, tnt, ev)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%V>", expErr, err)
	}

}

func TestBlockerFromDynamicsErr(t *testing.T) {

	ctx := context.Background()
	dBs := []*utils.DynamicBlocker{
		{
			FilterIDs: []string{"*stirng:~*req.Account:1001"},
			Blocker:   true,
		},
	}
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cM := NewConnManager(cfg)
	fltrs := NewFilterS(cfg, cM, dm)
	tnt := utils.CGRateSorg
	ev := utils.MapStorage{}

	expErr := "NOT_IMPLEMENTED:*stirng"
	if _, err := BlockerFromDynamics(ctx, dBs, fltrs, tnt, ev); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received error <%V>", expErr, err)
	}

}
