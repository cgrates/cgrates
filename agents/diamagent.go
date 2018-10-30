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
	"reflect"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
)

func NewDiameterAgent(cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	sessionS rpcclient.RpcClientConnection) (*DiameterAgent, error) {
	if sessionS != nil && reflect.ValueOf(sessionS).IsNil() {
		sessionS = nil
	}
	da := &DiameterAgent{cgrCfg: cgrCfg, filterS: filterS, sessionS: sessionS}
	dictsPath := cgrCfg.DiameterAgentCfg().DictionariesPath
	if len(dictsPath) != 0 {
		if err := loadDictionaries(dictsPath, utils.DiameterAgent); err != nil {
			return nil, err
		}
	}
	msgTemplates := da.cgrCfg.DiameterAgentCfg().Templates
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

type DiameterAgent struct {
	cgrCfg   *config.CGRConfig
	filterS  *engine.FilterS
	sessionS rpcclient.RpcClientConnection // Connection towards CGR-SessionS component
}

// ListenAndServe is called when DiameterAgent is started, usually from within cmd/cgr-engine
func (da *DiameterAgent) ListenAndServe() error {
	return diam.ListenAndServe(da.cgrCfg.DiameterAgentCfg().Listen, da.handlers(), nil)
}

// Creates the message handlers
func (da *DiameterAgent) handlers() diam.Handler {
	settings := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(da.cgrCfg.DiameterAgentCfg().OriginHost),
		OriginRealm:      datatype.DiameterIdentity(da.cgrCfg.DiameterAgentCfg().OriginRealm),
		VendorID:         datatype.Unsigned32(da.cgrCfg.DiameterAgentCfg().VendorId),
		ProductName:      datatype.UTF8String(da.cgrCfg.DiameterAgentCfg().ProductName),
		FirmwareRevision: datatype.Unsigned32(utils.DIAMETER_FIRMWARE_REVISION),
	}
	dSM := sm.New(settings)
	dSM.HandleFunc("ALL", da.handleMessage) // route all commands to one dispatcher
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
		writeOnConn(c, m.Answer(diam.NoCommonApplication))
		return
	}
	dCmd, err := m.Dictionary().FindCommand(
		m.Header.ApplicationID,
		m.Header.CommandCode)
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> decoding app: %d, command %d, err: %s",
			utils.DiameterAgent, m.Header.ApplicationID, m.Header.CommandCode, err.Error()))
		writeOnConn(c, m.Answer(diam.CommandUnsupported))
		return
	}
	reqVars := map[string]interface{}{
		utils.OriginHost:  da.cgrCfg.DiameterAgentCfg().OriginHost, // used in templates
		utils.OriginRealm: da.cgrCfg.DiameterAgentCfg().OriginRealm,
		utils.ProductName: da.cgrCfg.DiameterAgentCfg().ProductName,
		utils.MetaApp:     dApp.Name,
		utils.MetaAppID:   dApp.ID,
		utils.MetaCmd:     dCmd.Short + "R",
	}
	rply := config.NewNavigableMap(nil) // share it among different processors
	var processed bool
	for _, reqProcessor := range da.cgrCfg.DiameterAgentCfg().RequestProcessors {
		var lclProcessed bool
		lclProcessed, err = da.processRequest(reqProcessor,
			newAgentRequest(
				newDADataProvider(m), reqVars, rply,
				reqProcessor.Tenant, da.cgrCfg.GeneralCfg().DefaultTenant,
				utils.FirstNonEmpty(reqProcessor.Timezone,
					config.CgrConfig().GeneralCfg().DefaultTimezone),
				da.filterS))
		if lclProcessed {
			processed = lclProcessed
		}
		if err != nil ||
			(lclProcessed && !reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing message: %s",
				utils.DiameterAgent, err.Error(), m))
		writeOnConn(c, m.Answer(diam.UnableToComply))
		return
	} else if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring message %s from %s",
				utils.DiameterAgent, m, c.RemoteAddr()))
		writeOnConn(c, m.Answer(diam.UnableToComply))
		return
	}
	a := m.Answer(diam.Success)
	// write reply into message
	pathIdx := make(map[string]int) // group items for same path
	for _, val := range rply.Values() {
		nmItms, isNMItems := val.([]*config.NMItem)
		if !isNMItems {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> cannot encode reply field: %s, ignoring message %s from %s",
					utils.DiameterAgent, utils.ToJSON(val), m, c.RemoteAddr()))
			writeOnConn(c, m.Answer(diam.UnableToComply))
			return
		}
		// find out the first itm which is not an attribute
		var itm *config.NMItem
		var itmStr string
		for _, cfgItm := range nmItms {
			if cfgItm.Config == nil || cfgItm.Config.AttributeID == "" {
				itmStr, err = utils.IfaceAsString(cfgItm.Data)
				if err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error: %s processing reply item: %s for message: %s",
							utils.DiameterAgent, err.Error(), utils.ToJSON(cfgItm), m))
					writeOnConn(c, m.Answer(diam.UnableToComply))
					return
				}
				// uniquePath is path contatenated with data (must be something unique)
				uniquePath := utils.ConcatenatedKey(strings.Join(cfgItm.Path, "."), itmStr)
				if _, has := pathIdx[uniquePath]; !has {
					pathIdx[uniquePath] += 1
					itm = cfgItm
					break
				} else {
					continue
				}
			}
		}

		if itm == nil {
			continue // all attributes, not writable to diameter packet
		}

		var newBranch bool
		if itm.Config != nil && itm.Config.NewBranch {
			newBranch = true
		}
		if err := messageSetAVPsWithPath(a, itm.Path,
			itmStr, newBranch, da.cgrCfg.GeneralCfg().DefaultTimezone); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s setting reply item: %s for message: %s",
					utils.DiameterAgent, err.Error(), utils.ToJSON(itm), m))
			writeOnConn(c, m.Answer(diam.UnableToComply))
			return
		}
	}
	writeOnConn(c, a)
}

