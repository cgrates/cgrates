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

	"github.com/cgrates/cgrates/utils"
)

func TestDispatcherHostProfileClone(t *testing.T) {
	dConn := &DispatcherHostProfile{
		ID:        "DSP_1",
		Weight:    30,
		FilterIDs: []string{"*string:Usage:10"},
	}
	eConn := &DispatcherHostProfile{
		ID:        "DSP_1",
		Weight:    30,
		FilterIDs: []string{"*string:Usage:10"},
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

type testRPCHost struct {
	serviceMethod string
	args          interface{}
	reply         interface{}
}

func (v *testRPCHost) Call(serviceMethod string, args interface{}, reply interface{}) error {
	v.serviceMethod = serviceMethod
	v.args = args
	v.reply = reply
	return nil
}

func TestDispatcherHostCall(t *testing.T) {
	tRPC := &testRPCHost{}
	dspHost := DispatcherHost{}
	etRPC := &testRPCHost{
		serviceMethod: utils.AttributeSv1Ping,
		args:          &utils.CGREvent{},
		reply:         utils.StringPointer(""),
	}
	var reply string
	dspHost.rpcConn = tRPC
	if err := dspHost.Call(utils.AttributeSv1Ping, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(*etRPC, *tRPC) {
		t.Errorf("Expected: %s , received: %s", utils.ToJSON(etRPC), utils.ToJSON(tRPC))
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
