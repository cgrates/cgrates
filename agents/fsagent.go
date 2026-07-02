/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package agents

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
)

const (
	FsParkEvent   = "CHANNEL_PARK"
	FsAnswerEvent = "CHANNEL_ANSWER"
	FsHangupEvent = "CHANNEL_HANGUP_COMPLETE"
	EventName     = "Event-Name"
	UUID          = "Unique-ID" // unique ID for this call leg
	CALL_DEST_NR  = "Caller-Destination-Number"
	SIP_REQ_USER  = "variable_sip_req_user"
	AUTH_OK       = "AUTH_OK"
	FsConnID      = "FsConnID" // used to share connID info in event for remote disconnects
)

func NewFSsessions(cgrcfg *config.CGRConfig, cache *engine.CacheS, filterS *engine.FilterS,
	timezone string, connMgr *engine.ConnManager, caps *engine.Caps) (*FSsessions, error) {
	fsAgentConfig := cgrcfg.FsAgentCfg()
	fsa := &FSsessions{
		cgrcfg:      cgrcfg,
		cache:       cache,
		cfg:         fsAgentConfig,
		conns:       make([]*fsock.FSock, len(fsAgentConfig.EventSocketConns)),
		senderPools: make([]*fsock.FSockPool, len(fsAgentConfig.EventSocketConns)),
		timezone:    timezone,
		connMgr:     connMgr,
		caps:        caps,
		fltrS:       filterS,
	}
	srv, _ := birpc.NewService(fsa, "", false)
	fsa.ctx = context.WithClient(context.TODO(), srv)
	msgTemplates := fsa.cgrcfg.TemplatesCfg()
	// Inflate *template field types
	for _, procsr := range fsa.cfg.RequestProcessors {
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
	return fsa, nil
}

// FSsessions is the freeswitch session manager
// it holds a buffer for the network connection
// and the active sessions
type FSsessions struct {
	cgrcfg      *config.CGRConfig
	cache       *engine.CacheS
	cfg         *config.FsAgentCfg
	conns       []*fsock.FSock     // Keep the list here for connection management purposes
	senderPools []*fsock.FSockPool // Keep sender pools here
	timezone    string
	connMgr     *engine.ConnManager
	caps        *engine.Caps
	ctx         *context.Context
	fltrS       *engine.FilterS
}

func (fsa *FSsessions) createHandlers() map[string][]func(string, int) {
	hdlr := func(body string, connIdx int) {
		fsa.handleFSEvent(fsock.EventToMap(body), connIdx)
	}
	handlers := map[string][]func(string, int){
		FsAnswerEvent: {hdlr},
		FsHangupEvent: {hdlr},
	}
	if fsa.cfg.SubscribePark {
		handlers[FsParkEvent] = []func(string, int){hdlr}
	}
	return handlers
}

func (fsa *FSsessions) handleFSEvent(fsev map[string]string, connIdx int) {
	eventName := fsev[EventName]
	uuid := fsev[UUID]
	if fsa.caps.IsLimited() {
		if err := fsa.caps.Allocate(); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> caps limit reached, rejecting %s for channel %s: %v",
					utils.FreeSWITCHAgent, fsev[EventName], fsev[UUID], err))
			if eventName == FsParkEvent {
				fsa.unparkCall(uuid, connIdx, utils.FirstNonEmpty(fsev[CALL_DEST_NR], fsev[SIP_REQ_USER]), err.Error())
			}
			return
		}
		defer fsa.caps.Deallocate()
	}
	if connIdx >= len(fsa.conns) {
		utils.Logger.Err(fmt.Sprintf("<%s> Index %v out of range",
			utils.FreeSWITCHAgent, connIdx))
		return
	}

	reqVars := &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			FsConnID:  utils.NewLeafNode(connIdx),
			EventName: utils.NewLeafNode(eventName),
			utils.OriginHost: utils.NewLeafNode(utils.FirstNonEmpty(
				fsa.cfg.EventSocketConns[connIdx].Alias,
				fsa.cfg.EventSocketConns[connIdx].Address)),
		},
	}
	dP := utils.MapStringDP(fsev)
	cgrRplyNM := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	opts := utils.MapStorage{}
	rply := utils.NewOrderedNavigableMap()
	sessConns, _ := engine.GetConnIDs(fsa.ctx, fsa.cfg.Conns, utils.MetaSessionS, utils.MetaAny, dP, nil, fsa.fltrS)

	var processed bool
	var err error
	for _, reqProcessor := range fsa.cfg.RequestProcessors {
		var lclProcessed bool
		lclProcessed, err = processRequest(
			fsa.ctx, reqProcessor,
			NewAgentRequest(
				dP, reqVars, cgrRplyNM, rply, opts,
				reqProcessor.Tenant, fsa.cgrcfg.GeneralCfg().DefaultTenant,
				utils.FirstNonEmpty(reqProcessor.Timezone,
					fsa.cgrcfg.GeneralCfg().DefaultTimezone),
				fsa.cgrcfg, fsa.cache, fsa.fltrS, nil),
			utils.FreeSWITCHAgent, fsa.connMgr,
			sessConns, nil, nil, fsa.fltrS)
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
			fmt.Sprintf("<%s> error: %v processing event %s",
				utils.FreeSWITCHAgent, err, eventName))
		fsa.dispatchErr(uuid, connIdx, eventName, fsev, err)
		return
	}
	if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring event %s",
				utils.FreeSWITCHAgent, eventName))
		fsa.dispatchErr(uuid, connIdx, eventName, fsev, utils.ErrNotFound)
		return
	}
	if err = fsa.setReplyVars(uuid, connIdx, rply); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %v applying reply variables for event %s",
				utils.FreeSWITCHAgent, err, eventName))
		fsa.dispatchErr(uuid, connIdx, eventName, fsev, err)
		return
	}
	if eventName == FsParkEvent {
		fsa.authorizePark(uuid, connIdx,
			utils.FirstNonEmpty(fsev[CALL_DEST_NR], fsev[SIP_REQ_USER]), cgrRplyNM)
	}
}

