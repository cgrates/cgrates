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

package rater

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/utils"
	"log"
	"time"
)

type DbReader struct {
	tpid              string
	storDB            DataStorage
	actions           map[string][]*Action
	actionsTimings    map[string][]*ActionTiming
	actionsTriggers   map[string][]*ActionTrigger
	accountActions    []*UserBalance
	destinations      []*Destination
	rates             map[string][]*Rate
	timings           map[string]*Timing
	activationPeriods map[string]*ActivationPeriod
	ratingProfiles    map[string]*RatingProfile
}

func NewDbReader(storDB DataStorage, tpid string) *DbReader {
	c := new(DbReader)
	c.storDB = storDB
	c.tpid = tpid
	c.activationPeriods = make(map[string]*ActivationPeriod)
	c.actionsTimings = make(map[string][]*ActionTiming)
	return c
}

func (dbr *DbReader) WriteToDatabase(storage DataStorage, flush, verbose bool) (err error) {
	if flush {
		storage.Flush()
	}
	if verbose {
		log.Print("Destinations")
	}
	for _, d := range dbr.destinations {
		err = storage.SetDestination(d)
		if err != nil {
			return err
		}
		if verbose {
			log.Print(d.Id, " : ", d.Prefixes)
		}
	}
	if verbose {
		log.Print("Rating profiles")
	}
	for _, rp := range dbr.ratingProfiles {
		err = storage.SetRatingProfile(rp)
		if err != nil {
			return err
		}
		if verbose {
			log.Print(rp.Id)
		}
	}
	if verbose {
		log.Print("Action timings")
	}
	for k, ats := range dbr.actionsTimings {
		err = storage.SetActionTimings(k, ats)
		if err != nil {
			return err
		}
		if verbose {
			log.Println(k)
		}
	}
	if verbose {
		log.Print("Actions")
	}
	for k, as := range dbr.actions {
		err = storage.SetActions(k, as)
		if err != nil {
			return err
		}
		if verbose {
			log.Println(k)
		}
	}
	if verbose {
		log.Print("Account actions")
	}
	for _, ub := range dbr.accountActions {
		err = storage.SetUserBalance(ub)
		if err != nil {
			return err
		}
		if verbose {
			log.Println(ub.Id)
		}
	}
	return
}

func (dbr *DbReader) LoadDestinations() (err error) {
	dbr.destinations, err = dbr.storDB.GetTpDestinations(dbr.tpid)
	return
}

func (dbr *DbReader) LoadRates() (err error) {
	dbr.rates, err = dbr.storDB.GetTpRates(dbr.tpid)
	return err
}

func (dbr *DbReader) LoadTimings() (err error) {
	dbr.timings, err = dbr.storDB.GetTpTimings(dbr.tpid)
	return err
}

func (dbr *DbReader) LoadRateTimings() error {
	rts, err := dbr.storDB.GetTpRateTimings(dbr.tpid)
	if err != nil {
		return err
	}
	for _, rt := range rts {
		t, exists := dbr.timings[rt.TimingsTag]
		if !exists {
			return errors.New(fmt.Sprintf("Could not get timing for tag %v", rt.TimingsTag))
		}
		rateTiming := &RateTiming{
			RatesTag: rt.RatesTag,
			Weight:   rt.Weight,
			timing:   t,
		}
		rs, exists := dbr.rates[rt.RatesTag]
		if !exists {
			return errors.New(fmt.Sprintf("Could not find rate for tag %v", rt.RatesTag))
		}
		for _, r := range rs {
			_, exists := dbr.activationPeriods[rt.Tag]
			if !exists {
				dbr.activationPeriods[rt.Tag] = &ActivationPeriod{}
			}
			dbr.activationPeriods[rt.Tag].AddIntervalIfNotPresent(rateTiming.GetInterval(r))
		}
	}
	return nil
}

func (dbr *DbReader) LoadRatingProfiles() error {
	rpfs, err := dbr.storDB.GetTpRatingProfiles(dbr.tpid)
	if err != nil {
		return err
	}
	for _, rp := range rpfs {
		at, err := time.Parse(time.RFC3339, rp.activationTime)
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot parse activation time from %v", rp.activationTime))
		}
		for _, d := range dbr.destinations {
			ap, exists := dbr.activationPeriods[rp.ratesTimingTag]
			if !exists {
				return errors.New(fmt.Sprintf("Could not load rating timing for tag: %v", rp.ratesTimingTag))
			}
			newAP := &ActivationPeriod{ActivationTime: at}
			//copy(newAP.Intervals, ap.Intervals)
			newAP.Intervals = append(newAP.Intervals, ap.Intervals...)
			rp.AddActivationPeriodIfNotPresent(d.Id, newAP)
			if rp.fallbackSubject != "" {
				rp.FallbackKey = fmt.Sprintf("%s:%s:%s:%s", rp.direction, rp.tenant, rp.tor, rp.fallbackSubject)
			}
		}
	}
	return nil
}

func (dbr *DbReader) LoadActions() (err error) {
	dbr.actions, err = dbr.storDB.GetTpActions(dbr.tpid)
	return err
}

func (dbr *DbReader) LoadActionTimings() (err error) {
	atsMap, err := dbr.storDB.GetTpActionTimings(dbr.tpid)
	if err != nil {
		return err
	}
	for tag, ats := range atsMap {
		for _, at := range ats {
			_, exists := dbr.actions[at.ActionsId]
			if !exists {
				return errors.New(fmt.Sprintf("ActionTiming: Could not load the action for tag: %v", at.ActionsId))
			}
			t, exists := dbr.timings[at.Tag]
			if !exists {
				return errors.New(fmt.Sprintf("ActionTiming: Could not load the timing for tag: %v", at.Tag))
			}
			actTmg := &ActionTiming{
				Id:     utils.GenUUID(),
				Tag:    at.Tag,
				Weight: at.Weight,
				Timing: &Interval{
					Months:    t.Months,
					MonthDays: t.MonthDays,
					WeekDays:  t.WeekDays,
					StartTime: t.StartTime,
				},
				ActionsId: at.ActionsId,
			}
			dbr.actionsTimings[tag] = append(dbr.actionsTimings[tag], actTmg)
		}
	}
	return err
}

func (dbr *DbReader) LoadActionTriggers() (err error) {
	dbr.actionsTriggers, err = dbr.storDB.GetTpActionTriggers(dbr.tpid)
	return err
}

func (dbr *DbReader) LoadAccountActions() (err error) {
	acs, err := dbr.storDB.GetTpAccountActions(dbr.tpid)
	if err != nil {
		return err
	}
	for _, aa := range acs {
		tag := fmt.Sprintf("%s:%s:%s", aa.Direction, aa.Tenant, aa.Account)
		aTriggers, exists := dbr.actionsTriggers[aa.ActionTriggersTag]
		if aa.ActionTriggersTag != "" && !exists {
			// only return error if there was something ther for the tag
			return errors.New(fmt.Sprintf("Could not get action triggers for tag %v", aa.ActionTriggersTag))
		}
		ub := &UserBalance{
			Type:           UB_TYPE_PREPAID,
			Id:             tag,
			ActionTriggers: aTriggers,
		}
		dbr.accountActions = append(dbr.accountActions, ub)

		aTimings, exists := dbr.actionsTimings[aa.ActionTimingsTag]
		if !exists {
			log.Printf("Could not get action timing for tag %v", aa.ActionTimingsTag)
			// must not continue here
		}
		for _, at := range aTimings {
			at.UserBalanceIds = append(at.UserBalanceIds, tag)
		}
	}
	return nil
}
