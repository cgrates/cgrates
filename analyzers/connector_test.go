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

package analyzers

import (
	"errors"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type mockConnector struct{}

func (c *mockConnector) Call(_ string, _, _ interface{}) (err error) {
	return errors.New("error")
}
func TestNewAnalyzeConnector(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	rpc := anz.NewAnalyzerConnector(new(mockConnector), utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012")
	if err = rpc.Call(utils.CoreSv1Ping, "args", "reply"); err == nil || err.Error() != "error" {
		t.Errorf("Expected 'error' received %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	runtime.Gosched()
	if cnt, err := anz.db.DocCount(); err != nil {
		t.Fatal(err)
	} else if cnt != 1 {
		t.Errorf("Expected only one document received:%v", cnt)
	}
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}

func (c *mockConnector) CallBiRPC(cl rpcclient.ClientConnector, serviceMethod string, args, reply interface{}) (err error) {
	return c.Call(serviceMethod, args, reply)
}
func (c *mockConnector) Handlers() map[string]interface{} { return make(map[string]interface{}) }
func TestNewAnalyzeBiRPCConnector1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	rpc := anz.NewAnalyzerBiRPCConnector(new(mockConnector), utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012")
	if err = rpc.Call(utils.CoreSv1Ping, "args", "reply"); err == nil || err.Error() != "error" {
		t.Errorf("Expected 'error' received %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	runtime.Gosched()
	if cnt, err := anz.db.DocCount(); err != nil {
		t.Fatal(err)
	} else if cnt != 1 {
		t.Errorf("Expected only one document received:%v", cnt)
	}
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}

func TestNewAnalyzeBiRPCConnector2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.AnalyzerSCfg().DBPath, 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	rpc := anz.NewAnalyzerBiRPCConnector(new(mockConnector), utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012")
	if err = rpc.CallBiRPC(nil, utils.CoreSv1Ping, "args", "reply"); err == nil || err.Error() != "error" {
		t.Errorf("Expected 'error' received %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	runtime.Gosched()
	if cnt, err := anz.db.DocCount(); err != nil {
		t.Fatal(err)
	} else if cnt != 1 {
		t.Errorf("Expected only one document received:%v", cnt)
	}
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}

	exp := make(map[string]interface{})
	if rply := rpc.Handlers(); !reflect.DeepEqual(rply, exp) {
		t.Errorf("Expected: %v ,received:%v", exp, rply)
	}
}
