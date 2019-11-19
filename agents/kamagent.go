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
	kamAuthReqRegexp       = regexp.MustCompile(CGR_AUTH_REQUEST)
	kamCallStartRegexp     = regexp.MustCompile(CGR_CALL_START)
	kamCallEndRegexp       = regexp.MustCompile(CGR_CALL_END)
	kamDlgListRegexp       = regexp.MustCompile(CGR_DLG_LIST)
	kamProcessMessageRegex = regexp.MustCompile(CGR_PROCESS_MESSAGE)
	kamProcessCDRRegex     = regexp.MustCompile(CGR_PROCESS_CDR)
)

func NewKamailioAgent(kaCfg *config.KamAgentCfg,
	sessionS rpcclient.RpcClientConnection, timezone string) (ka *KamailioAgent) {
	ka = &KamailioAgent{
		cfg:              kaCfg,
		sessionS:         sessionS,
		timezone:         timezone,
		conns:            make([]*kamevapi.KamEvapi, len(kaCfg.EvapiConns)),
		activeSessionIDs: make(chan []*sessions.SessionID),
	}
	return
}

type KamailioAgent struct {
	cfg              *config.KamAgentCfg
	sessionS         rpcclient.RpcClientConnection
	timezone         string
	conns            []*kamevapi.KamEvapi
	activeSessionIDs chan []*sessions.SessionID
}

func (self *KamailioAgent) Connect() (err error) {
	eventHandlers := map[*regexp.Regexp][]func([]byte, int){
		kamAuthReqRegexp:       {self.onCgrAuth},
		kamCallStartRegexp:     {self.onCallStart},
		kamCallEndRegexp:       {self.onCallEnd},
		kamDlgListRegexp:       {self.onDlgList},
		kamProcessMessageRegex: {self.onCgrProcessMessage},
		kamProcessCDRRegex:     {self.onCgrProcessCDR},
	}
	errChan := make(chan error)
	for connIdx, connCfg := range self.cfg.EvapiConns {
		logger := log.New(utils.Logger, "kamevapi:", 2)
		if self.conns[connIdx], err = kamevapi.NewKamEvapi(connCfg.Address, connIdx, connCfg.Reconnects, eventHandlers, logger); err != nil {
			return
		}
		go func(conn *kamevapi.KamEvapi) { // Start reading in own goroutine, return on error
			if err := conn.ReadEvents(); err != nil {
				errChan <- err
			}
		}(self.conns[connIdx])
	}
	err = <-errChan // Will keep the Connect locked until the first error in one of the connections
	return
}

func (self *KamailioAgent) Shutdown() (err error) {
	for conIndx, conn := range self.conns {
		if err = conn.Disconnect(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> can't disconnect connection at index %v because: %s",
				utils.KamailioAgent, conIndx, err))
			continue
		}
	}
	return
}

// rpcclient.RpcClientConnection interface
func (ka *KamailioAgent) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(ka, serviceMethod, args, reply)
}

// onCgrAuth is called when new event of type CGR_AUTH_REQUEST is coming
func (ka *KamailioAgent) onCgrAuth(evData []byte, connIdx int) {
	if connIdx >= len(ka.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(ka.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.KamailioAgent, err.Error()))
		return
	}
	kev, err := NewKamEvent(evData, ka.cfg.EvapiConns[connIdx].Alias, ka.conns[connIdx].RemoteAddr().String())
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
		} else if err = ka.conns[connIdx].Send(kRply.String()); err != nil {
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
	authArgs.CGREvent.Event[EvapiConnID] = connIdx // Attach the connection ID
	var authReply sessions.V1AuthorizeReply
	if err = ka.sessionS.Call(utils.SessionSv1AuthorizeEvent, authArgs, &authReply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to auth event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	} else if kar, err := kev.AsKamAuthReply(authArgs, &authReply, err); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed building auth reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	} else if err = ka.conns[connIdx].Send(kar.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed sending auth reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	}
}

