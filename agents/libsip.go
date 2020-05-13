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
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/sipd"
)

// updateSIPMsgFromNavMap will update the diameter message with items from navigable map
func updateSIPMsgFromNavMap(m sipd.Message, navMp *utils.OrderedNavigableMap) (err error) {
	// write reply into message
	for el := navMp.GetFirstElement(); el != nil; el = el.Next() {
		val := el.Value
		var nmIt utils.NMInterface
		if nmIt, err = navMp.Field(val); err != nil {
			return
		}
		itm, isNMItem := nmIt.(*config.NMItem)
		if !isNMItem {
			return fmt.Errorf("cannot encode reply value: %s, err: not NMItems", utils.ToJSON(val))
		}
		if itm == nil {
			continue // all attributes, not writable to diameter packet
		}
		m[strings.Join(itm.Path, utils.NestingSep)] = utils.IfaceAsString(itm.Data)
	}
	return
}
