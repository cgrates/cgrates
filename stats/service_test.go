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
package stats

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestReqFilterPassStatS(t *testing.T) {
	if cgrCfg := config.CgrConfig(); cgrCfg == nil {
		cgrCfg, _ = config.NewDefaultCGRConfig()
		config.SetCgrConfig(cgrCfg)
	}
	dataStorage, _ := engine.NewMapStorage()
	dataStorage.SetStatsQueue(
		&engine.StatsQueue{ID: "CDRST1",
			Filters: []*engine.RequestFilter{
				&engine.RequestFilter{Type: engine.MetaString, FieldName: "Tenant",
					Values: []string{"cgrates.org"}}},
			Metrics: []string{utils.MetaASR}})
	statS, err := NewStatService(dataStorage, dataStorage.Marshaler(), 0)
	if err != nil {
		t.Fatal(err)
	}
	var replyStr string
	if err := statS.Call("StatSV1.LoadQueues", ArgsLoadQueues{},
		&replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Errorf("reply received: %s", replyStr)
	}
	cdr := &engine.CDR{
		Tenant:          "cgrates.org",
		Category:        "call",
		AnswerTime:      time.Now(),
		SetupTime:       time.Now(),
		Usage:           10 * time.Second,
		Cost:            10,
		Supplier:        "suppl1",
		DisconnectCause: "NORMAL_CLEARNING",
	}
	cdrMp, _ := cdr.AsMapStringIface()
	cdrMp[utils.ID] = "event1"
	if err := statS.processEvent(cdrMp); err != nil {
		t.Error(err)
	}
	rf, err := engine.NewRequestFilter(engine.MetaStatS, "",
		[]string{"CDRST1:*min_asr:20"})
	if err != nil {
		t.Fatal(err)
	}
	if passes, err := rf.Pass(cdr, "", statS); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
}
