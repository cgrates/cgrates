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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/fsock"
)

func NewFSsessions(fsAgentConfig *config.FsAgentConfig,
	smg *utils.BiRPCInternalClient, timezone string) (fsa *FSsessions) {
	fsa = &FSsessions{
		cfg:         fsAgentConfig,
		conns:       make(map[string]*fsock.FSock),
		senderPools: make(map[string]*fsock.FSockPool),
		smg:         smg,
		timezone:    timezone,
	}
	fsa.smg.SetClientConn(fsa) // pass the connection to FsA back into smg so we can receive the disconnects
	return
}

// The freeswitch session manager type holding a buffer for the network connection
// and the active sessions
type FSsessions struct {
	cfg         *config.FsAgentConfig
	conns       map[string]*fsock.FSock     // Keep the list here for connection management purposes
	senderPools map[string]*fsock.FSockPool // Keep sender pools here
	smg         *utils.BiRPCInternalClient
	timezone    string
}

func (sm *FSsessions) createHandlers() map[string][]func(string, string) {
	ca := func(body, connId string) {
		sm.onChannelAnswer(
			NewFSEvent(body), connId)
	}
	ch := func(body, connId string) {
		sm.onChannelHangupComplete(
			NewFSEvent(body), connId)
	}
	handlers := map[string][]func(string, string){
		"CHANNEL_ANSWER":          []func(string, string){ca},
		"CHANNEL_HANGUP_COMPLETE": []func(string, string){ch},
	}
	if sm.cfg.SubscribePark {
		cp := func(body, connId string) {
			sm.onChannelPark(
				NewFSEvent(body), connId)
		}
		handlers["CHANNEL_PARK"] = []func(string, string){cp}
	}
	return handlers
}

// Sets the call timeout valid of starting of the call
func (sm *FSsessions) setMaxCallDuration(uuid, connId string,
	maxDur time.Duration, destNr string) error {
	if len(sm.cfg.EmptyBalanceContext) != 0 {
		_, err := sm.conns[connId].SendApiCmd(
			fmt.Sprintf("uuid_setvar %s execute_on_answer sched_transfer +%d %s XML %s\n\n",
				uuid, int(maxDur.Seconds()), destNr, sm.cfg.EmptyBalanceContext))
		if err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not transfer the call to empty balance context, error: <%s>, connId: %s",
					utils.FreeSWITCHAgent, err.Error(), connId))
			return err
		}
		return nil
	} else if len(sm.cfg.EmptyBalanceAnnFile) != 0 {
		if _, err := sm.conns[connId].SendApiCmd(
			fmt.Sprintf("sched_broadcast +%d %s playback!manager_request::%s aleg\n\n",
				int(maxDur.Seconds()), uuid, sm.cfg.EmptyBalanceAnnFile)); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not send uuid_broadcast to freeswitch, error: <%s>, connId: %s",
					utils.FreeSWITCHAgent, err.Error(), connId))
			return err
		}
		return nil
	} else {
		_, err := sm.conns[connId].SendApiCmd(
			fmt.Sprintf("uuid_setvar %s execute_on_answer sched_hangup +%d alloted_timeout\n\n",
				uuid, int(maxDur.Seconds())))
		if err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not send sched_hangup command to freeswitch, error: <%s>, connId: %s",
					utils.FreeSWITCHAgent, err.Error(), connId))
			return err
		}
		return nil
	}
	return nil
}

// Sends the transfer command to unpark the call to freeswitch
func (sm *FSsessions) unparkCall(uuid, connId, call_dest_nb, notify string) (err error) {
	_, err = sm.conns[connId].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", uuid, notify))
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send unpark api notification to freeswitch, error: <%s>, connId: %s",
				utils.FreeSWITCHAgent, err.Error(), connId))
		return
	}
	if _, err = sm.conns[connId].SendApiCmd(
		fmt.Sprintf("uuid_transfer %s %s\n\n", uuid, call_dest_nb)); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send unpark api call to freeswitch, error: <%s>, connId: %s",
				utils.FreeSWITCHAgent, err.Error(), connId))
	}
	return
}

