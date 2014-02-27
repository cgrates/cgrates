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
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

type TPLoader interface {
	LoadDestinations() error
	LoadRates() error
	LoadDestinationRates() error
	LoadTimings() error
	LoadRatingPlans() error
	LoadRatingProfiles() error
	LoadSharedGroups() error
	LoadActions() error
	LoadActionTimings() error
	LoadActionTriggers() error
	LoadAccountActions() error
	LoadAll() error
	GetLoadedIds(string) ([]string, error)
	ShowStatistics()
	WriteToDatabase(bool, bool) error
}

func NewLoadRate(tag, connectFee, price, ratedUnits, rateIncrements, groupInterval, roundingMethod, roundingDecimals string) (r *utils.TPRate, err error) {
	cf, err := strconv.ParseFloat(connectFee, 64)
	if err != nil {
		log.Printf("Error parsing connect fee from: %v", connectFee)
		return
	}
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		log.Printf("Error parsing price from: %v", price)
		return
	}
	rd, err := strconv.Atoi(roundingDecimals)
	if err != nil {
		log.Printf("Error parsing rounding decimals: %s", roundingDecimals)
		return
	}
	rs, err := utils.NewRateSlot(cf, p, ratedUnits, rateIncrements, groupInterval, roundingMethod, rd)
	if err != nil {
		return nil, err
	}
	r = &utils.TPRate{
		RateId:    tag,
		RateSlots: []*utils.RateSlot{rs},
	}
	return
}

func ValidNextGroup(present, next *utils.RateSlot) error {
	if next.GroupIntervalStartDuration() <= present.GroupIntervalStartDuration() {
		return errors.New(fmt.Sprintf("Next rate group interval start must be heigher than the last one: %#v", next))
	}
	if math.Mod(next.GroupIntervalStartDuration().Seconds(), present.RateIncrementDuration().Seconds()) != 0 {
		return errors.New(fmt.Sprintf("GroupIntervalStart of %#v must be a multiple of RateIncrement of %#v", next, present))
	}
	if present.RoundingMethod != next.RoundingMethod || present.RoundingDecimals != next.RoundingDecimals {
		return errors.New(fmt.Sprintf("Rounding stuff must be equal for sam rate tag: %#v, %#v", present, next))
	}
	return nil
}

func NewTiming(timingInfo ...string) (rt *utils.TPTiming) {
	rt = &utils.TPTiming{}
	rt.Id = timingInfo[0]
	rt.Years.Parse(timingInfo[1], ";")
	rt.Months.Parse(timingInfo[2], ";")
	rt.MonthDays.Parse(timingInfo[3], ";")
	rt.WeekDays.Parse(timingInfo[4], ";")
	rt.StartTime = timingInfo[5]
	return
}

func NewRatingPlan(timing *utils.TPTiming, weight string) (drt *utils.TPRatingPlanBinding) {
	w, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		log.Printf("Error parsing weight unit from: %v", weight)
		return
	}
	drt = &utils.TPRatingPlanBinding{
		Weight: w,
	}
	drt.SetTiming(timing)
	return
}

func GetRateInterval(rpl *utils.TPRatingPlanBinding, dr *utils.DestinationRate) (i *RateInterval) {
	i = &RateInterval{
		Timing: &RITiming{
			Years:     rpl.Timing().Years,
			Months:    rpl.Timing().Months,
			MonthDays: rpl.Timing().MonthDays,
			WeekDays:  rpl.Timing().WeekDays,
			StartTime: rpl.Timing().StartTime,
		},
		Weight: rpl.Weight,
		Rating: &RIRate{
			ConnectFee:       dr.Rate.RateSlots[0].ConnectFee,
			RoundingMethod:   dr.Rate.RateSlots[0].RoundingMethod,
			RoundingDecimals: dr.Rate.RateSlots[0].RoundingDecimals,
		},
	}
	for _, rl := range dr.Rate.RateSlots {
		i.Rating.Rates = append(i.Rating.Rates, &Rate{
			GroupIntervalStart: rl.GroupIntervalStartDuration(),
			Value:              rl.Rate,
			RateIncrement:      rl.RateIncrementDuration(),
			RateUnit:           rl.RateUnitDuration(),
		})
	}
	return
}

