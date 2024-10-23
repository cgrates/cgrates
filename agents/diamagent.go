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
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/avp"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
	"github.com/fiorix/go-diameter/v4/diam/dict"
	"github.com/fiorix/go-diameter/v4/diam/sm"
	"github.com/fiorix/go-diameter/v4/diam/sm/smpeer"
)

const (
	all = "ALL"
	raa = "RAA"
	dpa = "DPA"
)

// NewDiameterAgent initializes a new DiameterAgent
func NewDiameterAgent(cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	connMgr *engine.ConnManager, caps *engine.Caps) (*DiameterAgent, error) {
	da := &DiameterAgent{
		cgrCfg:  cgrCfg,
		filterS: filterS,
		connMgr: connMgr,
		caps:    caps,
		raa:     make(map[string]chan *diam.Message),
		dpa:     make(map[string]chan *diam.Message),
		peers:   make(map[string]diam.Conn),
	}
	srv, err := birpc.NewServiceWithMethodsRename(da, utils.AgentV1, true, func(oldFn string) (newFn string) {
		return strings.TrimPrefix(oldFn, "V1")
	})
	if err != nil {
		return nil, err
	}
	da.ctx = context.WithClient(context.TODO(), srv)
	dictsPath := cgrCfg.DiameterAgentCfg().DictionariesPath
	if len(dictsPath) != 0 {
		if err := loadDictionaries(dictsPath, utils.DiameterAgent); err != nil {
			return nil, err
		}
	}
	msgTemplates := da.cgrCfg.TemplatesCfg()
	// Inflate *template field types
	for _, procsr := range da.cgrCfg.DiameterAgentCfg().RequestProcessors {
		if tpls, err := config.InflateTemplates(procsr.RequestFields, msgTemplates); err != nil {
			return nil, err
		} else if tpls != nil {
			procsr.RequestFields = tpls
		}
		if tpls, err := config.InflateTemplates(procsr.ReplyFields, msgTemplates); err != nil {
			return nil, err
		} else if tpls != nil {
			procsr.ReplyFields = tpls
		}
	}
	return da, nil
}

// DiameterAgent describes the diameter server
type DiameterAgent struct {
	cgrCfg  *config.CGRConfig
	filterS *engine.FilterS
	connMgr *engine.ConnManager
	caps    *engine.Caps

	raaLck   sync.RWMutex
	raa      map[string]chan *diam.Message
	peersLck sync.Mutex
	peers    map[string]diam.Conn // peer index by OriginHost;OriginRealm
	dpaLck   sync.RWMutex
	dpa      map[string]chan *diam.Message

	ctx *context.Context
}

// ListenAndServe is called when DiameterAgent is started, usually from within cmd/cgr-engine
func (da *DiameterAgent) ListenAndServe(stopChan <-chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> Start listening on <%s>", utils.DiameterAgent, da.cgrCfg.DiameterAgentCfg().Listen))
	srv := &diam.Server{
		Network: da.cgrCfg.DiameterAgentCfg().ListenNet,
		Addr:    da.cgrCfg.DiameterAgentCfg().Listen,
		Handler: da.handlers(),
		Dict:    nil,
	}
	// used to control the server state
	var lsn net.Listener
	if lsn, err = diam.MultistreamListen(utils.FirstNonEmpty(srv.Network, utils.TCP),
		utils.FirstNonEmpty(srv.Addr, ":3868")); err != nil {
		return
	}
	errChan := make(chan error)
	go func() {
		errChan <- srv.Serve(lsn)
	}()
	select {
	case err = <-errChan:
		return
	case <-stopChan:
		return lsn.Close()
	}
}

