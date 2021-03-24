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
		val := el.Value
		var itm *utils.DataLeaf
		if itm, err = navMp.Field(val); err != nil {
			return
		}
		if itm == nil {
			continue // all attributes, not writable to diameter packet
		}
		m[strings.Join(itm.Path, utils.NestingSep)] = utils.IfaceAsString(itm.Data)
	}
	return
}

func sipErr(m utils.DataProvider, sipMessage sipingo.Message,
	reqVars *utils.DataNode,
	tpl []*config.FCTemplate, tnt, tmz string,
	filterS *engine.FilterS) (a sipingo.Message, err error) {
	aReq := NewAgentRequest(
		m, reqVars,
		nil, nil, nil, nil,
		tnt, tmz, filterS, nil, nil)
	if err = aReq.SetFields(tpl); err != nil {
		return
	}
	if err = updateSIPMsgFromNavMap(sipMessage, aReq.Reply); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s encoding out %s",
				utils.SIPAgent, err.Error(), utils.ToJSON(aReq.Reply)))
		return
	}
	sipMessage.PrepareReply()
	return sipMessage, nil
}

func bareSipErr(m sipingo.Message, err string) sipingo.Message {
	m[requestHeader] = err
	m.PrepareReply()
	return m
}
