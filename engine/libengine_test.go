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
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

// For the purpose of this test, we don't need our client to establish a connection
// we only want to check if the client loaded with the given config where needed
func TestLibengineNewRPCConnection(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := &config.RemoteHost{
		ID:              "a4f3f",
		Address:         "localhost:6012",
		Transport:       "*json",
		ConnectAttempts: 2,
		Reconnects:      5,
		ConnectTimeout:  2 * time.Minute,
		ReplyTimeout:    3 * time.Minute,
		TLS:             true,
		ClientKey:       "key1",
	}
	expectedErr := "dial tcp [::1]:6012: connect: connection refused"
	cM := NewConnManager(config.NewDefaultCGRConfig(), nil)
	exp, err := rpcclient.NewRPCClient(context.Background(), utils.TCP, cfg.Address, cfg.TLS, cfg.ClientKey, cM.cfg.TLSCfg().ClientCerificate,
		cM.cfg.TLSCfg().CaCertificate, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
		cfg.Transport, nil, false, nil)

	if err.Error() != expectedErr {
		t.Errorf("Expected %v \n but received \n %v", expectedErr, err)
	}

	conn, err := NewRPCConnection(context.Background(), cfg, cM.cfg.TLSCfg().ClientKey, cM.cfg.TLSCfg().ClientCerificate, cM.cfg.TLSCfg().CaCertificate,
		cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects, cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
		nil, false, nil, "*localhost", "a4f3f", new(ltcache.Cache))
	if err.Error() != expectedErr {
		t.Errorf("Expected %v \n but received \n %v", expectedErr, err)
	}
	if !reflect.DeepEqual(exp, conn) {
		t.Error("Connections don't match")
	}
}

func TestLibengineNewRPCConnectionInternal(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := &config.RemoteHost{
		ID:              "a4f3f",
		Address:         rpcclient.InternalRPC,
		Transport:       "",
		ConnectAttempts: 2,
		Reconnects:      5,
		ConnectTimeout:  2 * time.Minute,
		ReplyTimeout:    3 * time.Minute,
		TLS:             true,
		ClientKey:       "key1",
	}
	cM := NewConnManager(config.NewDefaultCGRConfig(), make(map[string]chan birpc.ClientConnector))
	exp, err := rpcclient.NewRPCClient(context.Background(), "", "", cfg.TLS, cfg.ClientKey, cM.cfg.TLSCfg().ClientCerificate,
		cM.cfg.TLSCfg().ClientCerificate, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
		rpcclient.InternalRPC, cM.rpcInternal["a4f3f"], false, nil)

	// We only want to check if the client loaded with the correct config,
	// therefore connection is not mandatory
	if err != rpcclient.ErrInternallyDisconnected {
		t.Error(err)
	}

	conn, err := NewRPCConnection(context.Background(), cfg, cM.cfg.TLSCfg().ClientKey, cM.cfg.TLSCfg().ClientCerificate, cM.cfg.TLSCfg().CaCertificate,
		cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects, cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
		cM.rpcInternal["a4f3f"], false, nil, "*internal", "a4f3f", new(ltcache.Cache))

	if err != rpcclient.ErrInternallyDisconnected {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, conn) {
		t.Error("Connections don't match")
	}
}
