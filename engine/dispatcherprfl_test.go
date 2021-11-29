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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestDispatcherHostProfileClone(t *testing.T) {
	dConn := &DispatcherHostProfile{
		ID:        "DSP_1",
		Weight:    30,
		FilterIDs: []string{"*string:Usage:10"},
		Params:    map[string]interface{}{"param1": 1},
	}
	eConn := &DispatcherHostProfile{
		ID:        "DSP_1",
		Weight:    30,
		FilterIDs: []string{"*string:Usage:10"},
		Params:    map[string]interface{}{"param1": 1},
	}
	d2Conn := dConn.Clone()
	d2Conn.ID = "DSP_4"
	if !reflect.DeepEqual(eConn, dConn) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eConn), utils.ToJSON(dConn))
	}
}

func TestDispatcherHostProfilesReorderFromIndex(t *testing.T) {
	dConns := DispatcherHostProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	eConns := DispatcherHostProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	if dConns.ReorderFromIndex(0); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", eConns, dConns)
	}
	dConns = DispatcherHostProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	if dConns.ReorderFromIndex(3); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", eConns, dConns)
	}
	dConns = DispatcherHostProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	eConns = DispatcherHostProfiles{
		{ID: "DSP_3", Weight: 10},
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
	}
	if dConns.ReorderFromIndex(2); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", eConns, dConns)
	}
	dConns = DispatcherHostProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	eConns = DispatcherHostProfiles{
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
		{ID: "DSP_1", Weight: 30},
	}
	if dConns.ReorderFromIndex(1); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eConns), utils.ToJSON(dConns))
	}
}

func TestDispatcherHostProfilesShuffle(t *testing.T) {
	dConns := DispatcherHostProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	oConns := DispatcherHostProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	if dConns.Shuffle(); dConns[0] == oConns[0] ||
		dConns[1] == oConns[1] || dConns[2] == oConns[2] {
		t.Errorf("received: %s", utils.ToJSON(dConns))
	}
}

func TestDispatcherHostProfilesSort(t *testing.T) {
	dConns := DispatcherHostProfiles{
		{ID: "DSP_3", Weight: 10},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_1", Weight: 30},
	}
	eConns := DispatcherHostProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	if dConns.Sort(); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eConns), utils.ToJSON(dConns))
	}
	dConns = DispatcherHostProfiles{
		{ID: "DSP_3", Weight: 10},
		{ID: "DSP_5", Weight: 50},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_4", Weight: 40},
		{ID: "DSP_1", Weight: 30},
	}
	eConns = DispatcherHostProfiles{
		{ID: "DSP_5", Weight: 50},
		{ID: "DSP_4", Weight: 40},
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	if dConns.Sort(); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eConns), utils.ToJSON(dConns))
	}
	dConns.Shuffle()
	if dConns.Sort(); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eConns), utils.ToJSON(dConns))
	}
}

func TestDispatcherHostProfilesClone(t *testing.T) {
	dConns := DispatcherHostProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10, FilterIDs: []string{"*string:Usage:10"}},
	}
	eConns := DispatcherHostProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10, FilterIDs: []string{"*string:Usage:10"}},
	}
	d2Conns := dConns.Clone()
	d2Conns[0].ID = "DSP_4"
	if !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eConns), utils.ToJSON(dConns))
	}
}

func TestDispatcherHostProfilesConnIDs(t *testing.T) {
	dConns := DispatcherHostProfiles{
		{ID: "DSP_5", Weight: 50},
		{ID: "DSP_4", Weight: 40},
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	eConnIDs := []string{"DSP_5", "DSP_4", "DSP_1", "DSP_2", "DSP_3"}
	if dConnIDs := dConns.HostIDs(); !reflect.DeepEqual(eConnIDs, dConnIDs) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eConnIDs), utils.ToJSON(dConnIDs))
	}
}

func TestDispatcherProfileTenantID(t *testing.T) {
	dProf := DispatcherProfile{
		Tenant: "cgrates.org",
		ID:     "DISP_1",
	}
	eTenantID := utils.ConcatenatedKey("cgrates.org", "DISP_1")
	if dTenantID := dProf.TenantID(); !reflect.DeepEqual(eTenantID, dTenantID) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eTenantID), utils.ToJSON(dTenantID))
	}
}

func TestDispatcherProfilesSort(t *testing.T) {
	dProf := DispatcherProfiles{
		{ID: "DSP_3", Weight: 10},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_1", Weight: 30},
	}
	eProf := DispatcherProfiles{
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	if dProf.Sort(); !reflect.DeepEqual(eProf, dProf) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eProf), utils.ToJSON(dProf))
	}
	dProf = DispatcherProfiles{
		{ID: "DSP_3", Weight: 10},
		{ID: "DSP_5", Weight: 50},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_4", Weight: 40},
		{ID: "DSP_1", Weight: 30},
	}
	eProf = DispatcherProfiles{
		{ID: "DSP_5", Weight: 50},
		{ID: "DSP_4", Weight: 40},
		{ID: "DSP_1", Weight: 30},
		{ID: "DSP_2", Weight: 20},
		{ID: "DSP_3", Weight: 10},
	}
	if dProf.Sort(); !reflect.DeepEqual(eProf, dProf) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eProf), utils.ToJSON(dProf))
	}
}

