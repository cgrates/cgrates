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

	"github.com/cgrates/birpc"
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

func TestCDRsAttrSProcessEventMock(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*AttrSProcessEventReply) = AttrSProcessEventReply{
					AlteredFields: []string{},
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes), utils.AttributeSv1, rpcInternal)

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
			utils.Subsys: utils.MetaAttributes,
		},
	}
	err := newCDRSrv.attrSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
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
			utils.OptsContext: utils.MetaCDRs,
			utils.Subsys:      utils.MetaCDRs,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsAttrSProcessEventMockNotFoundErr(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*AttrSProcessEventReply) = AttrSProcessEventReply{
					AlteredFields: []string{},
				}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes), utils.AttributeSv1, rpcInternal)

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
			utils.Subsys: utils.MetaAttributes,
		},
	}
	err := newCDRSrv.attrSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
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
			utils.OptsContext: utils.MetaCDRs,
			utils.Subsys:      utils.MetaCDRs,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsAttrSProcessEventMockNotEmptyAF(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*AttrSProcessEventReply) = AttrSProcessEventReply{
					AlteredFields: []string{utils.AccountField},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "testID",
						Event: map[string]interface{}{
							utils.AccountField: "1001",
							"Resources":        "ResourceProfile1",
							utils.AnswerTime:   time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
							"UsageInterval":    "1s",
							"PddInterval":      "1s",
							utils.Weight:       "20.0",
							utils.Usage:        135 * time.Second,
							utils.Cost:         123.0,
						},
						APIOpts: map[string]interface{}{
							utils.Subsys:       utils.MetaAttributes,
							utils.AccountField: "1001",
						},
					},
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAttributes), utils.AttributeSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Resources":        "ResourceProfile1",
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":    "1s",
			"PddInterval":      "1s",
			utils.Weight:       "20.0",
			utils.Usage:        135 * time.Second,
			utils.Cost:         123.0,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAttributes,
		},
	}
	err := newCDRSrv.attrSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"Resources":        "ResourceProfile1",
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":    "1s",
			"PddInterval":      "1s",
			utils.Weight:       "20.0",
			utils.Usage:        135 * time.Second,
			utils.Cost:         123.0,
		},
		APIOpts: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Subsys:       utils.MetaAttributes,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsChrgrSProcessEvent(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()
	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaChargers)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*[]*ChrgSProcessEventReply) = []*ChrgSProcessEventReply{
					{
						ChargerSProfile: "string",
					},
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaChargers), utils.ChargerSv1, rpcInternal)

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
	result, err := newCDRSrv.chrgrSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	var expecte *utils.CGREvent
	expected := []*utils.CGREvent{
		expecte,
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", expected, result)
	}

}

func TestCDRsRateProcessEventMock(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaRateS)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.RateSv1CostForEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*utils.RateProfileCost) = utils.RateProfileCost{}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaRateS), utils.RateSv1, rpcInternal)

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
			utils.Subsys: utils.MetaRateS,
		},
	}
	err := newCDRSrv.rateSCostForEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]interface{}{
			utils.MetaRateSCost: utils.RateProfileCost{},
			"Resources":         "ResourceProfile1",
			utils.AnswerTime:    time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":     "1s",
			"PddInterval":       "1s",
			utils.Weight:        "20.0",
			utils.Usage:         135 * time.Second,
			utils.Cost:          123.0,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaRateS,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsAccountProcessEventMock(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAccounts)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AccountSv1DebitAbstracts: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*utils.EventCharges) = utils.EventCharges{}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaAccounts), utils.AccountSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]interface{}{
			utils.MetaAccountSCost: &utils.EventCharges{},
			"Resources":            "ResourceProfile1",
			utils.AnswerTime:       time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":        "1s",
			"PddInterval":          "1s",
			utils.Weight:           "20.0",
			utils.Usage:            135 * time.Second,
			utils.Cost:             123.0,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	err := newCDRSrv.accountSDebitEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]interface{}{
			utils.MetaAccountSCost: cgrEv.Event[utils.MetaAccountSCost],
			"Resources":            "ResourceProfile1",
			utils.AnswerTime:       time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":        "1s",
			"PddInterval":          "1s",
			utils.Weight:           "20.0",
			utils.Usage:            135 * time.Second,
			utils.Cost:             123.0,
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAccounts,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsThdSProcessEventMock(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaThresholds)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*[]string) = []string{"testID"}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)

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
		APIOpts: nil,
	}
	err := newCDRSrv.thdSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
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
		APIOpts: nil,
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsThdSProcessEventMockNotfound(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaThresholds)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)

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
		APIOpts: nil,
	}
	err := newCDRSrv.thdSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
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
		APIOpts: nil,
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsStatSProcessEventMock(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaStats)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.StatSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*[]string) = []string{"testID"}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaStats), utils.ThresholdSv1, rpcInternal)

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
		APIOpts: nil,
	}
	err := newCDRSrv.statSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
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
		APIOpts: nil,
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsEESProcessEventMock(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			APIOpts: nil,
		},
	}
	err := newCDRSrv.eeSProcessEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREventWithEeIDs{
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
			APIOpts: nil,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsProcessEventMock(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
		APIOpts: nil,
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
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
		APIOpts: nil,
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsProcessEventMockSkipOpts(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
		APIOpts: nil,
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
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
		APIOpts: nil,
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsNewMapEventFromReqFormErr(t *testing.T) {
	httpReq := &http.Request{
		URL: &url.URL{
			RawQuery: "%0x",
		},
	}
	_, err := newMapEventFromReqForm(httpReq)
	errExpect := `invalid URL escape "%0x"`
	if err == nil || err.Error() != errExpect {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", errExpect, err)
	}

}

func TestCDRsProcessEventMockAttrsErr(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsAttributeS: true,
			"*context":           utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "ATTRIBUTES_ERROR:MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "ATTRIBUTES_ERROR:MANDATORY_IE_MISSING: [connIDs]", err)
	}
}

func TestCDRsProcessEventMockAttrsErrBoolOpts(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsAttributeS: time.Second,
			"*context":           utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "cannot convert field: 1s to bool", err)
	}
}