type AccountAction struct {
	Tenant, Account, Direction, ActionTimingsTag, ActionTriggersTag string
}

func ValidateCSVData(fn string, re *regexp.Regexp) (err error) {
	fin, err := os.Open(fn)
	if err != nil {
		// do not return the error, the file might be not needed
		return nil
	}
	defer fin.Close()
	r := bufio.NewReader(fin)
	line_number := 1
	for {
		line, truncated, err := r.ReadLine()
		if err != nil {
			break
		}
		if truncated {
			return errors.New("line too long")
		}
		// skip the header line
		if line_number > 1 {
			if !re.Match(line) {
				return errors.New(fmt.Sprintf("%s: error on line %d: %s", fn, line_number, line))
			}
		}
		line_number++
	}
	return
}

type FileLineRegexValidator struct {
	FieldsPerRecord int            // Number of fields in one record, useful for crosschecks
	Rule            *regexp.Regexp // Regexp rule
	Message         string         // Pass this message as helper
}

var FileValidators = map[string]*FileLineRegexValidator{
	utils.DESTINATIONS_CSV: &FileLineRegexValidator{utils.DESTINATIONS_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\+?\d+.?\d*){1}$`),
		"Tag([0-9A-Za-z_]),Prefix([0-9])"},
	utils.TIMINGS_CSV: &FileLineRegexValidator{utils.TIMINGS_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\*any\s*,\s*|(?:\d{1,4};?)+\s*,\s*|\s*,\s*){4}(?:\d{2}:\d{2}:\d{2}|\*asap){1}$`),
		"Tag([0-9A-Za-z_]),Years([0-9;]|*any|<empty>),Months([0-9;]|*any|<empty>),MonthDays([0-9;]|*any|<empty>),WeekDays([0-9;]|*any|<empty>),Time([0-9:]|*asap)"},
	utils.RATES_CSV: &FileLineRegexValidator{utils.RATES_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\d+\.?\d*,){2}(?:\d+s*,){3}(?:\*\w+,){1}(?:\d+\.?\d*,?){1}$`),
		"Tag([0-9A-Za-z_]),ConnectFee([0-9.]),Rate([0-9.]),RateUnit([0-9.]),RateIncrementStart([0-9.])"},
	utils.DESTINATION_RATES_CSV: &FileLineRegexValidator{utils.DESTINATION_RATES_NRCOLS,
		regexp.MustCompile(`^(?:\w+\s*),(?:\w+\s*),(?:\w+\s*)$`),
		"Tag([0-9A-Za-z_]),DestinationsTag([0-9A-Za-z_]),RateTag([0-9A-Za-z_])"},
	utils.RATING_PLANS_CSV: &FileLineRegexValidator{utils.DESTRATE_TIMINGS_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:\d+.?\d*){1}$`),
		"Tag([0-9A-Za-z_]),DestinationRatesTag([0-9A-Za-z_]),TimingProfile([0-9A-Za-z_]),Weight([0-9.])"},
	utils.RATING_PROFILES_CSV: &FileLineRegexValidator{utils.RATE_PROFILES_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,\s*){2}(?:\*out\s*,\s*){1}(?:\*any\s*,\s*|\w+\s*,\s*){1}(?:\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z){1}(?:\w*\s*,?\s*){2}$`),
		"Tenant([0-9A-Za-z_]),TOR([0-9A-Za-z_]),Direction(*out),Subject([0-9A-Za-z_]|*all),RatesFallbackSubject([0-9A-Za-z_]|<empty>),RatesTimingTag([0-9A-Za-z_]),ActivationTime([0-9T:X])"},
	utils.SHARED_GROUPS_CSV: &FileLineRegexValidator{utils.SHARED_GROUPS_NRCOLS,
		regexp.MustCompile(``),
		""},
	utils.ACTIONS_CSV: &FileLineRegexValidator{utils.ACTIONS_NRCOLS,
		regexp.MustCompile(`^(?:\w+\s*),(?:\*\w+\s*),(?:\*\w+\s*)?,(?:\*out\s*)?,(?:\d+\s*)?,(?:\*\w+\s*|\+\d+[smh]\s*|\d+\s*)?,(?:\*any|\w+\s*)?,(?:\*?\w+\s*)?,(?:\w+\s*)?,(?:\d+\.?\d*\s*)?,(?:\S+\s*)?,(?:\d+\.?\d*\s*)$`),
		"Tag([0-9A-Za-z_]),Action([0-9A-Za-z_]),BalanceType([*a-z_]),Direction(*out),Units([0-9]),ExpiryTime(*[a-z_]|+[0-9][smh]|[0-9])DestinationTag([0-9A-Za-z_]|*all),RatingSubject([0-9A-Za-z_]),BalanceWeight([0-9.]),ExtraParameters([0-9A-Za-z_:;]),Weight([0-9.])"},
	utils.ACTION_PLANS_CSV: &FileLineRegexValidator{utils.ACTION_PLANS_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:\d+\.?\d*){1}`),
		"Tag([0-9A-Za-z_]),ActionsTag([0-9A-Za-z_]),TimingTag([0-9A-Za-z_]),Weight([0-9.])"},
	utils.ACTION_TRIGGERS_CSV: &FileLineRegexValidator{utils.ACTION_TRIGGERS_NRCOLS,
		regexp.MustCompile(`(?:\w+),(?:\*\w+),(?:\*out),(?:\*\w+),(?:\d+\.?\d*),(?:\w+|\*any)?,(?:\w+),(?:\d+\.?\d*)$`),
		"Tag([0-9A-Za-z_]),BalanceType(*[a-z_]),Direction(*out),ThresholdType(*[a-z_]),ThresholdValue([0-9]+),DestinationTag([0-9A-Za-z_]|*all),ActionsTag([0-9A-Za-z_]),Weight([0-9]+)"},
	utils.ACCOUNT_ACTIONS_CSV: &FileLineRegexValidator{utils.ACCOUNT_ACTIONS_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\w+\s*,\s*){1}(?:\*out\s*,\s*){1}(?:\w+\s*,?\s*){2}$`),
		"Tenant([0-9A-Za-z_]),Account([0-9A-Za-z_.]),Direction(*out),ActionTimingsTag([0-9A-Za-z_]),ActionTriggersTag([0-9A-Za-z_])"},
}

func NewTPCSVFileParser(dirPath, fileName string) (*TPCSVFileParser, error) {
	validator, hasValidator := FileValidators[fileName]
	if !hasValidator {
		return nil, fmt.Errorf("No validator found for file <%s>", fileName)
	}
	// Open the file here
	fin, err := os.Open(path.Join(dirPath, fileName))
	if err != nil {
		return nil, err
	}
	//defer fin.Close()
	reader := bufio.NewReader(fin)
	return &TPCSVFileParser{validator, reader}, nil
}

// Opens the connection to a file and returns the parsed lines one by one when ParseNextLine() is called
type TPCSVFileParser struct {
	validator *FileLineRegexValidator // Row validator
	reader    *bufio.Reader           // Reader to the file we are interested in
}

func (self *TPCSVFileParser) ParseNextLine() ([]string, error) {
	line, truncated, err := self.reader.ReadLine()
	if err != nil {
		return nil, err
	} else if truncated {
		return nil, errors.New("Line too long.")
	}
	// skip commented lines
	if strings.HasPrefix(string(line), string(utils.COMMENT_CHAR)) {
		return nil, errors.New("Line starts with comment character.")
	}
	// Validate here string line
	if !self.validator.Rule.Match(line) {
		return nil, fmt.Errorf("Invalid line, <%s>", self.validator.Message)
	}
	// Open csv reader directly on string line
	csvReader, _, err := openStringCSVReader(string(line), ',', self.validator.FieldsPerRecord)
	if err != nil {
		return nil, err
	}
	record, err := csvReader.Read() // if no errors, record should be good to go having right format and length
	if err != nil {
		return nil, err
	}
	return record, nil
}
