/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM

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
	"bufio"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

var timingsSample = `#Tag,Years,Months,MonthDays,WeekDays,Time
ALWAYS,*any,*any,*any,*any,00:00:00
DUMMY,INVALID;DATA
ASAP,*any,*any,*any,*any,*asap
`

var destsSample = `#Tag,Prefix
GERMANY,+49
DUMMY,INVALID;DATA
GERMANY_MOBILE,+4915
`
var ratesSample = `#Tag,DestinationRatesTag,TimingTag,Weight
RT_1CENT,0,1,1s,1s,0s
DUMMY,INVALID;DATA
RT_DATA_2c,0,0.002,10,10,0
`

var destRatesSample = `#Tag,DestinationsTag,RatesTag
DR_RETAIL,GERMANY,RT_1CENT,*up,0
DUMMY,INVALID;DATA
DR_DATA_1,*any,RT_DATA_2c,*up,2
`
var ratingPlansSample = `#Tag,DestinationRatesTag,TimingTag,Weight
RP_RETAIL,DR_RETAIL,ALWAYS,10
DUMMY,INVALID;DATA
`

var ratingProfilesSample = `#Tenant,TOR,Direction,Subject,ActivationTime,RatingPlanTag,FallbackSubject
*out,cgrates.org,call,*any,2012-01-01T00:00:00Z,RP_RETAIL,
DUMMY,INVALID;DATA
*out,cgrates.org,call,subj1;alias1,2012-01-01T00:00:00Z,RP_RETAIL,
`

var actionsSample = `#ActionsTag[0],Action[1],ExtraParameters[2],BalanceType[3],Direction[4],Category[5],DestinationTag[6],RatingSubject[7],SharedGroup[8],ExpiryTime[9],Units[10],BalanceWeight[11],Weight[12]
PREPAID_10,*topup_reset,,*monetary,*out,,*any,,,*unlimited,5,10,10
WARN_HTTP,*call_url,http://localhost:8000,,,,,,,,,,10
LOG_BALANCE,*log,,,,,,,,,,,10
DUMMY,INVALID;DATA
PREPAID_10,*topup_reset,,*monetary,*out,,*any,,*unlimited,5,10,10
TOPUP_RST_SHARED_5,*topup_reset,param&some,*monetary,*out,,*any,subj,SHARED_A,*unlimited,5,20,10
`

var actionTimingsSample = `#Tag,ActionsTag,TimingTag,Weight
PREPAID_10,PREPAID_10,ASAP,10
DUMMY,INVALID;DATA
`

var actionTriggersSample = `#Tag,BalanceTag,Direction,ThresholdType,ThresholdValue,Recurrent,MinSleep,BalanceDestinationTag,BalanceWeight,BalanceExpiryTime,BalanceRatingSubject,BalanceCategory,BalanceSharedGroup,StatsMinQueuedItems,ActionsTag,Weight
STANDARD_TRIGGERS,*min_balance,2,false,0,*monetary,*out,,,,,,,,LOG_WARNING,10
STANDARD_TRIGGERS,*max_balance,20,false,0,*monetary,*out,,,,,,,,LOG_WARNING,10
STANDARD_TRIGGERS,*max_counter,15,false,0,*monetary,*out,,FS_USERS,,,,,,LOG_WARNING,10
CDRST1_WARN_ASR,*min_asr,45,true,1h,,,,,,,,,3,CDRST_WARN_HTTP,10
DUMMY,INVALID;DATA
`

var sharedGroupsSample = `#Id,Account,Strategy,RatingSubject
SHARED_A,*any,*lowest_first,
DUMMY,INVALID;DATA
`

