// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
	"github.com/cgrates/rpcclient"
)

func TestDAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPCClient(new(DiameterAgent))
}

type testMockSessionConn struct {
	calls map[string]func(_ *context.Context, _, _ any) error
}

func (s *testMockSessionConn) Call(ctx *context.Context, method string, arg, rply any) error {
	if call, has := s.calls[method]; has {
		return call(ctx, arg, rply)
	}
	return rpcclient.ErrUnsupporteServiceMethod
}

func (s *testMockSessionConn) Handlers() (b map[string]any) {
	b = make(map[string]any)
	for n, f := range s.calls {
		b[n] = f
	}
	return
}

func TestProcessRequest(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm.SetCache(cacheS)
	filters := engine.NewFilterS(cfg, nil, dm) // no need for filterS but still try to configure the dm :D

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
		Tenant:  utils.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
		Filters: []string{},
		RequestFields: []*config.FCTemplate{
			{Tag: utils.ToR,
				Type: utils.MetaConstant, Path: utils.MetaCgreq + utils.NestingSep + utils.ToR,
				Value: utils.NewRSRParsersMustCompile(utils.MetaVoice, utils.InfieldSep)},
			{Tag: utils.OriginID,
				Type: utils.MetaComposed, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginID,
				Value: utils.NewRSRParsersMustCompile("~*req.SessionId", utils.InfieldSep), Mandatory: true},
			{Tag: utils.OriginHost,
				Type: utils.MetaVariable, Path: utils.MetaCgreq + utils.NestingSep + utils.OriginHost,
				Value: utils.NewRSRParsersMustCompile("~*vars.RemoteHost", utils.InfieldSep), Mandatory: true},
			{Tag: utils.Category,
				Type: utils.MetaConstant, Path: utils.MetaCgreq + utils.NestingSep + utils.Category,
				Value: utils.NewRSRParsersMustCompile(utils.Call, utils.InfieldSep)},
			{Tag: utils.AccountField,
				Type: utils.MetaComposed, Path: utils.MetaCgreq + utils.NestingSep + utils.AccountField,
				Value: utils.NewRSRParsersMustCompile("~*req.Account", utils.InfieldSep), Mandatory: true},
			{Tag: utils.Destination,
				Type: utils.MetaComposed, Path: utils.MetaCgreq + utils.NestingSep + utils.Destination,
				Value: utils.NewRSRParsersMustCompile("~*req.Destination", utils.InfieldSep), Mandatory: true},
			{Tag: utils.Usage,
				Type: utils.MetaComposed, Path: utils.MetaCgreq + utils.NestingSep + utils.Usage,
				Value: utils.NewRSRParsersMustCompile("~*req.Usage", utils.InfieldSep), Mandatory: true},
		},
		ReplyFields: []*config.FCTemplate{
			{Tag: "ResultCode",
				Type: utils.MetaConstant, Path: utils.MetaRep + utils.NestingSep + "ResultCode",
				Value: utils.NewRSRParsersMustCompile("2001", utils.InfieldSep)},
			{Tag: "GrantedUnits",
				Type: utils.MetaVariable, Path: utils.MetaRep + utils.NestingSep + "Granted-Service-Unit.CC-Time",
				Value:     utils.NewRSRParsersMustCompile("~*cgrep.MaxUsage.*primary{*durationSeconds}", utils.InfieldSep),
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
		utils.OriginHost:  utils.NewLeafNode(cfg.DiameterAgentCfg().OriginHost),
		utils.OriginRealm: utils.NewLeafNode(cfg.DiameterAgentCfg().OriginRealm),
		utils.ProductName: utils.NewLeafNode(cfg.DiameterAgentCfg().ProductName),
		utils.MetaApp:     utils.NewLeafNode("appName"),
		utils.MetaAppID:   utils.NewLeafNode("appID"),
		utils.MetaCmd:     utils.NewLeafNode("cmdR"),
		utils.RemoteHost:  utils.NewLeafNode(utils.LocalAddr().String()),
	}}

	expAPIOpts := map[string]any{}
	sS := &testMockSessionConn{calls: map[string]func(_ *context.Context, _, _ any) error{
		utils.SessionSv1RegisterInternalBiJSONConn: func(_ *context.Context, _, _ any) error {
			return nil
		},
		utils.SessionSv1ProcessEvent: func(_ *context.Context, arg, rply any) error {
			var id string
			if arg == nil {
				t.Errorf("args is nil")
			} else if rargs, can := arg.(*utils.CGREvent); !can {
				t.Errorf("args is not of utils.CGREvent type")
			} else {
				id = rargs.ID
			}
			expargs := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     id,
				Event: map[string]any{
					"Account":     "1001",
					"Category":    "call",
					"Destination": "1003",
					"OriginHost":  "local",
					"OriginID":    "123456",
					"ToR":         "*voice",
					"Usage":       "10s",
				},
				APIOpts: expAPIOpts,
			}
			if !reflect.DeepEqual(expargs, arg) {
				t.Errorf("Expected:%s ,received: %s", utils.ToJSON(expargs), utils.ToJSON(arg))
			}
			prply, can := rply.(*sessions.V1ProcessEventReply)
			if !can {
				t.Errorf("Wrong argument type : %T", rply)
				return nil
			}
			*prply = sessions.V1ProcessEventReply{
				AccountsUsage: map[string]time.Duration{utils.MetaPrimary: 10 * time.Second},
			}
			return nil
		},
	}}
	reqProcessor.Flags = utils.FlagsWithParamsFromSlice([]string{utils.MetaSessionS})
	agReq := NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply, nil,
		reqProcessor.Tenant, cfg.GeneralCfg().DefaultTenant,
		cfg.GeneralCfg().DefaultTimezone, cfg, cacheS, filters, nil)

	internalSessionSChan := make(chan birpc.ClientConnector, 1)
	internalSessionSChan <- sS
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), utils.SessionSv1, internalSessionSChan)
	connMgr.AddInternalConn(utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS), utils.SessionSv1, internalSessionSChan)
	da := &DiameterAgent{
		cgrCfg:  cfg,
		fltrS:   filters,
		connMgr: connMgr,
	}
	srv, _ := birpc.NewService(da, "", false)
	da.ctx = context.WithClient(context.Background(), srv)
	tmpls := []*config.FCTemplate{
		{Tag: "ResultCode",
			Type: utils.MetaConstant, Path: utils.MetaRep + utils.NestingSep + "ResultCode",
			Value: utils.NewRSRParsersMustCompile("2001", utils.InfieldSep)},
		{Tag: "GrantedUnits",
			Type: utils.MetaVariable, Path: utils.MetaRep + utils.NestingSep + "Granted-Service-Unit.CC-Time",
			Value:     utils.NewRSRParsersMustCompile("~*cgrep.MaxUsage.*primary{*durationSeconds}", utils.InfieldSep),
			Mandatory: true},
	}
	for _, v := range tmpls {
		v.ComputePath()
	}
	clnReq := reqProcessor.Clone()
	clnReq.ReplyFields = tmpls
	pr, err := processRequest(da.ctx, clnReq, agReq, utils.DiameterAgent, connMgr, []string{"*internal"}, nil, nil, da.fltrS)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 2 {
		t.Errorf("Expected the reply to have 2 values received: %s", rply.String())
	}

	reqProcessor.Flags = utils.FlagsWithParamsFromSlice([]string{utils.MetaSessionS})

	tmpls2 := []*config.FCTemplate{
		{Type: utils.MetaConstant, Path: utils.MetaOpts + utils.NestingSep + utils.MetaSession,
			Value: utils.NewRSRParsersMustCompile("true", utils.InfieldSep)},
		{Type: utils.MetaConstant, Path: utils.MetaOpts + utils.NestingSep + utils.MetaAttributes,
			Value: utils.NewRSRParsersMustCompile("true", utils.InfieldSep)},
	}
	for _, v := range tmpls2 {
		v.ComputePath()
	}
	clnReq2 := reqProcessor.Clone()
	clnReq2.RequestFields = append(clnReq2.RequestFields, tmpls2...)
	cgrRplyNM = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply, nil,
		reqProcessor.Tenant, cfg.GeneralCfg().DefaultTenant,
		cfg.GeneralCfg().DefaultTimezone, cfg, cacheS, filters, nil)

	expAPIOpts = map[string]any{
		utils.MetaAttributes: "true",
		utils.MetaSession:    "true",
	}
	pr, err = processRequest(da.ctx, clnReq2, agReq, utils.DiameterAgent, connMgr, []string{"*internal"}, nil, nil, da.fltrS)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 2 {
		t.Errorf("Expected the reply to have 2 values received: %s", rply.String())
	}

	reqProcessor.Flags = utils.FlagsWithParamsFromSlice([]string{utils.MetaSessionS})

	tmpls = []*config.FCTemplate{
		{Type: utils.MetaConstant, Path: utils.MetaOpts + utils.NestingSep + utils.MetaSession,
			Value: utils.NewRSRParsersMustCompile("true", utils.InfieldSep)},
		{Type: utils.MetaConstant, Path: utils.MetaOpts + utils.NestingSep + utils.MetaAttributes,
			Value: utils.NewRSRParsersMustCompile("true", utils.InfieldSep)},
	}
	for _, v := range tmpls {
		v.ComputePath()
	}
	clnReq = reqProcessor.Clone()
	clnReq.RequestFields = append(clnReq.RequestFields, tmpls...)
	cgrRplyNM = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply, nil,
		reqProcessor.Tenant, cfg.GeneralCfg().DefaultTenant,
		cfg.GeneralCfg().DefaultTimezone, cfg, cacheS, filters, nil)

	expAPIOpts = map[string]any{
		utils.MetaAttributes: "true",
		utils.MetaSession:    "true",
	}
	pr, err = processRequest(da.ctx, clnReq, agReq, utils.DiameterAgent, connMgr, []string{"*internal"}, nil, nil, da.fltrS)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 2 {
		t.Errorf("Expected the reply to have 2 values received: %s", rply.String())
	}
	reqProcessor.Flags = utils.FlagsWithParamsFromSlice([]string{utils.MetaSessionS})
	tmpls = []*config.FCTemplate{
		{Type: utils.MetaConstant, Path: utils.MetaOpts + utils.NestingSep + utils.MetaTerminate,
			Value: utils.NewRSRParsersMustCompile("true", utils.InfieldSep)},
	}
	for _, v := range tmpls {
		v.ComputePath()
	}
	reqProcessor.ReplyFields = []*config.FCTemplate{{Tag: "ResultCode",
		Type: utils.MetaConstant, Path: utils.MetaRep + utils.NestingSep + "ResultCode",
		Value: utils.NewRSRParsersMustCompile("2001", utils.InfieldSep)}}
	for _, v := range reqProcessor.ReplyFields {
		v.ComputePath()
	}
	clnReq = reqProcessor.Clone()
	clnReq.RequestFields = append(clnReq.RequestFields, tmpls...)
	cgrRplyNM = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply, nil,
		reqProcessor.Tenant, cfg.GeneralCfg().DefaultTenant,
		cfg.GeneralCfg().DefaultTimezone, cfg, cacheS, filters, nil)

	expAPIOpts = map[string]any{
		utils.MetaTerminate: "true",
	}
	pr, err = processRequest(da.ctx, clnReq, agReq, utils.DiameterAgent, connMgr, []string{"*internal"}, nil, nil, da.fltrS)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 1 {
		t.Errorf("Expected the reply to have one value received: %s", rply.String())
	}

	reqProcessor.Flags = utils.FlagsWithParamsFromSlice([]string{utils.MetaSessionS})
	tmpls = []*config.FCTemplate{
		{Type: utils.MetaConstant, Path: utils.MetaOpts + utils.NestingSep + utils.MetaSession,
			Value: utils.NewRSRParsersMustCompile("true", utils.InfieldSep)},
		{Type: utils.MetaConstant, Path: utils.MetaOpts + utils.NestingSep + utils.MetaTerminate,
			Value: utils.NewRSRParsersMustCompile("true", utils.InfieldSep)},
		{Type: utils.MetaConstant, Path: utils.MetaOpts + utils.NestingSep + utils.MetaAttributes,
			Value: utils.NewRSRParsersMustCompile("true", utils.InfieldSep)},
	}
	for _, v := range tmpls {
		v.ComputePath()
	}
	clnReq = reqProcessor.Clone()
	clnReq.RequestFields = append(clnReq.RequestFields, tmpls...)
	cgrRplyNM = &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rply = utils.NewOrderedNavigableMap()

	agReq = NewAgentRequest(diamDP, reqVars, cgrRplyNM, rply, nil,
		reqProcessor.Tenant, cfg.GeneralCfg().DefaultTenant,
		cfg.GeneralCfg().DefaultTimezone, cfg, cacheS, filters, nil)

	expAPIOpts = map[string]any{
		utils.MetaSession:    "true",
		utils.MetaTerminate:  "true",
		utils.MetaAttributes: "true",
	}
	pr, err = processRequest(da.ctx, clnReq, agReq, utils.DiameterAgent, connMgr, []string{"*internal"}, nil, nil, da.fltrS)
	if err != nil {
		t.Error(err)
	} else if !pr {
		t.Errorf("Expected the request to be processed")
	} else if len(rply.GetOrder()) != 1 {
		t.Errorf("Expected the reply to have one value received: %s", rply.String())
	}
}

