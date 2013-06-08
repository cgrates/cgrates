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
	timings           map[string][]*Timing
	activationPeriods map[string]*ActivationPeriod
	ratingProfiles    map[string]*RatingProfile
}

func NewDbReader(storDB DataStorage) *DbReader {
	c := new(DbReader)
	c.storDB = storDB
	/*c.actions = make(map[string][]*Action)
	c.actionsTimings = make(map[string][]*ActionTiming)
	c.actionsTriggers = make(map[string][]*ActionTrigger)
	c.rates = make(map[string][]*Rate)
	c.timings = make(map[string][]*Timing)*/
	c.activationPeriods = make(map[string]*ActivationPeriod)
	//c.ratingProfiles = make(map[string]*RatingProfile)
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

func (dbr *DbReader) LoadDestinations(tpid string) (err error) {
	dbr.destinations, err = dbr.storDB.GetAllDestinations(tpid)
	return
}

func (dbr *DbReader) LoadRates(tpid string) (err error) {
	dbr.rates, err = dbr.storDB.GetAllRates(tpid)
	return err
}

func (dbr *DbReader) LoadTimings(tpid string) (err error) {
	dbr.timings, err = dbr.storDB.GetAllTimings(tpid)
	return err
}

func (dbr *DbReader) LoadRateTimings(tpid string) error {
	rts, err := dbr.storDB.GetAllRateTimings(tpid)
	if err != nil {
		return err
	}
	for _, rt := range rts {
		ts, exists := dbr.timings[rt.TimingsTag]
		if !exists {
			return errors.New(fmt.Sprintf("Could not get timing for tag %v", rt.TimingsTag))
		}
		for _, t := range ts {
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
	}
	return nil
}

func (dbr *DbReader) LoadRatingProfiles(tpid string) error {
	rpfs, err := dbr.storDB.GetAllRatingProfiles(tpid)
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

func (dbr *DbReader) LoadActions(tpid string) (err error) {
	dbr.actions, err = dbr.storDB.GetAllActions(tpid)
	return err
}

func (dbr *DbReader) LoadActionTimings(tpid string) (err error) {
	atsMap, err := dbr.storDB.GetAllActionTimings(tpid)
	if err != nil {
		return err
	}
	for tag, ats := range atsMap {
		for _, at := range ats {
			_, exists := dbr.actions[at.ActionsId]
			if !exists {
				return errors.New(fmt.Sprintf("ActionTiming: Could not load the action for tag: %v", at.ActionsId))
			}
			ts, exists := dbr.timings[at.Tag]
			if !exists {
				return errors.New(fmt.Sprintf("ActionTiming: Could not load the timing for tag: %v", at.Tag))
			}
			for _, t := range ts {
				actTmg := &ActionTiming{
					Id:     GenUUID(),
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
	}
	return err
}

func (dbr *DbReader) LoadActionTriggers(tpid string) (err error) {
	dbr.actionsTriggers, err = dbr.storDB.GetAllActionTriggers(tpid)
	return err
}

func (dbr *DbReader) LoadAccountActions(tpid string) (err error) {
	dbr.accountActions, err = dbr.storDB.GetAllUserBalances(tpid)
	for _, ub := range dbr.accountActions {
		aTimings, exists := dbr.actionsTimings[ub.actionTimingsTag]
		if !exists {
			log.Printf("Could not get action timing for tag %v", ub.actionTimingsTag)
			// must not continue here
		}
		for _, at := range aTimings {
			aTriggers, exists := dbr.actionsTriggers[ub.actionTriggersTag]
			if ub.actionTriggersTag != "" && !exists {
				// only return error if there was something ther for the tag
				return errors.New(fmt.Sprintf("Could not get action triggers for tag %v", ub.actionTriggersTag))
			}
			ub.ActionTriggers = aTriggers
			at.UserBalanceIds = append(at.UserBalanceIds, ub.Id)
		}
	}
	return nil
}
