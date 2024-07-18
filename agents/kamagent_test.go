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
	"github.com/cgrates/kamevapi"
)

func TestKAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPClient(new(KamailioAgent))
}

func TestKAReload(t *testing.T) {
	cfg := &config.KamAgentCfg{}
	ka := &KamailioAgent{
		cfg: cfg,
	}
	ka.Reload()
	expectedLength := len(cfg.EvapiConns)
	if len(ka.conns) != expectedLength {
		t.Errorf("Expected ka.conns to have length %d, got %d", expectedLength, len(ka.conns))
	}
	for _, conn := range ka.conns {
		if conn != nil {
			t.Errorf("Expected nil KamEvapi instance in ka.conns, got non-nil")
		}
	}
	for i := range ka.conns {
		ka.conns[i] = &kamevapi.KamEvapi{}
	}
	for _, conn := range ka.conns {
		if conn == nil {
			t.Errorf("Expected non-nil KamEvapi instance in ka.conns, got nil")
		}
	}
}

func TestKACall(t *testing.T) {
	ka := &KamailioAgent{}
	ctx := &context.Context{}
	serviceMethod := "SomeService.Method"
	args := struct{ Key string }{"value"}
	reply := struct{ Result string }{}
	err := ka.Call(ctx, serviceMethod, args, &reply)
	if err == nil {
		t.Errorf("Call didn't return an error: %v", err)
	}
	expected := ""
	if reply.Result != expected {
		t.Errorf("Expected reply.Result to be %q, got %q", expected, reply.Result)
	}
}

func TestKAShutdown(t *testing.T) {
	agent := &KamailioAgent{}
	err := agent.Shutdown()
	if err != nil {
		t.Errorf("Shutdown returned an error: %v", err)
	}

}
