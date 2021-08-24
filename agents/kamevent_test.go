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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var kamEv = KamEvent{KamTRIndex: "29223", KamTRLabel: "698469260",
	"callid": "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ", "from_tag": "eb082607", "to_tag": "4ea9687f", "cgr_account": "dan",
	"cgr_reqtype": utils.MetaPrepaid, "cgr_subject": "dan", "cgr_destination": "+4986517174963", "cgr_tenant": "itsyscom.com",
	"cgr_duration": "20", utils.CGRRoute: "suppl2", utils.CGRDisconnectCause: "200", "extra1": "val1", "extra2": "val2"}

func TestNewKamEvent(t *testing.T) {
	evStr := `{"event":"CGR_CALL_END",
		"callid":"46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag":"bf71ad59",
		"to_tag":"7351fecf",
		"cgr_reqtype":"*postpaid",
		"cgr_account":"1001",
		"cgr_destination":"1002",
		"cgr_answertime":"1419839310",
		"cgr_duration":"3",
		"cgr_route":"supplier2",
		"cgr_disconnectcause": "200",
		"cgr_pdd": "4"}`
	eKamEv := KamEvent{
		"event":                  "CGR_CALL_END",
		"callid":                 "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag":               "bf71ad59",
		"to_tag":                 "7351fecf",
		"cgr_reqtype":            utils.MetaPostpaid,
		"cgr_account":            "1001",
		"cgr_destination":        "1002",
		"cgr_answertime":         "1419839310",
		"cgr_duration":           "3",
		"cgr_pdd":                "4",
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

func TestKamEvMissingParameter(t *testing.T) {
	kamEv = KamEvent{EVENT: CGR_CALL_END,
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_answertime": "1419839310", "cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	if missingParam := kamEv.MissingParameter(); missingParam != true {
		t.Errorf("Expecting: true, received:%+v ", missingParam)
	}
}

func TestKamEvAsMapStringInterface(t *testing.T) {
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	expMp := make(map[string]interface{})
	expMp["cgr_account"] = "1001"
	expMp["cgr_duration"] = "3"
	expMp["cgr_pdd"] = "4"
	expMp["cgr_destination"] = "1002"
	expMp[utils.CGRRoute] = "supplier2"
	expMp["cgr_answertime"] = "1419839310"
	expMp[utils.CGRDisconnectCause] = "200"
	expMp["callid"] = "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0"
	expMp["from_tag"] = "bf71ad59"
	expMp["to_tag"] = "7351fecf"
	expMp["cgr_reqtype"] = utils.MetaPostpaid
	expMp[utils.Source] = utils.KamailioAgent
	expMp[utils.RequestType] = utils.MetaRated
	rcv := kamEv.AsMapStringInterface()
	if !reflect.DeepEqual(expMp, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expMp, rcv)
	}
}

func TestKamEvAsCGREvent(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	expected := &utils.CGREvent{
		Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
			config.CgrConfig().GeneralCfg().DefaultTenant),
		ID:    utils.UUIDSha1Prefix(),
		Event: kamEv.AsMapStringInterface(),
	}
	if rcv := kamEv.AsCGREvent(timezone); !reflect.DeepEqual(expected.Tenant, rcv.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Tenant, rcv.Tenant)
	} else if !reflect.DeepEqual(expected.Event, rcv.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Event, rcv.Event)
	}
}

func TestKamEvV1AuthorizeArgs(t *testing.T) {
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200",
		utils.CGRFlags:           "*accounts;*routes;*routesEventCost;*routesIgnoreErrors"}
	expected := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Event: kamEv.AsMapStringInterface(),
		},
		GetRoutes:          true,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      utils.MetaEventCost,
	}
	rcv := kamEv.V1AuthorizeArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.GetMaxUsage, rcv.GetMaxUsage) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetMaxUsage, rcv.GetMaxUsage)
	} else if !reflect.DeepEqual(expected.GetRoutes, rcv.GetRoutes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetRoutes, rcv.GetRoutes)
	} else if !reflect.DeepEqual(expected.GetAttributes, rcv.GetAttributes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetAttributes, rcv.GetAttributes)
	} else if !reflect.DeepEqual(expected.RoutesMaxCost, rcv.RoutesMaxCost) {
		t.Errorf("Expecting: %+v, received: %+v", expected.RoutesMaxCost, rcv.RoutesMaxCost)
	} else if !reflect.DeepEqual(expected.RoutesIgnoreErrors, rcv.RoutesIgnoreErrors) {
		t.Errorf("Expecting: %+v, received: %+v", expected.RoutesIgnoreErrors, rcv.RoutesIgnoreErrors)
	}
}

