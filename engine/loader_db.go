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
	"log"

	"github.com/cgrates/cgrates/utils"
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
	c.actions = make(map[string][]*Action)
	c.actionsTimings = make(map[string][]*ActionTiming)
	c.actionsTriggers = make(map[string][]*ActionTrigger)
	c.ratingPlans = make(map[string]*RatingPlan)
	c.ratingProfiles = make(map[string]*RatingProfile)
	return c
}

func (dbr *DbReader) ShowStatistics() {
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
	mpRpls, err := dbr.storDb.GetTpRatingPlans(dbr.tpid, "")
	if err != nil {
		return err
	}
	for tag, rplBnds := range mpRpls {
		for _, rplBnd := range rplBnds {
			t, exists := dbr.timings[rplBnd.TimingId]
			if !exists {
				return errors.New(fmt.Sprintf("Could not get timing for tag %v", rplBnd.TimingId))
			}
			rplBnd.SetTiming(t)
			drs, exists := dbr.destinationRates[rplBnd.DestinationRatesId]
			if !exists {
				return errors.New(fmt.Sprintf("Could not find destination rate for tag %v", rplBnd.DestinationRatesId))
			}
			plan, exists := dbr.ratingPlans[tag]
			if !exists {
				plan = &RatingPlan{Id: tag}
				dbr.ratingPlans[plan.Id] = plan
			}
			for _, dr := range drs.DestinationRates {
				plan.AddRateInterval(dr.DestinationId, GetRateInterval(rplBnd, dr))
			}
		}
	}
	return nil
}

func (dbr *DbReader) LoadRatingProfiles() error {
	mpTpRpfs, err := dbr.storDb.GetTpRatingProfiles(&utils.TPRatingProfile{TPid: dbr.tpid}) //map[string]*utils.TPRatingProfile
	if err != nil {
		return err
	}
	for _, tpRpf := range mpTpRpfs {
		rpf := &RatingProfile{Id: tpRpf.KeyId()}
		for _, tpRa := range tpRpf.RatingPlanActivations {
			at, err := utils.ParseDate(tpRa.ActivationTime)
			if err != nil {
				return errors.New(fmt.Sprintf("Cannot parse activation time from %v", tpRa.ActivationTime))
			}
			_, exists := dbr.ratingPlans[tpRa.RatingPlanId]
			if !exists {
				if dbExists, err := dbr.dataDb.ExistsData(RATING_PLAN_PREFIX, tpRa.RatingPlanId); err != nil {
					return err
				} else if !dbExists {
					return errors.New(fmt.Sprintf("Could not load rating plans for tag: %v", tpRa.RatingPlanId))
				}
			}
			rpf.RatingPlanActivations = append(rpf.RatingPlanActivations,
				&RatingPlanActivation{
					ActivationTime: at,
					RatingPlanId:   tpRa.RatingPlanId,
					FallbackKeys:   utils.FallbackSubjKeys(tpRpf.Direction, tpRpf.Tenant, tpRpf.TOR, tpRa.FallbackSubjects),
				})
		}
		dbr.ratingProfiles[tpRpf.KeyId()] = rpf
	}
	return nil
}

