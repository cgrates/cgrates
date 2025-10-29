/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package engine

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewSureTaxRequest(t *testing.T) {
	CGRID := utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String())
	cdr := &CDR{CGRID: CGRID, OrderID: 123, ToR: utils.VOICE,
		OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001",
		Subject: "1001", Destination: "1002",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(12) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01, PreRated: true,
	}
	cfg, _ := config.NewDefaultCGRConfig()
	stCfg := cfg.SureTaxCfg()
	stCfg.ClientNumber = "000000000"
	stCfg.ValidationKey = "19491161-F004-4F44-BDB3-E976D6739A64"
	stCfg.Timezone = time.UTC
	eSTRequest := &STRequest{
		ClientNumber:   "000000000",
		ValidationKey:  "19491161-F004-4F44-BDB3-E976D6739A64",
		DataYear:       "2013",
		DataMonth:      "11",
		TotalRevenue:   1.01,
		ReturnFileCode: "0",
		ClientTracking: CGRID,
		ResponseGroup:  "03",
		ResponseType:   "D4",
		ItemList: []*STRequestItem{
			{
				CustomerNumber:       "1001",
				OrigNumber:           "1001",
				TermNumber:           "1002",
				BillToNumber:         "",
				TransDate:            "2013-11-07T08:42:26",
				Revenue:              1.01,
				Units:                1,
				UnitType:             "00",
				Seconds:              12,
				TaxIncludedCode:      "0",
				TaxSitusRule:         "04",
				TransTypeCode:        "010101",
				SalesTypeCode:        "R",
				RegulatoryCode:       "03",
				TaxExemptionCodeList: []string{},
			},
		},
	}
	jsnReq, _ := json.Marshal(eSTRequest)
	eSureTaxRequest := &SureTaxRequest{Request: string(jsnReq)}
	if stReq, err := NewSureTaxRequest(cdr, stCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSureTaxRequest, stReq) {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", string(eSureTaxRequest.Request), string(stReq.Request))
	}
}

func TestSureTaxNewSureTaxRequest(t *testing.T) {
	cdr := &CDR{}
	rcv, err := NewSureTaxRequest(cdr, nil)

	if err != nil {
		if err.Error() != "invalid SureTax config" {
			t.Fatal(err)
		}
	}

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestSureTaxSureTaxProcessCdr(t *testing.T) {
	str := "test"
	var nm int64 = 1
	fl := 1.2
	bl := false
	td := 1 * time.Second
	tm := time.Date(2023, 8, 17, 20, 34, 58, 651387237, time.UTC)
	tm2 := time.Date(2023, 8, 17, 20, 34, 59, 651387237, time.UTC)
	nm2 := 1
	ct := &ChargedTiming{
		Years:     utils.Years{1, 2},
		Months:    utils.Months{2, 3},
		MonthDays: utils.MonthDays{4, 5},
		WeekDays:  utils.WeekDays{2, 3},
		StartTime: "Time",
	}
	r := &Rate{
		GroupIntervalStart: td,
		Value:              fl,
		RateIncrement:      td,
		RateUnit:           td,
	}
	rg := RateGroups{r}
	rm := RatingMatchedFilters{str: nm2}
	bc := &BalanceCharge{
		AccountID:     str,
		BalanceUUID:   str,
		RatingID:      str,
		Units:         fl,
		ExtraChargeID: str,
	}
	ru := &RatingUnit{
		ConnectFee:       fl,
		RoundingMethod:   str,
		RoundingDecimals: nm2,
		MaxCost:          fl,
		MaxCostStrategy:  str,
		TimingID:         str,
		RatesID:          str,
		RatingFiltersID:  str,
	}
	cdr := &CDR{
		CGRID:       str,
		RunID:       str,
		OrderID:     nm,
		OriginHost:  str,
		Source:      str,
		OriginID:    str,
		ToR:         str,
		RequestType: str,
		Tenant:      str,
		Category:    str,
		Account:     str,
		Subject:     str,
		Destination: str,
		SetupTime:   tm,
		AnswerTime:  tm2,
		Usage:       td,
		ExtraFields: map[string]string{str: str},
		ExtraInfo:   str,
		Partial:     bl,
		PreRated:    bl,
		CostSource:  str,
		Cost:        fl,
		CostDetails: &EventCost{
			CGRID:     str,
			RunID:     str,
			StartTime: tm,
			Usage:     &td,
			Cost:      &fl,
			Charges: []*ChargingInterval{{
				RatingID: str,
				Increments: []*ChargingIncrement{{
					Usage:          td,
					Cost:           fl,
					AccountingID:   str,
					CompressFactor: nm2,
				}},
				CompressFactor: nm2,
				usage:          &td,
				ecUsageIdx:     &td,
				cost:           &fl,
			}},
			AccountSummary: &AccountSummary{
				Tenant: str,
				ID:     str,
				BalanceSummaries: BalanceSummaries{{
					UUID:     str,
					ID:       str,
					Type:     str,
					Value:    fl,
					Disabled: bl,
				}},
				AllowNegative: bl,
				Disabled:      bl,
			},
			Rating:        Rating{str: ru},
			Accounting:    Accounting{str: bc},
			RatingFilters: RatingFilters{str: rm},
			Rates:         ChargedRates{str: rg},
			Timings:       ChargedTimings{str: ct},
		},
	}

	err := SureTaxProcessCdr(cdr)

	if err != nil {
		if err.Error() != `Post "": unsupported protocol scheme ""` {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}
}
