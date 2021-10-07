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
	"net/http"
	"net/url"
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
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}
}

func TestCDRsAttrSProcessEventNoOpts(t *testing.T) {
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
	}
	err := newCDRSrv.attrSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}
}

func TestCDRsAttrSProcessEvent(t *testing.T) {
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
	err := newCDRSrv.attrSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}
}

func TestCDRsRateSCostForEventErr(t *testing.T) {
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
	err := newCDRSrv.rateSCostForEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}
}

func TestCDRsAccountSDebitEventErr(t *testing.T) {
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
	err := newCDRSrv.accountSDebitEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}
}

func TestCDRsThdSProcessEventErr(t *testing.T) {
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
	err := newCDRSrv.thdSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}

}

func TestCDRsStatSProcessEventErrMsnConnIDs(t *testing.T) {
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
	err := newCDRSrv.statSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}

}

func TestCDRsEESProcessEventErrMsnConnIDs(t *testing.T) {
	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	dm := &DataManager{}
	fltrs := &FilterS{}
	connMng := &ConnManager{}
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)

	cgrEv := &utils.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
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
		},
	}
	err := newCDRSrv.eeSProcessEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [connIDs]> \n, received <%+v>", err)
	}

}

func TestCDRsNewMapEventFromReqForm(t *testing.T) {
	httpReq := &http.Request{
		Form: url.Values{
			"value1": {"value2"},
		},
	}
	result, err := newMapEventFromReqForm(httpReq)
	if err != nil {
		t.Errorf("\nExpected <nil> \n, received <%+v>", err)
	}
	expected := MapEvent{
		"value1": "value2",
		"Source": "",
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", expected, result)
	}
}

func TestCDRsNewMapEventFromReqFormErr(t *testing.T) {
	httpReq := &http.Request{
		URL: &url.URL{
			RawQuery: ";",
		},
	}
	_, err := newMapEventFromReqForm(httpReq)
	if err == nil || err.Error() != "invalid semicolon separator in query" {
		t.Errorf("\nExpected <invalid semicolon separator in query> \n, received <%+v>", err)
	}

}

// func TestCDRsChrgrSProcessEvent(t *testing.T) {
// 	Cache.Clear(nil)
// 	var sent StorDB
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
// 		utils.MetaChargers)}
// 	storDBChan := make(chan StorDB, 1)
// 	storDBChan <- sent
// 	data := NewInternalDB(nil, nil, true)
// 	connMng := NewConnManager(cfg)
// 	dm := NewDataManager(data, cfg.CacheCfg(), nil)
// 	fltrs := NewFilterS(cfg, nil, dm)
// 	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
// 	ccM := &ccMock{
// 		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
// 			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
// 				*reply.(*[]*ChrgSProcessEventReply) = []*ChrgSProcessEventReply{
// 					{
// 						ChargerSProfile: "string",
// 					},
// 				}
// 				return nil
// 			},
// 		},
// 	}
// 	rpcInternal := make(chan birpc.ClientConnector, 1)
// 	rpcInternal <- ccM
// 	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
// 		utils.MetaChargers), utils.ChargerSv1, rpcInternal)

// 	cgrEv := &utils.CGREvent{
// 		Tenant: "cgrates.org",
// 		ID:     "testID",
// 		Event: map[string]interface{}{
// 			"Resources":      "ResourceProfile1",
// 			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
// 			"UsageInterval":  "1s",
// 			"PddInterval":    "1s",
// 			utils.Weight:     "20.0",
// 			utils.Usage:      135 * time.Second,
// 			utils.Cost:       123.0,
// 		},
// 		APIOpts: map[string]interface{}{
// 			utils.Subsys: utils.MetaChargers,
// 		},
// 	}
// 	result, err := newCDRSrv.chrgrSProcessEvent(context.Background(), cgrEv)
// 	if err != nil {
// 		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
// 	}
// 	var expecte *utils.CGREvent
// 	expected := []*utils.CGREvent{
// 		expecte,
// 	}
// 	if !reflect.DeepEqual(result, expected) {
// 		t.Errorf("\nExpected <%+v> \n, received <%+v>", expected, result)
// 	}
// 	if err := dm.DataDB().Flush(""); err != nil {
// 		t.Error(err)
// 	}
// }

// func TestCDRsChrgrSProcessEventEmptyChrgrs(t *testing.T) {
// 	Cache.Clear(nil)
// 	var sent StorDB
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
// 		utils.MetaChargers)}
// 	storDBChan := make(chan StorDB, 1)
// 	storDBChan <- sent
// 	data := NewInternalDB(nil, nil, true)
// 	connMng := NewConnManager(cfg)
// 	dm := NewDataManager(data, cfg.CacheCfg(), nil)
// 	fltrs := NewFilterS(cfg, nil, dm)
// 	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
// 	ccM := &ccMock{
// 		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
// 			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
// 				return nil
// 			},
// 		},
// 	}
// 	rpcInternal := make(chan birpc.ClientConnector, 1)
// 	rpcInternal <- ccM
// 	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
// 		utils.MetaChargers), utils.ChargerSv1, rpcInternal)

// 	cgrEv := &utils.CGREvent{
// 		Tenant: "cgrates.org",
// 		ID:     "testID",
// 		Event: map[string]interface{}{
// 			"Resources":      "ResourceProfile1",
// 			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
// 			"UsageInterval":  "1s",
// 			"PddInterval":    "1s",
// 			utils.Weight:     "20.0",
// 			utils.Usage:      135 * time.Second,
// 			utils.Cost:       123.0,
// 		},
// 		APIOpts: map[string]interface{}{
// 			utils.Subsys: utils.MetaChargers,
// 		},
// 	}
// 	_, err := newCDRSrv.chrgrSProcessEvent(context.Background(), cgrEv)
// 	if err != nil {
// 		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
// 	}
// 	if err := dm.DataDB().Flush(""); err != nil {
// 		t.Error(err)
// 	}
// }
