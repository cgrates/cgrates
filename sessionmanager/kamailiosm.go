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

package sessionmanager

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/kamevapi"
	"github.com/cgrates/rpcclient"
)

func NewKamailioSessionManager(smKamCfg *config.SmKamConfig, rater, cdrsrv,
	rlS rpcclient.RpcClientConnection, timezone string) (ksm *KamailioSessionManager, err error) {
	if rlS != nil && reflect.ValueOf(rlS).IsNil() {
		rlS = nil
	}
	ksm = &KamailioSessionManager{cfg: smKamCfg, rater: rater, cdrsrv: cdrsrv, rlS: rlS,
		timezone: timezone, conns: make(map[string]*kamevapi.KamEvapi), sessions: NewSessions()}
	return
}

type KamailioSessionManager struct {
	cfg      *config.SmKamConfig
	rater    rpcclient.RpcClientConnection
	cdrsrv   rpcclient.RpcClientConnection
	rlS      rpcclient.RpcClientConnection
	timezone string
	conns    map[string]*kamevapi.KamEvapi
	sessions *Sessions
}

func (self *KamailioSessionManager) getSuppliers(kev KamEvent) (string, error) {
	cd, err := kev.AsCallDescriptor()
	cd.CgrID = kev.GetCgrId(self.timezone)
	if err != nil {
		utils.Logger.Info(fmt.Sprintf("<SM-Kamailio> LCR_PREPROCESS_ERROR error: %s", err.Error()))
		return "", errors.New("LCR_PREPROCESS_ERROR")
	}
	var lcr engine.LCRCost
	if err = self.Rater().Call("Responder.GetLCR", &engine.AttrGetLcr{CallDescriptor: cd}, &lcr); err != nil {
		utils.Logger.Info(fmt.Sprintf("<SM-Kamailio> LCR_API_ERROR error: %s", err.Error()))
		return "", errors.New("LCR_API_ERROR")
	}
	if lcr.HasErrors() {
		lcr.LogErrors()
		return "", errors.New("LCR_COMPUTE_ERROR")
	}
	return lcr.SuppliersString()
}

func (self *KamailioSessionManager) allocateResources(kev KamEvent) (err error) {
	if self.rlS == nil {
		return errors.New("no RLs connection")
	}
	var ev map[string]interface{}
	if ev, err = kev.AsMapStringIface(); err != nil {
		return
	}
	attrRU := utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: kev.GetTenant(utils.META_DEFAULT),
			Event:  ev,
		},
		UsageID: kev.GetUUID(),
		Units:   1, // One channel reserved
	}
	var reply string
	return self.rlS.Call(utils.ResourceSv1AllocateResources, attrRU, &reply)
}

func (self *KamailioSessionManager) onCgrAuth(evData []byte, connId string) {
	kev, err := NewKamEvent(evData)
	if err != nil {
		utils.Logger.Info(fmt.Sprintf("<SM-Kamailio> ERROR unmarshalling event: %s, error: %s", evData, err.Error()))
		return
	}
	if kev.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	if kev.MissingParameter(self.timezone) {
		if kar, err := kev.AsKamAuthReply(0.0, "", false, "", utils.ErrMandatoryIeMissing); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed building auth reply %s", err.Error()))
		} else if err = self.conns[connId].Send(kar.String()); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed sending auth reply %s", err.Error()))
		}
		return
	}
	var remainingDuration float64
	var errReply error
	if errReply = self.rater.Call("Responder.GetDerivedMaxSessionTime",
		kev.AsCDR(self.timezone), &remainingDuration); errReply != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Could not get max session time, error: %s", errReply.Error()))
	}
	var supplStr string
	var errSuppl error
	if kev.ComputeLcr() {
		if supplStr, errSuppl = self.getSuppliers(kev); errSuppl != nil {
			utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Could not get suppliers, error: %s", errSuppl.Error()))
		}
	}
	if errReply == nil { // Overwrite the error from maxSessionTime with the one from suppliers if nil
		errReply = errSuppl
	}
	resourceAllowed := true
	if self.rlS != nil {
		if err := self.allocateResources(kev); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> RLs error: %s", err.Error()))
			resourceAllowed = false
		}
	}
	if kar, err := kev.AsKamAuthReply(remainingDuration, supplStr, resourceAllowed, "", errReply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed building auth reply %s", err.Error()))
	} else if err = self.conns[connId].Send(kar.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed sending auth reply %s", err.Error()))
	}
}

func (self *KamailioSessionManager) onCgrLcrReq(evData []byte, connId string) {
	kev, err := NewKamEvent(evData)
	if err != nil {
		utils.Logger.Info(fmt.Sprintf("<SM-Kamailio> ERROR unmarshalling event: %s, error: %s", string(evData), err.Error()))
		return
	}
	supplStr, err := self.getSuppliers(kev)
	kamLcrReply, errReply := kev.AsKamAuthReply(0, supplStr, false, "", err)
	kamLcrReply.Event = CGR_LCR_REPLY // Hit the CGR_LCR_REPLY event route on Kamailio side
	if errReply != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed building LCR reply %s", errReply.Error()))
	} else if err = self.conns[connId].Send(kamLcrReply.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed sending LCR reply %s", err.Error()))
	}
}

