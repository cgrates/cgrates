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
	"strings"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
	"github.com/cgrates/rpcclient"
)

func NewFSsessions(fsAgentConfig *config.FsAgentCfg,
	timezone string, connMgr *engine.ConnManager) (fsa *FSsessions) {
	return &FSsessions{
		cfg:         fsAgentConfig,
		conns:       make([]*fsock.FSock, len(fsAgentConfig.EventSocketConns)),
		senderPools: make([]*fsock.FSockPool, len(fsAgentConfig.EventSocketConns)),
		timezone:    timezone,
		connMgr:     connMgr,
	}
}

// FSsessions is the freeswitch session manager
// it holds a buffer for the network connection
// and the active sessions
type FSsessions struct {
	cfg         *config.FsAgentCfg
	conns       []*fsock.FSock     // Keep the list here for connection management purposes
	senderPools []*fsock.FSockPool // Keep sender pools here
	timezone    string
	connMgr     *engine.ConnManager
}

func (fsa *FSsessions) createHandlers() map[string][]func(string, int) {
	ca := func(body string, connIdx int) {
		fsa.onChannelAnswer(
			NewFSEvent(body), connIdx)
	}
	ch := func(body string, connIdx int) {
		fsa.onChannelHangupComplete(
			NewFSEvent(body), connIdx)
	}
	handlers := map[string][]func(string, int){
		"CHANNEL_ANSWER":          {ca},
		"CHANNEL_HANGUP_COMPLETE": {ch},
	}
	if fsa.cfg.SubscribePark {
		cp := func(body string, connIdx int) {
			fsa.onChannelPark(
				NewFSEvent(body), connIdx)
		}
		handlers["CHANNEL_PARK"] = []func(string, int){cp}
	}
	return handlers
}

// Sets the call timeout valid of starting of the call
func (fsa *FSsessions) setMaxCallDuration(uuid string, connIdx int,
	maxDur time.Duration, destNr string) error {
	if len(fsa.cfg.EmptyBalanceContext) != 0 {
		_, err := fsa.conns[connIdx].SendApiCmd(
			fmt.Sprintf("uuid_setvar %s execute_on_answer sched_transfer +%d %s XML %s\n\n",
				uuid, int(maxDur.Seconds()), destNr, fsa.cfg.EmptyBalanceContext))
		if err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not transfer the call to empty balance context, error: <%s>, connIdx: %v",
					utils.FreeSWITCHAgent, err.Error(), connIdx))
			return err
		}
		return nil
	}
	if len(fsa.cfg.EmptyBalanceAnnFile) != 0 {
		if _, err := fsa.conns[connIdx].SendApiCmd(
			fmt.Sprintf("sched_broadcast +%d %s playback!manager_request::%s aleg\n\n",
				int(maxDur.Seconds()), uuid, fsa.cfg.EmptyBalanceAnnFile)); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not send uuid_broadcast to freeswitch, error: <%s>, connIdx: %v",
					utils.FreeSWITCHAgent, err.Error(), connIdx))
			return err
		}
		return nil
	}
	_, err := fsa.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s execute_on_answer sched_hangup +%d alloted_timeout\n\n",
			uuid, int(maxDur.Seconds())))
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send sched_hangup command to freeswitch, error: <%s>, connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
		return err
	}
	return nil
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

