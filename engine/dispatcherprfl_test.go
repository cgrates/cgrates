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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestDispatcherHostProfileClone(t *testing.T) {
	dConn := &DispatcherHostProfile{
		ID:        "DSP_1",
		Weight:    30,
		FilterIDs: []string{"*string:Usage:10"},
		Params:    map[string]any{"param1": 1},
	}
	eConn := &DispatcherHostProfile{
		ID:        "DSP_1",
		Weight:    30,
		FilterIDs: []string{"*string:Usage:10"},
		Params:    map[string]any{"param1": 1},
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

func TestDispatcherHostSet(t *testing.T) {
	dp := DispatcherHost{RemoteHost: &config.RemoteHost{}}
	exp := DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:                   "ID",
			Address:              "127.0.0.1",
			Transport:            utils.MetaJSON,
			ConnectAttempts:      1,
			Reconnects:           1,
			MaxReconnectInterval: 1,
			ConnectTimeout:       time.Nanosecond,
			ReplyTimeout:         time.Nanosecond,
			TLS:                  true,
			ClientKey:            "key",
			ClientCertificate:    "ce",
			CaCertificate:        "ca",
		},
	}
	if err := dp.Set([]string{}, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{"NotAField"}, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{"NotAField", "1"}, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := dp.Set([]string{utils.Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Address}, "127.0.0.1", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Transport}, utils.MetaJSON, false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ConnectAttempts}, 1, false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Reconnects}, 1, false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.MaxReconnectInterval}, 1, false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ConnectTimeout}, 1, false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ReplyTimeout}, 1, false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.TLS}, true, false); err != nil {
		t.Error(err)
	}

	if err := dp.Set([]string{utils.ClientKey}, "key", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ClientCertificate}, "ce", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.CaCertificate}, "ca", false); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}

func TestDispatcherHostAsInterface(t *testing.T) {
	dh := DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:                   "ID",
			Address:              "127.0.0.1",
			Transport:            utils.MetaJSON,
			ConnectAttempts:      1,
			Reconnects:           1,
			MaxReconnectInterval: 1,
			ConnectTimeout:       time.Nanosecond,
			ReplyTimeout:         time.Nanosecond,
			TLS:                  true,
			ClientKey:            "key",
			ClientCertificate:    "ce",
			CaCertificate:        "ca",
		},
	}
	if _, err := dh.FieldAsInterface(nil); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := dh.FieldAsInterface([]string{"field"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := dh.FieldAsInterface([]string{"field", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := dh.FieldAsInterface([]string{utils.Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := utils.ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if val, err := dh.FieldAsInterface([]string{utils.Address}); err != nil {
		t.Fatal(err)
	} else if exp := dh.Address; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.Transport}); err != nil {
		t.Fatal(err)
	} else if exp := dh.Transport; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.ConnectAttempts}); err != nil {
		t.Fatal(err)
	} else if exp := dh.ConnectAttempts; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.Reconnects}); err != nil {
		t.Fatal(err)
	} else if exp := dh.Reconnects; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.MaxReconnectInterval}); err != nil {
		t.Fatal(err)
	} else if exp := dh.MaxReconnectInterval; exp != val {
		t.Errorf("Expected %+v \n but received \n %+v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.ConnectTimeout}); err != nil {
		t.Fatal(err)
	} else if exp := dh.ConnectTimeout; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.ReplyTimeout}); err != nil {
		t.Fatal(err)
	} else if exp := dh.ReplyTimeout; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.TLS}); err != nil {
		t.Fatal(err)
	} else if exp := dh.TLS; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.ClientKey}); err != nil {
		t.Fatal(err)
	} else if exp := dh.ClientKey; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.ClientCertificate}); err != nil {
		t.Fatal(err)
	} else if exp := dh.ClientCertificate; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := dh.FieldAsInterface([]string{utils.CaCertificate}); err != nil {
		t.Fatal(err)
	} else if exp := dh.CaCertificate; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if _, err := dh.FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := dh.FieldAsString([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := dh.String(), utils.ToJSON(dh); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestDispatcherHostMerge(t *testing.T) {
	dp := &DispatcherHost{
		RemoteHost: &config.RemoteHost{},
	}
	exp := &DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:                   "ID",
			Address:              "127.0.0.1",
			Transport:            utils.MetaJSON,
			ConnectAttempts:      1,
			Reconnects:           1,
			MaxReconnectInterval: 1,
			ConnectTimeout:       time.Nanosecond,
			ReplyTimeout:         time.Nanosecond,
			TLS:                  true,
			ClientKey:            "key",
			ClientCertificate:    "ce",
			CaCertificate:        "ca",
		},
	}
	if dp.Merge(&DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:                   "ID",
			Address:              "127.0.0.1",
			Transport:            utils.MetaJSON,
			ConnectAttempts:      1,
			Reconnects:           1,
			MaxReconnectInterval: 1,
			ConnectTimeout:       time.Nanosecond,
			ReplyTimeout:         time.Nanosecond,
			TLS:                  true,
			ClientKey:            "key",
			ClientCertificate:    "ce",
			CaCertificate:        "ca",
		},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}

func TestDispatcherHostGetConnErr(t *testing.T) {
	dH := &DispatcherHost{
		RemoteHost: &config.RemoteHost{},
	}
	cfg := config.NewDefaultCGRConfig()
	_, err := dH.GetConn(context.Background(), cfg, make(chan birpc.ClientConnector, 1))
	if err == nil || err.Error() != "dial tcp: missing address" {
		t.Errorf("Expected %v \n but received \n %v", "dial tcp: missing address", err)
	}

}
func TestDispatcherHostProfileMerge(t *testing.T) {

	dspHost := &DispatcherHostProfile{
		Params: map[string]any{
			"opt1": "val1",
		},
	}

	dspHostV2 := &DispatcherHostProfile{
		ID: "DispatcherId",
		Params: map[string]any{

			"opt1": "val1",
			"opt2": "val1",
			"opt3": "val1",
		},
		Weight:    10,
		Blocker:   true,
		FilterIDs: []string{"FltrId"},
	}
	exp := dspHostV2

	dspHost.Merge(dspHostV2)
	if !reflect.DeepEqual(dspHost, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dspHost))
	}
}

type cMock struct {
	rcvM string
}

func (*cMock) Call(ctx *context.Context, serviceMethod string, args, reply any) error {
	return nil
}
func TestDispatcherHostGetConnExistingConn(t *testing.T) {
	Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	chanRPC := make(chan birpc.ClientConnector, 1)
	chanRPC <- &cMock{
		rcvM: "testM",
	}
	connMgr := NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, chanRPC)
	dH := &DispatcherHost{
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
		rpcConn: <-chanRPC,
	}

	exp := &cMock{rcvM: "testM"}

	if rcv, err := dH.GetConn(context.Background(), cfg, make(chan birpc.ClientConnector, 1)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %+v %T \n but received \n %+v %T", rcv, rcv, exp, exp)
	}

}