// Creates the message handlers
func (da *DiameterAgent) handlers() diam.Handler {
	settings := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(da.cgrCfg.DiameterAgentCfg().OriginHost),
		OriginRealm:      datatype.DiameterIdentity(da.cgrCfg.DiameterAgentCfg().OriginRealm),
		VendorID:         datatype.Unsigned32(da.cgrCfg.DiameterAgentCfg().VendorID),
		ProductName:      datatype.UTF8String(da.cgrCfg.DiameterAgentCfg().ProductName),
		FirmwareRevision: datatype.Unsigned32(utils.DiameterFirmwareRevision),
	}
	hosts := disectDiamListen(da.cgrCfg.DiameterAgentCfg().Listen)
	if len(hosts) == 0 {
		interfaces, err := net.Interfaces()
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error : %v, when querying interfaces for address", utils.DiameterAgent, err))
		}
		for _, inter := range interfaces {
			addrs, err := inter.Addrs()
			if err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> error: %+v, when taking address from interface: %+v",
					utils.DiameterAgent, err, inter.Name))
				continue
			}
			for _, iAddr := range addrs {
				hosts = append(hosts, net.ParseIP(strings.Split(iAddr.String(), utils.HDRValSep)[0])) // address came in form x.y.z.t/24
			}
		}
	}
	settings.HostIPAddresses = make([]datatype.Address, len(hosts))
	for i, host := range hosts {
		settings.HostIPAddresses[i] = datatype.Address(host)
	}

	dSM := sm.New(settings)
	if da.cgrCfg.DiameterAgentCfg().SyncedConnReqs {
		dSM.HandleFunc(all, da.handleMessage)
		dSM.HandleFunc(raa, da.handleRAA)
		dSM.HandleFunc(dpa, da.handleDPA)
	} else {
		dSM.HandleFunc(all, func(c diam.Conn, m *diam.Message) { go da.handleMessage(c, m) })
		dSM.HandleFunc(raa, func(c diam.Conn, m *diam.Message) { go da.handleRAA(c, m) })
		dSM.HandleFunc(dpa, func(c diam.Conn, m *diam.Message) { go da.handleDPA(c, m) })
	}
	go da.handleConns(dSM.HandshakeNotify())
	go func() {
		for err := range dSM.ErrorReports() {
			utils.Logger.Err(fmt.Sprintf("<%s> sm error: %v", utils.DiameterAgent, err))
		}
	}()
	return dSM
}

