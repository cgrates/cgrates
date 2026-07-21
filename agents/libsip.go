// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/sipingo"
)

// updateSIPMsgFromNavMap will update the diameter message with items from navigable map
func updateSIPMsgFromNavMap(m sipingo.Message, navMp *utils.OrderedNavigableMap) (err error) {
	// write reply into message
	for el := navMp.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		itm, _ := navMp.Field(path)
		if itm == nil {
			continue // all attributes, not writable to diameter packet
		}
		path = utils.StripTrailingIndex(path)
		m[strings.Join(path, utils.NestingSep)] = utils.IfaceAsString(itm.Data)
	}
	return
}

func sipErr(m utils.DataProvider, sipMessage sipingo.Message,
	reqVars *utils.DataNode,
	tpl []*config.FCTemplate, tnt, tmz string,
	cfg *config.CGRConfig, cache *engine.CacheS, filterS *engine.FilterS) (sipingo.Message, error) {
	aReq := NewAgentRequest(
		m, reqVars,
		nil, nil, nil, nil,
		tnt, tmz, cfg, cache, filterS, nil)
	if err := aReq.SetFields(tpl); err != nil {
		return nil, err
	}
	if err := updateSIPMsgFromNavMap(sipMessage, aReq.Reply); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s encoding out %s",
				utils.SIPAgent, err.Error(), utils.ToJSON(aReq.Reply)))
		return nil, err
	}
	sipMessage.PrepareReply()
	return sipMessage, nil
}

func bareSipErr(m sipingo.Message, err string) sipingo.Message {
	m[requestHeader] = err
	m.PrepareReply()
	return m
}
