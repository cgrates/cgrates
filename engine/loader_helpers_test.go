/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/utils"
	"io"
	"reflect"
	"strings"
	"testing"
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
RT_1CENT,0,1,1s,1s,0s,*up,2
DUMMY,INVALID;DATA
`

var destRatesSample = `#Tag,DestinationsTag,RatesTag
DR_RETAIL,GERMANY,RT_1CENT
DUMMY,INVALID;DATA
`
var ratingPlansSample = `#Tag,DestinationRatesTag,TimingTag,Weight
RP_RETAIL,DR_RETAIL,ALWAYS,10
DUMMY,INVALID;DATA
`

var ratingProfilesSample = `#Tenant,TOR,Direction,Subject,ActivationTime,RatingPlanTag,FallbackSubject
cgrates.org,call,*out,*any,2012-01-01T00:00:00Z,RP_RETAIL,
DUMMY,INVALID;DATA
`

var actionsSample = `#ActionsTag,Action,BalanceType,Direction,Units,ExpiryTime,DestinationTag,RatingSubject,BalanceWeight,SharedGroup,ExtraParameters,Weight
PREPAID_10,*topup_reset,*monetary,*out,5,*unlimited,*any,,,10,,10
WARN_HTTP,*call_url,,,,,,,,,http://localhost:8000,10
LOG_BALANCE,*log,,,,,,,,,,10
DUMMY,INVALID;DATA
PREPAID_10,*topup_reset,*monetary,*out,5,*unlimited,*any,,10,,10
TOPUP_RST_SHARED_5,*topup_reset,*monetary,*out,5,*unlimited,*any,subj,20,SHARED_A,param&some,10
`

var actionTimingsSample = `#Tag,ActionsTag,TimingTag,Weight
PREPAID_10,PREPAID_10,ASAP,10
DUMMY,INVALID;DATA
`

var actionTriggersSample = `#Tag,BalanceType,Direction,ThresholdType,ThresholdValue,DestinationTag,ActionsTag,Weight
STANDARD_TRIGGERS,*monetary,*out,*min_balance,2,,LOG_BALANCE,10
STANDARD_TRIGGERS,*monetary,*out,*max_balance,20,,LOG_BALANCE,10
STANDARD_TRIGGERS,*monetary,*out,*max_counter,15,FS_USERS,LOG_BALANCE,10
DUMMY,INVALID;DATA
`

var accountActionsSample = `#Tenant,Account,Direction,ActionTimingsTag,ActionTriggersTag
cgrates.org,1001,*out,PREPAID_10,STANDARD_TRIGGERS
DUMMY,INVALID;DATA
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
		case 2:
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
		case 2:
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
		case 2:
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
		case 2, 3, 4:
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
		case 1, 5:
			if valid {
				t.Error("Validation passed for invalid line", string(ln))
			}
		case 2, 3, 4:
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
		case 2:
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
			if !reflect.DeepEqual(record, []string{"RT_1CENT", "0", "1", "1s", "1s", "0s", "*up", "2"}) {
				t.Error("Unexpected record extracted", record)
			}
		case 3:
			if err == nil {
				t.Error("Expecting invalid line at row 3")
			}
		}
	}
}
