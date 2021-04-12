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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestDAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPClient(new(DiameterAgent))
}

type testMockSessionConn struct {
	calls map[string]func(_ *context.Context, _, _ interface{}) error
}

func (s *testMockSessionConn) Call(ctx *context.Context, method string, arg, rply interface{}) error {
	if call, has := s.calls[method]; has {
		return call(ctx, arg, rply)
	}
	return rpcclient.ErrUnsupporteServiceMethod
}

func (s *testMockSessionConn) Handlers() (b map[string]interface{}) {
	b = make(map[string]interface{})
	for n, f := range s.calls {
		b[n] = f
	}
	return
}

func TestProcessRequest(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(config.CgrConfig(), nil, dm) // no need for filterS but still try to configure the dm :D

	cgrRplyNM := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rply := utils.NewOrderedNavigableMap()
	diamDP := utils.MapStorage{
		"SessionId":   "123456",
		"Account":     "1001",
		"Destination": "1003",
		"Usage":       10 * time.Second,
	}
	reqProcessor := &config.RequestProcessor{
		ID:      "Default",
		Tenant:  config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
		Filters: []string{},
		RequestFields: []*config.FCTemplate{
			{Tag: utils.ToR,
				Type: utils.MetaConstant, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR,
				Value: config.NewRSRParsersMustCompile(utils.MetaVoice, utils.InfieldSep)},
			{Tag: utils.OriginID,
				Type: utils.MetaComposed, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID,
				Value: config.NewRSRParsersMustCompile("~*req.SessionId", utils.InfieldSep), Mandatory: true},
			{Tag: utils.OriginHost,
				Type: utils.MetaRemoteHost, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginHost, Mandatory: true},
			{Tag: utils.Category,
				Type: utils.MetaConstant, Path: utils.MetaCgreq + utils.NestingSep + utils.Category,
				Value: config.NewRSRParsersMustCompile(utils.Call, utils.InfieldSep)},
			{Tag: utils.AccountField,
				Type: utils.MetaComposed, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField,
				Value: config.NewRSRParsersMustCompile("~*req.Account", utils.InfieldSep), Mandatory: true},
			{Tag: utils.Destination,
				Type: utils.MetaComposed, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination,
				Value: config.NewRSRParsersMustCompile("~*req.Destination", utils.InfieldSep), Mandatory: true},
			{Tag: utils.Usage,
				Type: utils.MetaComposed, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage,
				Value: config.NewRSRParsersMustCompile("~*req.Usage", utils.InfieldSep), Mandatory: true},
		},
		ReplyFields: []*config.FCTemplate{
			{Tag: "ResultCode",
				Type: utils.MetaConstant, Path: utils.MetaRep + utils.NestingSep + "ResultCode",
				Value: config.NewRSRParsersMustCompile("2001", utils.InfieldSep)},
			{Tag: "GrantedUnits",
				Type: utils.MetaVariable, Path: utils.MetaRep + utils.NestingSep + "Granted-Service-Unit.CC-Time",
				Value:     config.NewRSRParsersMustCompile("~*cgrep.MaxUsage{*duration_seconds}", utils.InfieldSep),
				Mandatory: true},
		},
	}
	for _, v := range reqProcessor.RequestFields {
		v.ComputePath()
	}
	for _, v := range reqProcessor.ReplyFields {
		v.ComputePath()
	}
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{
		utils.OriginHost:  utils.NewLeafNode(config.CgrConfig().DiameterAgentCfg().OriginHost),
		utils.OriginRealm: utils.NewLeafNode(config.CgrConfig().DiameterAgentCfg().OriginRealm),
		utils.ProductName: utils.NewLeafNode(config.CgrConfig().DiameterAgentCfg().ProductName),
		utils.MetaApp:     utils.NewLeafNode("appName"),
		utils.MetaAppID:   utils.NewLeafNode("appID"),
		utils.MetaCmd:     utils.NewLeafNode("cmdR"),
	}}

	sS := &testMockSessionConn{calls: map[string]func(_ *context.Context, _, _ interface{}) error{
		utils.SessionSv1RegisterInternalBiJSONConn: func(_ *context.Context, _, _ interface{}) error {
			return nil
		},
		utils.SessionSv1AuthorizeEvent: func(_ *context.Context, arg, rply interface{}) error {
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
					APIOpts: map[string]interface{}{},
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
				MaxUsage: utils.DurationPointer(-1),
			}
			return nil
		},
		utils.SessionSv1InitiateSession: func(_ *context.Context, arg, rply interface{}) error {
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
					APIOpts: map[string]interface{}{},
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
		utils.SessionSv1UpdateSession: func(_ *context.Context, arg, rply interface{}) error {
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
					APIOpts: map[string]interface{}{},
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
		utils.SessionSv1ProcessCDR: func(_ *context.Context, arg, rply interface{}) error {
			var tm *time.Time
			var id string
			if arg == nil {
				t.Errorf("args is nil")
			} else if rargs, can := arg.(*utils.CGREvent); !can {
				t.Errorf("args is not of utils.CGREventWithOpts type")
			} else {
				tm = rargs.Time // need time
				id = rargs.ID
			}
			expargs := &utils.CGREvent{
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
				APIOpts: make(map[string]interface{}),
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
		utils.SessionSv1TerminateSession: func(_ *context.Context, arg, rply interface{}) error {
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
					APIOpts: map[string]interface{}{},
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
		utils.SessionSv1ProcessMessage: func(_ *context.Context, arg, rply interface{}) error {
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
					APIOpts: map[string]interface{}{},
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
	reqProcessor.Flags = utils.FlagsWithParamsFromSlice([]string{utils.MetaAuthorize, utils.MetaAccounts})
	agReq := NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply, nil,
		reqProcessor.Tenant, config.CgrConfig().GeneralCfg().DefaultTenant,
		config.CgrConfig().GeneralCfg().DefaultTimezone, filters, nil, nil)

	internalSessionSChan := make(chan birpc.ClientConnector, 1)
	internalSessionSChan <- sS
	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS):      internalSessionSChan,
		utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS): internalSessionSChan,
	})
	da := &DiameterAgent{
		cgrCfg:  config.CgrConfig(),
		filterS: filters,
		connMgr: connMgr,
	}
	da.ctx = context.WithClient(context.Background(), da)
	pr, err := da.processRequest(reqProcessor, agReq)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 2 {
		t.Errorf("Expected the reply to have 2 values received: %s", rply.String())
	}

	reqProcessor.Flags = utils.FlagsWithParamsFromSlice([]string{utils.MetaInitiate, utils.MetaAccounts, utils.MetaAttributes})
	cgrRplyNM = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply, nil,
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

	reqProcessor.Flags = utils.FlagsWithParamsFromSlice([]string{utils.MetaUpdate, utils.MetaAccounts, utils.MetaAttributes})
	cgrRplyNM = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply, nil,
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

	reqProcessor.Flags = utils.FlagsWithParamsFromSlice([]string{utils.MetaTerminate, utils.MetaAccounts, utils.MetaAttributes, utils.MetaCDRs})
	reqProcessor.ReplyFields = []*config.FCTemplate{{Tag: "ResultCode",
		Type: utils.MetaConstant, Path: utils.MetaRep + utils.NestingSep + "ResultCode",
		Value: config.NewRSRParsersMustCompile("2001", utils.InfieldSep)}}
	for _, v := range reqProcessor.ReplyFields {
		v.ComputePath()
	}
	cgrRplyNM = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply, nil,
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

	reqProcessor.Flags = utils.FlagsWithParamsFromSlice([]string{utils.MetaMessage, utils.MetaAccounts, utils.MetaAttributes})
	cgrRplyNM = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply, nil,
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
