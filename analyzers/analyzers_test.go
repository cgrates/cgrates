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
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewAnalyzerService(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(path.Dir(cfg.AnalyzerSCfg().DBPath), 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// no need to DeepEqual
	if err = anz.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if err = anz.initDB(); err != nil {
		t.Fatal(err)
	}
	exitChan := make(chan bool, 1)
	exitChan <- true
	if err := anz.ListenAndServe(exitChan); err != nil {
		t.Fatal(err)
	}
	anz.db.Close()
}

func TestAnalyzerSLogTraffic(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(path.Dir(cfg.AnalyzerSCfg().DBPath), 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t1 := time.Now().Add(-time.Hour)
	if err = anz.logTrafic(0, utils.AnalyzerSv1Ping, "status", "result", "error", &extraInfo{
		enc:  utils.MetaJSON,
		from: "127.0.0.1:5565",
		to:   "127.0.0.1:2012",
	}, t1, t1.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err = anz.logTrafic(0, utils.CoreSv1Status, "status", "result", "error", &extraInfo{
		enc:  utils.MetaJSON,
		from: "127.0.0.1:5565",
		to:   "127.0.0.1:2012",
	}, t1, t1.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	t1 = time.Now().Add(-10 * time.Minute)
	if err = anz.logTrafic(0, utils.CoreSv1Status, "status", "result", "error", &extraInfo{
		enc:  utils.MetaJSON,
		from: "127.0.0.1:5565",
		to:   "127.0.0.1:2012",
	}, t1, t1.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if cnt, err := anz.db.DocCount(); err != nil {
		t.Fatal(err)
	} else if cnt != 2 {
		t.Errorf("Expected only 2 documents received:%v", cnt)
	}
	if err = anz.clenaUp(); err != nil {
		t.Fatal(err)
	}
	if cnt, err := anz.db.DocCount(); err != nil {
		t.Fatal(err)
	} else if cnt != 1 {
		t.Errorf("Expected only one document received:%v", cnt)
	}

	if err = anz.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err = anz.clenaUp(); err != bleve.ErrorIndexClosed {
		t.Errorf("Expected error: %v,received: %+v", bleve.ErrorIndexClosed, err)
	}
}

func TestAnalyzersDeleteHits(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(path.Dir(cfg.AnalyzerSCfg().DBPath), 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = anz.deleteHits(search.DocumentMatchCollection{&search.DocumentMatch{}}); err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected error: %v,received: %+v", utils.ErrPartiallyExecuted, err)
	}
}

func TestAnalyzersListenAndServe(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.AnalyzerSCfg().DBPath = "/tmp/analyzers"
	cfg.AnalyzerSCfg().TTL = 30 * time.Minute
	if err := os.RemoveAll(cfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(path.Dir(cfg.AnalyzerSCfg().DBPath), 0700); err != nil {
		t.Fatal(err)
	}
	anz, err := NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := anz.db.Close(); err != nil {
		t.Fatal(err)
	}
	anz.ListenAndServe(make(chan bool))

	cfg.AnalyzerSCfg().CleanupInterval = 1
	anz, err = NewAnalyzerService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(1)
		runtime.Gosched()
		anz.db.Close()
	}()
	anz.ListenAndServe(make(chan bool))
}