// handleALL is the handler of all messages coming in via Diameter
func (da *DiameterAgent) handleMessage(c diam.Conn, m *diam.Message) {
	dApp, err := m.Dictionary().App(m.Header.ApplicationID)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> decoding app: %d, err: %s",
			utils.DiameterAgent, m.Header.ApplicationID, err.Error()))
		writeOnConn(c, diamErrMsg(m, diam.NoCommonApplication, err.Error()))
		return
	}
	dCmd, err := m.Dictionary().FindCommand(
		m.Header.ApplicationID,
		m.Header.CommandCode)
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> decoding app: %d, command %d, err: %s",
			utils.DiameterAgent, m.Header.ApplicationID, m.Header.CommandCode, err.Error()))
		writeOnConn(c, diamErrMsg(m, diam.CommandUnsupported, err.Error()))
		return
	}
	diamDP := newDADataProvider(c, m)
	reqVars := &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			utils.OriginHost:  utils.NewLeafNode(da.cgrCfg.DiameterAgentCfg().OriginHost), // used in templates
			utils.OriginRealm: utils.NewLeafNode(da.cgrCfg.DiameterAgentCfg().OriginRealm),
			utils.ProductName: utils.NewLeafNode(da.cgrCfg.DiameterAgentCfg().ProductName),
			utils.MetaApp:     utils.NewLeafNode(dApp.Name),
			utils.MetaAppID:   utils.NewLeafNode(dApp.ID),
			utils.MetaCmd:     utils.NewLeafNode(dCmd.Short + "R"),
			utils.RemoteHost:  utils.NewLeafNode(c.RemoteAddr().String()),
		},
	}
	if da.caps.IsLimited() {
		if err := da.caps.Allocate(); err != nil {
			diamErr(c, m, diam.TooBusy, reqVars, da.cgrCfg, da.filterS)
			return
		}
		defer da.caps.Deallocate()
	}

	// cache message for ASR
	if da.cgrCfg.DiameterAgentCfg().ASRTemplate != "" ||
		da.cgrCfg.DiameterAgentCfg().RARTemplate != "" {
		sessID, err := diamDP.FieldAsString([]string{"Session-Id"})
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed retrieving Session-Id err: %s, message: %s",
					utils.DiameterAgent, err.Error(), m))
			diamErr(c, m, diam.UnableToComply, reqVars, da.cgrCfg, da.filterS)
			return
		}
		// cache message data needed for building up the ASR
		if errCh := engine.Cache.Set(utils.CacheDiameterMessages, sessID, &diamMsgData{c, m, reqVars},
			nil, true, utils.NonTransactional); errCh != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed message: %s to set Cache: %s", utils.DiameterAgent, m, errCh.Error()))
			diamErr(c, m, diam.UnableToComply, reqVars, da.cgrCfg, da.filterS)
			return
		}
	}

	cgrRplyNM := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	opts := utils.MapStorage{}
	rply := utils.NewOrderedNavigableMap() // share it among different processors
	var processed bool
	for _, reqProcessor := range da.cgrCfg.DiameterAgentCfg().RequestProcessors {
		var lclProcessed bool
		lclProcessed, err = processRequest(
			da.ctx,
			reqProcessor,
			NewAgentRequest(
				diamDP, reqVars, cgrRplyNM, rply,
				opts, reqProcessor.Tenant,
				da.cgrCfg.GeneralCfg().DefaultTenant,
				utils.FirstNonEmpty(
					reqProcessor.Timezone,
					da.cgrCfg.GeneralCfg().DefaultTimezone,
				),
				da.filterS, nil),
			utils.DiameterAgent, da.connMgr,
			da.cgrCfg.DiameterAgentCfg().SessionSConns,
			da.filterS)
		if lclProcessed {
			processed = lclProcessed
		}
		if err != nil ||
			(lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue)) {
			break
		}
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing message: %s",
				utils.DiameterAgent, err.Error(), m))
		diamErr(c, m, diam.UnableToComply, reqVars, da.cgrCfg, da.filterS)
		return
	}
	if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring message %s from %s",
				utils.DiameterAgent, m, c.RemoteAddr()))
		diamErr(c, m, diam.UnableToComply, reqVars, da.cgrCfg, da.filterS)
		return
	}
	a, err := diamAnswer(m, 0, false,
		rply, da.cgrCfg.GeneralCfg().DefaultTimezone)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> err: %s, replying to message: %+v",
				utils.DiameterAgent, err.Error(), m))
		diamErr(c, m, diam.UnableToComply, reqVars, da.cgrCfg, da.filterS)
		return
	}
	writeOnConn(c, a)
}

// V1DisconnectSession is part of the sessions.BiRPClient
func (da *DiameterAgent) V1DisconnectSession(ctx *context.Context, cgrEv utils.CGREvent, reply *string) (err error) {
	ssID, has := cgrEv.Event[utils.OriginID]
	if !has {
		utils.Logger.Info(
			fmt.Sprintf("<%s> cannot disconnect session, missing OriginID in event: %s",
				utils.DiameterAgent, utils.ToJSON(cgrEv.Event)))
		return utils.ErrMandatoryIeMissing
	}
	originID := ssID.(string)
	switch da.cgrCfg.DiameterAgentCfg().ForcedDisconnect {
	case utils.MetaNone:
		*reply = utils.OK
		return
	case utils.MetaASR:
		return da.sendASR(originID, reply)
	case utils.MetaRAR:
		return da.V1AlterSession(ctx, utils.CGREvent{Event: cgrEv.Event}, reply)
	default:
		return fmt.Errorf("Unsupported request type <%s>", da.cgrCfg.DiameterAgentCfg().ForcedDisconnect)
	}
}

