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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cgrates/aringo"
	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// constants used by AsteriskAgent
const (
	CGRAuthAPP            = "cgratesAuth"
	CGRMaxSessionTime     = "CGRMaxSessionTime"
	CGRRoute              = "CGRRoute"
	ARIStasisStart        = "StasisStart"
	ARIChannelStateChange = "ChannelStateChange"
	ARIChannelDestroyed   = "ChannelDestroyed"
	ARIRESTResponse       = "RESTResponse"
)

// NewAsteriskAgent constructs a new Asterisk Agent
func NewAsteriskAgent(cgrCfg *config.CGRConfig, astConnIdx int,
	connMgr *engine.ConnManager, caps *engine.Caps, fltrS *engine.FilterS) (*AsteriskAgent, error) {
	sma := &AsteriskAgent{
		cgrCfg:     cgrCfg,
		astConnIdx: astConnIdx,
		connMgr:    connMgr,
		caps:       caps,
		fltrS:      fltrS,
	}
	srv, _ := birpc.NewService(sma, "", false)
	sma.ctx = context.WithClient(context.TODO(), srv)
	msgTemplates := cgrCfg.TemplatesCfg()
	for _, procsr := range cgrCfg.AsteriskAgentCfg().RequestProcessors {
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
	return sma, nil
}

// ARIConnector abstracts the transport layer (HTTP or WebSocket) for sending ARI commands.
type ARIConnector interface {
	Call(method, uri string, queryStr map[string]string, body []byte) (aringo.RESTResponse, error)
}

// AsteriskAgent used to cominicate with asterisk
type AsteriskAgent struct {
	cgrCfg     *config.CGRConfig // Separate from smCfg since there can be multiple
	connMgr    *engine.ConnManager
	caps       *engine.Caps
	fltrS      *engine.FilterS
	astConnIdx int
	astConn    ARIConnector
	astEvChan  chan map[string]any
	astErrChan chan error
	ctx        *context.Context
}

func (sma *AsteriskAgent) connectAsterisk(stopChan <-chan struct{}) (err error) {
	connCfg := sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx]
	sma.astEvChan = make(chan map[string]any)
	sma.astErrChan = make(chan error)
	if connCfg.AriWebSocket {
		sma.astConn, err = aringo.NewARInGO(fmt.Sprintf("ws://%s/ari/events?api_key=%s:%s&app=%s&subscribeAll=true",
			connCfg.Address, connCfg.User, connCfg.Password, CGRAuthAPP), "http://cgrates.org",
			connCfg.User, connCfg.Password, fmt.Sprintf("%s@%s", utils.CGRateS, utils.Version),
			sma.astEvChan, sma.astErrChan, stopChan, connCfg.ConnectAttempts, connCfg.Reconnects,
			connCfg.MaxReconnectInterval, utils.FibDuration)
	} else {
		sma.astConn, err = aringo.NewARInGOV1(fmt.Sprintf("ws://%s/ari/events?api_key=%s:%s&app=%s&subscribeAll=true",
			connCfg.Address, connCfg.User, connCfg.Password, CGRAuthAPP), "http://cgrates.org",
			connCfg.User, connCfg.Password, connCfg.Address, fmt.Sprintf("%s@%s", utils.CGRateS, utils.Version),
			sma.astEvChan, sma.astErrChan, stopChan, connCfg.ConnectAttempts, connCfg.Reconnects,
			connCfg.MaxReconnectInterval, utils.FibDuration)
	}
	return
}

func (sma *AsteriskAgent) filterEventTypes() {
	filterEv := struct {
		Allowed []map[string]string `json:"allowed"`
	}{
		Allowed: []map[string]string{
			{"type": ARIStasisStart},
			{"type": ARIChannelStateChange},
			{"type": ARIChannelDestroyed},
			{"type": ARIRESTResponse},
		}}
	body, err := json.Marshal(filterEv)
	if err != nil {
		utils.Logger.Warning(err.Error())
		return
	}
	if _, err := sma.astConn.Call(aringo.HTTP_PUT,
		fmt.Sprintf("applications/%s/eventFilter", CGRAuthAPP), nil, body); err != nil {
		utils.Logger.Warning(err.Error())
	}
}