// Returns true, nil in case of load success, false, nil in case of RatingPlan  not found in storDb
func (dbr *DbReader) LoadRatingPlanByTag(tag string) (bool, error) {
	mpRpls, err := dbr.storDb.GetTpRatingPlans(dbr.tpid, tag)
	if err != nil {
		return false, err
	} else if len(mpRpls) == 0 {
		return false, nil
	}
	for tag, rplBnds := range mpRpls {
		ratingPlan := &RatingPlan{Id: tag}
		for _, rp := range rplBnds {
			// Logger.Debug(fmt.Sprintf("Rating Plan binding: %v", rp))
			tm, err := dbr.storDb.GetTpTimings(dbr.tpid, rp.TimingId)
			// Logger.Debug(fmt.Sprintf("Timing: %v", tm))
			if err != nil || len(tm) == 0 {
				return false, fmt.Errorf("No Timings profile with id %s: %v", rp.TimingId, err)
			}
			rp.SetTiming(tm[rp.TimingId])
			drm, err := dbr.storDb.GetTpDestinationRates(dbr.tpid, rp.DestinationRatesId)
			if err != nil || len(drm) == 0 {
				return false, fmt.Errorf("No DestinationRates profile with id %s: %v", rp.DestinationRatesId, err)
			}
			for _, drate := range drm[rp.DestinationRatesId].DestinationRates {
				// Logger.Debug(fmt.Sprintf("Destination rate: %v", drate))
				rt, err := dbr.storDb.GetTpRates(dbr.tpid, drate.RateId)
				if err != nil || len(rt) == 0 {
					return false, fmt.Errorf("No Rates profile with id %s: %v", drate.RateId, err)
				}
				// Logger.Debug(fmt.Sprintf("Rate: %v", rt))
				drate.Rate = rt[drate.RateId]
				ratingPlan.AddRateInterval(drate.DestinationId, GetRateInterval(rp, drate))

				dms, err := dbr.storDb.GetTpDestinations(dbr.tpid, drate.DestinationId)
				if err != nil {
					return false, err
				} else if len(dms) == 0 {
					if dbExists, err := dbr.dataDb.ExistsData(DESTINATION_PREFIX, drate.DestinationId); err != nil {
						return false, err
					} else if !dbExists {
						return false, fmt.Errorf("Could not get destination for tag %v", drate.DestinationId)
					}
					continue
				}
				// Logger.Debug(fmt.Sprintf("Tag: %s Destinations: %v", drate.DestinationId, dms))
				for _, destination := range dms {
					// Logger.Debug(fmt.Sprintf("Destination: %v", destination))
					dbr.dataDb.SetDestination(destination)
				}
			}
		}
		if err := dbr.dataDb.SetRatingPlan(ratingPlan); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (dbr *DbReader) LoadRatingProfileFiltered(qriedRpf *utils.TPRatingProfile) error {
	var resultRatingProfile *RatingProfile
	mpTpRpfs, err := dbr.storDb.GetTpRatingProfiles(qriedRpf) //map[string]*utils.TPRatingProfile
	if err != nil {
		return fmt.Errorf("No RateProfile for filter %v, error: %s", qriedRpf, err.Error())
	}
	for _, tpRpf := range mpTpRpfs {
		// Logger.Debug(fmt.Sprintf("Rating profile: %v", tpRpf))
		resultRatingProfile = &RatingProfile{Id: tpRpf.KeyId()}
		for _, tpRa := range tpRpf.RatingPlanActivations {
			at, err := utils.ParseDate(tpRa.ActivationTime)
			if err != nil {
				return errors.New(fmt.Sprintf("Cannot parse activation time from %v", tpRa.ActivationTime))
			}
			_, exists := dbr.ratingPlans[tpRa.RatingPlanId]
			if !exists {
				if dbExists, err := dbr.dataDb.ExistsData(RATING_PLAN_PREFIX, tpRa.RatingPlanId); err != nil {
					return err
				} else if !dbExists {
					return errors.New(fmt.Sprintf("Could not load rating plans for tag: %v", tpRa.RatingPlanId))
				}
			}
			resultRatingProfile.RatingPlanActivations = append(resultRatingProfile.RatingPlanActivations,
				&RatingPlanActivation{at, tpRa.RatingPlanId,
					utils.FallbackSubjKeys(tpRpf.Direction, tpRpf.Tenant, tpRpf.TOR, tpRa.FallbackSubjects)})
		}
		if err := dbr.dataDb.SetRatingProfile(resultRatingProfile); err != nil {
			return err
		}
	}
	return nil
}

func (dbr *DbReader) LoadActions() (err error) {
	storActs, err := dbr.storDb.GetTpActions(dbr.tpid, "")
	if err != nil {
		return err
	}
	// map[string][]*Action
	for tag, tpacts := range storActs {
		acts := make([]*Action, len(tpacts))
		for idx, tpact := range tpacts {
			acts[idx] = &Action{
				Id:               utils.GenUUID(),
				ActionType:       tpact.Identifier,
				BalanceId:        tpact.BalanceType,
				Direction:        tpact.Direction,
				Weight:           tpact.Weight,
				ExtraParameters:  tpact.ExtraParameters,
				ExpirationString: tpact.ExpiryTime,
				Balance: &Balance{
					Uuid:          utils.GenUUID(),
					Value:         tpact.Units,
					Weight:        tpact.BalanceWeight,
					RateSubject:   tpact.RatingSubject,
					DestinationId: tpact.DestinationId,
				},
			}
		}
		dbr.actions[tag] = acts
	}
	return nil
}

func (dbr *DbReader) LoadActionTimings() (err error) {
	atsMap, err := dbr.storDb.GetTPActionTimings(dbr.tpid, "")
	if err != nil {
		return err
	}
	for atId, ats := range atsMap {
		for _, at := range ats {
			_, exists := dbr.actions[at.ActionsId]
			if !exists {
				return errors.New(fmt.Sprintf("ActionTiming: Could not load the action for tag: %v", at.ActionsId))
			}
			t, exists := dbr.timings[at.TimingId]
			if !exists {
				return errors.New(fmt.Sprintf("ActionTiming: Could not load the timing for tag: %v", at.TimingId))
			}
			actTmg := &ActionTiming{
				Id:     utils.GenUUID(),
				Tag:    atId,
				Weight: at.Weight,
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     t.Years,
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
				BalanceId:      apiAtr.BalanceType,
				Direction:      apiAtr.Direction,
				ThresholdType:  apiAtr.ThresholdType,
				ThresholdValue: apiAtr.ThresholdValue,
				DestinationId:  apiAtr.DestinationId,
				Weight:         apiAtr.Weight,
				ActionsId:      apiAtr.ActionsId,
			}
		}
		dbr.actionsTriggers[key] = atrs
	}
	return err
}

func (dbr *DbReader) LoadAccountActions() (err error) {
	acs, err := dbr.storDb.GetTpAccountActions(&utils.TPAccountActions{TPid: dbr.tpid})
	if err != nil {
		return err
	}
	for _, aa := range acs {
		aTriggers, exists := dbr.actionsTriggers[aa.ActionTriggersId]
		if !exists {
			return errors.New(fmt.Sprintf("Could not get action triggers for tag %v", aa.ActionTriggersId))
		}
		ub := &UserBalance{
			Type:           UB_TYPE_PREPAID,
			Id:             aa.KeyId(),
			ActionTriggers: aTriggers,
		}
		dbr.accountActions = append(dbr.accountActions, ub)
		aTimings, exists := dbr.actionsTimings[aa.ActionTimingsId]
		if !exists {
			log.Printf("Could not get action timing for tag %v", aa.ActionTimingsId)
			// must not continue here
		}
		for _, at := range aTimings {
			at.UserBalanceIds = append(at.UserBalanceIds, aa.KeyId())
		}
	}
	return nil
}

func (dbr *DbReader) LoadAccountActionsFiltered(qriedAA *utils.TPAccountActions) error {
	accountActions, err := dbr.storDb.GetTpAccountActions(qriedAA)
	if err != nil {
		return err
	}
	for _, accountAction := range accountActions {
		id := accountAction.KeyId()
		var actionsIds []string // collects action ids
		// action timings
		if accountAction.ActionTimingsId != "" {
			// get old userBalanceIds
			var exitingUserBalanceIds []string
			existingActionTimings, err := dbr.dataDb.GetActionTimings(accountAction.ActionTimingsId)
			if err == nil && len(existingActionTimings) > 0 {
				// all action timings from a specific tag shuld have the same list of user balances from the first one
				exitingUserBalanceIds = existingActionTimings[0].UserBalanceIds
			}

			actionTimingsMap, err := dbr.storDb.GetTPActionTimings(dbr.tpid, accountAction.ActionTimingsId)
			if err != nil {
				return err
			} else if len(actionTimingsMap) == 0 {
				return fmt.Errorf("No ActionTimings with id <%s>", accountAction.ActionTimingsId)
			}
			var actionTimings []*ActionTiming
			ats := actionTimingsMap[accountAction.ActionTimingsId]
			for _, at := range ats {
				// Check action exists before saving it inside actionTiming key
				// ToDo: try saving the key after the actions was retrieved in order to save one query here.
				if actions, err := dbr.storDb.GetTpActions(dbr.tpid, at.ActionsId); err != nil {
					return err
				} else if len(actions) == 0 {
					return fmt.Errorf("No Action with id <%s>", at.ActionsId)
				}
				timingsMap, err := dbr.storDb.GetTpTimings(dbr.tpid, at.TimingId)
				if err != nil {
					return err
				} else if len(timingsMap) == 0 {
					return fmt.Errorf("No Timing with id <%s>", at.TimingId)
				}
				t := timingsMap[at.TimingId]
				actTmg := &ActionTiming{
					Id:     utils.GenUUID(),
					Tag:    accountAction.ActionTimingsId,
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
			err = dbr.dataDb.SetActionTimings(accountAction.ActionTimingsId, actionTimings)
			if err != nil {
				return err
			}
		}
		// action triggers
		var actionTriggers ActionTriggerPriotityList
		//ActionTriggerPriotityList []*ActionTrigger
		if accountAction.ActionTriggersId != "" {
			apiAtrsMap, err := dbr.storDb.GetTpActionTriggers(dbr.tpid, accountAction.ActionTriggersId)
			if err != nil {
				return err
			}
			atrsMap := make(map[string][]*ActionTrigger)
			for key, atrsLst := range apiAtrsMap {
				atrs := make([]*ActionTrigger, len(atrsLst))
				for idx, apiAtr := range atrsLst {
					atrs[idx] = &ActionTrigger{Id: utils.GenUUID(),
						BalanceId:      apiAtr.BalanceType,
						Direction:      apiAtr.Direction,
						ThresholdType:  apiAtr.ThresholdType,
						ThresholdValue: apiAtr.ThresholdValue,
						DestinationId:  apiAtr.DestinationId,
						Weight:         apiAtr.Weight,
						ActionsId:      apiAtr.ActionsId,
					}
				}
				atrsMap[key] = atrs
			}
			actionTriggers = atrsMap[accountAction.ActionTriggersId]
			// collect action ids from triggers
			for _, atr := range actionTriggers {
				actionsIds = append(actionsIds, atr.ActionsId)
			}
		}

		// actions
		acts := make(map[string][]*Action)
		for _, actId := range actionsIds {
			storActs, err := dbr.storDb.GetTpActions(dbr.tpid, actId)
			if err != nil {
				return err
			}
			for tag, tpacts := range storActs {
				enacts := make([]*Action, len(tpacts))
				for idx, tpact := range tpacts {
					enacts[idx] = &Action{
						Id:               utils.GenUUID(),
						ActionType:       tpact.Identifier,
						BalanceId:        tpact.BalanceType,
						Direction:        tpact.Direction,
						Weight:           tpact.Weight,
						ExtraParameters:  tpact.ExtraParameters,
						ExpirationString: tpact.ExpiryTime,
						Balance: &Balance{
							Uuid:          utils.GenUUID(),
							Value:         tpact.Units,
							Weight:        tpact.BalanceWeight,
							RateSubject:   tpact.RatingSubject,
							DestinationId: tpact.DestinationId,
						},
					}
				}
				acts[tag] = enacts
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

		if err := dbr.dataDb.SetUserBalance(ub); err != nil {
			return err
		}
	}
	return nil
}

// Automated loading
func (dbr *DbReader) LoadAll() error {
	var err error
	if err = dbr.LoadDestinations(); err != nil {
		return err
	}
	if err = dbr.LoadTimings(); err != nil {
		return err
	}
	if err = dbr.LoadRates(); err != nil {
		return err
	}
	if err = dbr.LoadDestinationRates(); err != nil {
		return err
	}
	if err = dbr.LoadRatingPlans(); err != nil {
		return err
	}
	if err = dbr.LoadRatingProfiles(); err != nil {
		return err
	}
	if err = dbr.LoadActions(); err != nil {
		return err
	}
	if err = dbr.LoadActionTimings(); err != nil {
		return err
	}
	if err = dbr.LoadActionTriggers(); err != nil {
		return err
	}
	if err = dbr.LoadAccountActions(); err != nil {
		return err
	}
	return nil
}

// Returns the identities loaded for a specific entity category
func (dbr *DbReader) GetLoadedIds(categ string) ([]string, error) {
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
