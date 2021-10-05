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
package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

func TestCDRsNewCDRServer(t *testing.T) {
	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	dm := &DataManager{}
	fltrs := &FilterS{}
	connMng := &ConnManager{}
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	expected := &CDRServer{
		cfg:        cfg,
		cdrDB:      sent,
		dm:         dm,
		guard:      guardian.Guardian,
		filterS:    fltrs,
		connMgr:    connMng,
		storDBChan: storDBChan,
	}
	if !reflect.DeepEqual(newCDRSrv, expected) {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", expected, newCDRSrv)
	}
}

func TestCDRsListenAndServeCaseStorDBChanOK(t *testing.T) {
	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	dm := &DataManager{}
	fltrs := &FilterS{}
	connMng := &ConnManager{}
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	stopChan := make(chan struct{}, 1)
	func() {
		storDBChan <- sent
		time.Sleep(10 * time.Millisecond)
		stopChan <- struct{}{}
	}()
	newCDRSrv.ListenAndServe(stopChan)
	if !reflect.DeepEqual(newCDRSrv.cdrDB, sent) {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", sent, newCDRSrv.cdrDB)
	}
}

func TestCDRsListenAndServeCaseStorDBChanNotOK(t *testing.T) {
	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	dm := &DataManager{}
	fltrs := &FilterS{}
	connMng := &ConnManager{}
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	stopChan := make(chan struct{}, 1)
	func() {
		time.Sleep(30 * time.Millisecond)
		close(storDBChan)
	}()
	newCDRSrv.ListenAndServe(stopChan)
	if !reflect.DeepEqual(newCDRSrv.cdrDB, nil) {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, newCDRSrv.cdrDB)
	}
}

func TestCDRsChrgrSProcessEventErrMsnConnIDs(t *testing.T) {
	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	dm := &DataManager{}
	fltrs := &FilterS{}
	connMng := &ConnManager{}
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]interface{}{
			"Resources":      "ResourceProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.Weight:     "20.0",
			utils.Usage:      135 * time.Second,
			utils.Cost:       123.0,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaChargers,
		},
	}
	_, err := newCDRSrv.chrgrSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}

}
