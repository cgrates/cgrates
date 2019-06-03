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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/kamevapi"
	"github.com/cgrates/rpcclient"
)

var (
	kamAuthReqRegexp     = regexp.MustCompile(CGR_AUTH_REQUEST)
	kamCallStartRegexp   = regexp.MustCompile(CGR_CALL_START)
	kamCallEndRegexp     = regexp.MustCompile(CGR_CALL_END)
	kamDlgListRegexp     = regexp.MustCompile(CGR_DLG_LIST)
	kamProcessEventRegex = regexp.MustCompile(CGR_PROCESS_EVENT)
)

func NewKamailioAgent(kaCfg *config.KamAgentCfg,
	sessionS rpcclient.RpcClientConnection, timezone string) (ka *KamailioAgent) {
	ka = &KamailioAgent{
		cfg:              kaCfg,
		sessionS:         sessionS,
		timezone:         timezone,
		conns:            make(map[string]*kamevapi.KamEvapi),
		activeSessionIDs: make(chan []*sessions.SessionID),
	}
	return
}

type KamailioAgent struct {
	cfg              *config.KamAgentCfg
	sessionS         rpcclient.RpcClientConnection
	timezone         string
	conns            map[string]*kamevapi.KamEvapi
	activeSessionIDs chan []*sessions.SessionID
}

func (self *KamailioAgent) Connect() error {
	var err error
	eventHandlers := map[*regexp.Regexp][]func([]byte, string){
		kamAuthReqRegexp:     {self.onCgrAuth},
		kamCallStartRegexp:   {self.onCallStart},
		kamCallEndRegexp:     {self.onCallEnd},
		kamDlgListRegexp:     {self.onDlgList},
		kamProcessEventRegex: {self.onCgrProcessEvent},
	}
	errChan := make(chan error)
	for _, connCfg := range self.cfg.EvapiConns {
		connID := utils.GenUUID()
		logger := log.New(utils.Logger, "kamevapi:", 2)
		if self.conns[connID], err = kamevapi.NewKamEvapi(connCfg.Address, connID, connCfg.Reconnects, eventHandlers, logger); err != nil {
			return err
		}
		go func() { // Start reading in own goroutine, return on error
			if err := self.conns[connID].ReadEvents(); err != nil {
				errChan <- err
			}
		}()
	}
	err = <-errChan // Will keep the Connect locked until the first error in one of the connections
	return err
}

func (self *KamailioAgent) Shutdown() error {
	return nil
}

// rpcclient.RpcClientConnection interface
func (ka *KamailioAgent) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(ka, serviceMethod, args, reply)
}

// onCgrAuth is called when new event of type CGR_AUTH_REQUEST is coming
func (ka *KamailioAgent) onCgrAuth(evData []byte, connID string) {
	kev, err := NewKamEvent(evData)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> unmarshalling event data: %s, error: %s",
			utils.KamailioAgent, evData, err.Error()))
		return
	}
	if kev[utils.RequestType] == utils.META_NONE { // Do not process this request
		return
	}
	if kev.MissingParameter() {
		if kRply, err := kev.AsKamAuthReply(nil, nil, utils.ErrMandatoryIeMissing); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed building auth reply for event: %s, error: %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		} else if err = ka.conns[connID].Send(kRply.String()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed sending auth reply for event: %s, error %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		}
		return
	}
	authArgs := kev.V1AuthorizeArgs()
	if authArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate auth session arguments",
			utils.KamailioAgent, kev[utils.OriginID]))
		return
	}
	authArgs.CGREvent.Event[utils.OriginHost] = utils.FirstNonEmpty(authArgs.CGREvent.Event[utils.OriginHost].(string), ka.conns[connID].RemoteAddr().String())
	authArgs.CGREvent.Event[EvapiConnID] = connID // Attach the connection ID
	var authReply sessions.V1AuthorizeReply
	err = ka.sessionS.Call(utils.SessionSv1AuthorizeEvent, authArgs, &authReply)
	if kar, err := kev.AsKamAuthReply(authArgs, &authReply, err); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed building auth reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	} else if err = ka.conns[connID].Send(kar.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed sending auth reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	}
}