func (fsa *FSsessions) onChannelPark(fsev FSEvent, connIdx int) {
	if fsev.GetReqType(utils.MetaDefault) == utils.MetaNone { // Not for us
		return
	}
	if connIdx >= len(fsa.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(fsa.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.FreeSWITCHAgent, err.Error()))
		return
	}
	fsev[VarCGROriginHost] = utils.FirstNonEmpty(fsev[VarCGROriginHost], fsa.cfg.EventSocketConns[connIdx].Alias) // rewrite the OriginHost variable if it is empty
	authArgs := fsev.V1AuthorizeArgs()
	authArgs.CGREvent.Event[FsConnID] = connIdx // Attach the connection ID
	var authReply sessions.V1AuthorizeReply
	if err := fsa.connMgr.Call(fsa.cfg.SessionSConns, fsa, utils.SessionSv1AuthorizeEvent, authArgs, &authReply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not authorize event %s, error: %s",
				utils.FreeSWITCHAgent, fsev.GetUUID(), err.Error()))
		fsa.unparkCall(fsev.GetUUID(), connIdx,
			fsev.GetCallDestNr(utils.MetaDefault), err.Error())
		return
	}
	if authReply.Attributes != nil {
		for _, fldName := range authReply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if _, has := authReply.Attributes.CGREvent.Event[fldName]; !has {
				continue //maybe removed
			}
			if _, err := fsa.conns[connIdx].SendApiCmd(
				fmt.Sprintf("uuid_setvar %s %s %s\n\n", fsev.GetUUID(), fldName,
					authReply.Attributes.CGREvent.Event[fldName])); err != nil {
				utils.Logger.Info(
					fmt.Sprintf("<%s> error %s setting channel variabile: %s",
						utils.FreeSWITCHAgent, err.Error(), fldName))
				fsa.unparkCall(fsev.GetUUID(), connIdx,
					fsev.GetCallDestNr(utils.MetaDefault), err.Error())
				return
			}
		}
	}
	if authArgs.GetMaxUsage {
		if authReply.MaxUsage == nil || *authReply.MaxUsage == 0 {
			fsa.unparkCall(fsev.GetUUID(), connIdx,
				fsev.GetCallDestNr(utils.MetaDefault), utils.ErrInsufficientCredit.Error())
			return
		}
		fsa.setMaxCallDuration(fsev.GetUUID(), connIdx,
			*authReply.MaxUsage, fsev.GetCallDestNr(utils.MetaDefault))
	}
	if authReply.ResourceAllocation != nil {
		if _, err := fsa.conns[connIdx].SendApiCmd(fmt.Sprintf("uuid_setvar %s %s %s\n\n",
			fsev.GetUUID(), CGRResourceAllocation, *authReply.ResourceAllocation)); err != nil {
			utils.Logger.Info(
				fmt.Sprintf("<%s> error %s setting channel variabile: %s",
					utils.FreeSWITCHAgent, err.Error(), CGRResourceAllocation))
			fsa.unparkCall(fsev.GetUUID(), connIdx,
				fsev.GetCallDestNr(utils.MetaDefault), err.Error())
			return
		}
	}
	if authReply.RouteProfiles != nil {
		fsArray := SliceAsFsArray(authReply.RouteProfiles.RoutesWithParams())
		if _, err := fsa.conns[connIdx].SendApiCmd(fmt.Sprintf("uuid_setvar %s %s %s\n\n",
			fsev.GetUUID(), utils.CGRRoutes, fsArray)); err != nil {
			utils.Logger.Info(fmt.Sprintf("<%s> error setting routes: %s",
				utils.FreeSWITCHAgent, err.Error()))
			fsa.unparkCall(fsev.GetUUID(), connIdx, fsev.GetCallDestNr(utils.MetaDefault), err.Error())
			return
		}
	}

	fsa.unparkCall(fsev.GetUUID(), connIdx,
		fsev.GetCallDestNr(utils.MetaDefault), AUTH_OK)
}

func (fsa *FSsessions) onChannelAnswer(fsev FSEvent, connIdx int) {
	if fsev.GetReqType(utils.MetaDefault) == utils.MetaNone { // Do not process this request
		return
	}
	if connIdx >= len(fsa.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(fsa.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.FreeSWITCHAgent, err.Error()))
		return
	}
	_, err := fsa.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s %s %s\n\n", fsev.GetUUID(),
			utils.CGROriginHost, utils.FirstNonEmpty(fsa.cfg.EventSocketConns[connIdx].Alias,
				fsa.cfg.EventSocketConns[connIdx].Address)))
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error %s setting channel variabile: %s",
				utils.FreeSWITCHAgent, err.Error(), VarCGROriginHost))
		return
	}
	fsev[VarCGROriginHost] = utils.FirstNonEmpty(fsev[VarCGROriginHost], fsa.cfg.EventSocketConns[connIdx].Alias) // rewrite the OriginHost variable if it is empty
	chanUUID := fsev.GetUUID()
	if missing := fsev.MissingParameter(fsa.timezone); missing != "" {
		fsa.disconnectSession(connIdx, chanUUID, "",
			utils.NewErrMandatoryIeMissing(missing).Error())
		return
	}
	initSessionArgs := fsev.V1InitSessionArgs()
	initSessionArgs.CGREvent.Event[FsConnID] = connIdx // Attach the connection ID so we can properly disconnect later
	var initReply sessions.V1InitSessionReply
	if err := fsa.connMgr.Call(fsa.cfg.SessionSConns, fsa, utils.SessionSv1InitiateSession,
		initSessionArgs, &initReply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> could not process answer for event %s, error: %s",
				utils.FreeSWITCHAgent, chanUUID, err.Error()))
		fsa.disconnectSession(connIdx, chanUUID, "", err.Error())
		return
	}
}