func TestDispatcherHostIDsProfilesReorderFromIndex(t *testing.T) {
	dConns := DispatcherHostIDs{"DSP_1", "DSP_2", "DSP_3"}
	eConns := DispatcherHostIDs{"DSP_1", "DSP_2", "DSP_3"}
	if dConns.ReorderFromIndex(0); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", eConns, dConns)
	}
	dConns = DispatcherHostIDs{"DSP_1", "DSP_2", "DSP_3"}
	if dConns.ReorderFromIndex(3); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", eConns, dConns)
	}
	dConns = DispatcherHostIDs{"DSP_1", "DSP_2", "DSP_3"}
	eConns = DispatcherHostIDs{"DSP_3", "DSP_1", "DSP_2"}
	if dConns.ReorderFromIndex(2); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", eConns, dConns)
	}
	dConns = DispatcherHostIDs{"DSP_1", "DSP_2", "DSP_3"}
	eConns = DispatcherHostIDs{"DSP_2", "DSP_3", "DSP_1"}
	if dConns.ReorderFromIndex(1); !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eConns), utils.ToJSON(dConns))
	}
}

func TestDispatcherHostIDsProfilesShuffle(t *testing.T) {
	dConns := DispatcherHostIDs{"DSP_1", "DSP_2", "DSP_3", "DSP_4", "DSP_5", "DSP_6", "DSP_7", "DSP_8"}
	oConns := DispatcherHostIDs{"DSP_1", "DSP_2", "DSP_3", "DSP_4", "DSP_5", "DSP_6", "DSP_7", "DSP_8"}
	if dConns.Shuffle(); reflect.DeepEqual(dConns, oConns) {
		t.Errorf("received: %s", utils.ToJSON(dConns))
	}
}

func TestDispatcherHostIDsProfilesClone(t *testing.T) {
	dConns := DispatcherHostIDs{"DSP_1", "DSP_2", "DSP_3"}
	eConns := DispatcherHostIDs{"DSP_1", "DSP_2", "DSP_3"}
	d2Conns := dConns.Clone()
	d2Conns[0] = "DSP_4"
	if !reflect.DeepEqual(eConns, dConns) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eConns), utils.ToJSON(dConns))
	}
}

func TestDispatcherProfileSet(t *testing.T) {
	dp := DispatcherProfile{}
	exp := DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weight:    10,
		Strategy:  utils.MetaRandom,
		StrategyParams: map[string]interface{}{
			"opt1": "val1",
			"opt2": "val1",
			"opt3": "val1",
		},
		Hosts: DispatcherHostProfiles{
			{
				ID:        "host1",
				FilterIDs: []string{"fltr1"},
				Weight:    10,
				Blocker:   true,
				Params: map[string]interface{}{
					"param1": "val1",
					"param2": "val1",
				},
			},
			{
				Params: map[string]interface{}{
					"param3": "val1",
				},
			},
		},
	}
	if err := dp.Set([]string{}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{"NotAField"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{"NotAField", "1"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := dp.Set([]string{utils.Tenant}, "cgrates.org", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ID}, "ID", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.FilterIDs}, "fltr1;*string:~*req.Account:1001", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Weight}, 10, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Strategy}, utils.MetaRandom, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.StrategyParams}, "opt1:val1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.StrategyParams + "[opt2]"}, "val1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.StrategyParams, "opt3"}, "val1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Hosts, utils.ID}, "host1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Hosts, utils.FilterIDs}, "fltr1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Hosts, utils.Weight}, "10", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Hosts, utils.Blocker}, "true", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Hosts, utils.Params}, "param1:val1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Hosts, utils.Params + "[param2]"}, "val1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Hosts, utils.Params, "param3"}, "val1", true, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if err := dp.Set([]string{utils.Hosts, "Wrong"}, "val1", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Hosts, "Wrong", "path"}, "", true, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}

func TestDispatcherHostSet(t *testing.T) {
	dp := DispatcherHost{RemoteHost: &config.RemoteHost{}}
	exp := DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:                "ID",
			Address:           "127.0.0.1",
			Transport:         utils.MetaJSON,
			ConnectAttempts:   1,
			Reconnects:        1,
			ConnectTimeout:    time.Nanosecond,
			ReplyTimeout:      time.Nanosecond,
			TLS:               true,
			ClientKey:         "key",
			ClientCertificate: "ce",
			CaCertificate:     "ca",
		},
	}
	if err := dp.Set([]string{}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{"NotAField"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{"NotAField", "1"}, "", false, utils.EmptyString); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := dp.Set([]string{utils.Tenant}, "cgrates.org", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ID}, "ID", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Address}, "127.0.0.1", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Transport}, utils.MetaJSON, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ConnectAttempts}, 1, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Reconnects}, 1, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ConnectTimeout}, 1, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ReplyTimeout}, 1, false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.TLS}, true, false, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if err := dp.Set([]string{utils.ClientKey}, "key", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ClientCertificate}, "ce", false, utils.EmptyString); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.CaCertificate}, "ca", false, utils.EmptyString); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}
