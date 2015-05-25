/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

type DbReader struct {
	tpid      string
	storDb    LoadStorage
	dataDb    RatingStorage
	accountDb AccountingStorage
	tp        *TPData
}

func NewDbReader(storDB LoadStorage, ratingDb RatingStorage, accountDb AccountingStorage, tpid string) *DbReader {
	c := new(DbReader)
	c.storDb = storDB
	c.dataDb = ratingDb
	c.accountDb = accountDb
	c.tpid = tpid
	c.tp = NewTPData()
	return c
}

// FIXME: this method is code duplication from csv loader
func (dbr *DbReader) ShowStatistics() {
	dbr.tp.ShowStatistics()
}

func (dbr *DbReader) IsDataValid() bool {
	return dbr.tp.IsValid()
}

func (dbr *DbReader) WriteToDatabase(flush, verbose bool) (err error) {
	return dbr.tp.WriteToDatabase(dbr.dataDb, dbr.accountDb, flush, verbose)
}

func (dbr *DbReader) LoadDestinations() (err error) {
	tpDests, err := dbr.storDb.GetTpDestinations(dbr.tpid, "")
	if err == nil {
		return
	}
	return csvr.tp.LoadDestinations(tpDests)
}

func (dbr *DbReader) LoadDestinationByTag(tag string) (bool, error) {
	tpDests, err := dbr.storDb.GetTpDestinations(dbr.tpid, tag)
	dest := &Destination{Id: tag}
	for _, tpDest := range tpDests {
		dest.AddPrefix(tpDest.Prefix)
	}
	dbr.dataDb.SetDestination(dest)
	return len(tpDests) > 0, err
}

func (dbr *DbReader) LoadTimings() (err error) {
	tpTmgs, err := dbr.storDb.GetTpTimings(dbr.tpid, "")
	if err != nil {
		return err
	}
	for _, tpTm := range tpTmgs {
		dbr.tp.timings[tpTm.TimingId] = NewTiming(tpTm.TimingId, tpTm.Years, tpTm.Months, tpTm.MonthDays, tpTm.WeekDays, tpTm.Time)
	}
	return nil
}

func (dbr *DbReader) LoadRates() (err error) {
	dbr.tp.rates, err = dbr.storDb.GetTpRates(dbr.tpid, "")
	return err
}

func (dbr *DbReader) LoadDestinationRates() (err error) {
	dbr.tp.destinationRates, err = dbr.storDb.GetTpDestinationRates(dbr.tpid, "", nil)
	if err != nil {
		return err
	}
	for _, drs := range dbr.tp.destinationRates {
		for _, dr := range drs.DestinationRates {
			rate, exists := dbr.tp.rates[dr.RateId]
			if !exists {
				return fmt.Errorf("Could not find rate for tag %v", dr.RateId)
			}
			dr.Rate = rate
			destinationExists := dr.DestinationId == utils.ANY
			if !destinationExists {
				_, destinationExists = dbr.tp.destinations[dr.DestinationId]
			}
			if !destinationExists {
				if dbExists, err := dbr.dataDb.HasData(DESTINATION_PREFIX, dr.DestinationId); err != nil {
					return err
				} else if !dbExists {
					return fmt.Errorf("Could not get destination for tag %v", dr.DestinationId)
				}
			}
		}
	}
	return nil
}