// ListenAndServe is called to start the service
func (sma *AsteriskAgent) ListenAndServe(stopChan <-chan struct{}) (err error) {
	if err = sma.connectAsterisk(stopChan); err != nil {
		return
	}
	sma.filterEventTypes()
	utils.Logger.Info(fmt.Sprintf("<%s> successfully connected to Asterisk at: <%s>",
		utils.AsteriskAgent, sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address))
	for {
		select {
		case <-stopChan:
		case err = <-sma.astErrChan:
			return
		case astRawEv := <-sma.astEvChan:
			ev := NewSMAsteriskEvent(astRawEv)
			go sma.handleSMAsteriskEvent(ev)

		}
	}
}
func (sma *AsteriskAgent) handleSMAsteriskEvent(ev *SMAsteriskEvent) {
	if sma.caps.IsLimited() {
		if err := sma.caps.Allocate(); err != nil {
			sma.hangupChannel(ev.ChannelID(),
				fmt.Sprintf("<%s> caps limit reached, rejecting %s for channel %s: %v",
					utils.AsteriskAgent, ev.EventType(), ev.ChannelID(), err))
			return
		}
		defer sma.caps.Deallocate()
	}
	evType := ev.EventType()
	chID := ev.ChannelID()
	reqVars := &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			utils.OriginHost: utils.NewLeafNode(
				strings.Split(sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address, ":")[0]),
		},
	}
	cgrRplyNM := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	opts := utils.MapStorage{}
	rply := utils.NewOrderedNavigableMap()
	sessConns, _ := engine.GetConnIDs(sma.ctx, sma.cgrCfg.AsteriskAgentCfg().Conns, utils.MetaSessionS,
		sma.cgrCfg.GeneralCfg().DefaultTenant, ev, nil, sma.fltrS)

	var processed bool
	var err error
	for _, reqProcessor := range sma.cgrCfg.AsteriskAgentCfg().RequestProcessors {
		var lclProcessed bool
		lclProcessed, err = processRequest(
			sma.ctx, reqProcessor,
			NewAgentRequest(
				ev, reqVars, cgrRplyNM, rply, opts,
				reqProcessor.Tenant, sma.cgrCfg.GeneralCfg().DefaultTenant,
				utils.FirstNonEmpty(reqProcessor.Timezone,
					sma.cgrCfg.GeneralCfg().DefaultTimezone),
				sma.cgrCfg, nil, sma.fltrS, nil),
			utils.AsteriskAgent, sma.connMgr,
			sessConns, nil, nil, sma.fltrS)
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
			fmt.Sprintf("<%s> error: %v processing event %s for channelID: %s",
				utils.AsteriskAgent, err, evType, chID))
		sma.dispatchErr(chID, evType)
		return
	}
	if !processed {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> no request processor enabled, ignoring event %s for channelID: %s",
				utils.AsteriskAgent, evType, chID))
		sma.dispatchErr(chID, evType)
		return
	}
	if err = sma.applyReplyFields(chID, rply); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %v applying reply variables for event %s for channelID: %s",
				utils.AsteriskAgent, err, evType, chID))
		sma.dispatchErr(chID, evType)
		return
	}
	if evType == ARIStasisStart {
		sma.authorizeStasis(chID, cgrRplyNM)
	}
}

func (sma *AsteriskAgent) dispatchErr(chID, evType string) {
	switch evType {
	case ARIStasisStart, ARIChannelStateChange:
		sma.hangupChannel(chID, "")
	}
}

func (sma *AsteriskAgent) authorizeStasis(chID string, cgrRply *utils.DataNode) {
	maxDur, hasUsage := minAccountUsage(cgrRply)
	if hasUsage && maxDur == 0 {
		sma.hangupChannel(chID, "")
		return
	}
	if hasUsage {
		if !sma.setChannelVar(chID, CGRMaxSessionTime,
			strconv.Itoa(int(maxDur.Milliseconds()))) {
			return
		}
	}
	sma.continueChannel(chID)
}