func (ka *KamailioAgent) onCallStart(evData []byte, connID string) {
	kev, err := NewKamEvent(evData)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> unmarshalling event: %s, error: %s",
			utils.KamailioAgent, evData, err.Error()))
		return
	}
	if kev[utils.RequestType] == utils.META_NONE { // Do not process this request
		return
	}
	if kev.MissingParameter() {
		ka.disconnectSession(connID,
			NewKamSessionDisconnect(kev[KamHashEntry], kev[KamHashID],
				utils.ErrMandatoryIeMissing.Error()))
		return
	}
	initSessionArgs := kev.V1InitSessionArgs()
	if initSessionArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate init session arguments",
			utils.KamailioAgent, kev[utils.OriginID]))
		return
	}
	initSessionArgs.CGREvent.Event[EvapiConnID] = connID // Attach the connection ID so we can properly disconnect later
	initSessionArgs.CGREvent.Event[utils.OriginHost] = utils.FirstNonEmpty(initSessionArgs.CGREvent.Event[utils.OriginHost].(string), ka.conns[connID].RemoteAddr().String())

	var initReply sessions.V1InitSessionReply
	if err := ka.sessionS.Call(utils.SessionSv1InitiateSession,
		initSessionArgs, &initReply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> could not process answer for event %s, error: %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		ka.disconnectSession(connID,
			NewKamSessionDisconnect(kev[KamHashEntry], kev[KamHashID],
				utils.ErrServerError.Error()))
		return
	}
}

func (ka *KamailioAgent) onCallEnd(evData []byte, connID string) {
	kev, err := NewKamEvent(evData)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> unmarshalling event: %s, error: %s",
			utils.KamailioAgent, evData, err.Error()))
		return
	}
	if kev[utils.RequestType] == utils.META_NONE { // Do not process this request
		return
	}
	if kev.MissingParameter() {
		utils.Logger.Err(fmt.Sprintf("<%s> mandatory IE missing out from event: %s",
			utils.KamailioAgent, kev[utils.OriginID]))
		return
	}
	tsArgs := kev.V1TerminateSessionArgs()
	if tsArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate terminate session arguments",
			utils.KamailioAgent, kev[utils.OriginID]))
		return
	}
	var reply string
	tsArgs.CGREvent.Event[utils.OriginHost] = utils.FirstNonEmpty(tsArgs.CGREvent.Event[utils.OriginHost].(string), ka.conns[connID].RemoteAddr().String())
	tsArgs.CGREvent.Event[EvapiConnID] = connID // Attach the connection ID in case we need to create a session and disconnect it
	if err := ka.sessionS.Call(utils.SessionSv1TerminateSession,
		tsArgs, &reply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> could not terminate session with event %s, error: %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		// no return here since we want CDR anyhow
	}
	if ka.cfg.CreateCdr || strings.Index(kev[utils.CGRSubsystems], utils.MetaCDRs) != -1 {
		cgrEv, err := kev.AsCGREvent(ka.timezone)
		if err != nil {
			return
		}
		cgrEv.Event[utils.OriginHost] = utils.FirstNonEmpty(cgrEv.Event[utils.OriginHost].(string), ka.conns[connID].RemoteAddr().String())
		cgrArgs := cgrEv.ConsumeArgs(strings.Index(kev[utils.CGRSubsystems], utils.MetaDispatchers) != -1, false)
		if err := ka.sessionS.Call(utils.SessionSv1ProcessCDR,
			&utils.CGREventWithArgDispatcher{CGREvent: cgrEv, ArgDispatcher: cgrArgs.ArgDispatcher}, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("%s> failed processing CGREvent: %s, error: %s",
				utils.KamailioAgent, utils.ToJSON(cgrEv), err.Error()))
		}
	}
}

