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
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/utils"
	"log"
)

type DbReader struct {
	tpid             string
	storDb           LoadStorage
	dataDb           DataStorage
	actions          map[string][]*Action
	actionsTimings   map[string][]*ActionTiming
	actionsTriggers  map[string][]*ActionTrigger
	accountActions   []*UserBalance
	destinations     []*Destination
	timings          map[string]*Timing
	rates            map[string][]*LoadRate
	destinationRates map[string][]*DestinationRate
	ratingPlans      map[string]*RatingPlan
	ratingProfiles   map[string]*RatingProfile
}

func NewDbReader(storDB LoadStorage, storage DataStorage, tpid string) *DbReader {
	c := new(DbReader)
	c.storDb = storDB
	c.dataDb = storage
	c.tpid = tpid
	c.actionsTimings = make(map[string][]*ActionTiming)
	c.ratingPlans = make(map[string]*RatingPlan)
	c.ratingProfiles = make(map[string]*RatingProfile)
	return c
}

func (dbr *DbReader) WriteToDatabase(flush, verbose bool) (err error) {
	storage := dbr.dataDb
	if flush {
		storage.(Storage).Flush()
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
		log.Print("Rating plans")
	}
	for _, rp := range dbr.ratingPlans {
		err = storage.SetRatingPlan(rp)
		if err != nil {
			return err
		}
		if verbose {
			log.Print(rp.Id)
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
			rates, exists := dbr.rates[dr.RateTag]
			if !exists {
				return errors.New(fmt.Sprintf("Could not find rate for tag %v", dr.RateTag))
			}
			dr.rates = rates
			destinationExists := false
			for _, d := range dbr.destinations {
				if d.Id == dr.DestinationsTag {
					destinationExists = true
					break
				}
			}
			if !destinationExists {
				if dbExists, err := dbr.dataDb.ExistsData(DESTINATION, dr.DestinationsTag); err != nil {
					return err
				} else if !dbExists {
					return errors.New(fmt.Sprintf("Could not get destination for tag %v", dr.DestinationsTag))
				}
			}
		}
	}
	return nil
}

func (dbr *DbReader) LoadRatingPlans() error {
	drts, err := dbr.storDb.GetTpRatingPlans(dbr.tpid, "")
	if err != nil {
		return err
	}
	for _, drt := range drts {
		t, exists := dbr.timings[drt.TimingTag]
		if !exists {
			return errors.New(fmt.Sprintf("Could not get timing for tag %v", drt.TimingTag))
		}
		drt.timing = t
		drs, exists := dbr.destinationRates[drt.DestinationRatesTag]
		if !exists {
			return errors.New(fmt.Sprintf("Could not find destination rate for tag %v", drt.DestinationRatesTag))
		}

		plan, exists := dbr.ratingPlans[drt.Tag]
		if !exists {
			plan = &RatingPlan{Id: drt.Tag}
			dbr.ratingPlans[drt.Tag] = plan
		}
		for _, dr := range drs {
			plan.AddRateInterval(dr.DestinationsTag, drt.GetRateInterval(dr))
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
		at, err := utils.ParseDate(rp.ActivationTime)
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot parse activation time from %v", rp.ActivationTime))
		}
		_, exists := dbr.ratingPlans[rp.DestRatesTimingTag]
		if !exists {
			if dbExists, err := dbr.dataDb.ExistsData(RATING_PLAN, rp.DestRatesTimingTag); err != nil {
				return err
			} else if !dbExists {
				return errors.New(fmt.Sprintf("Could not load rating plans for tag: %v", rp.DestRatesTimingTag))
			}
		}
		rp.RatingPlanActivations = append(rp.RatingPlanActivations,
			&RatingPlanActivation{
				ActivationTime: at,
				RatingPlanId:   rp.DestRatesTimingTag,
				FallbackKeys:   rp.FallbackKeys,
			})
		dbr.ratingProfiles[rp.Id] = rp
	}
	return nil
}

