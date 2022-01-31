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

package apis

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestEeSProcessEvent(t *testing.T) {
	filePath := "/tmp/TestV1ProcessEvent"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsCfg().Exporters[0].Type = "*fileCSV"
	cfg.EEsCfg().Exporters[0].ID = "SQLExporterFull"
	cfg.EEsCfg().Exporters[0].ExportPath = filePath
	newIDb := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	newDM := engine.NewDataManager(newIDb, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	eeS := ees.NewEventExporterS(cfg, filterS, nil)
	cS := NewEeSv1(eeS)
	cgrEv := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]interface{}{

				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.MetaRunID:    utils.MetaDefault,
				utils.Cost:         1.01,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]interface{}{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}
	var reply map[string]map[string]interface{}
	replyExpect := map[string]map[string]interface{}{
		"SQLExporterFull": {},
	}
	if err := cS.ProcessEvent(context.Background(), cgrEv, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, replyExpect) {
		t.Errorf("Expected %v \n but received \n %v", replyExpect, reply)
	}

	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}