func (da *DiameterAgent) sendASR(originID string, reply *string) (err error) {
	msg, has := engine.Cache.Get(utils.CacheDiameterMessages, originID)
	if !has {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot retrieve message from cache with OriginID: <%s>",
				utils.DiameterAgent, originID))
		return utils.ErrMandatoryIeMissing
	}
	dmd := msg.(*diamMsgData)
	aReq := NewAgentRequest(
		newDADataProvider(dmd.c, dmd.m),
		dmd.vars, nil, nil, nil, nil,
		da.cgrCfg.GeneralCfg().DefaultTenant,
		da.cgrCfg.GeneralCfg().DefaultTimezone, da.filterS, nil)
	if err = aReq.SetFields(da.cgrCfg.TemplatesCfg()[da.cgrCfg.DiameterAgentCfg().ASRTemplate]); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot disconnect session with OriginID: <%s>, err: %s",
				utils.DiameterAgent, originID, err.Error()))
		return utils.ErrServerError
	}
	m := diam.NewRequest(dmd.m.Header.CommandCode,
		dmd.m.Header.ApplicationID, dmd.m.Dictionary())
	if err = updateDiamMsgFromNavMap(m, aReq.diamreq,
		da.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot disconnect session with OriginID: <%s>, err: %s",
				utils.DiameterAgent, originID, err.Error()))
		return utils.ErrServerError
	}
	if err = writeOnConn(dmd.c, m); err != nil {
		return utils.ErrServerError
	}
	*reply = utils.OK
	return
}

// V1AlterSession  sends a rar message to diameter client
func (da *DiameterAgent) V1AlterSession(ctx *context.Context, cgrEv utils.CGREvent, reply *string) (err error) {
	originID, err := cgrEv.FieldAsString(utils.OriginID)
	if err != nil {
		return fmt.Errorf("could not retrieve OriginID: %w", err)
	}
	if originID == "" {
		utils.Logger.Info(
			fmt.Sprintf("<%s> cannot send RAR, missing session ID",
				utils.DiameterAgent))
		return utils.ErrMandatoryIeMissing
	}
	msg, has := engine.Cache.Get(utils.CacheDiameterMessages, originID)
	if !has {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot retrieve message from cache with OriginID: <%s>",
				utils.DiameterAgent, originID))
		return utils.ErrMandatoryIeMissing
	}
	dmd := msg.(*diamMsgData)
	aReq := NewAgentRequest(
		newDADataProvider(dmd.c, dmd.m),
		dmd.vars, nil, nil, nil, nil,
		da.cgrCfg.GeneralCfg().DefaultTenant,
		da.cgrCfg.GeneralCfg().DefaultTimezone, da.filterS, nil)
	if err = aReq.SetFields(da.cgrCfg.TemplatesCfg()[da.cgrCfg.DiameterAgentCfg().RARTemplate]); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot send RAR with OriginID: <%s>, err: %s",
				utils.DiameterAgent, originID, err.Error()))
		return utils.ErrServerError
	}
	m := diam.NewRequest(diam.ReAuth,
		dmd.m.Header.ApplicationID, dmd.m.Dictionary())
	if err = updateDiamMsgFromNavMap(m, aReq.diamreq,
		da.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> cannot send RAR with OriginID: <%s>, err: %s",
				utils.DiameterAgent, originID, err.Error()))
		return utils.ErrServerError
	}
	raaCh := make(chan *diam.Message, 1)
	da.raaLck.Lock()
	da.raa[originID] = raaCh
	da.raaLck.Unlock()
	defer func() {
		da.raaLck.Lock()
		delete(da.raa, originID)
		da.raaLck.Unlock()
	}()
	if err = writeOnConn(dmd.c, m); err != nil {
		return utils.ErrServerError
	}
	select {
	case raa := <-raaCh:
		var avps []*diam.AVP
		if avps, err = raa.FindAVPsWithPath([]any{avp.ResultCode}, dict.UndefinedVendorID); err != nil {
			return
		}
		if len(avps) == 0 {
			return fmt.Errorf("Missing AVP")
		}
		var data any
		if data, err = diamAVPAsIface(avps[0]); err != nil {
			return
		} else if data != uint32(diam.Success) {
			return fmt.Errorf("Wrong result code: <%v>", data)
		}
	case <-time.After(time.Second):
		return utils.ErrTimedOut
	}
	*reply = utils.OK
	return
}

// handleRAA is used to handle all Re-Authorize Answers that are received
func (da *DiameterAgent) handleRAA(c diam.Conn, m *diam.Message) {
	avp, err := m.FindAVP(avp.SessionID, dict.UndefinedVendorID)
	if err != nil {
		return
	}
	originID, err := diamAVPAsString(avp)
	if err != nil {
		return
	}
	da.raaLck.Lock()
	ch, has := da.raa[originID]
	da.raaLck.Unlock()
	if !has {
		return
	}
	ch <- m
}

