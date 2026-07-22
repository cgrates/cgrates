// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package analyzers

import (
	"errors"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type mockConnector struct{}

func (c *mockConnector) Call(_ *context.Context, _ string, _, _ any) (err error) {
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
	if err = rpc.Call(context.Background(), utils.CoreSv1Ping, "args", "reply"); err == nil || err.Error() != "error" {
		t.Errorf("Expected 'error' received %v", err)
	}
	time.Sleep(15 * time.Millisecond)
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