func (fsa *FSsessions) dispatchErr(uuid string, connIdx int, eventName string, fsev map[string]string, err error) {
	switch eventName {
	case FsParkEvent:
		fsa.unparkCall(uuid, connIdx,
			utils.FirstNonEmpty(fsev[CALL_DEST_NR], fsev[SIP_REQ_USER]), err.Error())
	case FsAnswerEvent:
		fsa.disconnectSession(connIdx, uuid, utils.EmptyString, err.Error())
	}
}

func (fsa *FSsessions) setReplyVars(uuid string, connIdx int, rply *utils.OrderedNavigableMap) error {
	for el := rply.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		itm, _ := rply.Field(path)
		if itm == nil {
			continue
		}
		val := utils.IfaceAsString(itm.Data)
		vrbl := strings.Join(utils.StripTrailingIndex(path), utils.NestingSep)
		if _, err := fsa.conns[connIdx].SendApiCmd(
			fmt.Sprintf("uuid_setvar %s %s %s\n\n", uuid, vrbl, val)); err != nil {
			return fmt.Errorf("uuid_setvar failed for %s: %w", vrbl, err)
		}
	}
	return nil
}

func (fsa *FSsessions) authorizePark(uuid string, connIdx int, destNr string, cgrRply *utils.DataNode) {
	maxDur, hasUsage := minAccountUsage(cgrRply)
	if hasUsage && maxDur == 0 {
		fsa.unparkCall(uuid, connIdx, destNr, utils.ErrInsufficientCredit.Error())
		return
	}
	if hasUsage {
		fsa.setMaxCallDuration(uuid, connIdx, maxDur, destNr)
	}
	fsa.unparkCall(uuid, connIdx, destNr, AUTH_OK)
}

func minAccountUsage(cgrRply *utils.DataNode) (min time.Duration, found bool) {
	node, has := cgrRply.Map[utils.CapMaxUsage]
	if !has || node == nil {
		return
	}
	for _, dataNode := range node.Map {
		if dataNode == nil || dataNode.Value == nil {
			continue
		}
		d, ok := dataNode.Value.Data.(time.Duration)
		if !ok {
			continue
		}
		if !found || d < min {
			min, found = d, true
		}
	}
	return
}

// Sets the call timeout valid of starting of the call
func (fsa *FSsessions) setMaxCallDuration(uuid string, connIdx int,
	maxDur time.Duration, destNr string) (err error) {
	if len(fsa.cfg.EmptyBalanceContext) != 0 {
		if _, err = fsa.conns[connIdx].SendApiCmd(
			fmt.Sprintf("uuid_setvar %s execute_on_answer sched_transfer +%d %s XML %s\n\n",
				uuid, int(maxDur.Seconds()), destNr, fsa.cfg.EmptyBalanceContext)); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not transfer the call to empty balance context, error: <%s>, connIdx: %v",
					utils.FreeSWITCHAgent, err.Error(), connIdx))
		}
		return
	}
	if len(fsa.cfg.EmptyBalanceAnnFile) != 0 {
		if _, err = fsa.conns[connIdx].SendApiCmd(
			fmt.Sprintf("sched_broadcast +%d %s playback!manager_request::%s aleg\n\n",
				int(maxDur.Seconds()), uuid, fsa.cfg.EmptyBalanceAnnFile)); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not send uuid_broadcast to freeswitch, error: <%s>, connIdx: %v",
					utils.FreeSWITCHAgent, err.Error(), connIdx))
		}
		return
	}
	if _, err = fsa.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s execute_on_answer sched_hangup +%d alloted_timeout\n\n",
			uuid, int(maxDur.Seconds()))); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send sched_hangup command to freeswitch, error: <%s>, connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
	}
	return
}