// onCgrRLReq is the handler for CGR_RL_REQUEST events coming from Kamailio
func (self *KamailioSessionManager) onCgrRLReq(evData []byte, connId string) {
	kev, err := NewKamEvent(evData)
	if err != nil {
		utils.Logger.Info(fmt.Sprintf("<SM-Kamailio> ERROR unmarshalling event: %s, error: %s", string(evData), err.Error()))
		return
	}
	resourceAllowed := true
	if err := self.allocateResources(kev); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> RLs error: %s", err.Error()))
		resourceAllowed = false
	}
	kamRLReply, errReply := kev.AsKamAuthReply(0, "", resourceAllowed, "", err)
	kamRLReply.Event = CGR_RL_REPLY // Hit the CGR_LCR_REPLY event route on Kamailio side
	if errReply != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed building RL reply %s", errReply.Error()))
	} else if err = self.conns[connId].Send(kamRLReply.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed sending RL reply %s", err.Error()))
	}
}

func (self *KamailioSessionManager) onCallStart(evData []byte, connId string) {
	kamEv, err := NewKamEvent(evData)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> ERROR unmarshalling event: %s, error: %s", evData, err.Error()))
		return
	}
	if kamEv.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	if kamEv.MissingParameter(self.timezone) {
		self.DisconnectSession(kamEv, connId, utils.ErrMandatoryIeMissing.Error())
		return
	}
	s := NewSession(kamEv, connId, self)
	if s != nil {
		self.sessions.indexSession(s)
	}
}

func (self *KamailioSessionManager) onCallEnd(evData []byte, connId string) {
	kev, err := NewKamEvent(evData)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> ERROR unmarshalling event: %s, error: %s", evData, err.Error()))
		return
	}
	if kev.GetReqType(utils.META_DEFAULT) == utils.META_NONE { // Do not process this request
		return
	}
	if kev.MissingParameter(self.timezone) {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Mandatory IE missing out of event: %+v", kev))
	}
	go self.ProcessCdr(kev.AsCDR(self.Timezone()))
	if self.rlS != nil { // Release RLs resource
		go func() {
			ev, err := kev.AsMapStringIface()
			if err != nil {
				utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> RLs error: %s", err.Error()))
				return
			}
			var reply string
			attrRU := utils.ArgRSv1ResourceUsage{
				CGREvent: utils.CGREvent{
					Tenant: kev.GetTenant(utils.META_DEFAULT),
					Event:  ev,
				},
				UsageID: kev.GetUUID(),
				Units:   1,
			}
			if err := self.rlS.Call(utils.ResourceSv1ReleaseResources, attrRU, &reply); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> RLs API error: %s", err.Error()))
			}
		}()
	}
	if s := self.sessions.getSession(kev.GetUUID()); s != nil {
		if err := self.sessions.removeSession(s, kev); err != nil {
			utils.Logger.Err(err.Error())
		}
	}

}

func (self *KamailioSessionManager) Connect() error {
	var err error
	eventHandlers := map[*regexp.Regexp][]func([]byte, string){
		regexp.MustCompile(CGR_AUTH_REQUEST): []func([]byte, string){self.onCgrAuth},
		regexp.MustCompile(CGR_LCR_REQUEST):  []func([]byte, string){self.onCgrLcrReq},
		regexp.MustCompile(CGR_RL_REQUEST):   []func([]byte, string){self.onCgrRLReq},
		regexp.MustCompile(CGR_CALL_START):   []func([]byte, string){self.onCallStart},
		regexp.MustCompile(CGR_CALL_END):     []func([]byte, string){self.onCallEnd},
	}
	errChan := make(chan error)
	for _, connCfg := range self.cfg.EvapiConns {
		connId := utils.GenUUID()
		logger := log.New(utils.Logger, "KamEvapi:", 2)
		if self.conns[connId], err = kamevapi.NewKamEvapi(connCfg.Address, connId, connCfg.Reconnects, eventHandlers, logger); err != nil {
			return err
		}
		go func() { // Start reading in own goroutine, return on error
			if err := self.conns[connId].ReadEvents(); err != nil {
				errChan <- err
			}
		}()
	}
	err = <-errChan // Will keep the Connect locked until the first error in one of the connections
	return err
}

func (self *KamailioSessionManager) DisconnectSession(ev engine.Event, connId, notify string) error {
	sessionIds := ev.GetSessionIds()
	disconnectEv := &KamSessionDisconnect{Event: CGR_SESSION_DISCONNECT, HashEntry: sessionIds[0], HashId: sessionIds[1], Reason: notify}
	if err := self.conns[connId].Send(disconnectEv.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed sending disconnect request, error %s, connection id: %s", err.Error(), connId))
		return err
	}
	return nil
}

func (self *KamailioSessionManager) DebitInterval() time.Duration {
	return self.cfg.DebitInterval
}
func (self *KamailioSessionManager) CdrSrv() rpcclient.RpcClientConnection {
	return self.cdrsrv
}
func (self *KamailioSessionManager) Rater() rpcclient.RpcClientConnection {
	return self.rater
}

func (self *KamailioSessionManager) ProcessCdr(cdr *engine.CDR) error {
	if !self.cfg.CreateCdr {
		return nil
	}
	var reply string
	if err := self.cdrsrv.Call("CdrsV1.ProcessCDR", cdr, &reply); err != nil {
		utils.Logger.Err(fmt.Sprintf("<SM-Kamailio> Failed processing CDR, cgrid: %s, accid: %s, error: <%s>", cdr.CGRID, cdr.OriginID, err.Error()))
	}
	return nil
}

func (sm *KamailioSessionManager) WarnSessionMinDuration(sessionUuid, connId string) {
}

func (self *KamailioSessionManager) Shutdown() error {
	return nil
}

func (self *KamailioSessionManager) Sessions() []*Session {
	return self.sessions.getSessions()
}

func (self *KamailioSessionManager) SyncSessions() error {
	return nil
}

func (self *KamailioSessionManager) Timezone() string {
	return self.timezone
}
