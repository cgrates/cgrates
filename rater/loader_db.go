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
	storDb            DataStorage
	dataDb            DataStorage
	actions           map[string][]*Action
	actionsTimings    map[string][]*ActionTiming
	actionsTriggers   map[string][]*ActionTrigger
	accountActions    []*UserBalance
	destinations      []*Destination
	timings           map[string]*Timing
	rates             map[string]*Rate
	destinationRates  map[string][]*DestinationRate
	activationPeriods map[string]*ActivationPeriod
	ratingProfiles    map[string]*RatingProfile
}

func NewDbReader(storDB DataStorage, storage DataStorage, tpid string) *DbReader {
	c := new(DbReader)
	c.storDb = storDB
	c.dataDb = storage
	c.tpid = tpid
	c.activationPeriods = make(map[string]*ActivationPeriod)
	c.actionsTimings = make(map[string][]*ActionTiming)
	return c
}

func (dbr *DbReader) WriteToDatabase(flush, verbose bool) (err error) {
	storage := dbr.dataDb
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
	dbr.destinations, err = dbr.storDb.GetTpDestinations(dbr.tpid, "")
	return
}

func (dbr *DbReader) LoadTimings() (err error) {
	dbr.timings, err = dbr.storDb.GetTpTimings(dbr.tpid, "")
	return err
}

func (dbr *DbReader) LoadRates() (err error) {
	dbr.rates, err = dbr.storDb.GetTpRates(dbr.tpid, "")
	return err
}

func (dbr *DbReader) LoadDestinationRates() (err error) {
	dbr.destinationRates, err = dbr.storDb.GetTpDestinationRates(dbr.tpid, "")
	if err != nil {
		return err
	}
	for _, drs := range dbr.destinationRates {
		for _, dr := range drs {
			rate, exists := dbr.rates[dr.RateTag]
			if !exists {
				return errors.New(fmt.Sprintf("Could not find rate for tag %v", dr.RateTag))
			}
			dr.Rate = rate
		}
	}
	return nil
}

func (dbr *DbReader) LoadDestinationRateTimings() error {
	rts, err := dbr.storDb.GetTpDestinationRateTimings(dbr.tpid, "")
	if err != nil {
		return err
	}
	for _, rt := range rts {
		t, exists := dbr.timings[rt.TimingsTag]
		if !exists {
			return errors.New(fmt.Sprintf("Could not get timing for tag %v", rt.TimingsTag))
		}
		rt.timing = t
		drs, exists := dbr.destinationRates[rt.DestinationRatesTag]
		if !exists {
			return errors.New(fmt.Sprintf("Could not find destination rate for tag %v", rt.DestinationRatesTag))
		}
		for _, dr := range drs {
			_, exists := dbr.activationPeriods[rt.Tag]
			if !exists {
				dbr.activationPeriods[rt.Tag] = &ActivationPeriod{}
			}
			dbr.activationPeriods[rt.Tag].AddIntervalIfNotPresent(rt.GetInterval(dr))
		}
	}
	return nil
}

func (dbr *DbReader) LoadRatingProfiles() error {
	rpfs, err := dbr.storDb.GetTpRatingProfiles(dbr.tpid, "")
	if err != nil {
		return err
	}
	for _, rp := range rpfs {
		at := time.Unix(rp.activationTime, 0)
		for _, d := range dbr.destinations {
			ap, exists := dbr.activationPeriods[rp.destRatesTimingTag]
			if !exists {
				return errors.New(fmt.Sprintf("Could not load rating timing for tag: %v", rp.destRatesTimingTag))
			}
			newAP := &ActivationPeriod{ActivationTime: at}
			//copy(newAP.Intervals, ap.Intervals)
			newAP.Intervals = append(newAP.Intervals, ap.Intervals...)
			rp.AddActivationPeriodIfNotPresent(d.Id, newAP)

		}
	}
	return nil
}