// Sends the transfer command to unpark the call to freeswitch
func (fsa *FSsessions) unparkCall(uuid string, connIdx int, callDestNb, notify string) (err error) {
	_, err = fsa.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", uuid, notify))
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send unpark api notification to freeswitch, error: <%s>, connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
		return
	}
	if _, err = fsa.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_transfer %s %s\n\n", uuid, callDestNb)); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send unpark api call to freeswitch, error: <%s>, connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
	}
	return
}

// Connect connects to the freeswitch mod_event_socket server and starts
// listening for events.
func (fsa *FSsessions) Connect() error {
	eventFilters := map[string][]string{"Call-Direction": {"inbound"}}
	connErr := make(chan error)
	for connIdx, connCfg := range fsa.cfg.EventSocketConns {
		fSock, err := fsock.NewFSock(
			connCfg.Address, connCfg.Password,
			connCfg.Reconnects, connCfg.MaxReconnectInterval,
			connCfg.ReplyTimeout, utils.FibDuration,
			fsa.createHandlers(), eventFilters,
			utils.Logger, connIdx, true, connErr)

		if !fSock.Connected() {
			return errors.New("Could not connect to FreeSWITCH")
		}
		fsa.conns[connIdx] = fSock
		utils.Logger.Info(fmt.Sprintf("<%s> successfully connected to FreeSWITCH at: <%s>", utils.FreeSWITCHAgent, connCfg.Address))
		fsSenderPool := fsock.NewFSockPool(5, connCfg.Address, connCfg.Password, 1, fsa.cfg.MaxWaitConnection,
			0, connCfg.ReplyTimeout, utils.FibDuration,
			make(map[string][]func(string, int)), make(map[string][]string),
			utils.Logger, connIdx, true, nil)
		if err != nil {
			return fmt.Errorf("Cannot connect FreeSWITCH senders pool, error: %s", err.Error())
		}
		if fsSenderPool == nil {
			return errors.New("Cannot connect FreeSWITCH senders pool")
		}
		fsa.senderPools[connIdx] = fsSenderPool
	}
	err := <-connErr // Will keep the Connect locked until the first error in one of the connections
	return err
}

// fsev.GetCallDestNr(utils.MetaDefault)
// Disconnects a session by sending hangup command to freeswitch
func (fsa *FSsessions) disconnectSession(connIdx int, uuid, redirectNr, notify string) error {
	if _, err := fsa.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", uuid, notify)); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error: %s when attempting to disconnect channelID: %s over connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), uuid, connIdx))
		return err
	}
	if notify == utils.ErrInsufficientCredit.Error() {
		if len(fsa.cfg.EmptyBalanceContext) != 0 {
			if _, err := fsa.conns[connIdx].SendApiCmd(fmt.Sprintf("uuid_transfer %s %s XML %s\n\n",
				uuid, redirectNr, fsa.cfg.EmptyBalanceContext)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Could not transfer the call to empty balance context, error: <%s>, connIdx: %v",
					utils.FreeSWITCHAgent, err.Error(), connIdx))
				return err
			}
			return nil
		}
		if len(fsa.cfg.EmptyBalanceAnnFile) != 0 {
			if _, err := fsa.conns[connIdx].SendApiCmd(fmt.Sprintf("uuid_broadcast %s playback!manager_request::%s aleg\n\n",
				uuid, fsa.cfg.EmptyBalanceAnnFile)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Could not send uuid_broadcast to freeswitch, error: <%s>, connIdx: %v",
					utils.FreeSWITCHAgent, err.Error(), connIdx))
				return err
			}
			return nil
		}
	}
	if err := fsa.conns[connIdx].SendMsgCmd(uuid,
		map[string]string{"call-command": "hangup", "hangup-cause": "MANAGER_REQUEST"}); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send disconect msg to freeswitch, error: <%s>, connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
		return err
	}
	return nil
}

// Shutdown stops all connected fsock connections
func (fsa *FSsessions) Shutdown() (err error) {
	for connIdx, fSock := range fsa.conns {
		if fSock == nil ||
			!fSock.Connected() {
			utils.Logger.Err(fmt.Sprintf("<%s> Cannot shutdown sessions, fsock not connected for connection index: %v", utils.FreeSWITCHAgent, connIdx))
			continue
		}
		utils.Logger.Info(fmt.Sprintf("<%s> Shutting down all sessions on connection index: %v", utils.FreeSWITCHAgent, connIdx))
		if _, err = fSock.SendApiCmd("hupall MANAGER_REQUEST cgr_reqtype *prepaid"); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error on calls shutdown: %s, connection index: %v", utils.FreeSWITCHAgent, err.Error(), connIdx))
		}
		if err = fSock.Disconnect(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error on disconnect: %s, connection index: %v", utils.FreeSWITCHAgent, err.Error(), connIdx))
		}

	}
	return
}

