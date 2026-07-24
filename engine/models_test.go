// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestModelsSetAccountActionId(t *testing.T) {
	aa := &TpAccountAction{}

	err := aa.SetAccountActionId(str)

	if err != nil {
		if err.Error() != "Wrong TP Account Action Id: test" {
			t.Error(err)
		}
	}

	err = aa.SetAccountActionId("test:test:test")
	if err != nil {
		t.Error(err)
	}

	if aa.Loadid != str && aa.Tenant != str && aa.Account != str {
		t.Error("didn't set account action id")
	}
}

func TestModelsGetAccountActionId(t *testing.T) {
	aa := &TpAccountAction{}

	err = aa.SetAccountActionId("test:test:test")
	if err != nil {
		t.Error(err)
	}

	rcv := aa.GetAccountActionId()
	exp := "test:test"

	if rcv != exp {
		t.Errorf("received %s, expected %s", rcv, exp)
	}
}

func TestModelsTableName(t *testing.T) {
	tn := CDRsql{}

	rcv := tn.TableName()

	if rcv != utils.CDRsTBL {
		t.Error(rcv)
	}
}

func TestModelsAsMapStringInterface(t *testing.T) {
	tn := CDRsql{
		ID:          1,
		Cgrid:       str,
		RunID:       str,
		OriginHost:  str,
		Source:      str,
		OriginID:    str,
		TOR:         str,
		RequestType: str,
		Tenant:      str,
		Category:    str,
		Account:     str,
		Subject:     str,
		Destination: str,
		SetupTime:   tm,
		AnswerTime:  tm,
		Usage:       1,
		ExtraFields: str,
		CostSource:  str,
		Cost:        fl,
		CostDetails: str,
		ExtraInfo:   str,
		CreatedAt:   tm,
		UpdatedAt:   tm,
		DeletedAt:   &tm,
	}

	rcv := tn.AsMapStringInterface()
	exp := make(map[string]any)
	exp["cgrid"] = tn.Cgrid
	exp["run_id"] = tn.RunID
	exp["origin_host"] = tn.OriginHost
	exp["source"] = tn.Source
	exp["origin_id"] = tn.OriginID
	exp["tor"] = tn.TOR
	exp["request_type"] = tn.RequestType
	exp["tenant"] = tn.Tenant
	exp["category"] = tn.Category
	exp["account"] = tn.Account
	exp["subject"] = tn.Subject
	exp["destination"] = tn.Destination
	exp["setup_time"] = tn.SetupTime
	exp["answer_time"] = tn.AnswerTime
	exp["usage"] = tn.Usage
	exp["extra_fields"] = tn.ExtraFields
	exp["cost_source"] = tn.CostSource
	exp["cost"] = tn.Cost
	exp["cost_details"] = tn.CostDetails
	exp["extra_info"] = tn.ExtraInfo
	exp["created_at"] = tn.CreatedAt
	exp["updated_at"] = tn.UpdatedAt

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("received %s, expected %s", rcv, exp)
	}
}

func TestModelsSessionCostsSQLTableName(t *testing.T) {
	tn := SessionCostsSQL{}

	rcv := tn.TableName()

	if rcv != utils.SessionCostsTBL {
		t.Error(rcv)
	}
}

func TestModelsTBLVersionTableName(t *testing.T) {
	tn := TBLVersion{}

	rcv := tn.TableName()

	if rcv != utils.TBLVersions {
		t.Error(rcv)
	}
}
