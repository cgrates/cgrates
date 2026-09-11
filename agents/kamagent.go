// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/kamevapi"
)

var (
	kamAllEvents     = regexp.MustCompile(".*")
	kamDlgListRegexp = regexp.MustCompile(CGR_DLG_LIST)
)

func NewKamailioAgent(cfg *config.CGRConfig,
	connMgr *engine.ConnManager, timezone string, caps *engine.Caps, fltrS *engine.FilterS) (ka *KamailioAgent, err error) {
	ka = &KamailioAgent{
		cfg:              cfg,
		kamCfg:           cfg.KamAgentCfg(),
		connMgr:          connMgr,
		timezone:         timezone,
		caps:             caps,
		fltrS:            fltrS,
		conns:            make([]*kamevapi.KamEvapi, len(cfg.KamAgentCfg().EvapiConns)),
		activeSessionIDs: make(chan []*sessions.SessionID),
	}
	srv, _ := birpc.NewService(ka, "", false)
	ka.ctx = context.WithClient(context.TODO(), srv)
	msgTemplates := cfg.TemplatesCfg()
	for _, procsr := range ka.kamCfg.RequestProcessors {
		var tpls []*config.FCTemplate
		if tpls, err = config.InflateTemplates(procsr.RequestFields, msgTemplates); err != nil {
			return nil, err
		} else if tpls != nil {
			procsr.RequestFields = tpls
		}
		if tpls, err = config.InflateTemplates(procsr.ReplyFields, msgTemplates); err != nil {
			return nil, err
		} else if tpls != nil {
			procsr.ReplyFields = tpls
		}
	}
	return
}

type KamailioAgent struct {
	cfg              *config.CGRConfig
	kamCfg           *config.KamAgentCfg
	connMgr          *engine.ConnManager
	timezone         string
	caps             *engine.Caps
	fltrS            *engine.FilterS
	conns            []*kamevapi.KamEvapi
	activeSessionIDs chan []*sessions.SessionID
	ctx              *context.Context
}

func (self *KamailioAgent) Connect() (err error) {
	eventHandlers := map[*regexp.Regexp][]func([]byte, int){
		kamAllEvents:     {self.onKamEvent},
		kamDlgListRegexp: {self.onDlgList},
	}
	errChan := make(chan error)
	for connIdx, connCfg := range self.kamCfg.EvapiConns {
		if self.conns[connIdx], err = kamevapi.NewKamEvapi(connCfg.Address, connIdx, connCfg.Reconnects, connCfg.MaxReconnectInterval,
			utils.FibDuration, eventHandlers, utils.Logger); err != nil {
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
		if conn == nil {
			break
		}
		if err = conn.Disconnect(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> can't disconnect connection at index %v because: %s",
				utils.KamailioAgent, conIndx, err))
			continue
		}
	}
	return
}

func (ka *KamailioAgent) onKamEvent(evData []byte, connIdx int) {
	if kamDlgListRegexp.Match(evData) {
		return
	}
	if ka.caps.IsLimited() {
		if err := ka.caps.Allocate(); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> caps limit reached, rejecting event: %v",
					utils.KamailioAgent, err))
			return
		}
		defer ka.caps.Deallocate()
	}
	if connIdx >= len(ka.conns) {
		utils.Logger.Err(fmt.Sprintf("<%s> Index out of range[0,%v): %v ",
			utils.KamailioAgent, len(ka.conns), connIdx))
		return
	}
	kev, err := NewKamEvent(evData, ka.kamCfg.EvapiConns[connIdx].Alias,
		ka.conns[connIdx].RemoteAddr().String())
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> unmarshalling event data: %s, error: %s",
			utils.KamailioAgent, evData, err.Error()))
		return
	}
	dP := utils.MapStringDP(kev)
	reqVars := &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			EVENT:            utils.NewLeafNode(kev[EVENT]),
			EvapiConnID:      utils.NewLeafNode(connIdx),
			KamTRIndex:       utils.NewLeafNode(kev[KamTRIndex]),
			KamTRLabel:       utils.NewLeafNode(kev[KamTRLabel]),
			utils.OriginHost: utils.NewLeafNode(kev[utils.OriginHost]),
		},
	}
	var processed bool
	opts := utils.MapStorage(kev.GetOptions())
	cgrRplyNM := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	rply := utils.NewOrderedNavigableMap()
	sessConns, _ := engine.GetConnIDs(ka.ctx, ka.kamCfg.Conns, utils.MetaSessionS,
		ka.cfg.GeneralCfg().DefaultTenant, dP, nil, ka.fltrS)
	for _, reqProcessor := range ka.kamCfg.RequestProcessors {
		agReq := NewAgentRequest(
			dP, reqVars, cgrRplyNM, rply, opts,
			reqProcessor.Tenant, ka.cfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(reqProcessor.Timezone, ka.timezone,
				ka.cfg.GeneralCfg().DefaultTimezone),
			ka.cfg, nil, ka.fltrS, nil)
		if len(reqProcessor.RequestFields) == 0 { // no templates, pass the event as it came from Kamailio
			if err = kev.setCGRRequest(agReq.CGRRequest, connIdx); err != nil {
				break
			}
		}
		var lclProcessed bool
		lclProcessed, err = processAgRequest(
			ka.ctx, reqProcessor, agReq,
			utils.KamailioAgent, ka.connMgr,
			sessConns, ka.fltrS)
		if lclProcessed {
			processed = lclProcessed
		}
		if err != nil ||
			(lclProcessed && !reqProcessor.Flags.GetBool(utils.MetaContinue)) {
			break
		}
	}
	if err == nil && !processed {
		err = errors.New("no request processor enabled")
	}
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %v processing event %s with OriginID: %s",
				utils.KamailioAgent, err, kev[EVENT], kev[utils.OriginID]))
		ka.sendReply(kev, connIdx, rply, err)
		return
	}
	ka.sendReply(kev, connIdx, rply, nil)
}