func (dbr *DbReader) LoadRatingProfileByTag(tag string) error {
	activationPeriods := make(map[string]*ActivationPeriod)
	resultRatingProfile := &RatingProfile{Id: tag}
	rpm, err := dbr.storDb.GetTpRatingProfiles(dbr.tpid, tag)
	if err != nil {
		return err
	} else if len(rpm) == 0 {
		return fmt.Errorf("No RateProfile with id: %s", tag)
	}
	for _, ratingProfile := range rpm {
		resultRatingProfile.FallbackKey = ratingProfile.FallbackKey // it will be the last fallback key
		at := time.Unix(ratingProfile.activationTime, 0)
		drtm, err := dbr.storDb.GetTpDestinationRateTimings(dbr.tpid, ratingProfile.destRatesTimingTag)
		if err != nil {
			return err
		} else if len(drtm) == 0 {
			return fmt.Errorf("No DestRateTimings profile with id: %s", ratingProfile.destRatesTimingTag)
		}
		for _, destrateTiming := range drtm {
			tm, err := dbr.storDb.GetTpTimings(dbr.tpid, destrateTiming.TimingsTag)
			if err != nil {
				return err
			} else if len(tm) == 0 {
				return fmt.Errorf("No Timings profile with id: %s", destrateTiming.TimingsTag)
			}
			destrateTiming.timing = tm[destrateTiming.TimingsTag]
			drm, err := dbr.storDb.GetTpDestinationRates(dbr.tpid, destrateTiming.DestinationRatesTag)
			if err != nil {
				return err
			} else if len(drm) == 0 {
				return fmt.Errorf("No Timings profile with id: %s", destrateTiming.DestinationRatesTag)
			}
			for _, drate := range drm[destrateTiming.DestinationRatesTag] {
				rt, err := dbr.storDb.GetTpRates(dbr.tpid, drate.RateTag)
				if err != nil {
					return err
				} else if len(rt) == 0 {
					return fmt.Errorf("No Rates profile with id: %s", drate.RateTag)
				}
				drate.Rate = rt[drate.RateTag]
				if _, exists := activationPeriods[destrateTiming.Tag]; !exists {
					activationPeriods[destrateTiming.Tag] = &ActivationPeriod{}
				}
				activationPeriods[destrateTiming.Tag].AddIntervalIfNotPresent(destrateTiming.GetInterval(drate))
				dm, err := dbr.storDb.GetTpDestinations(dbr.tpid, drate.DestinationsTag)
				if err != nil {
					return err
				}
				for _, destination := range dm {
					ap := activationPeriods[ratingProfile.destRatesTimingTag]
					newAP := &ActivationPeriod{ActivationTime: at}
					newAP.Intervals = append(newAP.Intervals, ap.Intervals...)
					resultRatingProfile.AddActivationPeriodIfNotPresent(destination.Id, newAP)
					dbr.dataDb.SetDestination(destination)
				}
			}
		}
	}
	return dbr.dataDb.SetRatingProfile(resultRatingProfile)
}

func (dbr *DbReader) LoadActions() (err error) {
	dbr.actions, err = dbr.storDb.GetTpActions(dbr.tpid, "")
	return err
}

func (dbr *DbReader) LoadActionTimings() (err error) {
	atsMap, err := dbr.storDb.GetTpActionTimings(dbr.tpid, "")
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
	dbr.actionsTriggers, err = dbr.storDb.GetTpActionTriggers(dbr.tpid, "")
	return err
}