func (dbr *DbReader) LoadRatingPlanByTag(tag string) error {
	ratingPlan := &RatingPlan{}
	rps, err := dbr.storDb.GetTpRatingPlans(dbr.tpid, tag)
	if err != nil || len(rps) == 0 {
		return fmt.Errorf("No DestRateTimings profile with id %s: %v", tag, err)
	}
	for _, rp := range rps {

		Logger.Debug(fmt.Sprintf("Rating Plan: %v", rp))
		tm, err := dbr.storDb.GetTpTimings(dbr.tpid, rp.TimingTag)
		Logger.Debug(fmt.Sprintf("Timing: %v", tm))
		if err != nil || len(tm) == 0 {
			return fmt.Errorf("No Timings profile with id %s: %v", rp.TimingTag, err)
		}
		rp.timing = tm[rp.TimingTag]
		drm, err := dbr.storDb.GetTpDestinationRates(dbr.tpid, rp.DestinationRatesTag)
		if err != nil || len(drm) == 0 {
			return fmt.Errorf("No DestinationRates profile with id %s: %v", rp.DestinationRatesTag, err)
		}
		for _, drate := range drm[rp.DestinationRatesTag] {
			Logger.Debug(fmt.Sprintf("Destination rate: %v", drate))
			rt, err := dbr.storDb.GetTpRates(dbr.tpid, drate.RateTag)
			if err != nil || len(rt) == 0 {
				return fmt.Errorf("No Rates profile with id %s: %v", drate.RateTag, err)
			}
			Logger.Debug(fmt.Sprintf("Rate: %v", rt))
			drate.rates = rt[drate.RateTag]
			ratingPlan.AddRateInterval(drate.DestinationsTag, rp.GetRateInterval(drate))

			dms, err := dbr.storDb.GetTpDestinations(dbr.tpid, drate.DestinationsTag)
			if err != nil {
				return err
			} else if len(dms) == 0 {
				if dbExists, err := dbr.dataDb.ExistsData(DESTINATION, drate.DestinationsTag); err != nil {
					return err
				} else if !dbExists {
					return fmt.Errorf("Could not get destination for tag %v", drate.DestinationsTag)
				}
				continue
			}
			Logger.Debug(fmt.Sprintf("Tag: %s Destinations: %v", drate.DestinationsTag, dms))
			for _, destination := range dms {
				Logger.Debug(fmt.Sprintf("Destination: %v", destination))
				dbr.dataDb.SetDestination(destination)
			}
		}
	}
	return dbr.dataDb.SetRatingPlan(ratingPlan)
}

func (dbr *DbReader) LoadRatingProfileByTag(tag string) error {
	resultRatingProfile := &RatingProfile{}
	rpm, err := dbr.storDb.GetTpRatingProfiles(dbr.tpid, tag)
	if err != nil || len(rpm) == 0 {
		return fmt.Errorf("No RateProfile with id %s: %v", tag, err)
	}
	for _, ratingProfile := range rpm {
		Logger.Debug(fmt.Sprintf("Rating profile: %v", rpm))
		resultRatingProfile.Id = ratingProfile.Id // idem
		at, err := utils.ParseDate(ratingProfile.ActivationTime)
		if err != nil {
			return fmt.Errorf("Cannot parse activation time from %v", ratingProfile.ActivationTime)
		}
		// Check if referenced RatingPlan exists
		_, exists := dbr.ratingPlans[ratingProfile.DestRatesTimingTag]
		if !exists {
			if dbExists, err := dbr.dataDb.ExistsData(RATING_PLAN, ratingProfile.DestRatesTimingTag); err != nil {
				return err
			} else if !dbExists {
				return errors.New(fmt.Sprintf("Could not load rating plans for tag: %v", ratingProfile.DestRatesTimingTag))
			}
		}
		resultRatingProfile.RatingPlanActivations = append(resultRatingProfile.RatingPlanActivations, &RatingPlanActivation{at, ratingProfile.DestRatesTimingTag, ratingProfile.FallbackKeys})
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
				Timing: &RateInterval{
					Timing: &RITiming{
						Months:    t.Months,
						MonthDays: t.MonthDays,
						WeekDays:  t.WeekDays,
						StartTime: t.StartTime,
					},
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
				Timing: &RateInterval{
					Timing: &RITiming{
						Months:    t.Months,
						MonthDays: t.MonthDays,
						WeekDays:  t.WeekDays,
						StartTime: t.StartTime,
					},
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
	acts := make(map[string][]*Action)
	for _, actId := range actionsIds {
		actions, err := dbr.storDb.GetTpActions(dbr.tpid, actId)
		if err != nil {
			return err
		}
		for id, act := range actions {
			acts[id] = act
		}
	}
	// writee actions
	for k, as := range acts {
		err = dbr.dataDb.SetActions(k, as)
		if err != nil {
			return err
		}
	}

	ub, err := dbr.dataDb.GetUserBalance(id)
	if err != nil {
		ub = &UserBalance{
			Type: UB_TYPE_PREPAID,
			Id:   id,
		}
	}
	ub.ActionTriggers = actionTriggers

	return dbr.dataDb.SetUserBalance(ub)
}
