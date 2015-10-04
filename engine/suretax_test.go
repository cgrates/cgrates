/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewSureTaxRequest(t *testing.T) {
	storedCdr := &StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(12) * time.Second, Pdd: time.Duration(7) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans", Rated: true,
	}
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.SureTaxCfg.ClientNumber = "000000000"
	cfg.SureTaxCfg.ValidationKey = "19491161-F004-4F44-BDB3-E976D6739A64"
	cfg.SureTaxCfg.Timezone = time.UTC
	eSureTaxRequest := &SureTaxRequest{
		ClientNumber:   cfg.SureTaxCfg.ClientNumber,
		ValidationKey:  cfg.SureTaxCfg.ValidationKey,
		DataYear:       "2013",
		DataMonth:      "11",
		TotalRevenue:   1.01,
		ReturnFileCode: "0",
		ClientTracking: storedCdr.CgrId,
		ResponseGroup:  "03",
		ResponseType:   "",
		ItemList: []*STRequestItem{
			&STRequestItem{
				OrigNumber:           "1001",
				TermNumber:           "1002",
				BillToNumber:         "1001",
				TransDate:            "2013-11-07T08:42:26",
				Revenue:              1.01,
				Units:                1,
				UnitType:             "00",
				Seconds:              12,
				TaxIncludedCode:      "0",
				TaxSitusRule:         "1",
				TransTypeCode:        "010101",
				SalesTypeCode:        "R",
				RegulatoryCode:       "01",
				TaxExemptionCodeList: []string{"00"},
			},
		},
	}
	if stReq, err := NewSureTaxRequest(cfg.SureTaxCfg.ClientNumber, cfg.SureTaxCfg.ValidationKey, cfg.SureTaxCfg.Timezone, cfg.SureTaxCfg.OriginationNumber,
		cfg.SureTaxCfg.TerminationNumber, storedCdr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSureTaxRequest, stReq) {
		t.Errorf("Expecting: %+v, received: %+v", eSureTaxRequest.ItemList[0], stReq.ItemList[0])
	}
}
