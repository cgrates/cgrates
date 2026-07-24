//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Sample HttpJsonPost, more for usage purposes
func TestHttpJsonPost(t *testing.T) {
	cdrOut := &ExternalCDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderID: 123,
		ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Tenant: "cgrates.org",
		Category: "call", Account: "account1", Subject: "tgooiscs0014", Destination: "1002",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String(), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String(),
		RunID: utils.MetaDefault,
		Usage: "0.00000001", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	jsn, _ := json.Marshal(cdrOut)
	if _, err := HttpJsonPost("http://localhost:8000", false, jsn); err == nil {
		t.Error(err)
	}
}
