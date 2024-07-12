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
	"encoding/json"
	"fmt"
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
	"cgr_reqtype": utils.META_PREPAID, "cgr_subject": "dan", "cgr_destination": "+4986517174963", "cgr_tenant": "itsyscom.com",
	"cgr_duration": "20", utils.CGR_SUPPLIER: "suppl2", utils.CGR_DISCONNECT_CAUSE: "200", "extra1": "val1", "extra2": "val2"}

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
		"cgr_supplier":"supplier2",
		"cgr_disconnectcause": "200",
		"cgr_pdd": "4"}`
	eKamEv := KamEvent{
		"event":                    "CGR_CALL_END",
		"callid":                   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag":                 "bf71ad59",
		"to_tag":                   "7351fecf",
		"cgr_reqtype":              utils.META_POSTPAID,
		"cgr_account":              "1001",
		"cgr_destination":          "1002",
		"cgr_answertime":           "1419839310",
		"cgr_duration":             "3",
		"cgr_pdd":                  "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200",
		utils.OriginHost:           utils.KamailioAgent,
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
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_answertime": "1419839310", "cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	if missingParam := kamEv.MissingParameter(); missingParam != true {
		t.Errorf("Expecting: true, received:%+v ", missingParam)
	}
}

func TestKamEvAsMapStringInterface(t *testing.T) {
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	expMp := make(map[string]any)
	expMp["cgr_account"] = "1001"
	expMp["cgr_duration"] = "3"
	expMp["cgr_pdd"] = "4"
	expMp["cgr_destination"] = "1002"
	expMp[utils.CGR_SUPPLIER] = "supplier2"
	expMp["cgr_answertime"] = "1419839310"
	expMp[utils.CGR_DISCONNECT_CAUSE] = "200"
	expMp["callid"] = "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0"
	expMp["from_tag"] = "bf71ad59"
	expMp["to_tag"] = "7351fecf"
	expMp["cgr_reqtype"] = utils.META_POSTPAID
	expMp[utils.Source] = utils.KamailioAgent
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
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &utils.CGREvent{
		Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
			config.CgrConfig().GeneralCfg().DefaultTenant),
		ID:    utils.UUIDSha1Prefix(),
		Time:  &sTime,
		Event: kamEv.AsMapStringInterface(),
	}
	if rcv, err := kamEv.AsCGREvent(timezone); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.Tenant, rcv.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Tenant, rcv.Tenant)
	} else if !reflect.DeepEqual(expected.Time, rcv.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Time, rcv.Time)
	} else if !reflect.DeepEqual(expected.Event, rcv.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Event, rcv.Event)
	}
}

func TestKamEvV1AuthorizeArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200",
		utils.CGRFlags:             "*accounts,*suppliers,*suppliers_event_cost,*suppliers_ignore_errors"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
		GetSuppliers:          true,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaEventCost,
	}
	rcv := kamEv.V1AuthorizeArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.GetMaxUsage, rcv.GetMaxUsage) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetMaxUsage, rcv.GetMaxUsage)
	} else if !reflect.DeepEqual(expected.GetSuppliers, rcv.GetSuppliers) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetSuppliers, rcv.GetSuppliers)
	} else if !reflect.DeepEqual(expected.GetAttributes, rcv.GetAttributes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetAttributes, rcv.GetAttributes)
	} else if !reflect.DeepEqual(expected.SuppliersMaxCost, rcv.SuppliersMaxCost) {
		t.Errorf("Expecting: %+v, received: %+v", expected.SuppliersMaxCost, rcv.SuppliersMaxCost)
	} else if !reflect.DeepEqual(expected.SuppliersIgnoreErrors, rcv.SuppliersIgnoreErrors) {
		t.Errorf("Expecting: %+v, received: %+v", expected.SuppliersIgnoreErrors, rcv.SuppliersIgnoreErrors)
	}
}

func TestKamEvV1AuthorizeArgs2(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200",
		utils.CGRFlags:             "*accounts,*suppliers,*suppliers_maxcost:100,*suppliers_ignore_errors"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
		GetSuppliers:          true,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      "100",
	}
	rcv := kamEv.V1AuthorizeArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.GetMaxUsage, rcv.GetMaxUsage) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetMaxUsage, rcv.GetMaxUsage)
	} else if !reflect.DeepEqual(expected.GetSuppliers, rcv.GetSuppliers) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetSuppliers, rcv.GetSuppliers)
	} else if !reflect.DeepEqual(expected.GetAttributes, rcv.GetAttributes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetAttributes, rcv.GetAttributes)
	} else if !reflect.DeepEqual(expected.SuppliersMaxCost, rcv.SuppliersMaxCost) {
		t.Errorf("Expecting: %+v, received: %+v", expected.SuppliersMaxCost, rcv.SuppliersMaxCost)
	} else if !reflect.DeepEqual(expected.SuppliersIgnoreErrors, rcv.SuppliersIgnoreErrors) {
		t.Errorf("Expecting: %+v, received: %+v", expected.SuppliersIgnoreErrors, rcv.SuppliersIgnoreErrors)
	}
}

func TestKamEvAsKamAuthReply(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	authArgs := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	authRply := &sessions.V1AuthorizeReply{
		MaxUsage: time.Duration(5 * time.Second),
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
			Time:  &sTime,
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
				Event: map[string]any{
					utils.Tenant:      "cgrates.org",
					utils.Account:     "1001",
					"Password":        "check123",
					utils.RequestType: utils.META_PREPAID,
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

func TestKamEvV1InitSessionArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	rcv := kamEv.V1InitSessionArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.InitSession, rcv.InitSession) {
		t.Errorf("Expecting: %+v, received: %+v", expected.InitSession, rcv.InitSession)
	}
}

func TestKamEvV1TerminateSessionArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_CALL_END",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	rcv := kamEv.V1TerminateSessionArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.TerminateSession, rcv.TerminateSession) {
		t.Errorf("Expecting: %+v, received: %+v", expected.TerminateSession, rcv.TerminateSession)
	}
}

func TestKamEvV1ProcessMessageArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_PROCESS_MESSAGE",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	expected := &sessions.V1ProcessMessageArgs{
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	rcv := kamEv.V1ProcessMessageArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	}
}

func TestKamEvAsKamProcessEventReply(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	kamEv := KamEvent{"event": "CGR_PROCESS_MESSAGE",
		"callid":   "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": utils.META_POSTPAID, "cgr_account": "1001",
		"cgr_destination": "1002", "cgr_answertime": "1419839310",
		"cgr_duration": "3", "cgr_pdd": "4",
		utils.CGR_SUPPLIER:         "supplier2",
		utils.CGR_DISCONNECT_CAUSE: "200"}
	sTime, err := utils.ParseTimeDetectLayout(kamEv[utils.AnswerTime], timezone)
	if err != nil {
		return
	}
	procEvArgs := &sessions.V1ProcessMessageArgs{
		Debit: true,
		CGREvent: &utils.CGREvent{
			Tenant: utils.FirstNonEmpty(kamEv[utils.Tenant],
				config.CgrConfig().GeneralCfg().DefaultTenant),
			ID:    utils.UUIDSha1Prefix(),
			Time:  &sTime,
			Event: kamEv.AsMapStringInterface(),
		},
	}
	procEvhRply := &sessions.V1ProcessMessageReply{
		MaxUsage: 5 * time.Second,
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
			Time:  &sTime,
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
				Event: map[string]any{
					utils.Tenant:      "cgrates.org",
					utils.Account:     "1001",
					"Password":        "check123",
					utils.RequestType: utils.META_PREPAID,
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

func TestKamEventNewKamSessionDisconnect(t *testing.T) {
	hEntry := "hEntry"
	hID := "hID"
	reason := "reason"
	disconnectEvent := NewKamSessionDisconnect(hEntry, hID, reason)
	if disconnectEvent.Event != CGR_SESSION_DISCONNECT {
		t.Errorf("Expected Event to be %q, got %q", CGR_SESSION_DISCONNECT, disconnectEvent.Event)
	}
	if disconnectEvent.HashEntry != hEntry {
		t.Errorf("Expected HashEntry to be %q, got %q", hEntry, disconnectEvent.HashEntry)
	}
	if disconnectEvent.HashId != hID {
		t.Errorf("Expected HashId to be %q, got %q", hID, disconnectEvent.HashId)
	}
	if disconnectEvent.Reason != reason {
		t.Errorf("Expected Reason to be %q, got %q", reason, disconnectEvent.Reason)
	}
}

func TestKamEventSessionDisconnectString(t *testing.T) {
	disconnectEvent := &KamSessionDisconnect{
		Event:     CGR_SESSION_DISCONNECT,
		HashEntry: "hEntry",
		HashId:    "hID",
		Reason:    "reason",
	}
	expectedJSON, err := json.Marshal(disconnectEvent)
	if err != nil {
		t.Fatalf("Failed to marshal KamSessionDisconnect to JSON: %v", err)
	}
	result := disconnectEvent.String()
	if result != string(expectedJSON) {
		t.Errorf("Expected String method to return %q, got %q", string(expectedJSON), result)
	}
}

func TestKamEventNewKamEventUnmarshalError(t *testing.T) {
	invalidData := []byte(`{"field1": "value1", "field2": "value2"`)
	_, err := NewKamEvent(invalidData, "alias", "address")
	if err == nil {
		t.Error("Expected an error but got nil")
	}

}

func TestKamEventMissingParameterDefaultCase(t *testing.T) {
	kev := KamEvent{
		EVENT: "unsupportedEvent",
	}
	result := kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for unsupported event, got false")
	}
}

func TestKamEventMissingParameterCgrProcessCDR(t *testing.T) {
	kev := KamEvent{
		EVENT:          CGR_PROCESS_CDR,
		KamTRIndex:     "1",
		KamTRLabel:     "Label",
		utils.OriginID: "origin1",
	}
	result := kev.MissingParameter()
	if result {
		t.Errorf("Expected MissingParameter() to return false for CGR_PROCESS_CDR with valid parameters, got true")
	}
	delete(kev, KamTRIndex)
	result = kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for CGR_PROCESS_CDR with missing KamTRIndex, got false")
	}
	kev = KamEvent{
		EVENT:          CGR_PROCESS_CDR,
		KamTRIndex:     "1",
		utils.OriginID: "origin1",
	}
	result = kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for CGR_PROCESS_CDR with missing KamTRLabel, got false")
	}
	kev = KamEvent{
		EVENT:      CGR_PROCESS_CDR,
		KamTRIndex: "1",
		KamTRLabel: "Label",
	}
	result = kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for CGR_PROCESS_CDR with missing OriginID, got false")
	}
}

func TestKamEventMissingParameterCgrProcessMessage(t *testing.T) {
	kev := KamEvent{
		EVENT:             CGR_PROCESS_MESSAGE,
		KamTRIndex:        "1",
		KamTRLabel:        "Label",
		utils.CGRFlags:    "flags",
		utils.OriginID:    "origin1",
		utils.AnswerTime:  "2024-07-15T12:00:00Z",
		utils.Account:     "user1",
		utils.Destination: "destination",
	}
	result := kev.MissingParameter()
	if result {
		t.Errorf("Expected MissingParameter() to return false for CGR_PROCESS_MESSAGE with valid parameters, got true")
	}
	delete(kev, KamTRIndex)
	result = kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for CGR_PROCESS_MESSAGE with missing KamTRIndex, got false")
	}
	kev = KamEvent{
		EVENT:             CGR_PROCESS_MESSAGE,
		KamTRIndex:        "1",
		utils.CGRFlags:    "flags",
		utils.OriginID:    "origin1",
		utils.AnswerTime:  "2024-07-15T12:00:00Z",
		utils.Account:     "user1",
		utils.Destination: "destination",
	}
	delete(kev, KamTRLabel)
	result = kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for CGR_PROCESS_MESSAGE with missing KamTRLabel, got false")
	}
	kev = KamEvent{
		EVENT:             CGR_PROCESS_MESSAGE,
		KamTRIndex:        "1",
		KamTRLabel:        "Label",
		utils.OriginID:    "origin1",
		utils.AnswerTime:  "2024-07-15T12:00:00Z",
		utils.Account:     "user1",
		utils.Destination: "destination",
	}
	result = kev.MissingParameter()
	if result {
		t.Errorf("false")
	}
}

func TestKamEventMissingParameterCgrCallStart(t *testing.T) {
	kev := KamEvent{
		EVENT:             CGR_CALL_START,
		KamHashEntry:      "hashEntry1",
		KamHashID:         "hashID1",
		utils.OriginID:    "origin1",
		utils.AnswerTime:  "2024-07-15T12:00:00Z",
		utils.Account:     "user1",
		utils.Destination: "destination",
	}
	result := kev.MissingParameter()
	if result {
		t.Errorf("Expected MissingParameter() to return false for CGR_CALL_START with valid parameters, got true")
	}
	delete(kev, KamHashEntry)
	result = kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for CGR_CALL_START with missing KamHashEntry, got false")
	}
	kev = KamEvent{
		EVENT:             CGR_CALL_START,
		KamHashEntry:      "hashEntry1",
		utils.OriginID:    "origin1",
		utils.AnswerTime:  "2024-07-15T12:00:00Z",
		utils.Account:     "user1",
		utils.Destination: "destination",
	}
	delete(kev, KamHashID)
	result = kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for CGR_CALL_START with missing KamHashID, got false")
	}
	kev = KamEvent{
		EVENT:             CGR_CALL_START,
		KamHashEntry:      "hashEntry1",
		KamHashID:         "hashID1",
		utils.AnswerTime:  "2024-07-15T12:00:00Z",
		utils.Account:     "user1",
		utils.Destination: "destination",
	}
	result = kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for CGR_CALL_START without setting utils.OriginID, got false")
	}
}

func TestKamEventMissingParameterCgrAuthRequest(t *testing.T) {
	kev := KamEvent{
		EVENT:      CGR_AUTH_REQUEST,
		KamTRIndex: "index1",
		KamTRLabel: "label1",
	}
	result := kev.MissingParameter()
	if result {
		t.Errorf("Expected MissingParameter() to return false for CGR_AUTH_REQUEST with valid parameters, got true")
	}
	delete(kev, KamTRIndex)
	result = kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for CGR_AUTH_REQUEST with missing KamTRIndex, got false")
	}
	kev = KamEvent{
		EVENT:      CGR_AUTH_REQUEST,
		KamTRIndex: "index1",
	}
	delete(kev, KamTRLabel)
	result = kev.MissingParameter()
	if !result {
		t.Errorf("Expected MissingParameter() to return true for CGR_AUTH_REQUEST with missing KamTRLabel, got false")
	}
}

func TestKamEventString(t *testing.T) {
	kev := KamEvent{
		"event":      "CGR_CALL_START",
		"KamTRIndex": "1",
		"KamTRLabel": "Label",
	}
	expectedJSON := `{"KamTRIndex":"1","KamTRLabel":"Label","event":"CGR_CALL_START"}`
	result := kev.String()
	if result != expectedJSON {
		t.Errorf("String() method did not produce expected JSON. Expected: %s, Got: %s", expectedJSON, result)
	}
	emptyKev := KamEvent{}
	emptyJSON := emptyKev.String()
	if emptyJSON != "{}" {
		t.Errorf("String() method did not produce expected empty JSON. Expected: {}, Got: %s", emptyJSON)
	}
	var unmarshaled KamEvent
	err := json.Unmarshal([]byte(result), &unmarshaled)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v", err)
	}
	if !reflect.DeepEqual(kev, unmarshaled) {
		t.Errorf("Unmarshaled KamEvent does not match original. Expected: %v, Got: %v", kev, unmarshaled)
	}
}

func TestKamEventAsKamProcessCDRReply(t *testing.T) {
	kev := KamEvent{
		EVENT:         CGR_PROCESS_CDR,
		KamTRIndex:    "1",
		KamTRLabel:    "label1",
		KamReplyRoute: CGR_PROCESS_CDR,
	}
	var rply string
	var rplyErr error
	kar, err := kev.AsKamProcessCDRReply(nil, &rply, rplyErr)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if kar == nil {
		t.Fatalf("Expected non-nil KamReply, got nil")
	}
	if kar.Event != CGR_PROCESS_CDR {
		t.Errorf("Unexpected Event name in KamReply: expected %s, got %s", CGR_PROCESS_CDR, kar.Event)
	}
	if kar.TransactionIndex != "1" {
		t.Errorf("Unexpected TransactionIndex in KamReply: expected %s, got %s", "1", kar.TransactionIndex)
	}
	if kar.TransactionLabel != "label1" {
		t.Errorf("Unexpected TransactionLabel in KamReply: expected %s, got %s", "label1", kar.TransactionLabel)
	}
	rplyErr = fmt.Errorf("reply error")
	kar, err = kev.AsKamProcessCDRReply(nil, &rply, rplyErr)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if kar == nil {
		t.Fatalf("Expected non-nil KamReply, got nil")
	}
	if kar.Error != "reply error" {
		t.Errorf("Unexpected Error field in KamReply: expected %s, got %s", "reply error", kar.Error)
	}
}

func TestKamEventAsKamProcessMessageEmptyReply(t *testing.T) {
	kev := KamEvent{
		EVENT:         CGR_PROCESS_MESSAGE,
		KamTRIndex:    "1",
		KamTRLabel:    "label",
		KamReplyRoute: CGR_PROCESS_MESSAGE,
	}
	kar := kev.AsKamProcessMessageEmptyReply()
	if kar == nil {
		t.Fatalf("Expected non-nil KamReply, got nil")
	}
	if kar.Event != CGR_PROCESS_MESSAGE {
		t.Errorf("Unexpected Event name in KamReply: expected %s, got %s", CGR_PROCESS_MESSAGE, kar.Event)
	}
	if kar.TransactionIndex != "1" {
		t.Errorf("Unexpected TransactionIndex in KamReply: expected %s, got %s", "54321", kar.TransactionIndex)
	}
	if kar.TransactionLabel != "label" {
		t.Errorf("Unexpected TransactionLabel in KamReply: expected %s, got %s", "label321", kar.TransactionLabel)
	}
}

func TestKamEventKamReplyString(t *testing.T) {
	reply := KamReply{
		Event:            CGR_PROCESS_CDR,
		TransactionIndex: "1",
		TransactionLabel: "label",
	}
	replyStr := reply.String()
	var parsedReply map[string]interface{}
	err := json.Unmarshal([]byte(replyStr), &parsedReply)
	if err != nil {
		t.Fatalf("Error unmarshalling KamReply string: %v", err)
	}
	if parsedReply["Event"] != CGR_PROCESS_CDR {
		t.Errorf("Unexpected Event in parsed KamReply: expected %s, got %v", CGR_PROCESS_CDR, parsedReply["Event"])
	}
	if parsedReply["TransactionIndex"] != "1" {
		t.Errorf("Unexpected TransactionIndex in parsed KamReply: expected %s, got %v", "1", parsedReply["TransactionIndex"])
	}
	if parsedReply["TransactionLabel"] != "label" {
		t.Errorf("Unexpected TransactionLabel in parsed KamReply: expected %s, got %v", "label", parsedReply["TransactionLabel"])
	}
}

func TestKamDlgReplyString(t *testing.T) {
	kdr := &KamDlgReply{}
	expectedJSON := `{"Event":"","Jsonrpl_body":null}`
	result := kdr.String()
	if result != expectedJSON {
		t.Errorf("Expected %s, but got %s", expectedJSON, result)
	}
}

func TestNewKamDlgReply(t *testing.T) {
	validJSON := []byte(`{"Field1":"test","Field2":123}`)
	expectedReply := KamDlgReply{}
	rpl, err := NewKamDlgReply(validJSON)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if rpl != expectedReply {
		t.Errorf("Expected %+v, but got %+v", expectedReply, rpl)
	}
	invalidJSON := []byte(`{"Field1":"test","Field2":}`)
	_, err = NewKamDlgReply(invalidJSON)
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
	emptyJSON := []byte(`{}`)
	expectedEmptyReply := KamDlgReply{}
	rpl, err = NewKamDlgReply(emptyJSON)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	if rpl != expectedEmptyReply {
		t.Errorf("Expected %+v, but got %+v", expectedEmptyReply, rpl)
	}
}

func TestAsKamAuthReplyProcessStats(t *testing.T) {
	kamEvData := KamEvent{}
	authArgs := &sessions.V1AuthorizeArgs{
		ProcessStats: true,
	}
	statQueueIDs := []string{"queue1", "queue2", "queue3"}
	authReply := &sessions.V1AuthorizeReply{
		StatQueueIDs: &statQueueIDs,
	}
	kar, err := kamEvData.AsKamAuthReply(authArgs, authReply, nil)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	expectedStatQueues := "queue1,queue2,queue3"
	if kar.StatQueues != expectedStatQueues {
		t.Errorf("Expected StatQueues to be %s, but got %s", expectedStatQueues, kar.StatQueues)
	}
}

func TestAsKamAuthReplyProcessThresholds(t *testing.T) {
	kamEvData := KamEvent{
		KamTRIndex: "index123",
		KamTRLabel: "label123",
	}
	authArgs := &sessions.V1AuthorizeArgs{
		ProcessThresholds: true,
	}
	thresholdIDs := []string{"threshold1", "threshold2", "threshold3"}
	authReply := &sessions.V1AuthorizeReply{
		ThresholdIDs: &thresholdIDs,
	}
	kar, err := kamEvData.AsKamAuthReply(authArgs, authReply, nil)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	expectedThresholds := "threshold1,threshold2,threshold3"
	if kar.Thresholds != expectedThresholds {
		t.Errorf("Expected Thresholds to be %s, but got %s", expectedThresholds, kar.Thresholds)
	}
}

func TestV1AuthorizeArgsParseFlags(t *testing.T) {
	kev := make(KamEvent)
	args := kev.V1AuthorizeArgs()
	if !args.GetMaxUsage {
		t.Error("Expected GetMaxUsage to be true when CGRFlags is not present")
	}

}