func (ka *KamailioAgent) onCallStart(evData []byte, connIdx int) {
	if connIdx >= len(ka.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(ka.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.KamailioAgent, err.Error()))
		return
	}
	kev, err := NewKamEvent(evData, ka.cfg.EvapiConns[connIdx].Alias, ka.conns[connIdx].RemoteAddr().String())
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> unmarshalling event: %s, error: %s",
			utils.KamailioAgent, evData, err.Error()))
		return
	}
	if kev[utils.RequestType] == utils.META_NONE { // Do not process this request
		return
	}
	if kev.MissingParameter() {
		ka.disconnectSession(connIdx,
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
	initSessionArgs.CGREvent.Event[EvapiConnID] = connIdx // Attach the connection ID so we can properly disconnect later

	var initReply sessions.V1InitSessionReply
	if err := ka.sessionS.Call(utils.SessionSv1InitiateSession,
		initSessionArgs, &initReply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> could not process answer for event %s, error: %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		ka.disconnectSession(connIdx,
			NewKamSessionDisconnect(kev[KamHashEntry], kev[KamHashID],
				utils.ErrServerError.Error()))
		return
	}
}

func (ka *KamailioAgent) onCallEnd(evData []byte, connIdx int) {
	if connIdx >= len(ka.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(ka.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.KamailioAgent, err.Error()))
		return
	}
	kev, err := NewKamEvent(evData, ka.cfg.EvapiConns[connIdx].Alias, ka.conns[connIdx].RemoteAddr().String())
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
	tsArgs.CGREvent.Event[EvapiConnID] = connIdx // Attach the connection ID in case we need to create a session and disconnect it
	if err := ka.sessionS.Call(utils.SessionSv1TerminateSession,
		tsArgs, &reply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> could not terminate session with event %s, error: %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		// no return here since we want CDR anyhow
	}
	if ka.cfg.CreateCdr || strings.Index(kev[utils.CGRFlags], utils.MetaCDRs) != -1 {
		cgrEv, err := kev.AsCGREvent(ka.timezone) // FixMe: do we need to create the event once again?
		if err != nil {
			return
		}
		cgrArgs := cgrEv.ExtractArgs(strings.Index(kev[utils.CGRFlags], utils.MetaDispatchers) != -1, false)
		if err := ka.sessionS.Call(utils.SessionSv1ProcessCDR,
			&utils.CGREventWithArgDispatcher{CGREvent: cgrEv, ArgDispatcher: cgrArgs.ArgDispatcher}, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("%s> failed processing CGREvent: %s, error: %s",
				utils.KamailioAgent, utils.ToJSON(cgrEv), err.Error()))
		}
	}
}

func (ka *KamailioAgent) onDlgList(evData []byte, connIdx int) {
	kamDlgRpl, err := NewKamDlgReply(evData)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> unmarshalling event data: %s, error: %s",
			utils.KamailioAgent, evData, err.Error()))
		return
	}
	var sIDs []*sessions.SessionID
	// FixMe: find way to add OriginHost from event also, to be compatible with above implementation
	for _, dlgInfo := range kamDlgRpl.Jsonrpl_body.Result {
		sIDs = append(sIDs, &sessions.SessionID{
			OriginHost: ka.conns[connIdx].RemoteAddr().String(),
			OriginID:   dlgInfo.CallId + ";" + dlgInfo.Caller.Tag,
		})
	}
	ka.activeSessionIDs <- sIDs
}

func (ka *KamailioAgent) onCgrProcessMessage(evData []byte, connIdx int) {
	if connIdx >= len(ka.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(ka.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.KamailioAgent, err.Error()))
		return
	}
	kev, err := NewKamEvent(evData, ka.cfg.EvapiConns[connIdx].Alias, ka.conns[connIdx].RemoteAddr().String())
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> unmarshalling event data: %s, error: %s",
			utils.KamailioAgent, evData, err.Error()))
		return
	}

	if kev.MissingParameter() {
		if kRply, err := kev.AsKamProcessMessageReply(nil, nil, utils.ErrMandatoryIeMissing); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed building process session event reply for event: %s, error: %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		} else if err = ka.conns[connIdx].Send(kRply.String()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed sending process session event reply for event: %s, error %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		}
		return
	}

	//in case that we don't reveice cgr_flags from kamailio
	//we consider this as ping-pong event
	if _, has := kev[utils.CGRFlags]; !has {
		if err = ka.conns[connIdx].Send(kev.AsKamProcessMessageEmptyReply().String()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed sending empty process message reply for event: %s, error %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		}
	}

	procEvArgs := kev.V1ProcessMessageArgs()
	if procEvArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate process message session arguments",
			utils.KamailioAgent, kev[utils.OriginID]))
		return
	}
	procEvArgs.CGREvent.Event[EvapiConnID] = connIdx // Attach the connection ID

	var processReply sessions.V1ProcessMessageReply
	if err = ka.sessionS.Call(utils.SessionSv1ProcessMessage, procEvArgs, &processReply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to proccess message: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	} else if kar, err := kev.AsKamProcessMessageReply(procEvArgs, &processReply, err); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed building process session event reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	} else if err = ka.conns[connIdx].Send(kar.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed sending auth reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	}
}