// V1DisconnectSession internal method to disconnect session in FreeSWITCH
func (fsa *FSsessions) V1DisconnectSession(ctx *context.Context, cgrEv utils.CGREvent, reply *string) (err error) {
	ev := engine.NewMapEvent(cgrEv.Event)
	channelID := ev.GetStringIgnoreErrors(utils.OriginID)
	disconnectCause := ev.GetStringIgnoreErrors(utils.DisconnectCause)
	connIdx, err := ev.GetTInt64(FsConnID)
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error: <%s:%s> when attempting to disconnect channelID: <%s>",
				utils.FreeSWITCHAgent, err.Error(), FsConnID, channelID))
		return
	}
	if int(connIdx) >= len(fsa.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(fsa.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.FreeSWITCHAgent, err.Error()))
		return err
	}
	if err = fsa.disconnectSession(int(connIdx), channelID,
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(CALL_DEST_NR), ev.GetStringIgnoreErrors(SIP_REQ_USER)),
		disconnectCause); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// V1GetActiveSessionIDs used to return all active sessions
func (fsa *FSsessions) V1GetActiveSessionIDs(ctx *context.Context, _ string,
	sessionIDs *[]*sessions.SessionID) (err error) {
	var sIDs []*sessions.SessionID
	for connIdx, senderPool := range fsa.senderPools {
		fsConn, err := senderPool.PopFSock()
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error on pop FSock: %s, connection index: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
			continue
		}
		apiCmd := "show channels"
		if fsa.cfg.ActiveSessionDelimiter != "," { // ',' delimiter is used by default
			apiCmd += " as delim " + fsa.cfg.ActiveSessionDelimiter
		}
		activeChanStr, err := fsConn.SendApiCmd(apiCmd)
		senderPool.PushFSock(fsConn)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf(
				"<%s> Error on push FSock: %s, connection index: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
			continue
		}
		for _, fsAChan := range fsock.MapChanData(activeChanStr, fsa.cfg.ActiveSessionDelimiter) {
			sIDs = append(sIDs, &sessions.SessionID{
				OriginHost: fsa.cfg.EventSocketConns[connIdx].Alias,
				OriginID:   fsAChan["uuid"]},
			)
		}
	}
	if len(sIDs) == 0 {
		return utils.ErrNoActiveSession
	}
	*sessionIDs = sIDs
	return
}

// Reload recreates the connection buffers
// only used on reload
func (fsa *FSsessions) Reload() {
	fsa.conns = make([]*fsock.FSock, len(fsa.cfg.EventSocketConns))
	fsa.senderPools = make([]*fsock.FSockPool, len(fsa.cfg.EventSocketConns))
}

// V1AlterSession is used to implement the sessions.BiRPClient interface
func (*FSsessions) V1AlterSession(*context.Context, utils.CGREvent, *string) error {
	return utils.ErrNotImplemented
}

// V1DisconnectPeer is used to implement the sessions.BiRPClient interface
func (*FSsessions) V1DisconnectPeer(*context.Context, *utils.DPRArgs, *string) error {
	return utils.ErrNotImplemented
}

// V1WarnDisconnect is called when call goes under the minimum duration threshold, so FreeSWITCH can play an announcement message
func (fsa *FSsessions) V1WarnDisconnect(ctx *context.Context, args map[string]any, reply *string) (err error) {
	if fsa.cfg.LowBalanceAnnFile == utils.EmptyString {
		*reply = utils.OK
		return
	}
	ev := engine.NewMapEvent(args)
	channelID := ev.GetStringIgnoreErrors(utils.OriginID)
	var connIdx int64
	if connIdx, err = ev.GetTInt64(FsConnID); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error: <%s:%s> when attempting to disconnect channelID: <%s>",
				utils.FreeSWITCHAgent, err.Error(), FsConnID, channelID))
		return
	}
	if int(connIdx) >= len(fsa.conns) { // protection against index out of range panic
		err = fmt.Errorf("Index out of range[0,%v): %v ", len(fsa.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.FreeSWITCHAgent, err.Error()))
		return
	}
	if _, err = fsa.conns[connIdx].SendApiCmd(fmt.Sprintf("uuid_broadcast %s  playback!manager_request::%s aleg\n\n", channelID, fsa.cfg.LowBalanceAnnFile)); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> Could not send uuid_broadcast to freeswitch, error: %s, connection id: %v",
			utils.FreeSWITCHAgent, err.Error(), connIdx))
		return
	}
	*reply = utils.OK
	return
}
