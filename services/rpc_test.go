// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestNewRPCServiceAddsPing(t *testing.T) {
	srv, err := newRPCService(new(rpcTestReceiver), "TestSv1")
	if err != nil {
		t.Fatal(err)
	}
	if _, has := srv.Methods[utils.Ping]; !has {
		t.Fatal("missing Ping")
	}
	var reply string
	if err := srv.Call(context.Background(), "TestSv1.Ping", &utils.CGREvent{}, &reply); err != nil {
		t.Fatal(err)
	}
	if reply != utils.Pong {
		t.Errorf("Ping reply = %q, want %q", reply, utils.Pong)
	}
}

type rpcTestReceiver struct{}

func (*rpcTestReceiver) Status(_ *context.Context, _ *utils.CGREvent, reply *string) error {
	*reply = utils.OK
	return nil
}
