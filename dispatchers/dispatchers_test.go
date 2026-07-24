// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func (dS *DispatcherService) DispatcherServicePing(ev *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

func TestDispatcherCall1(t *testing.T) {
	dS := &DispatcherService{}
	var reply string
	if err := dS.Call(context.Background(), utils.DispatcherServicePing, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Expected: %s , received: %s", utils.Pong, reply)
	}
}

func TestDispatcherCall2(t *testing.T) {
	dS := &DispatcherService{}
	var reply string
	if err := dS.Call(context.Background(), "DispatcherServicePing", &utils.CGREvent{}, &reply); err == nil || err.Error() != rpcclient.ErrUnsupporteServiceMethod.Error() {
		t.Error(err)
	}
	if err := dS.Call(context.Background(), "DispatcherService.Pong", &utils.CGREvent{}, &reply); err == nil || err.Error() != rpcclient.ErrUnsupporteServiceMethod.Error() {
		t.Error(err)
	}
}