var accountActionsSample = `#Tenant,Account,Direction,ActionTimingsTag,ActionTriggersTag
cgrates.org,1001,*out,PREPAID_10,STANDARD_TRIGGERS
DUMMY,INVALID;DATA
cgrates.org,1002;1006,*out,PACKAGE_10,STANDARD_TRIGGERS
`
var derivedChargesSample = `#Direction,Tenant,Tor,Account,Subject,RunId,ReqTypeField,DirectionField,TenantField,TorField,AccountField,SubjectField,DestinationField,SetupTimeField,AnswerTimeField,DurationField
*out,cgrates.org,call,dan,dan,extra1,^filteredHeader1/filterValue1,^prepaid,,,,rif,rif,,,,
*out,cgrates.org,,dan,dan,extra1,^filteredHeader1/filterValue1,^prepaid,,,,rif,rif,,,,
*in,cgrates.org,call,dan,dan,extra1,^filteredHeader1/filterValue1,^prepaid,,,,rif,rif,,,,
DUMMY_DATA
*out,cgrates.org,call,dan,dan,extra2,,,,,,ivo,ivo,,,,
*out,cgrates.org,call,dan,*any,extra1,,,,,,rif2,rif2,,,,
*out,cgrates.org,call,dan,*any,*any,,,,,,rif2,rif2,,,,
*out,cgrates.org,call,dan,*any,*default,,*default,*default,*default,*default,rif2,rif2,*default,*default,*default,*default
*out,cgrates.org,call,dan,*any,test,^test,^test,^test,^test,^test,^test,^test,^test,^test,^test,^test
*out,cgrates.org,call,dan,*any,run1,,,,,,,,,,,
*out,cgrates.org,call,dan,*default,,,,,,,,,,,,
*out,cgrates.org,call,dan,dan,extra3,~filterhdr1:s/(.+)/special_run3/,,,,,^runusr3,^runusr3,,,,
`

func TestTimingsValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(timingsSample))
	lnValidator := FileValidators[utils.TIMINGS_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 3:
			if valid {
				t.Error("Validation passed for invalid line", ln)
			}
		case 2, 4:
			if !valid {
				t.Error("Validation did not pass for valid line", ln)
			}
		}
	}
}

func TestDestinationsValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(destsSample))
	lnValidator := FileValidators[utils.DESTINATIONS_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 3:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2, 4:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestRatesValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(ratesSample))
	lnValidator := FileValidators[utils.RATES_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 3:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2, 4:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestDestRatesValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(destRatesSample))
	lnValidator := FileValidators[utils.DESTINATION_RATES_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 3:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2, 4:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestRatingPlansValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(ratingPlansSample))
	lnValidator := FileValidators[utils.RATING_PLANS_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 3:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestRatingProfilesValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(ratingProfilesSample))
	lnValidator := FileValidators[utils.RATING_PROFILES_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 3:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2, 4:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestActionsValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(actionsSample))
	lnValidator := FileValidators[utils.ACTIONS_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 5, 6:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2, 3, 4, 7:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestActionTimingsValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(actionTimingsSample))
	lnValidator := FileValidators[utils.ACTION_PLANS_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 3:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestActionTriggersValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(actionTriggersSample))
	lnValidator := FileValidators[utils.ACTION_TRIGGERS_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 6:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2, 3, 4, 5:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestSharedGroupsValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(sharedGroupsSample))
	lnValidator := FileValidators[utils.SHARED_GROUPS_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 3:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestAccountActionsValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(accountActionsSample))
	lnValidator := FileValidators[utils.ACCOUNT_ACTIONS_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 3:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2, 4:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestDerivedChargersValidator(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(derivedChargesSample))
	lnValidator := FileValidators[utils.DERIVED_CHARGERS_CSV]
	lineNr := 0
	for {
		lineNr++
		ln, _, err := reader.ReadLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		valid := lnValidator.Rule.Match(ln)
		switch lineNr {
		case 1, 3, 4, 5, 8, 9, 12:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2, 6, 7, 10, 11, 13:
			if !valid {
				t.Error("Validation did not pass for valid line", string(ln))
			}
		}
	}
}

func TestTPCSVFileParser(t *testing.T) {
	bfRdr := bufio.NewReader(strings.NewReader(ratesSample))
	fParser := &TPCSVFileParser{FileValidators[utils.RATES_CSV], bfRdr}
	lineNr := 0
	for {
		lineNr++
		record, err := fParser.ParseNextLine()
		if err == io.EOF { // Reached end of the string
			break
		}
		switch lineNr {
		case 1:
			if err == nil || err.Error() != "Line starts with comment character." {
				t.Error("Failed to detect comment character")
			}
		case 2:
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(record, []string{"RT_1CENT", "0", "1", "1s", "1s", "0s"}) {
				t.Error("Unexpected record extracted", record)
			}
		case 3:
			if err == nil {
				t.Error("Expecting invalid line at row 3")
			}
		}
	}
}

func TestValueOrDefault(t *testing.T) {
	if res := ValueOrDefault("someval", "*any"); res != "someval" {
		t.Error("Unexpected value received", res)
	}
	if res := ValueOrDefault("", "*any"); res != "*any" {
		t.Error("Unexpected value received", res)
	}
}