func (fsa *FSsessions) onChannelHangupComplete(fsev FSEvent, connIdx int) {
	if fsev.GetReqType(utils.MetaDefault) == utils.MetaNone { // Do not process this request
		return
	}
	if connIdx >= len(fsa.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(fsa.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.FreeSWITCHAgent, err.Error()))
		return
	}
	var reply string
	fsev[VarCGROriginHost] = utils.FirstNonEmpty(fsev[VarCGROriginHost], fsa.cfg.EventSocketConns[connIdx].Alias) // rewrite the OriginHost variable if it is empty
	if fsev[VarAnswerEpoch] != "0" {                                                                              // call was answered
		terminateSessionArgs := fsev.V1TerminateSessionArgs()
		terminateSessionArgs.CGREvent.Event[FsConnID] = connIdx // Attach the connection ID in case we need to create a session and disconnect it
		if err := fsa.connMgr.Call(fsa.cfg.SessionSConns, fsa, utils.SessionSv1TerminateSession,
			terminateSessionArgs, &reply); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not terminate session with event %s, error: %s",
					utils.FreeSWITCHAgent, fsev.GetUUID(), err.Error()))
		}
	}
	if fsa.cfg.CreateCdr {
		cgrEv, err := fsev.AsCGREvent(fsa.timezone)
		if err != nil {
			return
		}
		if err := fsa.connMgr.Call(fsa.cfg.SessionSConns, fsa, utils.SessionSv1ProcessCDR,
			cgrEv, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed processing CGREvent: %s,  error: <%s>",
				utils.FreeSWITCHAgent, utils.ToJSON(cgrEv), err.Error()))
		}
	}
}

// Connect connects to the freeswitch mod_event_socket server and starts
// listening for events.
func (fsa *FSsessions) Connect() error {
	eventFilters := map[string][]string{"Call-Direction": {"inbound"}}
	errChan := make(chan error)
	for connIdx, connCfg := range fsa.cfg.EventSocketConns {
		fSock, err := fsock.NewFSock(connCfg.Address, connCfg.Password, connCfg.Reconnects, connCfg.MaxReconnectInterval, utils.FibDuration,
			fsa.createHandlers(), eventFilters, utils.Logger, connIdx, true)
		if err != nil {
			return err
		}
		if !fSock.Connected() {
			return errors.New("Could not connect to FreeSWITCH")
		}
		fsa.conns[connIdx] = fSock
		utils.Logger.Info(fmt.Sprintf("<%s> successfully connected to FreeSWITCH at: <%s>", utils.FreeSWITCHAgent, connCfg.Address))
		go func(fsock *fsock.FSock) { // Start reading in own goroutine, return on error
			if err := fsock.ReadEvents(); err != nil {
				errChan <- err
			}
		}(fSock)
		fsSenderPool := fsock.NewFSockPool(5, connCfg.Address, connCfg.Password, 1, fsa.cfg.MaxWaitConnection,
			0, utils.FibDuration, make(map[string][]func(string, int)), make(map[string][]string), utils.Logger, connIdx, true)
		if fsSenderPool == nil {
			return errors.New("Cannot connect FreeSWITCH senders pool")
		}
		fsa.senderPools[connIdx] = fsSenderPool
	}
	err := <-errChan // Will keep the Connect locked until the first error in one of the connections
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

// Call implements rpcclient.ClientConnector interface
func (fsa *FSsessions) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(fsa, serviceMethod, args, reply)
}

// V1DisconnectSession internal method to disconnect session in FreeSWITCH
func (fsa *FSsessions) V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) (err error) {
	ev := engine.NewMapEvent(args.EventStart)
	channelID := ev.GetStringIgnoreErrors(utils.OriginID)
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
		args.Reason); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// V1GetActiveSessionIDs used to return all active sessions
