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
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

func TestLibengineNewRPCPoolNoAddress(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	connID := "connID"
	intChan := make(chan rpcclient.ClientConnector)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.RPCConns()[connID] = config.NewDfltRPCConn()
	defaultCfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			ID:      connID,
			Address: utils.EmptyString,
		},
	}
	connCache := ltcache.NewCache(-1, 0, true, nil)

	exp := &rpcclient.RPCPool{}
	experr := rpcclient.ErrDisconnected
	rcv, err := NewRPCPool("", "", "", "", defaultCfg.GeneralCfg().ConnectAttempts,
		defaultCfg.GeneralCfg().Reconnects, defaultCfg.GeneralCfg().ConnectTimeout,
		0, defaultCfg.RPCConns()[connID].Conns, intChan, false, nil, connID, connCache)

	if err == nil || err != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

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
		Synchronous:     false,
		ConnectAttempts: 2,
		Reconnects:      5,
		ConnectTimeout:  2 * time.Minute,
		ReplyTimeout:    3 * time.Minute,
		TLS:             true,
		ClientKey:       "key1",
	}
	expectedErr := "dial tcp [::1]:6012: connect: connection refused"
	cM := NewConnManager(config.NewDefaultCGRConfig(), nil)
	exp, err := rpcclient.NewRPCClient(utils.TCP, cfg.Address, cfg.TLS, cfg.ClientKey, cM.cfg.TLSCfg().ClientCerificate,
		cM.cfg.TLSCfg().CaCertificate, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
		cfg.Transport, nil, false, nil)

	if err.Error() != expectedErr {
		t.Errorf("Expected %v \n but received \n %v", expectedErr, err)
	}

	conn, err := NewRPCConnection(cfg, cM.cfg.TLSCfg().ClientKey, cM.cfg.TLSCfg().ClientCerificate, cM.cfg.TLSCfg().CaCertificate,
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
		Synchronous:     false,
		ConnectAttempts: 2,
		Reconnects:      5,
		ConnectTimeout:  2 * time.Minute,
		ReplyTimeout:    3 * time.Minute,
		TLS:             true,
		ClientKey:       "key1",
	}
	cM := NewConnManager(config.NewDefaultCGRConfig(), make(map[string]chan rpcclient.ClientConnector))
	exp, err := rpcclient.NewRPCClient("", "", cfg.TLS, cfg.ClientKey, cM.cfg.TLSCfg().ClientCerificate,
		cM.cfg.TLSCfg().ClientCerificate, cfg.ConnectAttempts, cfg.Reconnects, cfg.ConnectTimeout, cfg.ReplyTimeout,
		rpcclient.InternalRPC, cM.rpcInternal["a4f3f"], false, nil)

	// We only want to check if the client loaded with the correct config,
	// therefore connection is not mandatory
	if err != rpcclient.ErrInternallyDisconnected {
		t.Error(err)
	}

	conn, err := NewRPCConnection(cfg, cM.cfg.TLSCfg().ClientKey, cM.cfg.TLSCfg().ClientCerificate, cM.cfg.TLSCfg().CaCertificate,
		cM.cfg.GeneralCfg().ConnectAttempts, cM.cfg.GeneralCfg().Reconnects, cM.cfg.GeneralCfg().ConnectTimeout, cM.cfg.GeneralCfg().ReplyTimeout,
		cM.rpcInternal["a4f3f"], false, nil, "*internal", "a4f3f", new(ltcache.Cache))

	if err != rpcclient.ErrInternallyDisconnected {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, conn) {
		t.Error("Connections don't match")
	}
}
func TestLibengineNewRPCPoolUnsupportedTransport(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	connID := "connID"
	intChan := make(chan rpcclient.ClientConnector)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.RPCConns()[connID] = config.NewDfltRPCConn()
	defaultCfg.RPCConns()[connID].Conns = []*config.RemoteHost{
		{
			ID:        connID,
			Address:   rpcclient.JSONrpc,
			Transport: "invalid",
		},
	}
	connCache := ltcache.NewCache(-1, 0, true, nil)

	var exp *rpcclient.RPCPool
	experr := fmt.Sprintf("Unsupported transport: <%s>",
		defaultCfg.RPCConns()[connID].Conns[0].Transport)
	rcv, err := NewRPCPool("", "", "", "", defaultCfg.GeneralCfg().ConnectAttempts,
		defaultCfg.GeneralCfg().Reconnects, defaultCfg.GeneralCfg().ConnectTimeout,
		0, defaultCfg.RPCConns()[connID].Conns, intChan, false, nil, connID, connCache)

	if err == nil || err.Error() != experr {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestLibengineNewRPCClientSet(t *testing.T) {
	exp := RPCClientSet{}
	rcv := NewRPCClientSet()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestLibengineAddInternalRPCClientSuccess(t *testing.T) {
	s := RPCClientSet{}
	name := "testName"
	connChan := make(chan rpcclient.ClientConnector)

	expClient, err := rpcclient.NewRPCClient(utils.EmptyString, utils.EmptyString, false,
		utils.EmptyString, utils.EmptyString, utils.EmptyString,
		config.CgrConfig().GeneralCfg().ConnectAttempts, config.CgrConfig().GeneralCfg().Reconnects,
		config.CgrConfig().GeneralCfg().ConnectTimeout, config.CgrConfig().GeneralCfg().ReplyTimeout,
		rpcclient.InternalRPC, connChan, true, nil)

	if err != nil {
		t.Fatal(err)
	}
	exp := RPCClientSet{
		"testName": expClient,
	}
	s.AddInternalRPCClient(name, connChan)

	if !reflect.DeepEqual(s, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, s)
	}
}

func TestLibengineAddInternalRPCClientLoggerErr(t *testing.T) {
	s := RPCClientSet{}
	name := "testName"

	utils.Logger.SetLogLevel(3)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	exp := "CGRateS <> [ERROR] <InternalRPCSet> Error adding testName to the set: INTERNALLY_DISCONNECTED\n"
	s.AddInternalRPCClient(name, nil)
	rcv := buf.String()[20:]

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	utils.Logger.SetLogLevel(0)
}

func TestLibengineCallInvalidMethod(t *testing.T) {
	s := RPCClientSet{
		"testField": &rpcclient.RPCClient{},
	}
	method := "invalid"
	args := "testArgs"
	reply := "testReply"

	experr := rpcclient.ErrUnsupporteServiceMethod
	err := s.Call(method, args, reply)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibengineCallMethodNotFound(t *testing.T) {
	connChan := make(chan rpcclient.ClientConnector)
	client, err := rpcclient.NewRPCClient(utils.EmptyString, utils.EmptyString, false,
		utils.EmptyString, utils.EmptyString, utils.EmptyString,
		config.CgrConfig().GeneralCfg().ConnectAttempts, config.CgrConfig().GeneralCfg().Reconnects,
		config.CgrConfig().GeneralCfg().ConnectTimeout, config.CgrConfig().GeneralCfg().ReplyTimeout,
		rpcclient.InternalRPC, connChan, true, nil)

	if err != nil {
		t.Fatal(err)
	}

	s := RPCClientSet{
		"testField": client,
	}
	method := "APIerSv1.Ping"
	args := "testArgs"
	reply := "testReply"

	experr := rpcclient.ErrUnsupporteServiceMethod
	err = s.Call(method, args, reply)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestLibengineCallNilArgument(t *testing.T) {
	connChan := make(chan rpcclient.ClientConnector)
	client, err := rpcclient.NewRPCClient(utils.EmptyString, utils.EmptyString, false,
		utils.EmptyString, utils.EmptyString, utils.EmptyString,
		config.CgrConfig().GeneralCfg().ConnectAttempts, config.CgrConfig().GeneralCfg().Reconnects,
		config.CgrConfig().GeneralCfg().ConnectTimeout, config.CgrConfig().GeneralCfg().ReplyTimeout,
		rpcclient.InternalRPC, connChan, true, nil)

	if err != nil {
		t.Fatal(err)
	}

	s := RPCClientSet{
		"APIerSv1": client,
	}
	method := "APIerSv1.Ping"
	var args int
	var reply *int

	experr := fmt.Sprintf("nil rpc in argument method: %s in: %v out: %v",
		method, args, reply)
	err = s.Call(method, args, reply)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}