func (sm *FSsessions) onChannelPark(fsev FSEvent, connId string) {
	if fsev.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Not for us
		return
	}
	authArgs := fsev.V1AuthorizeArgs()
	var authReply sessions.V1AuthorizeReply
	if err := sm.smg.Call(utils.SessionSv1AuthorizeEvent, authArgs, &authReply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not authorize event %s, error: %s",
				utils.FreeSWITCHAgent, fsev.GetUUID(), err.Error()))
		sm.unparkCall(fsev.GetUUID(), connId,
			fsev.GetCallDestNr(utils.META_DEFAULT), utils.ErrServerError.Error())
		return
	}
	if authArgs.GetMaxUsage {
		if *authReply.MaxUsage != -1 { // For calls different than unlimited, set limits
			if *authReply.MaxUsage == 0 {
				sm.unparkCall(fsev.GetUUID(), connId,
					fsev.GetCallDestNr(utils.META_DEFAULT), utils.ErrInsufficientCredit.Error())
				return
			}
			sm.setMaxCallDuration(fsev.GetUUID(), connId,
				*authReply.MaxUsage, fsev.GetCallDestNr(utils.META_DEFAULT))
		}
	}
	if authArgs.AuthorizeResources {
		if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_setvar %s %s %s\n\n",
			fsev.GetUUID(), CGRResourceAllocation, *authReply.ResourceAllocation)); err != nil {
			utils.Logger.Info(
				fmt.Sprintf("<%s> error %s setting channel variabile: %s",
					utils.FreeSWITCHAgent, err.Error(), CGRResourceAllocation))
			sm.unparkCall(fsev.GetUUID(), connId,
				fsev.GetCallDestNr(utils.META_DEFAULT), utils.ErrServerError.Error())
			return
		}
	}
	if authArgs.GetSuppliers {
		fsArray := SliceAsFsArray(authReply.Suppliers.SuppliersWithParams())
		if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_setvar %s %s %s\n\n",
			fsev.GetUUID(), utils.CGR_SUPPLIERS, fsArray)); err != nil {
			utils.Logger.Info(fmt.Sprintf("<%s> error setting suppliers: %s", utils.FreeSWITCHAgent, err.Error()))
			sm.unparkCall(fsev.GetUUID(), connId, fsev.GetCallDestNr(utils.META_DEFAULT), utils.ErrServerError.Error())
			return
		}
	}
	if authArgs.GetAttributes {
		if authReply.Attributes != nil {
			for _, fldName := range authReply.Attributes.AlteredFields {
				if _, has := authReply.Attributes.CGREvent.Event[fldName]; !has {
					continue //maybe removed
				}
				if _, err := sm.conns[connId].SendApiCmd(
					fmt.Sprintf("uuid_setvar %s %s %s\n\n", fsev.GetUUID(), fldName,
						authReply.Attributes.CGREvent.Event[fldName])); err != nil {
					utils.Logger.Info(
						fmt.Sprintf("<%s> error %s setting channel variabile: %s",
							utils.FreeSWITCHAgent, err.Error(), fldName))
					sm.unparkCall(fsev.GetUUID(), connId,
						fsev.GetCallDestNr(utils.META_DEFAULT), utils.ErrServerError.Error())
					return
				}
			}
		}
	}
	sm.unparkCall(fsev.GetUUID(), connId,
		fsev.GetCallDestNr(utils.META_DEFAULT), AUTH_OK)
}

func (sm *FSsessions) onChannelAnswer(fsev FSEvent, connId string) {
	if fsev.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	chanUUID := fsev.GetUUID()
	if missing := fsev.MissingParameter(sm.timezone); missing != "" {
		sm.disconnectSession(connId, chanUUID, "",
			utils.NewErrMandatoryIeMissing(missing).Error())
		return
	}
	initSessionArgs := fsev.V1InitSessionArgs()
	initSessionArgs.CGREvent.Event[FsConnID] = connId // Attach the connection ID so we can properly disconnect later
	var initReply sessions.V1InitSessionReply
	if err := sm.smg.Call(utils.SessionSv1InitiateSession,
		initSessionArgs, &initReply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> could not process answer for event %s, error: %s",
				utils.FreeSWITCHAgent, chanUUID, err.Error()))
		sm.disconnectSession(connId, chanUUID, "", utils.ErrServerError.Error())
		return
	}
}

func (sm *FSsessions) onChannelHangupComplete(fsev FSEvent, connId string) {
	if fsev.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	var reply string
	if fsev[VarAnswerEpoch] != "0" { // call was answered
		if err := sm.smg.Call(utils.SessionSv1TerminateSession,
			fsev.V1TerminateSessionArgs(), &reply); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not terminate session with event %s, error: %s",
					utils.FreeSWITCHAgent, fsev.GetUUID(), err.Error()))
		}
	}
	if sm.cfg.CreateCdr {
		cgrEv, err := fsev.AsCGREvent(sm.timezone)
		if err != nil {
			return
		}
		if err := sm.smg.Call(utils.SessionSv1ProcessCDR, *cgrEv, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed processing CGREvent: %s,  error: <%s>",
				utils.FreeSWITCHAgent, utils.ToJSON(cgrEv), err.Error()))
		}
	}
}

