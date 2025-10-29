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
package agents

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestKAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPCClient(new(KamailioAgent))
}

func TestKamailioAgentV1WarnDisconnect(t *testing.T) {
	agent := KamailioAgent{}
	ctx := context.Background()
	args := make(map[string]any)
	var reply string
	err := agent.V1WarnDisconnect(ctx, args, &reply)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

func TestKamailioAgentV1DisconnectPeer(t *testing.T) {
	agent := KamailioAgent{}
	ctx := context.Background()
	dprArgs := &utils.DPRArgs{}
	var reply string

	err := agent.V1DisconnectPeer(ctx, dprArgs, &reply)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

func TestKamailioAgentV1AlterSession(t *testing.T) {
	agent := KamailioAgent{}
	ctx := context.Background()
	cgrEvent := utils.CGREvent{}
	var reply string
	err := agent.V1AlterSession(ctx, cgrEvent, &reply)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected ErrNotImplemented, got %v", err)
	}
}

func TestKamailioAgentReload(t *testing.T) {
	cfg := config.KamAgentCfg{
		EvapiConns: []*config.KamConnCfg{
			{},
			{},
			{},
		},
	}
	ka := &KamailioAgent{
		cfg: &cfg,
	}
	ka.Reload()
	if len(ka.conns) != len(cfg.EvapiConns) {
		t.Errorf("Expected conns length %d, but got %d", len(cfg.EvapiConns), len(ka.conns))
	}
	for i, conn := range ka.conns {
		if conn != nil {
			t.Errorf("Expected ka.conns[%d] to be nil, but got  value", i)
		}
	}
}
