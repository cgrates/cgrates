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
	"github.com/cgrates/cgrates/utils"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TPLoader interface {
	LoadDestinations() error
	LoadRates() error
	LoadDestinationRates() error
	LoadTimings() error
	LoadDestinationRateTimings() error
	LoadRatingProfiles() error
	LoadActions() error
	LoadActionTimings() error
	LoadActionTriggers() error
	LoadAccountActions() error
	WriteToDatabase(bool, bool) error
}

type Rate struct {
	Tag                                         string
	ConnectFee, Price                           float64
	RateUnit, RateIncrement, GroupIntervalStart time.Duration
	RoundingMethod                              string
	RoundingDecimals                            int
	Weight                                      float64
}

func NewRate(tag, connectFee, price, ratedUnits, rateIncrements, groupInterval, roundingMethod, roundingDecimals, weight string) (r *Rate, err error) {
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
	gi, err := time.ParseDuration(groupInterval)
	if err != nil {
		log.Printf("Error parsing group interval from: %v", price)
		return
	}
	ru, err := time.ParseDuration(ratedUnits)
	if err != nil {
		log.Printf("Error parsing rated units from: %v", ratedUnits)
		return
	}
	ri, err := time.ParseDuration(rateIncrements)
	if err != nil {
		log.Printf("Error parsing rates increments from: %v", rateIncrements)
		return
	}
	wght, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		log.Printf("Error parsing weight from: %s", weight)
		return
	}
	rd, err := strconv.Atoi(roundingDecimals)
	if err != nil {
		log.Printf("Error parsing rounding decimals: %s", roundingDecimals)
		return
	}

	r = &Rate{
		Tag:                tag,
		ConnectFee:         cf,
		Price:              p,
		GroupIntervalStart: gi,
		RateUnit:           ru,
		RateIncrement:      ri,
		Weight:             wght,
		RoundingMethod:     roundingMethod,
		RoundingDecimals:   rd,
	}
	return
}

type DestinationRate struct {
	Tag             string
	DestinationsTag string
	RateTag         string
	Rate            *Rate
}

type Timing struct {
	Id        string
	Years     Years
	Months    Months
	MonthDays MonthDays
	WeekDays  WeekDays
	StartTime string
}

func NewTiming(timingInfo ...string) (rt *Timing) {
	rt = &Timing{}
	rt.Id = timingInfo[0]
	rt.Years.Parse(timingInfo[1], ";")
	rt.Months.Parse(timingInfo[2], ";")
	rt.MonthDays.Parse(timingInfo[3], ";")
	rt.WeekDays.Parse(timingInfo[4], ";")
	rt.StartTime = timingInfo[5]
	return
}

type DestinationRateTiming struct {
	Tag                 string
	DestinationRatesTag string
	Weight              float64
	TimingsTag          string // intermediary used when loading from db
	timing              *Timing
}

func NewDestinationRateTiming(destinationRatesTag string, timing *Timing, weight string) (rt *DestinationRateTiming) {
	w, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		log.Printf("Error parsing weight unit from: %v", weight)
		return
	}
	rt = &DestinationRateTiming{
		DestinationRatesTag: destinationRatesTag,
		Weight:              w,
		timing:              timing,
	}
	return
}

func (rt *DestinationRateTiming) GetInterval(dr *DestinationRate) (i *Interval) {
	i = &Interval{
		Years:      rt.timing.Years,
		Months:     rt.timing.Months,
		MonthDays:  rt.timing.MonthDays,
		WeekDays:   rt.timing.WeekDays,
		StartTime:  rt.timing.StartTime,
		Weight:     rt.Weight,
		ConnectFee: dr.Rate.ConnectFee,
		Prices: PriceGroups{&Price{
			GroupIntervalStart: dr.Rate.GroupIntervalStart,
			Value:              dr.Rate.Price,
			RateIncrement:      dr.Rate.RateIncrement,
			RateUnit:           dr.Rate.RateUnit,
		}},
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
		"Tag([0-9A-Za-z_]),Years([0-9;]|*all|<empty>),Months([0-9;]|*all|<empty>),MonthDays([0-9;]|*all|<empty>),WeekDays([0-9;]|*all|<empty>),Time([0-9:]|*asap)"},
	utils.RATES_CSV: &FileLineRegexValidator{utils.RATES_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,\s*){1}(?:\d+\.?\d*,){2}(?:\d+s*,){3}(?:\*\w+,){1}(?:\d+\.?\d*,?){2}$`),
		"Tag([0-9A-Za-z_]),ConnectFee([0-9.]),Rate([0-9.]),RateUnit([0-9.]),RateIncrementStart([0-9.])"},
	utils.DESTINATION_RATES_CSV: &FileLineRegexValidator{utils.DESTINATION_RATES_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,?\s*){3}$`),
		"Tag([0-9A-Za-z_]),DestinationsTag([0-9A-Za-z_]),RateTag([0-9A-Za-z_])"},
	utils.DESTRATE_TIMINGS_CSV: &FileLineRegexValidator{utils.DESTRATE_TIMINGS_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,\s*){3}(?:\d+.?\d*){1}$`),
		"Tag([0-9A-Za-z_]),DestinationRatesTag([0-9A-Za-z_]),TimingProfile([0-9A-Za-z_]),Weight([0-9.])"},
	utils.RATE_PROFILES_CSV: &FileLineRegexValidator{utils.RATE_PROFILES_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*,\s*){2}(?:\*out\s*,\s*){1}(?:\*any\s*,\s*|\w+\s*,\s*){1}(?:\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z){1}(?:\w*\s*,?\s*){2}$`),
		"Tenant([0-9A-Za-z_]),TOR([0-9A-Za-z_]),Direction(*out),Subject([0-9A-Za-z_]|*all),RatesFallbackSubject([0-9A-Za-z_]|<empty>),RatesTimingTag([0-9A-Za-z_]),ActivationTime([0-9T:X])"},
	utils.ACTIONS_CSV: &FileLineRegexValidator{utils.ACTIONS_NRCOLS,
		regexp.MustCompile(`(?:\w+\s*),(?:\*\w+\s*),(?:\*\w+\s*),(?:\*out\s*),(?:\d+\s*),(?:\*\w+\s*|\+\d+[smh]\s*|\d+\s*),(?:\*any|\w+\s*),(?:\*\w+\s*)?,(?:\d+\.?\d*\s*)?,(?:\d+\.?\d*\s*)?,(?:\d+\.?\d*\s*)$`),
		"Tag([0-9A-Za-z_]),Action([0-9A-Za-z_]),BalanceType([*a-z_]),Direction(*out),Units([0-9]),ExpiryTime(*[a-z_]|+[0-9][smh]|[0-9])DestinationTag([0-9A-Za-z_]|*all),RateType(*[a-z_]),RateValue([0-9.]),MinutesWeight([0-9.]),Weight([0-9.])"},
	utils.ACTION_TIMINGS_CSV: &FileLineRegexValidator{utils.ACTION_TIMINGS_NRCOLS,
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