func TestCDRsProcessEventMockChrgsErr(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsChargerS: true,
			"*context":         utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "CHARGERS_ERROR:MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "CHARGERS_ERROR:MANDATORY_IE_MISSING: [connIDs]", err)
	}

}

func TestCDRsProcessEventMockChrgsErrBoolOpts(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsChargerS: time.Second,
			"*context":         utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "cannot convert field: 1s to bool", err)
	}

}

func TestCDRsProcessEventMockRateSErr(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsRateS: true,
			"*context":      utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}

}

func TestCDRsProcessEventMockRateSErrBoolOpts(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsRateS: time.Second,
			"*context":      utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "cannot convert field: 1s to bool", err)
	}

}

func TestCDRsProcessEventMockAcntsErr(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsAccountS: true,
			"*context":         utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}

}

func TestCDRsProcessEventMockAcntsErrBoolOpts(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsAccountS: time.Second,
			"*context":         utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "cannot convert field: 1s to bool", err)
	}

}

func TestCDRsProcessEventMockExportErr(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{

			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsCDRsExport: true,
			"*context":           utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}

}

func TestCDRsProcessEventMockExportErrBoolOpts(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsCDRsExport: time.Second,
			"*context":           utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "cannot convert field: 1s to bool", err)
	}

}

func TestCDRsProcessEventMockThdsErr(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsThresholdS: true,
			"*context":           utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}

}

func TestCDRsProcessEventMockThdsErrBoolOpts(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsThresholdS: time.Second,
			"*context":           utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "cannot convert field: 1s to bool", err)
	}

}

func TestCDRsProcessEventMockStatsErr(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsStatS: true,
			"*context":      utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}

}

func TestCDRsProcessEventMockStatsErrGetBoolOpts(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	cfg.CdrsCfg().Opts.Attributes = []*utils.DynamicBoolOpt{
		{
			Value: false,
		},
	}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrExists
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsStatS: time.Second,
			"*context":      utils.MetaCDRs,
		},
	}
	_, err := newCDRSrv.processEvent(context.Background(), cgrEv)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "cannot convert field: 1s to bool", err)
	}

}

func TestCDRsV1ProcessEventMock(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
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
			utils.OptsCDRsExport: true,
			"*context":           utils.MetaCDRs,
		},
	}
	var rply string
	err := newCDRSrv.V1ProcessEvent(context.Background(), cgrEv, &rply)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
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
			utils.OptsCDRsExport: true,
			"*context":           utils.MetaCDRs,
		},
	}
	cgrEv.ID = "testID"
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", utils.ToJSON(expected), utils.ToJSON(cgrEv))
	}
}