func (ka *KamailioAgent) onDlgList(evData []byte, connID string) {
	kamDlgRpl, err := NewKamDlgReply(evData)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> unmarshalling event data: %s, error: %s",
			utils.KamailioAgent, evData, err.Error()))
		return
	}
	var sIDs []*sessions.SessionID
	for _, dlgInfo := range kamDlgRpl.Jsonrpl_body.Result {
		sIDs = append(sIDs, &sessions.SessionID{
			OriginHost: ka.conns[connID].RemoteAddr().String(),
			OriginID:   dlgInfo.CallId + ";" + dlgInfo.Caller.Tag,
		})
	}
	ka.activeSessionIDs <- sIDs
}

func (ka *KamailioAgent) onCgrProcessEvent(evData []byte, connID string) {
	kev, err := NewKamEvent(evData)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> unmarshalling event data: %s, error: %s",
			utils.KamailioAgent, evData, err.Error()))
		return
	}

	if kev.MissingParameter() {
		if kRply, err := kev.AsKamProcessEventReply(nil, nil, utils.ErrMandatoryIeMissing); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed building process session event reply for event: %s, error: %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		} else if err = ka.conns[connID].Send(kRply.String()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed sending process session event reply for event: %s, error %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		}
		return
	}

	//in case that we don't reveice cgr_subsystems from kamailio
	//we consider this as ping-pong event
	if _, has := kev[utils.CGRSubsystems]; !has {
		if err = ka.conns[connID].Send(kev.AsKamProcessEventEmptyReply().String()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed sending empty process event reply for event: %s, error %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		}
	}

	procEvArgs := kev.V1ProcessEventArgs()
	if procEvArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate process event session arguments",
			utils.KamailioAgent, kev[utils.OriginID]))
		return
	}
	procEvArgs.CGREvent.Event[utils.OriginHost] = utils.FirstNonEmpty(procEvArgs.CGREvent.Event[utils.OriginHost].(string), ka.conns[connID].RemoteAddr().String())
	procEvArgs.CGREvent.Event[EvapiConnID] = connID // Attach the connection ID

	var processReply sessions.V1ProcessEventReply
	err = ka.sessionS.Call(utils.SessionSv1ProcessEvent, procEvArgs, &processReply)

	if kar, err := kev.AsKamProcessEventReply(procEvArgs, &processReply, err); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed building process session event reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	} else if err = ka.conns[connID].Send(kar.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed sending auth reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	}
}

func (self *KamailioAgent) disconnectSession(connID string, dscEv *KamSessionDisconnect) error {
	if err := self.conns[connID].Send(dscEv.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed sending disconnect request: %s,  connection id: %s, error %s",
			utils.KamailioAgent, utils.ToJSON(dscEv), err.Error(), connID))
		return err
	}
	return nil
}

// Internal method to disconnect session in Kamailio
func (ka *KamailioAgent) V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) (err error) {
	hEntry, err := utils.IfaceAsString(args.EventStart[KamHashEntry])
	if err != nil {
		return err
	}
	hID, err := utils.IfaceAsString(args.EventStart[KamHashID])
	if err != nil {
		return err
	}
	connIDIface, has := args.EventStart[EvapiConnID]
	if !has {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error: <%s:%s> when attempting to disconnect <%s:%s> and <%s:%s>",
				utils.KamailioAgent, utils.ErrNotFound.Error(), EvapiConnID,
				KamHashEntry, hEntry, KamHashID, hID))
		return
	}
	connID, err := utils.IfaceAsString(connIDIface)
	if err != nil {
		return err
	}
	if err = ka.disconnectSession(connID,
		NewKamSessionDisconnect(hEntry, hID,
			utils.ErrInsufficientCredit.Error())); err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (ka *KamailioAgent) V1GetActiveSessionIDs(ignParam string, sessionIDs *[]*sessions.SessionID) (err error) {
	for _, evapi := range ka.conns {
		kamEv, _ := json.Marshal(map[string]string{utils.Event: CGR_DLG_LIST})
		if err = evapi.Send(string(kamEv)); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed sending event, error %s",
				utils.KamailioAgent, err.Error()))
			return
		}
	}
	select {
	case *sessionIDs = <-ka.activeSessionIDs:
	case <-time.After(5 * time.Second):
		return errors.New("timeout executing dialog list")
	}
	return
}