type capturingConn struct {
	diam.Conn
	written [][]byte
}

func (c *capturingConn) Write(b []byte) (int, error) {
	c.written = append(c.written, append([]byte(nil), b...))
	return len(b), nil
}

func (c *capturingConn) WriteStream(b []byte, _ uint) (int, error) { return c.Write(b) }

// Test to check the flags of RAR where REQ and PXY flags are present
func TestServerInitiatedRequestsAreProxiable(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	conn := &capturingConn{}

	// The cached CCR the agent builds the RAR from.
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("sess1"))
	ccr.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("pgw.example.org"))
	ccr.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("example.org"))

	locker := engine.NewLocker(cfg)
	da := &DiameterAgent{
		cgrCfg: cfg,
		cache:  engine.NewCacheS(cfg, nil, nil, nil, locker),
		raa:    make(map[string]chan *diam.Message),
	}
	da.cache.Set(context.TODO(), utils.CacheDiameterMessages, "sess1",
		&diamMsgData{c: conn, m: ccr, vars: &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}}, nil, true, utils.NonTransactional)
	defer da.cache.Remove(context.TODO(), utils.CacheDiameterMessages, "sess1",
		true, utils.NonTransactional)
	cgrEv := utils.CGREvent{Event: map[string]any{utils.OriginID: "sess1"}}
	var reply string
	if err := da.V1AlterSession(nil, cgrEv, &reply); err != utils.ErrTimedOut {
		t.Fatalf("expected ErrTimedOut waiting for the RAA, got: %v", err)
	}

	if len(conn.written) != 1 {
		t.Fatalf("expected exactly one request on the wire, got %d", len(conn.written))
	}

	var hdr diam.Header
	if err := hdr.DecodeFromBytes(conn.written[0]); err != nil {
		t.Fatalf("could not decode the request we wrote: %v", err)
	}
	if hdr.CommandCode != diam.ReAuth {
		t.Errorf("expected a RAR (258), got command code %d", hdr.CommandCode)
	}
	if hdr.CommandFlags&diam.RequestFlag == 0 {
		t.Error("the R bit is not set; this is not a request")
	}
	if hdr.CommandFlags&diam.ProxiableFlag == 0 {
		t.Error("the P bit is not set: a relay agent answers 3001 and the RAR never reaches the gateway")
	}
}