func (dbr *DbReader) LoadRatingPlans() error {
	mpRpls, err := dbr.storDb.GetTpRatingPlans(dbr.tpid, "", nil)
	if err != nil {
		return err
	}
	for tag, rplBnds := range mpRpls {
		for _, rplBnd := range rplBnds {
			t, exists := dbr.tp.timings[rplBnd.TimingId]
			if !exists {
				return fmt.Errorf("Could not get timing for tag %v", rplBnd.TimingId)
			}
			rplBnd.SetTiming(t)
			drs, exists := dbr.tp.destinationRates[rplBnd.DestinationRatesId]
			if !exists {
				return fmt.Errorf("Could not find destination rate for tag %v", rplBnd.DestinationRatesId)
			}
			plan, exists := dbr.tp.ratingPlans[tag]
			if !exists {
				plan = &RatingPlan{Id: tag}
				dbr.tp.ratingPlans[plan.Id] = plan
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
		// extract aliases from subject
		aliases := strings.Split(tpRpf.Subject, ";")
		dbr.tp.dirtyRpAliases = append(dbr.tp.dirtyRpAliases, &TenantRatingSubject{Tenant: tpRpf.Tenant, Subject: aliases[0]})
		if len(aliases) > 1 {
			tpRpf.Subject = aliases[0]
			for _, alias := range aliases[1:] {
				dbr.tp.rpAliases[utils.RatingSubjectAliasKey(tpRpf.Tenant, alias)] = tpRpf.Subject
			}
		}
		rpf := &RatingProfile{Id: tpRpf.KeyId()}
		for _, tpRa := range tpRpf.RatingPlanActivations {
			at, err := utils.ParseDate(tpRa.ActivationTime)
			if err != nil {
				return fmt.Errorf("Cannot parse activation time from %v", tpRa.ActivationTime)
			}
			_, exists := dbr.tp.ratingPlans[tpRa.RatingPlanId]
			if !exists {
				if dbExists, err := dbr.dataDb.HasData(RATING_PLAN_PREFIX, tpRa.RatingPlanId); err != nil {
					return err
				} else if !dbExists {
					return fmt.Errorf("Could not load rating plans for tag: %v", tpRa.RatingPlanId)
				}
			}
			rpf.RatingPlanActivations = append(rpf.RatingPlanActivations,
				&RatingPlanActivation{
					ActivationTime:  at,
					RatingPlanId:    tpRa.RatingPlanId,
					FallbackKeys:    utils.FallbackSubjKeys(tpRpf.Direction, tpRpf.Tenant, tpRpf.Category, tpRa.FallbackSubjects),
					CdrStatQueueIds: strings.Split(tpRa.CdrStatQueueIds, utils.INFIELD_SEP),
				})
		}
		dbr.tp.ratingProfiles[tpRpf.KeyId()] = rpf
	}
	return nil
}

// Returns true, nil in case of load success, false, nil in case of RatingPlan  not found storDb
func (dbr *DbReader) LoadRatingPlanByTag(tag string) (bool, error) {
	mpRpls, err := dbr.storDb.GetTpRatingPlans(dbr.tpid, tag, nil)
	if err != nil {
		return false, err
	} else if len(mpRpls) == 0 {
		return false, nil
	}
	for tag, rplBnds := range mpRpls {
		ratingPlan := &RatingPlan{Id: tag}
		for _, rp := range rplBnds {
			tm, err := dbr.storDb.GetTpTimings(dbr.tpid, rp.TimingId)
			if err != nil || len(tm) == 0 {
				return false, fmt.Errorf("No Timings profile with id %s: %v", rp.TimingId, err)
			}
			tpTmng := NewTiming(tm[rp.TimingId].TimingId, tm[rp.TimingId].Years, tm[rp.TimingId].Months, tm[rp.TimingId].MonthDays, tm[rp.TimingId].WeekDays, tm[rp.TimingId].Time)
			rp.SetTiming(tpTmng)
			drm, err := dbr.storDb.GetTpDestinationRates(dbr.tpid, rp.DestinationRatesId, nil)
			if err != nil || len(drm) == 0 {
				return false, fmt.Errorf("No DestinationRates profile with id %s: %v", rp.DestinationRatesId, err)
			}
			for _, drate := range drm[rp.DestinationRatesId].DestinationRates {
				rt, err := dbr.storDb.GetTpRates(dbr.tpid, drate.RateId)
				if err != nil || len(rt) == 0 {
					return false, fmt.Errorf("No Rates profile with id %s: %v", drate.RateId, err)
				}
				drate.Rate = rt[drate.RateId]
				ratingPlan.AddRateInterval(drate.DestinationId, GetRateInterval(rp, drate))
				if drate.DestinationId == utils.ANY {
					continue // no need of loading the destinations in this case
				}
				dms, err := dbr.storDb.GetTpDestinations(dbr.tpid, drate.DestinationId)
				if err != nil {
					return false, err
				} else if len(dms) == 0 {
					if dbExists, err := dbr.dataDb.HasData(DESTINATION_PREFIX, drate.DestinationId); err != nil {
						return false, err
					} else if !dbExists {
						return false, fmt.Errorf("Could not get destination for tag %v", drate.DestinationId)
					}
					continue
				}
				for _, destination := range dms {
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
		return fmt.Errorf("No RateProfile for filter %v, error: %v", qriedRpf, err)
	}
	for _, tpRpf := range mpTpRpfs {
		resultRatingProfile = &RatingProfile{Id: tpRpf.KeyId()}
		for _, tpRa := range tpRpf.RatingPlanActivations {
			at, err := utils.ParseDate(tpRa.ActivationTime)
			if err != nil {
				return fmt.Errorf("Cannot parse activation time from %v", tpRa.ActivationTime)
			}
			_, exists := dbr.tp.ratingPlans[tpRa.RatingPlanId]
			if !exists {
				if dbExists, err := dbr.dataDb.HasData(RATING_PLAN_PREFIX, tpRa.RatingPlanId); err != nil {
					return err
				} else if !dbExists {
					return fmt.Errorf("Could not load rating plans for tag: %v", tpRa.RatingPlanId)
				}
			}
			resultRatingProfile.RatingPlanActivations = append(resultRatingProfile.RatingPlanActivations,
				&RatingPlanActivation{
					ActivationTime:  at,
					RatingPlanId:    tpRa.RatingPlanId,
					FallbackKeys:    utils.FallbackSubjKeys(tpRpf.Direction, tpRpf.Tenant, tpRpf.Category, tpRa.FallbackSubjects),
					CdrStatQueueIds: strings.Split(tpRa.CdrStatQueueIds, utils.INFIELD_SEP),
				})
		}
		if err := dbr.dataDb.SetRatingProfile(resultRatingProfile); err != nil {
			return err
		}
	}
	return nil
}

func (dbr *DbReader) LoadSharedGroupByTag(tag string, save bool) error {
	storSgs, err := dbr.storDb.GetTpSharedGroups(dbr.tpid, tag)
	if err != nil {
		return err
	}
	var loadedTags []string
	for tag, tpSgs := range storSgs {
		sg, exists := dbr.tp.sharedGroups[tag]
		if !exists {
			sg = &SharedGroup{
				Id:                tag,
				AccountParameters: make(map[string]*SharingParameters, len(tpSgs)),
			}
		}
		for _, tpSg := range tpSgs {
			sg.AccountParameters[tpSg.Account] = &SharingParameters{
				Strategy:      tpSg.Strategy,
				RatingSubject: tpSg.RatingSubject,
			}
		}
		dbr.tp.sharedGroups[tag] = sg
		loadedTags = append(loadedTags, tag)
	}
	if save {
		for _, tag := range loadedTags {
			if err := dbr.accountDb.SetSharedGroup(dbr.tp.sharedGroups[tag]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (dbr *DbReader) LoadSharedGroups() error {
	return dbr.LoadSharedGroupByTag("", false)
}

func (dbr *DbReader) LoadLCRs() (err error) {
	dbr.tp.lcrs, err = dbr.storDb.GetTpLCRs(dbr.tpid, "")
	return err
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
				Id:               tag + strconv.Itoa(idx),
				ActionType:       tpact.Identifier,
				BalanceType:      tpact.BalanceType,
				Direction:        tpact.Direction,
				Weight:           tpact.Weight,
				ExtraParameters:  tpact.ExtraParameters,
				ExpirationString: tpact.ExpiryTime,
				Balance: &Balance{
					Uuid:           utils.GenUUID(),
					Id:             tpact.BalanceId,
					Value:          tpact.Units,
					Weight:         tpact.BalanceWeight,
					TimingIDs:      tpact.TimingTags,
					RatingSubject:  tpact.RatingSubject,
					Category:       tpact.Category,
					DestinationIds: tpact.DestinationIds,
				},
			}
			// load action timings from tags
			if acts[idx].Balance.TimingIDs != "" {
				timingIds := strings.Split(acts[idx].Balance.TimingIDs, utils.INFIELD_SEP)
				for _, timingID := range timingIds {
					if timing, found := dbr.tp.timings[timingID]; found {
						acts[idx].Balance.Timings = append(acts[idx].Balance.Timings, &RITiming{
							Years:     timing.Years,
							Months:    timing.Months,
							MonthDays: timing.MonthDays,
							WeekDays:  timing.WeekDays,
							StartTime: timing.StartTime,
							EndTime:   timing.EndTime,
						})
					} else {
						return fmt.Errorf("Could not find timing: %v", timingID)
					}
				}
			}
		}
		dbr.tp.actions[tag] = acts
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

			_, exists := dbr.tp.actions[at.ActionsId]
			if !exists {
				return fmt.Errorf("ActionTiming: Could not load the action for tag: %v", at.ActionsId)
			}
			t, exists := dbr.tp.timings[at.TimingId]
			if !exists {
				return fmt.Errorf("ActionTiming: Could not load the timing for tag: %v", at.TimingId)
			}
			actTmg := &ActionTiming{
				Uuid:   utils.GenUUID(),
				Id:     atId,
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
			dbr.tp.actionsTimings[atId] = append(dbr.tp.actionsTimings[atId], actTmg)
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
			balance_expiration_date, _ := utils.ParseTimeDetectLayout(apiAtr.BalanceExpirationDate)
			id := apiAtr.Id
			if id == "" {
				id = utils.GenUUID()
			}
			minSleep, err := utils.ParseDurationWithSecs(apiAtr.MinSleep)
			if err != nil {
				return err
			}
			atrs[idx] = &ActionTrigger{
				Id:                    id,
				ThresholdType:         apiAtr.ThresholdType,
				ThresholdValue:        apiAtr.ThresholdValue,
				Recurrent:             apiAtr.Recurrent,
				MinSleep:              minSleep,
				BalanceId:             apiAtr.BalanceId,
				BalanceType:           apiAtr.BalanceType,
				BalanceDirection:      apiAtr.BalanceDirection,
				BalanceDestinationIds: apiAtr.BalanceDestinationIds,
				BalanceWeight:         apiAtr.BalanceWeight,
				BalanceExpirationDate: balance_expiration_date,
				BalanceTimingTags:     apiAtr.BalanceTimingTags,
				BalanceRatingSubject:  apiAtr.BalanceRatingSubject,
				BalanceCategory:       apiAtr.BalanceCategory,
				BalanceSharedGroup:    apiAtr.BalanceSharedGroup,
				Weight:                apiAtr.Weight,
				ActionsId:             apiAtr.ActionsId,
				MinQueuedItems:        apiAtr.MinQueuedItems,
			}
			if atrs[idx].Id == "" {
				atrs[idx].Id = utils.GenUUID()
			}
		}
		dbr.tp.actionsTriggers[key] = atrs
	}
	return err
}

func (dbr *DbReader) LoadAccountActions() (err error) {
	acs, err := dbr.storDb.GetTpAccountActions(&utils.TPAccountActions{TPid: dbr.tpid})
	if err != nil {
		return err
	}
	for _, aa := range acs {
		if _, alreadyDefined := dbr.tp.accountActions[aa.KeyId()]; alreadyDefined {
			return fmt.Errorf("Duplicate account action found: %s", aa.KeyId())
		}

		// extract aliases from subject
		aliases := strings.Split(aa.Account, ";")
		dbr.tp.dirtyAccAliases = append(dbr.tp.dirtyAccAliases, &TenantAccount{Tenant: aa.Tenant, Account: aliases[0]})
		if len(aliases) > 1 {
			aa.Account = aliases[0]
			for _, alias := range aliases[1:] {
				dbr.tp.accAliases[utils.AccountAliasKey(aa.Tenant, alias)] = aa.Account
			}
		}
		aTriggers, exists := dbr.tp.actionsTriggers[aa.ActionTriggersId]
		if !exists {
			return fmt.Errorf("Could not get action triggers for tag %v", aa.ActionTriggersId)
		}
		ub := &Account{
			Id:             aa.KeyId(),
			ActionTriggers: aTriggers,
		}
		dbr.tp.accountActions[aa.KeyId()] = ub
		aTimings, exists := dbr.tp.actionsTimings[aa.ActionPlanId]
		if !exists {
			log.Printf("Could not get action timing for tag %v", aa.ActionPlanId)
			// must not continue here
		}
		for _, at := range aTimings {
			at.AccountIds = append(at.AccountIds, aa.KeyId())
		}
	}
	return nil
}

func (dbr *DbReader) LoadAccountActionsFiltered(qriedAA *utils.TPAccountActions) error {
	accountActions, err := dbr.storDb.GetTpAccountActions(qriedAA)
	if err != nil {
		return errors.New(err.Error() + ": " + fmt.Sprintf("%+v", qriedAA))
	}
	for _, accountAction := range accountActions {
		id := accountAction.KeyId()
		var actionsIds []string // collects action ids
		// action timings
		if accountAction.ActionPlanId != "" {
			// get old userBalanceIds
			var exitingAccountIds []string
			existingActionTimings, err := dbr.accountDb.GetActionTimings(accountAction.ActionPlanId)
			if err == nil && len(existingActionTimings) > 0 {
				// all action timings from a specific tag shuld have the same list of user balances from the first one
				exitingAccountIds = existingActionTimings[0].AccountIds
			}

			actionTimingsMap, err := dbr.storDb.GetTPActionTimings(dbr.tpid, accountAction.ActionPlanId)
			if err != nil {
				return errors.New(err.Error() + " (ActionPlan): " + accountAction.ActionPlanId)
			} else if len(actionTimingsMap) == 0 {
				return fmt.Errorf("No ActionTimings with id <%s>", accountAction.ActionPlanId)
			}
			var actionTimings []*ActionTiming
			ats := actionTimingsMap[accountAction.ActionPlanId]
			for _, at := range ats {
				// Check action exists before saving it inside actionTiming key
				// ToDo: try saving the key after the actions was retrieved in order to save one query here.
				if actions, err := dbr.storDb.GetTpActions(dbr.tpid, at.ActionsId); err != nil {
					return errors.New(err.Error() + " (Actions): " + at.ActionsId)
				} else if len(actions) == 0 {
					return fmt.Errorf("No Action with id <%s>", at.ActionsId)
				}
				timingsMap, err := dbr.storDb.GetTpTimings(dbr.tpid, at.TimingId)
				if err != nil {
					return errors.New(err.Error() + " (Timing): " + at.TimingId)
				} else if len(timingsMap) == 0 {
					return fmt.Errorf("No Timing with id <%s>", at.TimingId)
				}
				t := NewTiming(timingsMap[at.TimingId].TimingId, timingsMap[at.TimingId].Years, timingsMap[at.TimingId].Months, timingsMap[at.TimingId].MonthDays, timingsMap[at.TimingId].WeekDays, timingsMap[at.TimingId].Time)
				actTmg := &ActionTiming{
					Uuid:   utils.GenUUID(),
					Id:     accountAction.ActionPlanId,
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
				for _, ubId := range exitingAccountIds {
					if ubId == id {
						found = true
						break
					}
				}
				if !found {
					actTmg.AccountIds = append(exitingAccountIds, id)
				}
				actionTimings = append(actionTimings, actTmg)
			}

			// write action timings
			err = dbr.accountDb.SetActionTimings(accountAction.ActionPlanId, actionTimings)
			if err != nil {
				return errors.New(err.Error() + " (SetActionPlan): " + accountAction.ActionPlanId)
			}
		}
		// action triggers
		var actionTriggers ActionTriggerPriotityList
		//ActionTriggerPriotityList []*ActionTrigger
		if accountAction.ActionTriggersId != "" {
			apiAtrsMap, err := dbr.storDb.GetTpActionTriggers(dbr.tpid, accountAction.ActionTriggersId)
			if err != nil {
				return errors.New(err.Error() + " (ActionTriggers): " + accountAction.ActionTriggersId)
			}
			atrsMap := make(map[string][]*ActionTrigger)
			for key, atrsLst := range apiAtrsMap {
				atrs := make([]*ActionTrigger, len(atrsLst))
				for idx, apiAtr := range atrsLst {
					expTime, _ := utils.ParseDate(apiAtr.BalanceExpirationDate)
					atrs[idx] = &ActionTrigger{Id: utils.GenUUID(),
						ThresholdType:         apiAtr.ThresholdType,
						ThresholdValue:        apiAtr.ThresholdValue,
						BalanceId:             apiAtr.BalanceId,
						BalanceType:           apiAtr.BalanceType,
						BalanceDirection:      apiAtr.BalanceDirection,
						BalanceDestinationIds: apiAtr.BalanceDestinationIds,
						BalanceWeight:         apiAtr.BalanceWeight,
						BalanceExpirationDate: expTime,
						BalanceRatingSubject:  apiAtr.BalanceRatingSubject,
						BalanceCategory:       apiAtr.BalanceCategory,
						BalanceSharedGroup:    apiAtr.BalanceSharedGroup,
						Weight:                apiAtr.Weight,
						ActionsId:             apiAtr.ActionsId,
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
						Id:               tag + strconv.Itoa(idx),
						ActionType:       tpact.Identifier,
						BalanceType:      tpact.BalanceType,
						Direction:        tpact.Direction,
						Weight:           tpact.Weight,
						ExtraParameters:  tpact.ExtraParameters,
						ExpirationString: tpact.ExpiryTime,
						Balance: &Balance{
							Uuid:           utils.GenUUID(),
							Value:          tpact.Units,
							Weight:         tpact.BalanceWeight,
							RatingSubject:  tpact.RatingSubject,
							DestinationIds: tpact.DestinationIds,
						},
					}
				}
				acts[tag] = enacts
			}
		}
		// writee actions
		for k, as := range acts {
			err = dbr.accountDb.SetActions(k, as)
			if err != nil {
				return err
			}
		}
		ub, err := dbr.accountDb.GetAccount(id)
		if err != nil {
			ub = &Account{
				Id: id,
			}
		}
		ub.ActionTriggers = actionTriggers

		if err := dbr.accountDb.SetAccount(ub); err != nil {
			return err
		}
	}
	return nil
}

func (dbr *DbReader) LoadDerivedChargers() (err error) {
	return dbr.LoadDerivedChargersFiltered(&utils.TPDerivedChargers{TPid: dbr.tpid}, false)
}

func (dbr *DbReader) LoadDerivedChargersFiltered(filter *utils.TPDerivedChargers, save bool) (err error) {
	tpDcses, err := dbr.storDb.GetTpDerivedChargers(filter)
	if err != nil {
		return err
	}
	for _, tpDcs := range tpDcses {
		tag := tpDcs.GetDerivedChargersKey()
		if _, hasIt := dbr.tp.derivedChargers[tag]; !hasIt {
			dbr.tp.derivedChargers[tag] = make(utils.DerivedChargers, 0) // Load object map since we use this method also from LoadDerivedChargers
		}
		for _, tpDc := range tpDcs.DerivedChargers {
			if dc, err := utils.NewDerivedCharger(tpDc.RunId, tpDc.RunFilters, tpDc.ReqTypeField, tpDc.DirectionField, tpDc.TenantField, tpDc.CategoryField,
				tpDc.AccountField, tpDc.SubjectField, tpDc.DestinationField, tpDc.SetupTimeField, tpDc.AnswerTimeField, tpDc.UsageField, tpDc.SupplierField,
				tpDc.DisconnectCauseField); err != nil {
				return err
			} else {
				dbr.tp.derivedChargers[tag] = append(dbr.tp.derivedChargers[tag], dc)
			}
		}
	}
	if save {
		for dcsKey, dcs := range dbr.tp.derivedChargers {
			if err := dbr.accountDb.SetDerivedChargers(dcsKey, dcs); err != nil {
				return err
			}
		}
	}
	return nil // Placeholder for now
}

func (dbr *DbReader) LoadCdrStatsByTag(tag string, save bool) error {
	storStats, err := dbr.storDb.GetTpCdrStats(dbr.tpid, tag)
	if err != nil {
		return err
	}
	if save && len(dbr.tp.actionsTriggers) == 0 {
		// load action triggers to check existence
		dbr.LoadActionTriggers()
	}
	var loadedTags []string
	for tag, tpStats := range storStats {
		for _, tpStat := range tpStats {
			var cs *CdrStats
			var exists bool
			if cs, exists = dbr.tp.cdrStats[tag]; !exists {
				cs = &CdrStats{Id: tag}
			}
			triggerTag := tpStat.ActionTriggers
			triggers, exists := dbr.tp.actionsTriggers[triggerTag]
			if triggerTag != "" && !exists {
				// only return error if there was something there for the tag
				return fmt.Errorf("Could not get action triggers for cdr stats id %s: %s", cs.Id, triggerTag)
			}
			UpdateCdrStats(cs, triggers, tpStat)
			dbr.tp.cdrStats[tag] = cs
			loadedTags = append(loadedTags, tag)
		}
	}
	if save {
		for _, tag := range loadedTags {
			if err := dbr.dataDb.SetCdrStats(dbr.tp.cdrStats[tag]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (dbr *DbReader) LoadCdrStats() error {
	return dbr.LoadCdrStatsByTag("", false)
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
	if err = dbr.LoadSharedGroups(); err != nil {
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
	if err = dbr.LoadDerivedChargers(); err != nil {
		return err
	}
	return nil
}

// Returns the identities loaded for a specific entity category
func (dbr *DbReader) GetLoadedIds(categ string) ([]string, error) {
	return dbr.tp.GetLoadedIds(categ)
}
