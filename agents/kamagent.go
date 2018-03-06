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
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/kamevapi"
)

func NewKamailioAgent(kaCfg *config.KamAgentCfg,
	sessionS *utils.BiRPCInternalClient, timezone string) (ka *KamailioAgent) {
	ka = &KamailioAgent{cfg: kaCfg, sessionS: sessionS,
		timezone: timezone,
		conns:    make(map[string]*kamevapi.KamEvapi)}
	ka.sessionS.SetClientConn(ka) // pass the connection to KA back into smg so we can receive the disconnects
	return
}

type KamailioAgent struct {
	cfg      *config.KamAgentCfg
	sessionS *utils.BiRPCInternalClient
	timezone string
	conns    map[string]*kamevapi.KamEvapi
}

func (self *KamailioAgent) Connect() error {
	var err error
	eventHandlers := map[*regexp.Regexp][]func([]byte, string){
		regexp.MustCompile(CGR_AUTH_REQUEST): []func([]byte, string){
			self.onCgrAuth},
		regexp.MustCompile(CGR_CALL_START): []func([]byte, string){
			self.onCallStart},
		regexp.MustCompile(CGR_CALL_END): []func([]byte, string){self.onCallEnd},
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
	if err := ka.sessionS.Call(utils.SessionSv1TerminateSession,
		tsArgs, &reply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> could not terminate session with event %s, error: %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		// no return here since we want CDR anyhow
	}
	if ka.cfg.CreateCdr || strings.Index(kev[KamCGRSubsystems], utils.MetaCDRs) != -1 {
		cgrEv, err := kev.AsCGREvent(ka.timezone)
		if err != nil {
			return
		}
		if err := ka.sessionS.Call(utils.SessionSv1ProcessCDR, *cgrEv, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("%s> failed processing CGREvent: %s, error: %s",
				utils.KamailioAgent, utils.ToJSON(cgrEv), err.Error()))
		}
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
	hEntry, _ := utils.CastFieldIfToString(args.EventStart[KamHashEntry])
	hID, _ := utils.CastFieldIfToString(args.EventStart[KamHashID])
	connID, _ := utils.CastFieldIfToString(args.EventStart[EvapiConnID])
	if err = ka.disconnectSession(connID,
		NewKamSessionDisconnect(hEntry, hID,
			utils.ErrInsufficientCredit.Error())); err != nil {
		return
	}
	*reply = utils.OK
	return
}