// setChannelVar will set the value of a variable
func (sma *AsteriskAgent) setChannelVar(chanID string, vrblName, vrblVal string) (success bool) {
	if _, err := sma.astConn.Call(aringo.HTTP_POST,
		fmt.Sprintf("channels/%s/variable", chanID), // Asterisk having issue with variable terminating empty so harcoding param in url
		map[string]string{"variable": vrblName, "value": vrblVal}, nil); err != nil {
		// Since we got error, disconnect channel
		sma.hangupChannel(chanID,
			fmt.Sprintf("<%s> error: <%s> setting <%s> for channelID: <%s>",
				utils.AsteriskAgent, err.Error(), vrblName, chanID))
		return
	}
	return true
}

// hangupChannel will disconnect from CGRateS side with congestion reason
func (sma *AsteriskAgent) hangupChannel(channelID, warnMsg string) {
	if warnMsg != "" {
		utils.Logger.Warning(warnMsg)
	}
	if _, err := sma.astConn.Call(aringo.HTTP_DELETE, fmt.Sprintf("channels/%s", channelID),
		map[string]string{"reason": "congestion"}, nil); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> failed disconnecting channel <%s>, err: %s",
				utils.AsteriskAgent, channelID, err.Error()))
	}
}

func (sma *AsteriskAgent) continueChannel(chID string) {
	if _, err := sma.astConn.Call(aringo.HTTP_POST,
		fmt.Sprintf("channels/%s/continue", chID), nil, nil); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s exiting stasis for channelID: %s",
				utils.AsteriskAgent, err.Error(), chID))
	}
}

// V1DisconnectSession is internal method to disconnect session in asterisk
func (sma *AsteriskAgent) V1DisconnectSession(_ *context.Context, cgrEv utils.CGREvent, reply *string) error {
	channelID := engine.NewMapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.OriginID)
	sma.hangupChannel(channelID, "")
	*reply = utils.OK
	return nil
}

// V1GetActiveSessionIDs is internal method to  get all active sessions in asterisk
func (sma *AsteriskAgent) V1GetActiveSessionIDs(_ *context.Context, _ string,
	sessionIDs *[]*sessions.SessionID) error {
	var slMpIface []map[string]any // decode the result from ari into a slice of map[string]any
	restResp, err := sma.astConn.Call(aringo.HTTP_GET, "channels", nil, nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(restResp.MessageBody), &slMpIface); err != nil {
		return err
	}
	var sIDs []*sessions.SessionID
	if len(slMpIface) == 0 {
		return utils.ErrNoActiveSession
	}
	for _, mpIface := range slMpIface {
		sIDs = append(sIDs, &sessions.SessionID{
			OriginHost: strings.Split(sma.cgrCfg.AsteriskAgentCfg().AsteriskConns[sma.astConnIdx].Address, ":")[0],
			OriginID:   utils.IfaceAsString(mpIface["id"]),
		})
	}
	*sessionIDs = sIDs
	return nil

}

// V1AlterSession is used to implement the sessions.BiRPClient interface
func (*AsteriskAgent) V1AlterSession(*context.Context, utils.CGREvent, *string) error {
	return utils.ErrNotImplemented
}

// V1DisconnectPeer is used to implement the sessions.BiRPClient interface
func (*AsteriskAgent) V1DisconnectPeer(*context.Context, *utils.DPRArgs, *string) error {
	return utils.ErrNotImplemented
}

// V1WarnDisconnect is used to implement the sessions.BiRPClient interface
func (sma *AsteriskAgent) V1WarnDisconnect(*context.Context, map[string]any, *string) error {
	return utils.ErrNotImplemented
}

// applyReplyFields runs the processor's ReplyFields template against the agent
func (sma *AsteriskAgent) applyReplyFields(chID string, rply *utils.OrderedNavigableMap) error {
	for el := rply.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		itm, _ := rply.Field(path)
		if itm == nil {
			continue
		}
		val := utils.IfaceAsString(itm.Data)
		if val == utils.EmptyString {
			continue
		}
		vrbl := strings.Join(utils.StripTrailingIndex(path), utils.NestingSep)
		if !sma.setChannelVar(chID, vrbl, val) {
			return fmt.Errorf("setChannelVar failed for %s", vrbl)
		}
	}
	return nil
}