func (ka *KamailioAgent) onCgrProcessCDR(evData []byte, connIdx int) {
	if connIdx >= len(ka.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(ka.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.KamailioAgent, err.Error()))
		return
	}
	kev, err := NewKamEvent(evData, ka.cfg.EvapiConns[connIdx].Alias, ka.conns[connIdx].RemoteAddr().String())
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> unmarshalling event data: %s, error: %s",
			utils.KamailioAgent, evData, err.Error()))
		return
	}

	if kev.MissingParameter() {
		if kRply, err := kev.AsKamProcessCDRReply(nil, nil, utils.ErrMandatoryIeMissing); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed building process session event reply for event: %s, error: %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		} else if err = ka.conns[connIdx].Send(kRply.String()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed sending process session event reply for event: %s, error %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		}
		return
	}

	procCDRArgs := kev.V1ProcessCDRArgs()
	if procCDRArgs == nil {
		utils.Logger.Err(fmt.Sprintf("<%s> event: %s cannot generate process cdr session arguments",
			utils.KamailioAgent, kev[utils.OriginID]))
		return
	}
	procCDRArgs.CGREvent.Event[EvapiConnID] = connIdx // Attach the connection ID

	var processReply string
	if err = ka.sessionS.Call(utils.SessionSv1ProcessCDR, procCDRArgs, &processReply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed to process CDR for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	} else if kar, err := kev.AsKamProcessCDRReply(procCDRArgs, &processReply, err); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed building process session event reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	} else if err = ka.conns[connIdx].Send(kar.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed sending auth reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
	}
}

func (self *KamailioAgent) disconnectSession(connIdx int, dscEv *KamSessionDisconnect) (err error) {
	if err = self.conns[connIdx].Send(dscEv.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed sending disconnect request: %s,  connection id: %v, error %s",
			utils.KamailioAgent, utils.ToJSON(dscEv), connIdx, err.Error()))
	}
	return
}

// Internal method to disconnect session in Kamailio
func (ka *KamailioAgent) V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) (err error) {
	hEntry := utils.IfaceAsString(args.EventStart[KamHashEntry])
	hID := utils.IfaceAsString(args.EventStart[KamHashID])
	connIdxIface, has := args.EventStart[EvapiConnID]
	if !has {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error: <%s:%s> when attempting to disconnect <%s:%s> and <%s:%s>",
				utils.KamailioAgent, utils.ErrNotFound.Error(), EvapiConnID,
				KamHashEntry, hEntry, KamHashID, hID))
		return
	}
	connIdx, err := utils.IfaceAsTInt64(connIdxIface)
	if err != nil {
		return err
	}
	if int(connIdx) >= len(ka.conns) { // protection against index out of range panic
		err = fmt.Errorf("Index out of range[0,%v): %v ", len(ka.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.KamailioAgent, err.Error()))
		return
	}
	if err = ka.disconnectSession(int(connIdx),
		NewKamSessionDisconnect(hEntry, hID,
			utils.ErrInsufficientCredit.Error())); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// V1GetActiveSessionIDs returns a list of CGRIDs based on active sessions from agent
func (ka *KamailioAgent) V1GetActiveSessionIDs(ignParam string, sessionIDs *[]*sessions.SessionID) (err error) {
	for _, evapi := range ka.conns {
		kamEv, _ := json.Marshal(map[string]string{utils.Event: CGR_DLG_LIST})
		if err = evapi.Send(string(kamEv)); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed sending event, error %s",
				utils.KamailioAgent, err.Error()))
			return
		}
	}
	for range ka.conns {
		select {
		case sIDs := <-ka.activeSessionIDs:
			*sessionIDs = append(*sessionIDs, sIDs...)
		case <-time.After(5 * time.Second):
			return errors.New("timeout executing dialog list")
		}
	}
	return
}

// SetSessionSConnection sets the new connection to the session service
// only used on reload
func (ka *KamailioAgent) SetSessionSConnection(sS rpcclient.RpcClientConnection) {
	ka.sessionS = sS
}

// Reload recreates the connection buffers
// only used on reload
func (ka *KamailioAgent) Reload() {
	ka.conns = make([]*kamevapi.KamEvapi, len(ka.cfg.EvapiConns))
}