// Connects to the freeswitch mod_event_socket server and starts
// listening for events.
func (sm *FSsessions) Connect() error {
	eventFilters := map[string][]string{"Call-Direction": []string{"inbound"}}
	errChan := make(chan error)
	for _, connCfg := range sm.cfg.EventSocketConns {
		connId := utils.GenUUID()
		fSock, err := fsock.NewFSock(connCfg.Address, connCfg.Password, connCfg.Reconnects,
			sm.createHandlers(), eventFilters, utils.Logger.GetSyslog(), connId)
		if err != nil {
			return err
		} else if !fSock.Connected() {
			return errors.New("Could not connect to FreeSWITCH")
		} else {
			sm.conns[connId] = fSock
		}
		utils.Logger.Info(fmt.Sprintf("<%s> successfully connected to FreeSWITCH at: <%s>", utils.FreeSWITCHAgent, connCfg.Address))
		go func() { // Start reading in own goroutine, return on error
			if err := sm.conns[connId].ReadEvents(); err != nil {
				errChan <- err
			}
		}()
		if fsSenderPool, err := fsock.NewFSockPool(5, connCfg.Address, connCfg.Password, 1, sm.cfg.MaxWaitConnection,
			make(map[string][]func(string, string)), make(map[string][]string), utils.Logger.GetSyslog(), connId); err != nil {
			return fmt.Errorf("Cannot connect FreeSWITCH senders pool, error: %s", err.Error())
		} else if fsSenderPool == nil {
			return errors.New("Cannot connect FreeSWITCH senders pool.")
		} else {
			sm.senderPools[connId] = fsSenderPool
		}
	}
	err := <-errChan // Will keep the Connect locked until the first error in one of the connections
	return err
}

// fsev.GetCallDestNr(utils.META_DEFAULT)
// Disconnects a session by sending hangup command to freeswitch
func (sm *FSsessions) disconnectSession(connId, uuid, redirectNr, notify string) error {
	if _, err := sm.conns[connId].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", uuid, notify)); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error: %s when attempting to disconnect channelID: %s over connID: %s",
				utils.FreeSWITCHAgent, err.Error(), uuid, connId))
		return err
	}
	if notify == utils.ErrInsufficientCredit.Error() {
		if len(sm.cfg.EmptyBalanceContext) != 0 {
			if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_transfer %s %s XML %s\n\n",
				uuid, redirectNr, sm.cfg.EmptyBalanceContext)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Could not transfer the call to empty balance context, error: <%s>, connId: %s",
					utils.FreeSWITCHAgent, err.Error(), connId))
				return err
			}
			return nil
		} else if len(sm.cfg.EmptyBalanceAnnFile) != 0 {
			if _, err := sm.conns[connId].SendApiCmd(fmt.Sprintf("uuid_broadcast %s playback!manager_request::%s aleg\n\n",
				uuid, sm.cfg.EmptyBalanceAnnFile)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Could not send uuid_broadcast to freeswitch, error: <%s>, connId: %s",
					utils.FreeSWITCHAgent, err.Error(), connId))
				return err
			}
			return nil
		}
	}
	if err := sm.conns[connId].SendMsgCmd(uuid,
		map[string]string{"call-command": "hangup", "hangup-cause": "MANAGER_REQUEST"}); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send disconect msg to freeswitch, error: <%s>, connId: %s",
				utils.FreeSWITCHAgent, err.Error(), connId))
		return err
	}
	return nil
}

func (sm *FSsessions) Shutdown() (err error) {
	for connId, fSock := range sm.conns {
		if !fSock.Connected() {
			utils.Logger.Err(fmt.Sprintf("<%s> Cannot shutdown sessions, fsock not connected for connection id: %s", utils.FreeSWITCHAgent, connId))
			continue
		}
		utils.Logger.Info(fmt.Sprintf("<%s> Shutting down all sessions on connection id: %s", utils.FreeSWITCHAgent, connId))
		if _, err = fSock.SendApiCmd("hupall MANAGER_REQUEST cgr_reqtype *prepaid"); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error on calls shutdown: %s, connection id: %s", utils.FreeSWITCHAgent, err.Error(), connId))
		}
	}
	return
}

// rpcclient.RpcClientConnection interface
func (sm *FSsessions) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(sm, serviceMethod, args, reply)
}

// Internal method to disconnect session in asterisk
func (fsa *FSsessions) V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) (err error) {
	fsEv := sessions.SMGenericEvent(args.EventStart)
	channelID := fsEv.GetOriginID(utils.META_DEFAULT)
	if err = fsa.disconnectSession(fsEv[FsConnID].(string), channelID, fsEv.GetCallDestNr(utils.META_DEFAULT),
		utils.ErrInsufficientCredit.Error()); err != nil {
		return
	}
	*reply = utils.OK
	return
}