func (da *DiameterAgent) processRequest(reqProcessor *config.DARequestProcessor,
	agReq *AgentRequest) (processed bool, err error) {
	if pass, err := da.filterS.Pass(agReq.tenant,
		reqProcessor.Filters, agReq); err != nil || !pass {
		return pass, err
	}
	if agReq.CGRRequest, err = agReq.AsNavigableMap(reqProcessor.RequestFields); err != nil {
		return
	}
	cgrEv := agReq.CGRRequest.AsCGREvent(agReq.tenant, utils.NestingSep)
	var reqType string
	for _, typ := range []string{
		utils.MetaDryRun, utils.MetaAuth,
		utils.MetaInitiate, utils.MetaUpdate,
		utils.MetaTerminate, utils.MetaEvent,
		utils.MetaCDRs} {
		if reqProcessor.Flags.HasKey(typ) { // request type is identified through flags
			reqType = typ
			break
		}
	}
	switch reqType {
	default:
		return false, fmt.Errorf("unknown request type: <%s>", reqType)
	case utils.MetaDryRun:
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, processorID: %s, CGREvent: %s",
				utils.DiameterAgent, reqProcessor.ID, utils.ToJSON(cgrEv)))
	case utils.MetaAuth:
		authArgs := sessions.NewV1AuthorizeArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			reqProcessor.Flags.HasKey(utils.MetaSuppliers),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersIgnoreErrors),
			reqProcessor.Flags.HasKey(utils.MetaSuppliersEventCost),
			*cgrEv)
		var authReply sessions.V1AuthorizeReply
		err = da.sessionS.Call(utils.SessionSv1AuthorizeEvent,
			authArgs, &authReply)
		if agReq.CGRReply, err = NewCGRReply(&authReply, err); err != nil {
			return
		}
	case utils.MetaInitiate:
		initArgs := sessions.NewV1InitSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats), *cgrEv)
		var initReply sessions.V1InitSessionReply
		err = da.sessionS.Call(utils.SessionSv1InitiateSession,
			initArgs, &initReply)
		if agReq.CGRReply, err = NewCGRReply(&initReply, err); err != nil {
			return
		}
	case utils.MetaUpdate:
		updateArgs := sessions.NewV1UpdateSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaAccounts), *cgrEv)
		var updateReply sessions.V1UpdateSessionReply
		err = da.sessionS.Call(utils.SessionSv1UpdateSession,
			updateArgs, &updateReply)
		if agReq.CGRReply, err = NewCGRReply(&updateReply, err); err != nil {
			return
		}
	case utils.MetaTerminate:
		terminateArgs := sessions.NewV1TerminateSessionArgs(
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats), *cgrEv)
		var tRply string
		err = da.sessionS.Call(utils.SessionSv1TerminateSession,
			terminateArgs, &tRply)
		if agReq.CGRReply, err = NewCGRReply(nil, err); err != nil {
			return
		}
	case utils.MetaEvent:
		evArgs := sessions.NewV1ProcessEventArgs(
			reqProcessor.Flags.HasKey(utils.MetaResources),
			reqProcessor.Flags.HasKey(utils.MetaAccounts),
			reqProcessor.Flags.HasKey(utils.MetaAttributes),
			reqProcessor.Flags.HasKey(utils.MetaThresholds),
			reqProcessor.Flags.HasKey(utils.MetaStats),
			*cgrEv)
		var eventRply sessions.V1ProcessEventReply
		err = da.sessionS.Call(utils.SessionSv1ProcessEvent,
			evArgs, &eventRply)
		if utils.ErrHasPrefix(err, utils.RalsErrorPrfx) {
			cgrEv.Event[utils.Usage] = 0 // avoid further debits
		} else if eventRply.MaxUsage != nil {
			cgrEv.Event[utils.Usage] = *eventRply.MaxUsage // make sure the CDR reflects the debit
		}
		if agReq.CGRReply, err = NewCGRReply(&eventRply, err); err != nil {
			return
		}
	case utils.MetaCDRs: // allow CDR processing
	}
	// separate request so we can capture the Terminate/Event also here
	if reqProcessor.Flags.HasKey(utils.MetaCDRs) &&
		!reqProcessor.Flags.HasKey(utils.MetaDryRun) {
		var rplyCDRs string
		if err = da.sessionS.Call(utils.SessionSv1ProcessCDR,
			cgrEv, &rplyCDRs); err != nil {
			agReq.CGRReply.Set([]string{utils.Error}, err.Error(), false)
		}
	}
	if nM, err := agReq.AsNavigableMap(reqProcessor.ReplyFields); err != nil {
		return false, err
	} else {
		agReq.Reply.Merge(nM)
	}
	if reqType == utils.MetaDryRun {
		utils.Logger.Info(
			fmt.Sprintf("<%s> DRY_RUN, Diameter reply: %s",
				utils.DiameterAgent, agReq.Reply))
	}
	return true, nil
}