func (ka *KamailioAgent) sendReply(kev KamEvent, connIdx int,
	rplyNM *utils.OrderedNavigableMap, rplyErr error) {
	var rply map[string]any
	if rplyErr != nil && kev[KamHashEntry] != utils.EmptyString {
		rply = map[string]any{
			KamReplyEvent: utils.FirstNonEmpty(kev[KamReplyRoute], kev[EVENT]),
			KamHashEntry:  kev[KamHashEntry],
			KamHashID:     kev[KamHashID],
		}
		if err := ka.conns[connIdx].Send(utils.ToJSON(rply)); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed sending reply for event: %s, error: %s",
				utils.KamailioAgent, kev[utils.OriginID], err.Error()))
		}
		return
	}
	if rplyNM.Empty() {
		return
	}
	rply = map[string]any{
		KamReplyEvent:   utils.FirstNonEmpty(kev[KamReplyRoute], kev[EVENT]),
		KamReplyTRIndex: kev[KamTRIndex],
		KamReplyTRLabel: kev[KamTRLabel],
	}

	for el := rplyNM.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		itm, _ := rplyNM.Field(path)
		if itm == nil {
			continue
		}
		rply[strings.Join(utils.StripTrailingIndex(path), utils.NestingSep)] = itm.Data
	}
	if rplyErr != nil {
		rply[utils.Error] = rplyErr.Error()
	}
	if err := ka.conns[connIdx].Send(utils.ToJSON(rply)); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed sending reply for event: %s, error: %s",
			utils.KamailioAgent, kev[utils.OriginID], err.Error()))
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
	for _, dlgInfo := range kamDlgRpl.Jsonrpl_body.Result {
		originHost := ka.conns[connIdx].RemoteAddr().String()
		originID := dlgInfo.CallId + ";" + dlgInfo.Caller.Tag
		for _, variable := range dlgInfo.Variables {
			if variable.CgrOriginHost != utils.EmptyString {
				originHost = variable.CgrOriginHost
			}
			if variable.CgrOriginID != utils.EmptyString {
				originID = variable.CgrOriginID
			}
		}
		sIDs = append(sIDs, &sessions.SessionID{
			OriginHost: originHost,
			OriginID:   originID,
		})
	}
	ka.activeSessionIDs <- sIDs
}

func (self *KamailioAgent) disconnectSession(connIdx int, dscEv *KamSessionDisconnect) (err error) {
	if err = self.conns[connIdx].Send(dscEv.String()); err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> failed sending disconnect request: %s,  connection id: %v, error %s",
			utils.KamailioAgent, utils.ToJSON(dscEv), connIdx, err.Error()))
	}
	return
}

// Internal method to disconnect session in Kamailio
func (ka *KamailioAgent) V1DisconnectSession(ctx *context.Context, cgrEv utils.CGREvent, reply *string) (err error) {
	hEntry := utils.IfaceAsString(cgrEv.Event[KamHashEntry])
	hID := utils.IfaceAsString(cgrEv.Event[KamHashID])
	connIdxIface, has := cgrEv.Event[EvapiConnID]
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

// V1GetActiveSessionIDs returns a list of originIDs based on active sessions from agent
func (ka *KamailioAgent) V1GetActiveSessionIDs(ctx *context.Context, _ string, sessionIDs *[]*sessions.SessionID) (err error) {
	kamEv := utils.ToJSON(map[string]string{utils.Event: CGR_DLG_LIST})
	var sentDLG int
	for i, evapi := range ka.conns {
		if err := evapi.Send(kamEv); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> failed sending event to connIdx<%v>, error %s",
				utils.KamailioAgent, i, err.Error()))
			continue
		}
		sentDLG++
	}
	if sentDLG == 0 {
		return
	}
	tm := time.NewTimer(ka.cfg.GeneralCfg().ReplyTimeout)
	for i := 0; i < sentDLG; i++ {
		select {
		case sIDs := <-ka.activeSessionIDs:
			*sessionIDs = append(*sessionIDs, sIDs...)
		case <-tm.C:
			return errors.New("timeout executing dialog list")
		}
	}
	if len(*sessionIDs) == 0 {
		return utils.ErrNoActiveSession
	}
	tm.Stop()
	return
}

// Reload recreates the connection buffers
// only used on reload
func (ka *KamailioAgent) Reload() {
	ka.conns = make([]*kamevapi.KamEvapi, len(ka.kamCfg.EvapiConns))
}

// V1AlterSession is used to implement the sessions.BiRPClient interface
func (*KamailioAgent) V1AlterSession(*context.Context, utils.CGREvent, *string) error {
	return utils.ErrNotImplemented
}

// V1DisconnectPeer is used to implement the sessions.BiRPClient interface
func (*KamailioAgent) V1DisconnectPeer(*context.Context, *utils.DPRArgs, *string) error {
	return utils.ErrNotImplemented
}

// V1WarnDisconnect is used to implement the sessions.BiRPClient interface
func (*KamailioAgent) V1WarnDisconnect(*context.Context, map[string]any, *string) error {
	return utils.ErrNotImplemented
}