func (dbr *DbReader) LoadAccountActions() (err error) {
	acs, err := dbr.storDb.GetTpAccountActions(dbr.tpid, "")
	if err != nil {
		return err
	}
	for _, aa := range acs {
		tag := fmt.Sprintf("%s:%s:%s", aa.Direction, aa.Tenant, aa.Account)
		aTriggers, exists := dbr.actionsTriggers[aa.ActionTriggersTag]
		if aa.ActionTriggersTag != "" && !exists {
			// only return error if there was something there for the tag
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

func (dbr *DbReader) LoadAccountActionsByTag(tag string) error {
	accountActions, err := dbr.storDb.GetTpAccountActions(dbr.tpid, tag)
	if err != nil {
		return err
	} else if len(accountActions) == 0 {
		return fmt.Errorf("No AccountActions with id <%s>", tag)
	} else if len(accountActions) > 1 {
		return fmt.Errorf("StorDb configuration error for AccountActions <%s>", tag)
	}
	accountAction := accountActions[tag]
	id := fmt.Sprintf("%s:%s:%s", accountAction.Direction, accountAction.Tenant, accountAction.Account)

	var actionsIds []string // collects action ids

	// action timings
	if accountAction.ActionTimingsTag != "" {
		// get old userBalanceIds
		var exitingUserBalanceIds []string
		existingActionTimings, err := dbr.dataDb.GetActionTimings(accountAction.ActionTimingsTag)
		if err == nil && len(existingActionTimings) > 0 {
			// all action timings from a specific tag shuld have the same list of user balances from the first one
			exitingUserBalanceIds = existingActionTimings[0].UserBalanceIds
		}

		actionTimingsMap, err := dbr.storDb.GetTpActionTimings(dbr.tpid, accountAction.ActionTimingsTag)
		if err != nil {
			return err
		} else if len(actionTimingsMap) == 0 {
			return fmt.Errorf("No ActionTimings with id <%s>", accountAction.ActionTimingsTag)
		}
		var actionTimings []*ActionTiming
		for _, at := range actionTimingsMap[accountAction.ActionTimingsTag] {
			existsAction, err := dbr.storDb.ExistsTPActions(dbr.tpid, at.ActionsId)
			if err != nil {
				return err
			} else if !existsAction {
				return fmt.Errorf("No Action with id <%s>", at.ActionsId)
			}
			timingsMap, err := dbr.storDb.GetTpTimings(dbr.tpid, at.Tag)
			if err != nil {
				return err
			} else if len(timingsMap) == 0 {
				return fmt.Errorf("No Timing with id <%s>", at.Tag)
			}
			t := timingsMap[at.Tag]
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
			// collect action ids from timings
			actionsIds = append(actionsIds, actTmg.ActionsId)
			//add user balance id if no already in
			found := false
			for _, ubId := range exitingUserBalanceIds {
				if ubId == id {
					found = true
					break
				}
			}
			if !found {
				at.UserBalanceIds = append(exitingUserBalanceIds, id)
			}
			actionTimings = append(actionTimings, actTmg)
		}

		// write action timings
		err = dbr.dataDb.SetActionTimings(accountAction.ActionTimingsTag, actionTimings)
		if err != nil {
			return err
		}
	}

	// action triggers
	var actionTriggers ActionTriggerPriotityList
	if accountAction.ActionTriggersTag != "" {
		actionTriggersMap, err := dbr.storDb.GetTpActionTriggers(dbr.tpid, accountAction.ActionTriggersTag)
		if err != nil {
			return err
		}
		actionTriggers = actionTriggersMap[accountAction.ActionTriggersTag]
		// collect action ids from triggers
		for _, atr := range actionTriggers {
			actionsIds = append(actionsIds, atr.ActionsId)
		}
	}

	// actions
	acts := make( map[string][]*Action )
	for _, actId := range actionsIds {
		actions, err := dbr.storDb.GetTpActions(dbr.tpid, actId)
		if err != nil {
			return err
		}
		for id, act := range actions {
			acts[id] =  act
		}
	}
	// write actions
	for k, as := range acts {
		err = dbr.dataDb.SetActions(k, as)
		if err != nil {
			return err
		}
	}

	ub, err := dbr.dataDb.GetUserBalance(id)
	if err != nil {
		return fmt.Errorf("Error when retrieving user balance <%s>: %s", id, err.Error())
	}
	ub.ActionTriggers = actionTriggers

	return dbr.dataDb.SetUserBalance(ub)
}
