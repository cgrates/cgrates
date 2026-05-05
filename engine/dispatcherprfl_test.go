/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
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
	args          any
	reply         any
}

func (v *testRPCHost) Call(ctx *context.Context, serviceMethod string, args any, reply any) error {
	v.serviceMethod = serviceMethod
	v.args = args
	v.reply = reply
	return nil
}

func TestDispatcherHostCall(t *testing.T) {
	tRPC := &testRPCHost{}
	dspHost := &DispatcherHost{}
	etRPC := &testRPCHost{
		serviceMethod: utils.AttributeSv1Ping,
		args:          &utils.CGREvent{},
		reply:         utils.StringPointer(""),
	}
	var reply string
	dspHost.rpcConn = tRPC
	if err := dspHost.Call(context.Background(), utils.AttributeSv1Ping, &utils.CGREvent{}, &reply); err != nil {
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

func TestDispatcherHostProfileCloneWithParams(t *testing.T) {
	dC := &DispatcherHostProfile{
		ID:      "testID",
		Weight:  10,
		Blocker: false,
		Params: map[string]any{
			"param1": "value of param1",
			"param2": "value of param2",
		},
	}

	exp := &DispatcherHostProfile{
		ID:      "testID",
		Weight:  10,
		Blocker: false,
		Params: map[string]any{
			"param1": "value of param1",
			"param2": "value of param2",
		},
	}
	rcv := dC.Clone()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v", exp, rcv)
	}
}

func TestDispatcherHostCallErr(t *testing.T) {
	dH := &DispatcherHost{
		Tenant: "testTenant",
		RemoteHost: &config.RemoteHost{
			ID:        "testID",
			Address:   "",
			Transport: "",
			TLS:       false,
		},
	}
	var reply string
	if err := dH.Call(context.Background(), utils.AttributeSv1Ping, &utils.CGREvent{}, &reply); err == nil || err.Error() != "dial tcp: missing address" {
		t.Error(err)
	}
}

func TestDispatcherHostIDsClone(t *testing.T) {
	tests := []struct {
		name              string
		dispatcherHostIDs DispatcherHostIDs
	}{
		{
			name:              "Complete DispatcherHostIDs",
			dispatcherHostIDs: DispatcherHostIDs{"testID1", "testID2"},
		},
		{
			name:              "Empty DispatcherHostIDs",
			dispatcherHostIDs: DispatcherHostIDs{},
		},
		{
			name:              "Nil DispatcherHostIDs",
			dispatcherHostIDs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result := tt.dispatcherHostIDs.Clone()

			if !reflect.DeepEqual(result, tt.dispatcherHostIDs) {
				t.Errorf("Clone() = %v, want %v", result, tt.dispatcherHostIDs)
			}
		})
		t.Run(tt.name, func(t *testing.T) {

			cache := tt.dispatcherHostIDs.CacheClone()

			if !reflect.DeepEqual(cache, tt.dispatcherHostIDs) {
				t.Errorf("CacheClone() = %v, want %v", cache, tt.dispatcherHostIDs)
			}

			_, ok := cache.(DispatcherHostIDs)
			if !ok {
				t.Errorf("CacheClone() returned type %T, want DispatcherHostIDs", cache)
				return
			}
		})
	}
}

func TestDispatcherProfileClone(t *testing.T) {
	tests := []struct {
		name              string
		dispatcherProfile *DispatcherProfile
	}{
		{
			name: "Complete DispatcherProfile",
			dispatcherProfile: &DispatcherProfile{
				Tenant:     "cgrates.org",
				ID:         "Dsp",
				Subsystems: []string{"*any"},
				FilterIDs:  []string{"FLTR_ACNT", "FLTR_DST"},
				Strategy:   utils.MetaFirst,
				ActivationInterval: &utils.ActivationInterval{
					ExpiryTime: time.Now(),
				},
				StrategyParams: map[string]any{},
				Weight:         20,
				Hosts: DispatcherHostProfiles{
					{
						ID:        "C1",
						FilterIDs: []string{},
						Weight:    10,
						Blocker:   false,
					},
					{
						ID:        "C2",
						FilterIDs: []string{},
						Weight:    10,
						Blocker:   false,
					},
				},
			},
		},
		{
			name: "Nil fields",
			dispatcherProfile: &DispatcherProfile{
				Tenant:             "cgrates.org",
				ID:                 "Dsp",
				Subsystems:         nil,
				FilterIDs:          nil,
				Strategy:           utils.MetaFirst,
				ActivationInterval: nil,
				StrategyParams:     nil,
				Weight:             20,
				Hosts:              nil,
			},
		},
		{
			name: "Empty fields",
			dispatcherProfile: &DispatcherProfile{
				Tenant:             "cgrates.org",
				ID:                 "Dsp",
				Subsystems:         []string{},
				FilterIDs:          []string{},
				Strategy:           utils.MetaFirst,
				ActivationInterval: &utils.ActivationInterval{},
				StrategyParams:     map[string]any{},
				Weight:             20,
				Hosts:              DispatcherHostProfiles{},
			},
		},
		{
			name:              "Nil Case",
			dispatcherProfile: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dispatcherProfile.Clone()

			if !reflect.DeepEqual(got, tt.dispatcherProfile) {
				t.Errorf("Clone() = %#+v, want %#+v", got, tt.dispatcherProfile)
			}

			if got != nil && got == tt.dispatcherProfile {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			cache := tt.dispatcherProfile.CacheClone()

			if !reflect.DeepEqual(cache, tt.dispatcherProfile) {
				t.Errorf("CacheClone() = %v, want %v", cache, tt.dispatcherProfile)
			}

			_, ok := cache.(*DispatcherProfile)

			if !ok {
				t.Errorf("CacheClone() returned type %T, want *DispatcherProfile", cache)
				return
			}
		})
	}
}
