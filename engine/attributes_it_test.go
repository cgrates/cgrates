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

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestAttrSProcessEventReplyDigest(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{"Account", "Subject"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaRating),
			Event: map[string]interface{}{
				"Account":     "1001",
				"Subject":     "1001",
				"Destination": "+491511231234",
			},
		},
	}
	expRpl := "Account:1001,Subject:1001"
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}

func TestAttrSProcessEventReplyDigest2(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaRating),
			Event: map[string]interface{}{
				"Account":     "1001",
				"Subject":     "1001",
				"Destination": "+491511231234",
			},
		},
	}
	expRpl := ""
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}

func TestAttrSProcessEventReplyDigest3(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{"Subject"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaRating),
			Event: map[string]interface{}{
				"Account":     "1001",
				"Subject":     "1001",
				"Destination": "+491511231234",
			},
		},
	}
	expRpl := "Subject:1001"
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}
