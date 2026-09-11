// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNewKamEvent(t *testing.T) {
	evStr := `{"event":"CGR_CALL_END",
		"callid":"46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag":"bf71ad59",
		"to_tag":"7351fecf",
		"cgrReqtype":"*postpaid",
		"cgrAccount":"1001",
		"cgrDestination":"1002",
		"cgrAnswertime":"1419839310",
		"cgrDuration":"3",
		"cgrRoute":"supplier2",
		"cgrDisconnectCause": "200",
		"cgrPdd": "4"}`
	eKamEv := KamEvent{
		"event":                  "CGR_CALL_END",
		"callid":                 "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag":               "bf71ad59",
		"to_tag":                 "7351fecf",
		"cgrReqtype":             utils.MetaPostpaid,
		"cgrAccount":             "1001",
		"cgrDestination":         "1002",
		"cgrAnswertime":          "1419839310",
		"cgrDuration":            "3",
		"cgrPdd":                 "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200",
		utils.OriginHost:         utils.KamailioAgent,
	}
	if kamEv, err := NewKamEvent([]byte(evStr), utils.KamailioAgent, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eKamEv, kamEv) {
		t.Error("Received: ", kamEv)
	}
}
