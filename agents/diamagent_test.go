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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/sessions"
)

func TestDAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPClient(new(DiameterAgent))
}

type testMockSessionConn struct {
	calls map[string]func(arg interface{}, rply interface{}) error
}

func (s *testMockSessionConn) Call(method string, arg interface{}, rply interface{}) error {
	if call, has := s.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(arg, rply)
	}
}

func (s *testMockSessionConn) CallBiRPC(_ rpcclient.ClientConnector, method string, arg interface{}, rply interface{}) error {
	if call, has := s.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(arg, rply)
	}
}

func TestProcessRequest(t *testing.T) {
	dfltCfg, _ := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true, dfltCfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(config.CgrConfig(), nil, dm) // no need for filterS but still try to configure the dm :D

	cgrRplyNM := &utils.NavigableMap2{}
	rply := utils.NewOrderedNavigableMap()
	diamDP := utils.NavigableMap(map[string]interface{}{
		"SessionId":   "123456",
		"Account":     "1001",
		"Destination": "1003",
		"Usage":       10 * time.Second,
	})
	reqProcessor := &config.RequestProcessor{
		ID:      "Default",
		Tenant:  config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		Filters: []string{},
		RequestFields: []*config.FCTemplate{
			&config.FCTemplate{Tag: utils.ToR,
				Type: utils.META_CONSTANT, Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: utils.ToR}},
				PathSlice: []string{utils.MetaCgreq, utils.ToR},
				Value:     config.NewRSRParsersMustCompile(utils.VOICE, true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: utils.OriginID,
				Type: utils.META_COMPOSED, Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: utils.OriginID}},
				PathSlice: []string{utils.MetaCgreq, utils.OriginID},
				Value:     config.NewRSRParsersMustCompile("~*req.SessionId", true, utils.INFIELD_SEP), Mandatory: true},
			&config.FCTemplate{Tag: utils.OriginHost,
				Type: utils.MetaRemoteHost, Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: utils.OriginHost}},
				PathSlice: []string{utils.MetaCgreq, utils.OriginHost}, Mandatory: true},
			&config.FCTemplate{Tag: utils.Category,
				Type: utils.META_CONSTANT, Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: utils.Category}},
				PathSlice: []string{utils.MetaCgreq, utils.Category},
				Value:     config.NewRSRParsersMustCompile(utils.CALL, true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: utils.Account,
				Type: utils.META_COMPOSED, Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: utils.Account}},
				PathSlice: []string{utils.MetaCgreq, utils.Account},
				Value:     config.NewRSRParsersMustCompile("~*req.Account", true, utils.INFIELD_SEP), Mandatory: true},
			&config.FCTemplate{Tag: utils.Destination,
				Type: utils.META_COMPOSED, Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: utils.Destination}},
				PathSlice: []string{utils.MetaCgreq, utils.Destination},
				Value:     config.NewRSRParsersMustCompile("~*req.Destination", true, utils.INFIELD_SEP), Mandatory: true},
			&config.FCTemplate{Tag: utils.Usage,
				Type: utils.META_COMPOSED, Path: utils.PathItems{{Field: utils.MetaCgreq}, {Field: utils.Usage}},
				PathSlice: []string{utils.MetaCgreq, utils.Usage},
				Value:     config.NewRSRParsersMustCompile("~*req.Usage", true, utils.INFIELD_SEP), Mandatory: true},
		},
		ReplyFields: []*config.FCTemplate{
			&config.FCTemplate{Tag: "ResultCode",
				Type: utils.META_CONSTANT, Path: utils.PathItems{{Field: utils.MetaRep}, {Field: "ResultCode"}},
				PathSlice: []string{utils.MetaRep, "ResultCode"},
				Value:     config.NewRSRParsersMustCompile("2001", true, utils.INFIELD_SEP)},
			&config.FCTemplate{Tag: "GrantedUnits",
				Type: utils.MetaVariable, Path: utils.PathItems{{Field: utils.MetaRep}, {Field: "Granted-Service-Unit.CC-Time"}},
				PathSlice: []string{utils.MetaRep, "Granted-Service-Unit.CC-Time"},
				Value:     config.NewRSRParsersMustCompile("~*cgrep.MaxUsage{*duration_seconds}", true, utils.INFIELD_SEP),
				Mandatory: true},
		},
	}
	reqVars := utils.NavigableMap2{
		utils.OriginHost:  utils.NewNMInterface(config.CgrConfig().DiameterAgentCfg().OriginHost),
		utils.OriginRealm: utils.NewNMInterface(config.CgrConfig().DiameterAgentCfg().OriginRealm),
		utils.ProductName: utils.NewNMInterface(config.CgrConfig().DiameterAgentCfg().ProductName),
		utils.MetaApp:     utils.NewNMInterface("appName"),
		utils.MetaAppID:   utils.NewNMInterface("appID"),
		utils.MetaCmd:     utils.NewNMInterface("cmdR"),
	}

	sS := &testMockSessionConn{calls: map[string]func(arg interface{}, rply interface{}) error{
		utils.SessionSv1RegisterInternalBiJSONConn: func(arg interface{}, rply interface{}) error {
			return nil
		},
		utils.SessionSv1AuthorizeEvent: func(arg interface{}, rply interface{}) error {
			var tm *time.Time
			var id string
			if arg == nil {
				t.Errorf("args is nil")
			} else if rargs, can := arg.(*sessions.V1AuthorizeArgs); !can {
				t.Errorf("args is not of sessions.V1AuthorizeArgs type")
			} else {
				tm = rargs.Time // need time
				id = rargs.ID
			}
			expargs := &sessions.V1AuthorizeArgs{
				GetMaxUsage: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     id,
					Time:   tm,
					Event: map[string]interface{}{
						"Account":     "1001",
						"Category":    "call",
						"Destination": "1003",
						"OriginHost":  "local",
						"OriginID":    "123456",
						"ToR":         "*voice",
						"Usage":       "10s",
					},
				},
			}
			if !reflect.DeepEqual(expargs, arg) {
				t.Errorf("Expected:%s ,received: %s", utils.ToJSON(expargs), utils.ToJSON(arg))
			}
			prply, can := rply.(*sessions.V1AuthorizeReply)
			if !can {
				t.Errorf("Wrong argument type : %T", rply)
				return nil
			}
			*prply = sessions.V1AuthorizeReply{
				MaxUsage: utils.DurationPointer(time.Duration(-1)),
			}
			return nil
		},
		utils.SessionSv1InitiateSession: func(arg interface{}, rply interface{}) error {
			var tm *time.Time
			var id string
			if arg == nil {
				t.Errorf("args is nil")
			} else if rargs, can := arg.(*sessions.V1InitSessionArgs); !can {
				t.Errorf("args is not of sessions.V1InitSessionArgs type")
			} else {
				tm = rargs.Time // need time
				id = rargs.ID
			}
			expargs := &sessions.V1InitSessionArgs{
				GetAttributes: true,
				InitSession:   true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     id,
					Time:   tm,
					Event: map[string]interface{}{
						"Account":     "1001",
						"Category":    "call",
						"Destination": "1003",
						"OriginHost":  "local",
						"OriginID":    "123456",
						"ToR":         "*voice",
						"Usage":       "10s",
					},
				},
			}
			if !reflect.DeepEqual(expargs, arg) {
				t.Errorf("Expected:%s ,received: %s", utils.ToJSON(expargs), utils.ToJSON(arg))
			}
			prply, can := rply.(*sessions.V1InitSessionReply)
			if !can {
				t.Errorf("Wrong argument type : %T", rply)
				return nil
			}
			*prply = sessions.V1InitSessionReply{
				Attributes: &engine.AttrSProcessEventReply{
					MatchedProfiles: []string{"ATTR_1001_SESSIONAUTH"},
					AlteredFields:   []string{"*req.Password", "*req.PaypalAccount", "*req.RequestType", "*req.LCRProfile"},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "e7d35bf",
						Event: map[string]interface{}{
							"Account":       "1001",
							"CGRID":         "1133dc80896edf5049b46aa911cb9085eeb27f4c",
							"Category":      "call",
							"Destination":   "1003",
							"LCRProfile":    "premium_cli",
							"OriginHost":    "local",
							"OriginID":      "123456",
							"Password":      "CGRateS.org",
							"PaypalAccount": "cgrates@paypal.com",
							"RequestType":   "*prepaid",
							"ToR":           "*voice",
							"Usage":         "10s",
						},
					},
				},
				MaxUsage: utils.DurationPointer(10 * time.Second),
			}
			return nil
		},
		utils.SessionSv1UpdateSession: func(arg interface{}, rply interface{}) error {
			var tm *time.Time
			var id string
			if arg == nil {
				t.Errorf("args is nil")
			} else if rargs, can := arg.(*sessions.V1UpdateSessionArgs); !can {
				t.Errorf("args is not of sessions.V1UpdateSessionArgs type")
			} else {
				tm = rargs.Time // need time
				id = rargs.ID
			}
			expargs := &sessions.V1UpdateSessionArgs{
				GetAttributes: true,
				UpdateSession: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     id,
					Time:   tm,
					Event: map[string]interface{}{
						"Account":     "1001",
						"Category":    "call",
						"Destination": "1003",
						"OriginHost":  "local",
						"OriginID":    "123456",
						"ToR":         "*voice",
						"Usage":       "10s",
					},
				},
			}
			if !reflect.DeepEqual(expargs, arg) {
				t.Errorf("Expected:%s ,received: %s", utils.ToJSON(expargs), utils.ToJSON(arg))
			}
			prply, can := rply.(*sessions.V1UpdateSessionReply)
			if !can {
				t.Errorf("Wrong argument type : %T", rply)
				return nil
			}
			*prply = sessions.V1UpdateSessionReply{
				Attributes: &engine.AttrSProcessEventReply{
					MatchedProfiles: []string{"ATTR_1001_SESSIONAUTH"},
					AlteredFields:   []string{"*req.Password", "*req.PaypalAccount", "*req.RequestType", "*req.LCRProfile"},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "e7d35bf",
						Event: map[string]interface{}{
							"Account":       "1001",
							"CGRID":         "1133dc80896edf5049b46aa911cb9085eeb27f4c",
							"Category":      "call",
							"Destination":   "1003",
							"LCRProfile":    "premium_cli",
							"OriginHost":    "local",
							"OriginID":      "123456",
							"Password":      "CGRateS.org",
							"PaypalAccount": "cgrates@paypal.com",
							"RequestType":   "*prepaid",
							"ToR":           "*voice",
							"Usage":         "10s",
						},
					},
				},
				MaxUsage: utils.DurationPointer(10 * time.Second),
			}
			return nil
		},
		utils.SessionSv1ProcessCDR: func(arg interface{}, rply interface{}) error {
			var tm *time.Time
			var id string
			if arg == nil {
				t.Errorf("args is nil")
			} else if rargs, can := arg.(*utils.CGREventWithArgDispatcher); !can {
				t.Errorf("args is not of utils.CGREventWithArgDispatcher type")
			} else {
				tm = rargs.Time // need time
				id = rargs.ID
			}
			expargs := &utils.CGREventWithArgDispatcher{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     id,
					Time:   tm,
					Event: map[string]interface{}{
						"Account":     "1001",
						"Category":    "call",
						"Destination": "1003",
						"OriginHost":  "local",
						"OriginID":    "123456",
						"ToR":         "*voice",
						"Usage":       "10s",
					},
				},
			}
			if !reflect.DeepEqual(expargs, arg) {
				t.Errorf("Expected:%s ,received: %s", utils.ToJSON(expargs), utils.ToJSON(arg))
			}
			prply, can := rply.(*string)
			if !can {
				t.Errorf("Wrong argument type : %T", rply)
				return nil
			}
			*prply = utils.OK
			return nil
		},
		utils.SessionSv1TerminateSession: func(arg interface{}, rply interface{}) error {
			var tm *time.Time
			var id string
			if arg == nil {
				t.Errorf("args is nil")
			} else if rargs, can := arg.(*sessions.V1TerminateSessionArgs); !can {
				t.Errorf("args is not of sessions.V1TerminateSessionArgs type")
			} else {
				tm = rargs.Time // need time
				id = rargs.ID
			}
			expargs := &sessions.V1TerminateSessionArgs{
				TerminateSession: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     id,
					Time:   tm,
					Event: map[string]interface{}{
						"Account":     "1001",
						"Category":    "call",
						"Destination": "1003",
						"OriginHost":  "local",
						"OriginID":    "123456",
						"ToR":         "*voice",
						"Usage":       "10s",
					},
				},
			}
			if !reflect.DeepEqual(expargs, arg) {
				t.Errorf("Expected:%s ,received: %s", utils.ToJSON(expargs), utils.ToJSON(arg))
			}
			prply, can := rply.(*string)
			if !can {
				t.Errorf("Wrong argument type : %T", rply)
				return nil
			}
			*prply = utils.OK
			return nil
		},
		utils.SessionSv1ProcessMessage: func(arg interface{}, rply interface{}) error {
			var tm *time.Time
			var id string
			if arg == nil {
				t.Errorf("args is nil")
			} else if rargs, can := arg.(*sessions.V1ProcessMessageArgs); !can {
				t.Errorf("args is not of sessions.V1ProcessMessageArgs type")
			} else {
				tm = rargs.Time // need time
				id = rargs.ID
			}
			expargs := &sessions.V1ProcessMessageArgs{
				GetAttributes: true,
				Debit:         true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     id,
					Time:   tm,
					Event: map[string]interface{}{
						"Account":     "1001",
						"Category":    "call",
						"Destination": "1003",
						"OriginHost":  "local",
						"OriginID":    "123456",
						"ToR":         "*voice",
						"Usage":       "10s",
					},
				},
			}
			if !reflect.DeepEqual(expargs, arg) {
				t.Errorf("Expected:%s ,received: %s", utils.ToJSON(expargs), utils.ToJSON(arg))
			}
			prply, can := rply.(*sessions.V1ProcessMessageReply)
			if !can {
				t.Errorf("Wrong argument type : %T", rply)
				return nil
			}
			*prply = sessions.V1ProcessMessageReply{
				Attributes: &engine.AttrSProcessEventReply{
					MatchedProfiles: []string{"ATTR_1001_SESSIONAUTH"},
					AlteredFields:   []string{"*req.Password", "*req.PaypalAccount", "*req.RequestType", "*req.LCRProfile"},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "e7d35bf",
						Event: map[string]interface{}{
							"Account":       "1001",
							"CGRID":         "1133dc80896edf5049b46aa911cb9085eeb27f4c",
							"Category":      "call",
							"Destination":   "1003",
							"LCRProfile":    "premium_cli",
							"OriginHost":    "local",
							"OriginID":      "123456",
							"Password":      "CGRateS.org",
							"PaypalAccount": "cgrates@paypal.com",
							"RequestType":   "*prepaid",
							"ToR":           "*voice",
							"Usage":         "10s",
						},
					},
				},
				MaxUsage: utils.DurationPointer(10 * time.Second),
			}
			return nil
		},
	}}
	reqProcessor.Flags, _ = utils.FlagsWithParamsFromSlice([]string{utils.MetaAuth, utils.MetaAccounts})
	agReq := NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply,
		reqProcessor.Tenant, config.CgrConfig().GeneralCfg().DefaultTenant,
		config.CgrConfig().GeneralCfg().DefaultTimezone, filters, nil, nil)

	internalSessionSChan := make(chan rpcclient.ClientConnector, 1)
	internalSessionSChan <- sS
	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS): internalSessionSChan,
	})
	da := &DiameterAgent{
		cgrCfg:  config.CgrConfig(),
		filterS: filters,
		connMgr: connMgr,
	}
	pr, err := da.processRequest(reqProcessor, agReq)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 2 {
		t.Errorf("Expected the reply to have 2 values received: %s", rply.String())
	}

	reqProcessor.Flags, _ = utils.FlagsWithParamsFromSlice([]string{utils.MetaInitiate, utils.MetaAccounts, utils.MetaAttributes})
	cgrRplyNM = &utils.NavigableMap2{}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply,
		reqProcessor.Tenant, config.CgrConfig().GeneralCfg().DefaultTenant,
		config.CgrConfig().GeneralCfg().DefaultTimezone, filters, nil, nil)

	pr, err = da.processRequest(reqProcessor, agReq)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 2 {
		t.Errorf("Expected the reply to have 2 values received: %s", rply.String())
	}

	reqProcessor.Flags, _ = utils.FlagsWithParamsFromSlice([]string{utils.MetaUpdate, utils.MetaAccounts, utils.MetaAttributes})
	cgrRplyNM = &utils.NavigableMap2{}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply,
		reqProcessor.Tenant, config.CgrConfig().GeneralCfg().DefaultTenant,
		config.CgrConfig().GeneralCfg().DefaultTimezone, filters, nil, nil)

	pr, err = da.processRequest(reqProcessor, agReq)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 2 {
		t.Errorf("Expected the reply to have 2 values received: %s", rply.String())
	}

	reqProcessor.Flags, _ = utils.FlagsWithParamsFromSlice([]string{utils.MetaTerminate, utils.MetaAccounts, utils.MetaAttributes, utils.MetaCDRs})
	reqProcessor.ReplyFields = []*config.FCTemplate{&config.FCTemplate{Tag: "ResultCode",
		PathSlice: []string{utils.MetaRep, "ResultCode"},
		Type:      utils.META_CONSTANT, Path: utils.PathItems{{Field: utils.MetaRep}, {Field: "ResultCode"}},
		Value: config.NewRSRParsersMustCompile("2001", true, utils.INFIELD_SEP)}}
	cgrRplyNM = &utils.NavigableMap2{}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply,
		reqProcessor.Tenant, config.CgrConfig().GeneralCfg().DefaultTenant,
		config.CgrConfig().GeneralCfg().DefaultTimezone, filters, nil, nil)

	pr, err = da.processRequest(reqProcessor, agReq)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 1 {
		t.Errorf("Expected the reply to have one value received: %s", rply.String())
	}

	reqProcessor.Flags, _ = utils.FlagsWithParamsFromSlice([]string{utils.MetaMessage, utils.MetaAccounts, utils.MetaAttributes})
	cgrRplyNM = &utils.NavigableMap2{}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply,
		reqProcessor.Tenant, config.CgrConfig().GeneralCfg().DefaultTenant,
		config.CgrConfig().GeneralCfg().DefaultTimezone, filters, nil, nil)

	pr, err = da.processRequest(reqProcessor, agReq)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 1 {
		t.Errorf("Expected the reply to have one value received: %s", rply.String())
	}

}
