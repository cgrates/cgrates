// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"testing"

	"github.com/cgrates/cgrates/config"
)

func TestAnalyzerServiceGetInternalCodec(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AnalyzerSCfg().Enabled = true
	anz := NewAnalyzerService(cfg)
	conn, err := newRPCService(new(rpcTestReceiver), "TestSv1")
	if err != nil {
		t.Fatal(err)
	}
	if got := anz.GetInternalCodec(conn, "TestSv1"); got != conn {
		t.Error("missing analyzer should leave connector unchanged")
	}
}
