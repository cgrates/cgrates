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
package migrator

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCdrsNewV1CDRFromCDRSql(t *testing.T) {
	testCdrSql := &engine.CDRsql{
		ID:          1,
		Cgrid:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		Source:      "testSource",
		TOR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
		Usage:       1,
		CostSource:  "testSource",
		Cost:        2,
		ExtraInfo:   "testExtraInfo",
		CreatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		UpdatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		DeletedAt:   utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
	}
	expected := &v1Cdrs{
		CGRID:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		OrderID:     1,
		Source:      "testSource",
		ToR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		Usage:       time.Nanosecond,
		ExtraInfo:   "testExtraInfo",
		Partial:     false,
		Rated:       false,
		CostSource:  "testSource",
		Cost:        2,
	}
	result, err := NewV1CDRFromCDRSql(testCdrSql)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestCdrsAsCDRsql(t *testing.T) {
	testV1Cdrs := &v1Cdrs{
		CGRID:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		OrderID:     1,
		Source:      "testSource",
		ToR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		Usage:       time.Nanosecond,
		ExtraInfo:   "testExtraInfo",
		Partial:     false,
		Rated:       false,
		CostSource:  "testSource",
		Cost:        2,
	}

	expected := &engine.CDRsql{
		ID:          0,
		Cgrid:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		Source:      "testSource",
		TOR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
		Usage:       1,
		CostSource:  "testSource",
		Cost:        2,
		ExtraInfo:   "testExtraInfo",
		CreatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		ExtraFields: "",
		CostDetails: "",
	}
	result := testV1Cdrs.AsCDRsql()
	result.CreatedAt = time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)
	result.ExtraFields = ""
	result.CostDetails = ""
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(result))
	}
}

func TestCdrsAsCDRsqlAnswertimeEmpty(t *testing.T) {
	var answTime time.Time
	testV1Cdrs := &v1Cdrs{
		AnswerTime: answTime,
	}

	expected := &engine.CDRsql{
		AnswerTime: nil,
		CreatedAt:  time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
	}
	result := testV1Cdrs.AsCDRsql()
	result.CreatedAt = time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)
	result.ExtraFields = ""
	result.CostDetails = ""
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(result))
	}
}

func TestCdrsNewV1CDRFromCDRSqlAnswerTimeNil(t *testing.T) {
	testCdrSql := &engine.CDRsql{
		ID:          1,
		Cgrid:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		Source:      "testSource",
		TOR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  nil,
		Usage:       1,
		CostSource:  "testSource",
		Cost:        2,
		ExtraInfo:   "testExtraInfo",
		CreatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		UpdatedAt:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		DeletedAt:   utils.TimePointer(time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC)),
	}
	var answTime time.Time
	expected := &v1Cdrs{
		CGRID:       "testID",
		RunID:       "testRunID",
		OriginHost:  "testOriginHost",
		OrderID:     1,
		Source:      "testSource",
		ToR:         "testTOR",
		RequestType: "testRequestType",
		Tenant:      "cgrates.org",
		Category:    "testCategory",
		Account:     "testAccount",
		Subject:     "testSubject",
		Destination: "testDestination",
		SetupTime:   time.Date(2021, 3, 3, 3, 3, 3, 3, time.UTC),
		AnswerTime:  answTime,
		Usage:       time.Nanosecond,
		ExtraInfo:   "testExtraInfo",
		Partial:     false,
		Rated:       false,
		CostSource:  "testSource",
		Cost:        2,
	}
	result, err := NewV1CDRFromCDRSql(testCdrSql)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
