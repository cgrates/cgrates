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
	timings          map[string]*utils.TPTiming
	rates            map[string]*utils.TPRate
	destinationRates map[string]*utils.TPDestinationRate
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
		for _, dr := range drs.DestinationRates {
			rate, exists := dbr.rates[dr.RateId]
			if !exists {
				return errors.New(fmt.Sprintf("Could not find rate for tag %v", dr.RateId))
			}
			dr.Rate = rate
			destinationExists := false
			for _, d := range dbr.destinations {
				if d.Id == dr.DestinationId {
					destinationExists = true
					break
				}
			}
			if !destinationExists {
				if dbExists, err := dbr.dataDb.ExistsData(DESTINATION_PREFIX, dr.DestinationId); err != nil {
					return err
				} else if !dbExists {
					return errors.New(fmt.Sprintf("Could not get destination for tag %v", dr.DestinationId))
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
	for _, drt := range drts.RatingPlans {
		t, exists := dbr.timings[drt.TimingId]
		if !exists {
			return errors.New(fmt.Sprintf("Could not get timing for tag %v", drt.TimingId))
		}
		drt.SetTiming(t)
		drs, exists := dbr.destinationRates[drt.DestinationRatesId]
		if !exists {
			return errors.New(fmt.Sprintf("Could not find destination rate for tag %v", drt.DestinationRatesId))
		}

		plan, exists := dbr.ratingPlans[drts.RatingPlanId]
		if !exists {
			plan = &RatingPlan{Id: drts.RatingPlanId}
			dbr.ratingPlans[drts.RatingPlanId] = plan
		}
		for _, dr := range drs.DestinationRates {
			plan.AddRateInterval(dr.DestinationId, GetRateInterval(drt, dr))
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
		rpf := &RatingProfile{Id: rp.RatingProfileId}
		at, err := utils.ParseDate(rp.ActivationTime)
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot parse activation time from %v", rp.ActivationTime))
		}
		_, exists := dbr.ratingPlans[rp.RatingPlanId]
		if !exists {
			if dbExists, err := dbr.dataDb.ExistsData(RATING_PLAN_PREFIX, rp.RatingPlanId); err != nil {
				return err
			} else if !dbExists {
				return errors.New(fmt.Sprintf("Could not load rating plans for tag: %v", rp.RatingPlanId))
			}
		}
		rpf.RatingPlanActivations = append(rpf.RatingPlanActivations,
			&RatingPlanActivation{
				ActivationTime: at,
				RatingPlanId:   rp.RatingPlanId,
				FallbackKeys:   rp.FallbackKeys,
			})
		dbr.ratingProfiles[rpf.Id] = rpf
	}
	return nil
}

func (dbr *DbReader) LoadRatingPlanByTag(tag string) error {
	ratingPlan := &RatingPlan{}
	rps, err := dbr.storDb.GetTpRatingPlans(dbr.tpid, tag)
	if err != nil || len(rps.RatingPlans) == 0 {
		return fmt.Errorf("No DestRateTimings profile with id %s: %v", tag, err)
	}
	for _, rp := range rps.RatingPlans {
		Logger.Debug(fmt.Sprintf("Rating Plan: %v", rp))
		tm, err := dbr.storDb.GetTpTimings(dbr.tpid, rp.TimingId)
		Logger.Debug(fmt.Sprintf("Timing: %v", tm))
		if err != nil || len(tm) == 0 {
			return fmt.Errorf("No Timings profile with id %s: %v", rp.TimingId, err)
		}
		rp.SetTiming(tm[rp.TimingId])
		drm, err := dbr.storDb.GetTpDestinationRates(dbr.tpid, rp.DestinationRatesId)
		if err != nil || len(drm) == 0 {
			return fmt.Errorf("No DestinationRates profile with id %s: %v", rp.DestinationRatesId, err)
		}
		for _, drate := range drm[rp.DestinationRatesId].DestinationRates {
			Logger.Debug(fmt.Sprintf("Destination rate: %v", drate))
			rt, err := dbr.storDb.GetTpRates(dbr.tpid, drate.RateId)
			if err != nil || len(rt) == 0 {
				return fmt.Errorf("No Rates profile with id %s: %v", drate.RateId, err)
			}
			Logger.Debug(fmt.Sprintf("Rate: %v", rt))
			drate.Rate = rt[drate.RateId]
			ratingPlan.AddRateInterval(drate.DestinationId, GetRateInterval(rp, drate))

			dms, err := dbr.storDb.GetTpDestinations(dbr.tpid, drate.DestinationId)
			if err != nil {
				return err
			} else if len(dms) == 0 {
				if dbExists, err := dbr.dataDb.ExistsData(DESTINATION_PREFIX, drate.DestinationId); err != nil {
					return err
				} else if !dbExists {
					return fmt.Errorf("Could not get destination for tag %v", drate.DestinationId)
				}
				continue
			}
			Logger.Debug(fmt.Sprintf("Tag: %s Destinations: %v", drate.DestinationId, dms))
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
	rpfs, err := dbr.storDb.GetTpRatingProfiles(dbr.tpid, tag)
	if err != nil || len(rpfs) == 0 {
		return fmt.Errorf("No RateProfile with id %s: %v", tag, err)
	}
	for _, rp := range rpfs {
		Logger.Debug(fmt.Sprintf("Rating profile: %v", rpfs))
		resultRatingProfile.Id = rp.RatingProfileId
		at, err := utils.ParseDate(rp.ActivationTime)
		if err != nil {
			return fmt.Errorf("Cannot parse activation time from %v", rp.ActivationTime)
		}
		// Check if referenced RatingPlan exists
		_, exists := dbr.ratingPlans[rp.RatingPlanId]
		if !exists {
			if dbExists, err := dbr.dataDb.ExistsData(RATING_PLAN_PREFIX, rp.RatingPlanId); err != nil {
				return err
			} else if !dbExists {
				return errors.New(fmt.Sprintf("Could not load rating plans for tag: %v", rp.RatingPlanId))
			}
		}
		resultRatingProfile.RatingPlanActivations = append(resultRatingProfile.RatingPlanActivations, &RatingPlanActivation{at, rp.RatingPlanId, rp.FallbackKeys})
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
	for atId, ats := range atsMap {
		for _, at := range ats {
			_, exists := dbr.actions[at.ActionsId]
			if !exists {
				return errors.New(fmt.Sprintf("ActionTiming: Could not load the action for tag: %v", at.ActionsId))
			}
			t, exists := dbr.timings[atId]
			if !exists {
				return errors.New(fmt.Sprintf("ActionTiming: Could not load the timing for tag: %v", atId))
			}
			actTmg := &ActionTiming{
				Id:     utils.GenUUID(),
				Tag:    atId,
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
			dbr.actionsTimings[atId] = append(dbr.actionsTimings[atId], actTmg)
		}
	}
	return err
}

func (dbr *DbReader) LoadActionTriggers() (err error) {
	atrsMap, err := dbr.storDb.GetTpActionTriggers(dbr.tpid, "")
	if err != nil {
		return err
	}
	for key, atrsLst := range atrsMap {
		atrs := make([]*ActionTrigger, len(atrsLst))
		for idx, apiAtr := range atrsLst {
			atrs[idx] = &ActionTrigger{Id: utils.GenUUID(),
						BalanceId: apiAtr.BalanceType,
						Direction: apiAtr.Direction,
						ThresholdType: apiAtr.ThresholdType,
						ThresholdValue: apiAtr.ThresholdValue,
						DestinationId: apiAtr.DestinationId,
						Weight: apiAtr.Weight,
						ActionsId: apiAtr.ActionsId,
						}
		}
		dbr.actionsTriggers[key] = atrs
	}
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
		ats := actionTimingsMap[accountAction.ActionTimingsTag]
		for _, at := range ats {
			existsAction, err := dbr.storDb.ExistsTPActions(dbr.tpid, at.ActionsId)
			if err != nil {
				return err
			} else if !existsAction {
				return fmt.Errorf("No Action with id <%s>", at.ActionsId)
			}
			timingsMap, err := dbr.storDb.GetTpTimings(dbr.tpid, accountAction.ActionTimingsTag)
			if err != nil {
				return err
			} else if len(timingsMap) == 0 {
				return fmt.Errorf("No Timing with id <%s>", accountAction.ActionTimingsTag)
			}
			t := timingsMap[accountAction.ActionTimingsTag]
			actTmg := &ActionTiming{
				Id:     utils.GenUUID(),
				Tag:    accountAction.ActionTimingsTag,
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
				actTmg.UserBalanceIds = append(exitingUserBalanceIds, id)
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
	//ActionTriggerPriotityList []*ActionTrigger
	if accountAction.ActionTriggersTag != "" {
		apiAtrsMap, err := dbr.storDb.GetTpActionTriggers(dbr.tpid, accountAction.ActionTriggersTag)
		if err != nil {
			return err
		}
		atrsMap := make( map[string][]*ActionTrigger )
		for key, atrsLst := range apiAtrsMap {
			atrs := make([]*ActionTrigger, len(atrsLst))
			for idx, apiAtr := range atrsLst {
				atrs[idx] = &ActionTrigger{Id: utils.GenUUID(),
						BalanceId: apiAtr.BalanceType,
						Direction: apiAtr.Direction,
						ThresholdType: apiAtr.ThresholdType,
						ThresholdValue: apiAtr.ThresholdValue,
						DestinationId: apiAtr.DestinationId,
						Weight: apiAtr.Weight,
						ActionsId: apiAtr.ActionsId,
						}
			}
			atrsMap[key] = atrs
		}
		actionTriggers = atrsMap[accountAction.ActionTriggersTag]
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


// Returns the identities loaded for a specific entity category
func (dbr *DbReader) GetLoadedIds( categ string ) ([]string, error) {
	switch categ {
	case DESTINATION_PREFIX:
		ids := make([]string, len(dbr.destinations))
		for idx, dst := range dbr.destinations {
			ids[idx] = dst.Id
		}
		return ids, nil
	case RATING_PLAN_PREFIX:
		keys := make([]string, len(dbr.ratingPlans))
		i := 0
		for k := range dbr.ratingPlans {
			keys[i] = k
			i++
		}
		return keys, nil
	case ACTION_TIMING_PREFIX: // actionsTimings
		keys := make([]string, len(dbr.actionsTimings))
		i := 0
		for k := range dbr.actionsTimings {
			keys[i] = k
			i++
		}
		return keys, nil
	}
	return nil, errors.New("Unsupported category")
}