func (fsa *FSsessions) V1GetActiveSessionIDs(_ string,
	sessionIDs *[]*sessions.SessionID) (err error) {
	var sIDs []*sessions.SessionID
	for connIdx, senderPool := range fsa.senderPools {
		fsConn, err := senderPool.PopFSock()
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error on pop FSock: %s, connection index: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
			continue
		}
		activeChanStr, err := fsConn.SendApiCmd("show channels")
		senderPool.PushFSock(fsConn)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error on push FSock: %s, connection index: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
			continue
		}
		for _, fsAChan := range fsock.MapChanData(activeChanStr) {
			sIDs = append(sIDs, &sessions.SessionID{
				OriginHost: fsa.cfg.EventSocketConns[connIdx].Alias,
				OriginID:   fsAChan["uuid"],
			})
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

// V1ReAuthorize is used to implement the sessions.BiRPClient interface
func (*FSsessions) V1ReAuthorize(originID string, reply *string) (err error) {
	return utils.ErrNotImplemented
}

// V1DisconnectPeer is used to implement the sessions.BiRPClient interface
func (*FSsessions) V1DisconnectPeer(args *utils.DPRArgs, reply *string) (err error) {
	return utils.ErrNotImplemented
}

// V1WarnDisconnect is called when call goes under the minimum duration threshold, so FreeSWITCH can play an announcement message
func (fsa *FSsessions) V1WarnDisconnect(args map[string]interface{}, reply *string) (err error) {
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

// CallBiRPC is part of utils.BiRPCServer interface to help internal connections do calls over rpcclient.ClientConnector interface
func (fsa *FSsessions) CallBiRPC(clnt rpcclient.ClientConnector, serviceMethod string, args interface{}, reply interface{}) error {
	return utils.BiRPCCall(fsa, clnt, serviceMethod, args, reply)
}

// BiRPCv1DisconnectSession is internal method to disconnect session in asterisk
func (fsa *FSsessions) BiRPCv1DisconnectSession(clnt rpcclient.ClientConnector, args utils.AttrDisconnectSession, reply *string) error {
	return fsa.V1DisconnectSession(args, reply)
}

// BiRPCv1GetActiveSessionIDs is internal method to  get all active sessions in asterisk
func (fsa *FSsessions) BiRPCv1GetActiveSessionIDs(clnt rpcclient.ClientConnector, ignParam string,
	sessionIDs *[]*sessions.SessionID) error {
	return fsa.V1GetActiveSessionIDs(ignParam, sessionIDs)

}

// BiRPCv1ReAuthorize is used to implement the sessions.BiRPClient interface
func (fsa *FSsessions) BiRPCv1ReAuthorize(clnt rpcclient.ClientConnector, originID string, reply *string) (err error) {
	return fsa.V1ReAuthorize(originID, reply)
}

// BiRPCv1DisconnectPeer is used to implement the sessions.BiRPClient interface
func (fsa *FSsessions) BiRPCv1DisconnectPeer(clnt rpcclient.ClientConnector, args *utils.DPRArgs, reply *string) (err error) {
	return fsa.V1DisconnectPeer(args, reply)
}

// BiRPCv1WarnDisconnect is used to implement the sessions.BiRPClient interface
func (fsa *FSsessions) BiRPCv1WarnDisconnect(clnt rpcclient.ClientConnector, args map[string]interface{}, reply *string) (err error) {
	return fsa.V1WarnDisconnect(args, reply)
}

// Handlers is used to implement the rpcclient.BiRPCConector interface
func (fsa *FSsessions) Handlers() map[string]interface{} {
	return map[string]interface{}{
		utils.SessionSv1DisconnectSession: func(clnt *rpc2.Client, args utils.AttrDisconnectSession, rply *string) error {
			return fsa.BiRPCv1DisconnectSession(clnt, args, rply)
		},
		utils.SessionSv1GetActiveSessionIDs: func(clnt *rpc2.Client, args string, rply *[]*sessions.SessionID) error {
			return fsa.BiRPCv1GetActiveSessionIDs(clnt, args, rply)
		},
		utils.SessionSv1ReAuthorize: func(clnt *rpc2.Client, args string, rply *string) (err error) {
			return fsa.BiRPCv1ReAuthorize(clnt, args, rply)
		},
		utils.SessionSv1DisconnectPeer: func(clnt *rpc2.Client, args *utils.DPRArgs, rply *string) (err error) {
			return fsa.BiRPCv1DisconnectPeer(clnt, args, rply)
		},
		utils.SessionSv1WarnDisconnect: func(clnt *rpc2.Client, args map[string]interface{}, rply *string) (err error) {
			return fsa.BiRPCv1WarnDisconnect(clnt, args, rply)
		},
	}
}