// handleConns is used to handle all conns that are connected to the agent
// it register the connection so it can be used to send a DPR
func (da *DiameterAgent) handleConns(peers <-chan diam.Conn) {
	for c := range peers {
		meta, _ := smpeer.FromContext(c.Context())
		key := string(meta.OriginHost + utils.ConcatenatedKeySep + meta.OriginRealm)
		da.peersLck.Lock()
		da.peers[key] = c // store in peers table
		da.peersLck.Unlock()
		go func(c diam.Conn, key string) {
			// wait for disconnect notification
			<-c.(diam.CloseNotifier).CloseNotify()
			da.peersLck.Lock()
			delete(da.peers, key) // remove from peers table
			da.peersLck.Unlock()
		}(c, key)
	}
}

// handleDPA is used to handle all DisconnectPeer Answers that are received
func (da *DiameterAgent) handleDPA(c diam.Conn, m *diam.Message) {
	meta, _ := smpeer.FromContext(c.Context())
	key := string(meta.OriginHost + utils.ConcatenatedKeySep + meta.OriginRealm)

	da.dpaLck.Lock()
	ch, has := da.dpa[key]
	da.dpaLck.Unlock()
	if !has {
		return
	}
	ch <- m
	c.Close()
}

// V1DisconnectPeer  sends a DPR meseage to diameter client
func (da *DiameterAgent) V1DisconnectPeer(ctx *context.Context, args *utils.DPRArgs, reply *string) (err error) {
	if args == nil {
		utils.Logger.Info(
			fmt.Sprintf("<%s> cannot send DPR, missing arrguments",
				utils.DiameterAgent))
		return utils.ErrMandatoryIeMissing
	}

	if args.DisconnectCause < 0 || args.DisconnectCause > 2 {
		return errors.New("WRONG_DISCONNECT_CAUSE")
	}
	m := diam.NewRequest(diam.DisconnectPeer,
		diam.CHARGING_CONTROL_APP_ID, dict.Default)
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity(args.OriginHost))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity(args.OriginRealm))
	m.NewAVP(avp.DisconnectCause, avp.Mbit, 0, datatype.Enumerated(args.DisconnectCause))

	key := args.OriginHost + utils.ConcatenatedKeySep + args.OriginRealm

	dpaCh := make(chan *diam.Message, 1)
	da.dpaLck.Lock()
	da.dpa[key] = dpaCh
	da.dpaLck.Unlock()
	defer func() {
		da.dpaLck.Lock()
		delete(da.dpa, key)
		da.dpaLck.Unlock()
	}()
	da.peersLck.Lock()
	conn, has := da.peers[key]
	da.peersLck.Unlock()
	if !has {
		return utils.ErrNotFound
	}
	if err = writeOnConn(conn, m); err != nil {
		return utils.ErrServerError
	}
	select {
	case dpa := <-dpaCh:
		var avps []*diam.AVP
		if avps, err = dpa.FindAVPsWithPath([]any{avp.ResultCode}, dict.UndefinedVendorID); err != nil {
			return
		}
		if len(avps) == 0 {
			return fmt.Errorf("Missing AVP")
		}
		var data any
		if data, err = diamAVPAsIface(avps[0]); err != nil {
			return
		} else if data != uint32(diam.Success) {
			return fmt.Errorf("Wrong result code: <%v>", data)
		}
	case <-time.After(10 * time.Second):
		return utils.ErrTimedOut
	}
	*reply = utils.OK
	return
}

// V1GetActiveSessionIDs is part of the sessions.BiRPClient
func (da *DiameterAgent) V1GetActiveSessionIDs(*context.Context, string,
	*[]*sessions.SessionID) error {
	return utils.ErrNotImplemented
}

// V1WarnDisconnect is used to implement the sessions.BiRPClient interface
func (*DiameterAgent) V1WarnDisconnect(*context.Context, map[string]any, *string) error {
	return utils.ErrNotImplemented
}