func TestKamEvV1AuthorizeArgs2(t *testing.T) {
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200",
		utils.CGRFlags:           "*accounts;*routes;*routesMaxcost:100;*routesIgnoreErrors"}
	expected := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Event: kamEv.AsMapStringInterface(),
		},
		GetRoutes:          true,
		RoutesIgnoreErrors: true,
		RoutesMaxCost:      "100",
	}
	rcv := kamEv.V1AuthorizeArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.GetMaxUsage, rcv.GetMaxUsage) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetMaxUsage, rcv.GetMaxUsage)
	} else if !reflect.DeepEqual(expected.GetRoutes, rcv.GetRoutes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetRoutes, rcv.GetRoutes)
	} else if !reflect.DeepEqual(expected.GetAttributes, rcv.GetAttributes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetAttributes, rcv.GetAttributes)
	} else if !reflect.DeepEqual(expected.RoutesMaxCost, rcv.RoutesMaxCost) {
		t.Errorf("Expecting: %+v, received: %+v", expected.RoutesMaxCost, rcv.RoutesMaxCost)
	} else if !reflect.DeepEqual(expected.RoutesIgnoreErrors, rcv.RoutesIgnoreErrors) {
		t.Errorf("Expecting: %+v, received: %+v", expected.RoutesIgnoreErrors, rcv.RoutesIgnoreErrors)
	}
}

func TestKamEvAsKamAuthReply(t *testing.T) {
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	authArgs := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Event: kamEv.AsMapStringInterface(),
		},
	}
	authRply := &sessions.V1AuthorizeReply{
		MaxUsage: utils.DurationPointer(5 * time.Second),
	}
	expected := &KamReply{
		Event:    CGR_AUTH_REPLY,
		MaxUsage: 5,
	}
	if rcv, err := kamEv.AsKamAuthReply(authArgs, authRply, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	kamEv = KamEvent{"event": "CGR_PROFILE_REQUEST",
		"Tenant": "cgrates.org", "Account": "1001",
		KamReplyRoute: "CGR_PROFILE_REPLY"}
	authArgs = &sessions.V1AuthorizeArgs{
		GetAttributes: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Event: kamEv.AsMapStringInterface(),
		},
	}
	authRply = &sessions.V1AuthorizeReply{
		Attributes: &engine.AttrSProcessEventReply{
			MatchedProfiles: []string{"ATTR_1001_ACCOUNT_PROFILE"},
			AlteredFields:   []string{"*req.Password", utils.MetaReq + utils.NestingSep + utils.RequestType},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestKamEvAsKamAuthReply",
				Event: map[string]interface{}{
					utils.Tenant:       "cgrates.org",
					utils.AccountField: "1001",
					"Password":         "check123",
					utils.RequestType:  utils.MetaPrepaid,
				},
			},
		},
	}
	expected = &KamReply{
		Event:      "CGR_PROFILE_REPLY",
		Attributes: "Password:check123,RequestType:*prepaid",
	}
	if rcv, err := kamEv.AsKamAuthReply(authArgs, authRply, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}

func TestKamEvV1ProcessMessageArgs(t *testing.T) {
	kamEv := KamEvent{"event": "CGR_PROCESS_MESSAGE",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	expected := &sessions.V1ProcessMessageArgs{
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Event: kamEv.AsMapStringInterface(),
		},
	}
	rcv := kamEv.V1ProcessMessageArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	}
}

func TestKamEvAsKamProcessEventReply(t *testing.T) {
	kamEv := KamEvent{"event": "CGR_PROCESS_MESSAGE",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.MetaPostpaid, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGRRoute:           "supplier2",
		utils.CGRDisconnectCause: "200"}
	procEvArgs := &sessions.V1ProcessMessageArgs{
		Debit: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Event: kamEv.AsMapStringInterface(),
		},
	}
	procEvhRply := &sessions.V1ProcessMessageReply{
		MaxUsage: utils.DurationPointer(5 * time.Second),
	}
	expected := &KamReply{
		Event:    CGR_PROCESS_MESSAGE,
		MaxUsage: 5,
	}
	if rcv, err := kamEv.AsKamProcessMessageReply(procEvArgs, procEvhRply, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
	kamEv = KamEvent{"event": "CGR_PROFILE_REQUEST",
		"Tenant": "cgrates.org", "Account": "1001",
		KamReplyRoute: "CGR_PROFILE_REPLY"}
	procEvArgs = &sessions.V1ProcessMessageArgs{
		GetAttributes: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Event: kamEv.AsMapStringInterface(),
		},
	}
	procEvhRply = &sessions.V1ProcessMessageReply{
		Attributes: &engine.AttrSProcessEventReply{
			MatchedProfiles: []string{"ATTR_1001_ACCOUNT_PROFILE"},
			AlteredFields:   []string{"*req.Password", utils.MetaReq + utils.NestingSep + utils.RequestType},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestKamEvAsKamAuthReply",
				Event: map[string]interface{}{
					utils.Tenant:       "cgrates.org",
					utils.AccountField: "1001",
					"Password":         "check123",
					utils.RequestType:  utils.MetaPrepaid,
				},
			},
		},
	}
	expected = &KamReply{
		Event:      "CGR_PROFILE_REPLY",
		Attributes: "Password:check123,RequestType:*prepaid",
	}
	if rcv, err := kamEv.AsKamProcessMessageReply(procEvArgs, procEvhRply, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", expected, rcv)
	}
}