func TestCDRsV1ProcessEventMockErr(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsStatS:      true,
			utils.OptsCDRsExport: true,
			"*context":           utils.MetaCDRs,
		},
	}
	var rply string
	err := newCDRSrv.V1ProcessEvent(context.Background(), cgrEv, &rply)
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "PARTIALLY_EXECUTED", err)
	}
	expected := &utils.CGREvent{
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
			utils.OptsStatS:      true,
			utils.OptsCDRsExport: true,
			"*context":           utils.MetaCDRs,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsV1ProcessEventMockCache(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsCDRsExport: true,
			"*context":           utils.MetaCDRs,
		},
	}
	defaultConf := config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses]
	config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	defer func() {
		config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses] = defaultConf
	}()
	var rply string
	err := newCDRSrv.V1ProcessEvent(context.Background(), cgrEv, &rply)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
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
			utils.OptsCDRsExport: true,
			"*context":           utils.MetaCDRs,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}
func TestCDRsV1ProcessEventWithGetMockCache(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

	cgrEv := &utils.CGREvent{
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
			utils.OptsCDRsExport: true,
			"*context":           utils.MetaCDRs,
		},
	}
	defaultConf := config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses]
	config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	defer func() {
		config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses] = defaultConf
	}()
	var rply []*utils.EventWithFlags
	err := newCDRSrv.V1ProcessEventWithGet(context.Background(), cgrEv, &rply)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
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
			utils.OptsCDRsExport: true,
			"*context":           utils.MetaCDRs,
		},
	}
	cgrEv.ID = "testID"
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}
func TestCDRsV1ProcessEventWithGetMockCacheErr(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				*reply.(*map[string]map[string]interface{}) = map[string]map[string]interface{}{}
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaEEs), utils.ThresholdSv1, rpcInternal)

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
			utils.OptsCDRsExport: true,
			utils.OptsAttributeS: time.Second,
			"*context":           utils.MetaCDRs,
		},
	}
	defaultConf := config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses]
	config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	defer func() {
		config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses] = defaultConf
	}()
	var rply []*utils.EventWithFlags
	err := newCDRSrv.V1ProcessEventWithGet(context.Background(), cgrEv, &rply)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", "cannot convert field: 1s to bool", err)
	}

}
func TestCDRsChrgrSProcessEventEmptyChrgrs(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()
	var sent StorDB
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaChargers)}
	storDBChan := make(chan StorDB, 1)
	storDBChan <- sent
	data := NewInternalDB(nil, nil, true)
	connMng := NewConnManager(cfg)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, connMng)
	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	newCDRSrv.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal,
		utils.MetaChargers), utils.ChargerSv1, rpcInternal)

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
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}

}

func TestCDRsV1ProcessEventCacheGet(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	storDBChan := make(chan StorDB, 1)
	storDBChan <- nil
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, nil)
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]interface{}{
			utils.Cost: 123,
		},
	}

	rply := "string"
	Cache.Set(context.Background(), utils.CacheRPCResponses, "CDRsV1.ProcessEvent:testID",
		&utils.CachedRPCResponse{Result: &rply, Error: nil},
		nil, true, utils.NonTransactional)

	err := newCDRSrv.V1ProcessEvent(context.Background(), cgrEv, &rply)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]interface{}{
			utils.Cost: 123,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}

func TestCDRsV1ProcessEventWithGetCacheGet(t *testing.T) {
	testCache := Cache
	tmpC := config.CgrConfig()
	tmpCM := connMgr
	defer func() {
		Cache = testCache
		config.SetCgrConfig(tmpC)
		connMgr = tmpCM
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	storDBChan := make(chan StorDB, 1)
	storDBChan <- nil
	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	newCDRSrv := NewCDRServer(cfg, storDBChan, dm, fltrs, nil)
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]interface{}{
			utils.Cost: 123,
		},
	}

	rply := []*utils.EventWithFlags{}
	Cache.Set(context.Background(), utils.CacheRPCResponses, "CDRsV1.ProcessEvent:testID",
		&utils.CachedRPCResponse{Result: &rply, Error: nil},
		nil, true, utils.NonTransactional)

	err := newCDRSrv.V1ProcessEventWithGet(context.Background(), cgrEv, &rply)
	if err != nil {
		t.Errorf("\nExpected <%+v> \n, received <%+v>", nil, err)
	}
	expected := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testID",
		Event: map[string]interface{}{
			utils.Cost: 123,
		},
	}
	if !reflect.DeepEqual(expected, cgrEv) {
		t.Errorf("\nExpected <%+v> \n,received <%+v>", expected, cgrEv)
	}
}
