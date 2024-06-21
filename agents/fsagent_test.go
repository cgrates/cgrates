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
package agents

import (
	"testing"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestFAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPCClient(new(FSsessions))
}

func TestFsAgentV1DisconnectPeer(t *testing.T) {
	ctx := context.Background()
	args := &utils.DPRArgs{}
	fss := &FSsessions{}
	err := fss.V1DisconnectPeer(ctx, args, nil)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %v, got: %v", utils.ErrNotImplemented, err)
	}
}

func TestFsAgentV1AlterSession(t *testing.T) {
	ctx := context.Background()
	cgrEv := utils.CGREvent{}
	fss := &FSsessions{}
	err := fss.V1AlterSession(ctx, cgrEv, nil)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error: %v, got: %v", utils.ErrNotImplemented, err)
	}
}

func TestFsAgentCreateHandlers(t *testing.T) {
	cfg := &config.FsAgentCfg{
		SubscribePark: true,
	}
	fs := &FSsessions{
		cfg: cfg,
	}
	handlers := fs.createHandlers()
	if _, ok := handlers["CHANNEL_ANSWER"]; !ok {
		t.Error("Expected CHANNEL_ANSWER handler, but not found")
	}
	if _, ok := handlers["CHANNEL_HANGUP_COMPLETE"]; !ok {
		t.Error("Expected CHANNEL_HANGUP_COMPLETE handler, but not found")
	}
	if _, ok := handlers["CHANNEL_PARK"]; !ok {
		t.Error("Expected CHANNEL_PARK handler, but not found")
	}
}
