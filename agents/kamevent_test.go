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

func TestKamEventSetCGRRequest(t *testing.T) {
	kev := KamEvent{
		"trIndex":       "4",
		"trLabel":       "label1",
		"replyRoute":    "CGR_AUTH_REPLY",
		"*interimUsage": "50",
		"*totalUsage":   "100",
		"*originID":     "kamailioOrigin",
	}

	nm := utils.NewOrderedNavigableMap()
	opts := make(utils.MapStorage)
	if err := kev.setCGRRequest(nm, opts); err != nil {
		t.Fatal(err)
	}
	eEv := map[string]any{
		"trIndex":    "4",
		"trLabel":    "label1",
		"replyRoute": "CGR_AUTH_REPLY",
	}
	if ev := nm.AsMap(); !reflect.DeepEqual(eEv, ev) {
		t.Errorf("expected  %s, received %s", utils.ToJSON(eEv), utils.ToJSON(ev))
	}
	eopts := utils.MapStorage{
		"*interimUsage": "50",
		"*totalUsage":   "100",
		"*originID":     "kamailioOrigin",
	}
	if !reflect.DeepEqual(eopts, opts) {
		t.Errorf("expected %s, received %s", utils.ToJSON(eopts), utils.ToJSON(opts))
	}
}
