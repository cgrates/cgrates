/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/structmatcher"
	"github.com/cgrates/cgrates/utils"
)

type TpReader struct {
	tpid               string
	timezone           string
	dm                 *DataManager
	lr                 LoadReader
	actions            map[string][]*Action
	actionPlans        map[string]*ActionPlan
	actionsTriggers    map[string]ActionTriggers
	accountActions     map[string]*Account
	destinations       map[string]*Destination
	timings            map[string]*utils.TPTiming
	rates              map[string]*utils.TPRate
	destinationRates   map[string]*utils.TPDestinationRate
	ratingPlans        map[string]*RatingPlan
	ratingProfiles     map[string]*RatingProfile
	sharedGroups       map[string]*SharedGroup
	resProfiles        map[utils.TenantID]*utils.TPResourceProfile
	sqProfiles         map[utils.TenantID]*utils.TPStatProfile
	thProfiles         map[utils.TenantID]*utils.TPThresholdProfile
	filters            map[utils.TenantID]*utils.TPFilterProfile
	sppProfiles        map[utils.TenantID]*utils.TPSupplierProfile
	attributeProfiles  map[utils.TenantID]*utils.TPAttributeProfile
	chargerProfiles    map[utils.TenantID]*utils.TPChargerProfile
	dispatcherProfiles map[utils.TenantID]*utils.TPDispatcherProfile
	dispatcherHosts    map[utils.TenantID]*utils.TPDispatcherHost
	resources          []*utils.TenantID // IDs of resources which need creation based on resourceProfiles
	statQueues         []*utils.TenantID // IDs of statQueues which need creation based on statQueueProfiles
	thresholds         []*utils.TenantID // IDs of thresholds which need creation based on thresholdProfiles
	revDests,
	acntActionPlans map[string][]string
	cacheConns     []string
	schedulerConns []string
}

func NewTpReader(db DataDB, lr LoadReader, tpid, timezone string,
	cacheConns, schedulerConns []string) (*TpReader, error) {

	tpr := &TpReader{
		tpid:           tpid,
		timezone:       timezone,
		dm:             NewDataManager(db, config.CgrConfig().CacheCfg(), connMgr), // ToDo: add ChacheCfg as parameter to the NewTpReader
		lr:             lr,
		cacheConns:     cacheConns,
		schedulerConns: schedulerConns,
	}
	tpr.Init()
	//add *any and *asap timing tag (in case of no timings file)
	tpr.timings[utils.ANY] = &utils.TPTiming{
		ID:        utils.ANY,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: "00:00:00",
		EndTime:   "",
	}
	tpr.timings[utils.ASAP] = &utils.TPTiming{
		ID:        utils.ASAP,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: utils.ASAP,
		EndTime:   "",
	}
	tpr.timings[utils.MetaEveryMinute] = &utils.TPTiming{
		ID:        utils.MetaEveryMinute,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: utils.MetaEveryMinute,
		EndTime:   "",
	}
	tpr.timings[utils.MetaHourly] = &utils.TPTiming{
		ID:        utils.MetaHourly,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: utils.MetaHourly,
		EndTime:   "",
	}
	return tpr, nil
}

func (tpr *TpReader) Init() {
	tpr.actions = make(map[string][]*Action)
	tpr.actionPlans = make(map[string]*ActionPlan)
	tpr.actionsTriggers = make(map[string]ActionTriggers)
	tpr.rates = make(map[string]*utils.TPRate)
	tpr.destinations = make(map[string]*Destination)
	tpr.destinationRates = make(map[string]*utils.TPDestinationRate)
	tpr.timings = make(map[string]*utils.TPTiming)
	tpr.ratingPlans = make(map[string]*RatingPlan)
	tpr.ratingProfiles = make(map[string]*RatingProfile)
	tpr.sharedGroups = make(map[string]*SharedGroup)
	tpr.accountActions = make(map[string]*Account)
	tpr.resProfiles = make(map[utils.TenantID]*utils.TPResourceProfile)
	tpr.sqProfiles = make(map[utils.TenantID]*utils.TPStatProfile)
	tpr.thProfiles = make(map[utils.TenantID]*utils.TPThresholdProfile)
	tpr.sppProfiles = make(map[utils.TenantID]*utils.TPSupplierProfile)
	tpr.attributeProfiles = make(map[utils.TenantID]*utils.TPAttributeProfile)
	tpr.chargerProfiles = make(map[utils.TenantID]*utils.TPChargerProfile)
	tpr.dispatcherProfiles = make(map[utils.TenantID]*utils.TPDispatcherProfile)
	tpr.dispatcherHosts = make(map[utils.TenantID]*utils.TPDispatcherHost)
	tpr.filters = make(map[utils.TenantID]*utils.TPFilterProfile)
	tpr.revDests = make(map[string][]string)
	tpr.acntActionPlans = make(map[string][]string)
}

func (tpr *TpReader) LoadDestinationsFiltered(tag string) (bool, error) {
	tpDests, err := tpr.lr.GetTPDestinations(tpr.tpid, tag)
	if err != nil {
		return false, err
	} else if len(tpDests) == 0 {
		return false, nil
	}
	transID := utils.GenUUID()
	for _, tpDst := range tpDests {
		dst := NewDestinationFromTPDestination(tpDst)
		// ToDo: Fix transactions at onlineDB level
		if err = tpr.dm.SetDestination(dst, transID); err != nil {
			Cache.RollbackTransaction(transID)
		}
		if err = tpr.dm.SetReverseDestination(dst, transID); err != nil {
			Cache.RollbackTransaction(transID)
		}
	}
	Cache.CommitTransaction(transID)
	return true, nil
}

func (tpr *TpReader) LoadDestinations() (err error) {
	tps, err := tpr.lr.GetTPDestinations(tpr.tpid, "")
	if err != nil {
		return
	}
	for _, tpDst := range tps {
		tpr.destinations[tpDst.ID] = NewDestinationFromTPDestination(tpDst)
		for _, prfx := range tpr.destinations[tpDst.ID].Prefixes {
			if _, hasIt := tpr.revDests[prfx]; !hasIt {
				tpr.revDests[prfx] = make([]string, 0)
			}
			tpr.revDests[prfx] = append(tpr.revDests[prfx], tpDst.ID)
		}
	}
	return
}

func (tpr *TpReader) LoadTimings() (err error) {
	tps, err := tpr.lr.GetTPTimings(tpr.tpid, "")
	if err != nil {
		return err
	}
	tpr.timings, err = MapTPTimings(tps)
	// add *any timing tag
	tpr.timings[utils.ANY] = &utils.TPTiming{
		ID:        utils.ANY,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: "00:00:00",
		EndTime:   "",
	}
	tpr.timings[utils.ASAP] = &utils.TPTiming{
		ID:        utils.ASAP,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: utils.ASAP,
		EndTime:   "",
	}
	return err
}

func (tpr *TpReader) LoadRates() (err error) {
	tps, err := tpr.lr.GetTPRates(tpr.tpid, "")
	if err != nil {
		return err
	}
	tpr.rates, err = MapTPRates(tps)
	return err
}

func (tpr *TpReader) LoadDestinationRates() (err error) {
	tps, err := tpr.lr.GetTPDestinationRates(tpr.tpid, "", nil)
	if err != nil {
		return err
	}
	tpr.destinationRates, err = MapTPDestinationRates(tps)
	if err != nil {
		return err
	}
	for _, drs := range tpr.destinationRates {
		for _, dr := range drs.DestinationRates {
			rate, exists := tpr.rates[dr.RateId]
			if !exists {
				return fmt.Errorf("could not find rate for tag %v", dr.RateId)
			}
			dr.Rate = rate
			destinationExists := dr.DestinationId == utils.ANY
			if !destinationExists {
				_, destinationExists = tpr.destinations[dr.DestinationId]
			}
			if !destinationExists && tpr.dm.dataDB != nil {
				if destinationExists, err = tpr.dm.HasData(utils.DESTINATION_PREFIX, dr.DestinationId, ""); err != nil {
					return err
				}
			}
			if !destinationExists {
				return fmt.Errorf("could not get destination for tag %v", dr.DestinationId)
			}
		}
	}
	return nil
}

// Returns true, nil in case of load success, false, nil in case of RatingPlan  not found dataStorage
func (tpr *TpReader) LoadRatingPlansFiltered(tag string) (bool, error) {
	mpRpls, err := tpr.lr.GetTPRatingPlans(tpr.tpid, tag, nil)
	if err != nil {
		return false, err
	} else if len(mpRpls) == 0 {
		return false, nil
	}

	bindings := MapTPRatingPlanBindings(mpRpls)

	for tag, rplBnds := range bindings {
		ratingPlan := &RatingPlan{Id: tag}
		for _, rp := range rplBnds {
			tm := tpr.timings
			_, exists := tpr.timings[rp.TimingId]
			if !exists {
				tptm, err := tpr.lr.GetTPTimings(tpr.tpid, rp.TimingId)
				if err != nil || len(tptm) == 0 {
					return false, fmt.Errorf("no timing with id %s: %v", rp.TimingId, err)
				}
				tm, err = MapTPTimings(tptm)
				if err != nil {
					return false, err
				}
			}

			rp.SetTiming(tm[rp.TimingId])
			tpdrm, err := tpr.lr.GetTPDestinationRates(tpr.tpid, rp.DestinationRatesId, nil)
			if err != nil || len(tpdrm) == 0 {
				return false, fmt.Errorf("no DestinationRates profile with id %s: %v", rp.DestinationRatesId, err)
			}
			drm, err := MapTPDestinationRates(tpdrm)
			if err != nil {
				return false, err
			}
			for _, drate := range drm[rp.DestinationRatesId].DestinationRates {
				tprt, err := tpr.lr.GetTPRates(tpr.tpid, drate.RateId)
				if err != nil || len(tprt) == 0 {
					return false, fmt.Errorf("no Rates profile with id %s: %v", drate.RateId, err)
				}
				rt, err := MapTPRates(tprt)
				if err != nil {
					return false, err
				}

				drate.Rate = rt[drate.RateId]
				ratingPlan.AddRateInterval(drate.DestinationId, GetRateInterval(rp, drate))
				if drate.DestinationId == utils.ANY {
					continue // no need of loading the destinations in this case
				}
				tpDests, err := tpr.lr.GetTPDestinations(tpr.tpid, drate.DestinationId)
				if err != nil {
					return false, err
				}
				dms := make([]*Destination, len(tpDests))
				for i, tpDst := range tpDests {
					dms[i] = NewDestinationFromTPDestination(tpDst)
				}
				destsExist := len(dms) != 0
				if !destsExist && tpr.dm.dataDB != nil {
					if dbExists, err := tpr.dm.HasData(utils.DESTINATION_PREFIX, drate.DestinationId, ""); err != nil {
						return false, err
					} else if dbExists {
						destsExist = true
					}
					continue
				}
				if !destsExist {
					return false, fmt.Errorf("could not get destination for tag %v", drate.DestinationId)
				}
				for _, destination := range dms {
					tpr.dm.SetDestination(destination, utils.NonTransactional)
					tpr.dm.SetReverseDestination(destination, utils.NonTransactional)
				}
			}
		}
		if err := tpr.dm.SetRatingPlan(ratingPlan, utils.NonTransactional); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (tpr *TpReader) LoadRatingPlans() (err error) {
	tps, err := tpr.lr.GetTPRatingPlans(tpr.tpid, "", nil)
	if err != nil {
		return err
	}
	bindings := MapTPRatingPlanBindings(tps)
	for tag, rplBnds := range bindings {
		for _, rplBnd := range rplBnds {
			t, exists := tpr.timings[rplBnd.TimingId]
			if !exists {
				return fmt.Errorf("could not get timing for tag %v", rplBnd.TimingId)
			}
			rplBnd.SetTiming(t)
			drs, exists := tpr.destinationRates[rplBnd.DestinationRatesId]
			if !exists {
				return fmt.Errorf("could not find destination rate for tag %v", rplBnd.DestinationRatesId)
			}
			plan, exists := tpr.ratingPlans[tag]
			if !exists {
				plan = &RatingPlan{Id: tag}
				tpr.ratingPlans[plan.Id] = plan
			}
			for _, dr := range drs.DestinationRates {
				plan.AddRateInterval(dr.DestinationId, GetRateInterval(rplBnd, dr))
			}
		}
	}
	return nil
}

func (tpr *TpReader) LoadRatingProfilesFiltered(qriedRpf *utils.TPRatingProfile) error {
	var resultRatingProfile *RatingProfile
	mpTpRpfs, err := tpr.lr.GetTPRatingProfiles(qriedRpf)
	if err != nil {
		return fmt.Errorf("no RateProfile for filter %v, error: %v", qriedRpf, err)
	}

	rpfs, err := MapTPRatingProfiles(mpTpRpfs)
	if err != nil {
		return err
	}
	for _, tpRpf := range rpfs {
		resultRatingProfile = &RatingProfile{Id: tpRpf.KeyId()}
		for _, tpRa := range tpRpf.RatingPlanActivations {
			at, err := utils.ParseTimeDetectLayout(tpRa.ActivationTime, tpr.timezone)
			if err != nil {
				return fmt.Errorf("cannot parse activation time from %v", tpRa.ActivationTime)
			}
			_, exists := tpr.ratingPlans[tpRa.RatingPlanId]
			if !exists && tpr.dm.dataDB != nil {
				if exists, err = tpr.dm.HasData(utils.RATING_PLAN_PREFIX, tpRa.RatingPlanId, ""); err != nil {
					return err
				}
			}
			if !exists {
				return fmt.Errorf("could not load rating plans for tag: %v", tpRa.RatingPlanId)
			}
			resultRatingProfile.RatingPlanActivations = append(resultRatingProfile.RatingPlanActivations,
				&RatingPlanActivation{
					ActivationTime: at,
					RatingPlanId:   tpRa.RatingPlanId,
					FallbackKeys: utils.FallbackSubjKeys(tpRpf.Tenant,
						tpRpf.Category, tpRa.FallbackSubjects),
				})
		}
		if err := tpr.dm.SetRatingProfile(resultRatingProfile, utils.NonTransactional); err != nil {
			return err
		}
	}
	return nil
}

func (tpr *TpReader) LoadRatingProfiles() (err error) {
	tps, err := tpr.lr.GetTPRatingProfiles(&utils.TPRatingProfile{TPid: tpr.tpid})
	if err != nil {
		return err
	}
	mpTpRpfs, err := MapTPRatingProfiles(tps)
	if err != nil {
		return err
	}
	for _, tpRpf := range mpTpRpfs {
		rpf := &RatingProfile{Id: tpRpf.KeyId()}
		for _, tpRa := range tpRpf.RatingPlanActivations {
			at, err := utils.ParseTimeDetectLayout(tpRa.ActivationTime, tpr.timezone)
			if err != nil {
				return fmt.Errorf("cannot parse activation time from %v", tpRa.ActivationTime)
			}
			_, exists := tpr.ratingPlans[tpRa.RatingPlanId]
			if !exists && tpr.dm.dataDB != nil { // Only query if there is a connection, eg on dry run there is none
				if exists, err = tpr.dm.HasData(utils.RATING_PLAN_PREFIX, tpRa.RatingPlanId, ""); err != nil {
					return err
				}
			}
			if !exists {
				return fmt.Errorf("could not load rating plans for tag: %v", tpRa.RatingPlanId)
			}
			rpf.RatingPlanActivations = append(rpf.RatingPlanActivations,
				&RatingPlanActivation{
					ActivationTime: at,
					RatingPlanId:   tpRa.RatingPlanId,
					FallbackKeys: utils.FallbackSubjKeys(tpRpf.Tenant,
						tpRpf.Category, tpRa.FallbackSubjects),
				})
		}
		tpr.ratingProfiles[tpRpf.KeyId()] = rpf
	}
	return nil
}

func (tpr *TpReader) LoadSharedGroupsFiltered(tag string, save bool) (err error) {
	tps, err := tpr.lr.GetTPSharedGroups(tpr.tpid, "")
	if err != nil {
		return err
	}
	storSgs := MapTPSharedGroup(tps)
	for tag, tpSgs := range storSgs {
		sg, exists := tpr.sharedGroups[tag]
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
		tpr.sharedGroups[tag] = sg
	}
	if save {
		for _, sg := range tpr.sharedGroups {
			if err := tpr.dm.SetSharedGroup(sg, utils.NonTransactional); err != nil {
				return err
			}
		}
	}
	return nil
}

func (tpr *TpReader) LoadSharedGroups() error {
	return tpr.LoadSharedGroupsFiltered(tpr.tpid, false)
}

func (tpr *TpReader) LoadActions() (err error) {
	tps, err := tpr.lr.GetTPActions(tpr.tpid, "")
	if err != nil {
		return err
	}
	storActs := MapTPActions(tps)
	// map[string][]*Action
	for tag, tpacts := range storActs {
		acts := make([]*Action, len(tpacts))
		for idx, tpact := range tpacts {
			// check filter field
			if len(tpact.Filter) > 0 {
				if _, err := structmatcher.NewStructMatcher(tpact.Filter); err != nil {
					return fmt.Errorf("error parsing action %s filter field: %v", tag, err)
				}
			}
			acts[idx] = &Action{
				Id:               tag,
				ActionType:       tpact.Identifier,
				Weight:           tpact.Weight,
				ExtraParameters:  tpact.ExtraParameters,
				ExpirationString: tpact.ExpiryTime,
				Filter:           tpact.Filter,
				Balance:          &BalanceFilter{},
			}
			if tpact.BalanceId != "" && tpact.BalanceId != utils.ANY {
				acts[idx].Balance.ID = utils.StringPointer(tpact.BalanceId)
			}
			if tpact.BalanceType != "" && tpact.BalanceType != utils.ANY {
				acts[idx].Balance.Type = utils.StringPointer(tpact.BalanceType)
			}

			if tpact.Units != "" && tpact.Units != utils.ANY {
				vf, err := utils.ParseBalanceFilterValue(tpact.BalanceType, tpact.Units)
				if err != nil {
					return err
				}
				acts[idx].Balance.Value = vf
			}

			if tpact.BalanceWeight != "" && tpact.BalanceWeight != utils.ANY {
				u, err := strconv.ParseFloat(tpact.BalanceWeight, 64)
				if err != nil {
					return err
				}
				acts[idx].Balance.Weight = utils.Float64Pointer(u)
			}

			if tpact.RatingSubject != "" && tpact.RatingSubject != utils.ANY {
				acts[idx].Balance.RatingSubject = utils.StringPointer(tpact.RatingSubject)
			}

			if tpact.Categories != "" && tpact.Categories != utils.ANY {
				acts[idx].Balance.Categories = utils.StringMapPointer(utils.ParseStringMap(tpact.Categories))
			}
			if tpact.DestinationIds != "" && tpact.DestinationIds != utils.ANY {
				acts[idx].Balance.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(tpact.DestinationIds))
			}
			if tpact.SharedGroups != "" && tpact.SharedGroups != utils.ANY {
				acts[idx].Balance.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(tpact.SharedGroups))
			}
			if tpact.TimingTags != "" && tpact.TimingTags != utils.ANY {
				acts[idx].Balance.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(tpact.TimingTags))
			}
			if tpact.BalanceBlocker != "" && tpact.BalanceBlocker != utils.ANY {
				u, err := strconv.ParseBool(tpact.BalanceBlocker)
				if err != nil {
					return err
				}
				acts[idx].Balance.Blocker = utils.BoolPointer(u)
			}
			if tpact.BalanceDisabled != "" && tpact.BalanceDisabled != utils.ANY {
				u, err := strconv.ParseBool(tpact.BalanceDisabled)
				if err != nil {
					return err
				}
				acts[idx].Balance.Disabled = utils.BoolPointer(u)
			}

			// load action timings from tags
			if tpact.TimingTags != "" {
				timingIds := strings.Split(tpact.TimingTags, utils.INFIELD_SEP)
				for _, timingID := range timingIds {
					timing, found := tpr.timings[timingID]
					if !found {
						if timing, err = tpr.dm.GetTiming(timingID, false,
							utils.NonTransactional); err != nil {
							return fmt.Errorf("error: <%s> querying timing with id: <%s>",
								err.Error(), timingID)
						}
					}
					acts[idx].Balance.Timings = append(acts[idx].Balance.Timings, &RITiming{
						Years:     timing.Years,
						Months:    timing.Months,
						MonthDays: timing.MonthDays,
						WeekDays:  timing.WeekDays,
						StartTime: timing.StartTime,
						EndTime:   timing.EndTime,
					})
				}
			}
		}
		tpr.actions[tag] = acts
	}
	return nil
}

func (tpr *TpReader) LoadActionPlans() (err error) {
	tps, err := tpr.lr.GetTPActionPlans(tpr.tpid, "")
	if err != nil {
		return err
	}
	storAps := MapTPActionTimings(tps)
	for atId, ats := range storAps {
		for _, at := range ats {

			_, exists := tpr.actions[at.ActionsId]
			if !exists && tpr.dm.dataDB != nil {
				if exists, err = tpr.dm.HasData(utils.ACTION_PREFIX, at.ActionsId, ""); err != nil {
					return fmt.Errorf("[ActionPlans] Error querying actions: %v - %s", at.ActionsId, err.Error())
				}
			}
			if !exists {
				return fmt.Errorf("[ActionPlans] Could not load the action for tag: %v", at.ActionsId)
			}
			t, exists := tpr.timings[at.TimingId]
			if !exists {
				return fmt.Errorf("[ActionPlans] Could not load the timing for tag: %v", at.TimingId)
			}
			var actPln *ActionPlan
			if actPln, exists = tpr.actionPlans[atId]; !exists {
				actPln = &ActionPlan{
					Id: atId,
				}
			}
			actPln.ActionTimings = append(actPln.ActionTimings, &ActionTiming{
				Uuid:   utils.GenUUID(),
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
				ActionsID: at.ActionsId,
			})

			tpr.actionPlans[atId] = actPln
		}
	}
	return nil
}

func (tpr *TpReader) LoadActionTriggers() (err error) {
	tps, err := tpr.lr.GetTPActionTriggers(tpr.tpid, "")
	if err != nil {
		return err
	}
	storAts := MapTPActionTriggers(tps)
	for key, atrsLst := range storAts {
		atrs := make([]*ActionTrigger, len(atrsLst))
		for idx, atr := range atrsLst {
			expirationDate, err := utils.ParseTimeDetectLayout(atr.ExpirationDate, tpr.timezone)
			if err != nil {
				return err
			}
			activationDate, err := utils.ParseTimeDetectLayout(atr.ActivationDate, tpr.timezone)
			if err != nil {
				return err
			}
			minSleep, err := utils.ParseDurationWithNanosecs(atr.MinSleep)
			if err != nil {
				return err
			}
			if atr.UniqueID == "" {
				atr.UniqueID = utils.GenUUID()
			}
			atrs[idx] = &ActionTrigger{
				ID:             key,
				UniqueID:       atr.UniqueID,
				ThresholdType:  atr.ThresholdType,
				ThresholdValue: atr.ThresholdValue,
				Recurrent:      atr.Recurrent,
				MinSleep:       minSleep,
				ExpirationDate: expirationDate,
				ActivationDate: activationDate,
				Balance:        &BalanceFilter{},
				Weight:         atr.Weight,
				ActionsID:      atr.ActionsId,
			}
			if atr.BalanceId != "" && atr.BalanceId != utils.ANY {
				atrs[idx].Balance.ID = utils.StringPointer(atr.BalanceId)
			}

			if atr.BalanceType != "" && atr.BalanceType != utils.ANY {
				atrs[idx].Balance.Type = utils.StringPointer(atr.BalanceType)
			}

			if atr.BalanceWeight != "" && atr.BalanceWeight != utils.ANY {
				u, err := strconv.ParseFloat(atr.BalanceWeight, 64)
				if err != nil {
					return err
				}
				atrs[idx].Balance.Weight = utils.Float64Pointer(u)
			}
			if atr.BalanceExpirationDate != "" && atr.BalanceExpirationDate != utils.ANY && atr.ExpirationDate != utils.UNLIMITED {
				u, err := utils.ParseTimeDetectLayout(atr.BalanceExpirationDate, tpr.timezone)
				if err != nil {
					return err
				}
				atrs[idx].Balance.ExpirationDate = utils.TimePointer(u)
			}
			if atr.BalanceRatingSubject != "" && atr.BalanceRatingSubject != utils.ANY {
				atrs[idx].Balance.RatingSubject = utils.StringPointer(atr.BalanceRatingSubject)
			}

			if atr.BalanceCategories != "" && atr.BalanceCategories != utils.ANY {
				atrs[idx].Balance.Categories = utils.StringMapPointer(utils.ParseStringMap(atr.BalanceCategories))
			}
			if atr.BalanceDestinationIds != "" && atr.BalanceDestinationIds != utils.ANY {
				atrs[idx].Balance.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(atr.BalanceDestinationIds))
			}
			if atr.BalanceSharedGroups != "" && atr.BalanceSharedGroups != utils.ANY {
				atrs[idx].Balance.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(atr.BalanceSharedGroups))
			}
			if atr.BalanceTimingTags != "" && atr.BalanceTimingTags != utils.ANY {
				atrs[idx].Balance.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(atr.BalanceTimingTags))
			}
			if atr.BalanceBlocker != "" && atr.BalanceBlocker != utils.ANY {
				u, err := strconv.ParseBool(atr.BalanceBlocker)
				if err != nil {
					return err
				}
				atrs[idx].Balance.Blocker = utils.BoolPointer(u)
			}
			if atr.BalanceDisabled != "" && atr.BalanceDisabled != utils.ANY {
				u, err := strconv.ParseBool(atr.BalanceDisabled)
				if err != nil {
					return err
				}
				atrs[idx].Balance.Disabled = utils.BoolPointer(u)
			}
		}
		tpr.actionsTriggers[key] = atrs
	}

	return nil
}

func (tpr *TpReader) LoadAccountActionsFiltered(qriedAA *utils.TPAccountActions) error {
	accountActions, err := tpr.lr.GetTPAccountActions(qriedAA)
	if err != nil {
		return errors.New(err.Error() + ": " + fmt.Sprintf("%+v", qriedAA))
	}
	storAas, err := MapTPAccountActions(accountActions)
	if err != nil {
		return err
	}
	for _, accountAction := range storAas {
		id := accountAction.KeyId()
		var actionIDs []string // collects action ids
		// action timings
		if accountAction.ActionPlanId != "" {
			// get old userBalanceIds
			exitingAccountIds := make(utils.StringMap)
			existingActionPlan, err := tpr.dm.GetActionPlan(accountAction.ActionPlanId, false, false, utils.NonTransactional)
			if err == nil && existingActionPlan != nil {
				exitingAccountIds = existingActionPlan.AccountIDs
			}

			tpap, err := tpr.lr.GetTPActionPlans(tpr.tpid, accountAction.ActionPlanId)
			if err != nil {
				return errors.New(err.Error() + " (ActionPlan): " + accountAction.ActionPlanId)
			} else if len(tpap) == 0 {
				return fmt.Errorf("no action plan with id <%s>", accountAction.ActionPlanId)
			}
			aps := MapTPActionTimings(tpap)
			var actionPlan *ActionPlan
			ats := aps[accountAction.ActionPlanId]
			for _, at := range ats {
				// Check action exists before saving it inside actionTiming key
				// ToDo: try saving the key after the actions was retrieved in order to save one query here.
				if actions, err := tpr.lr.GetTPActions(tpr.tpid, at.ActionsId); err != nil {
					return errors.New(err.Error() + " (Actions): " + at.ActionsId)
				} else if len(actions) == 0 {
					return fmt.Errorf("no action with id <%s>", at.ActionsId)
				}
				var t *utils.TPTiming
				if at.TimingId != utils.ASAP {
					tptm, err := tpr.lr.GetTPTimings(tpr.tpid, at.TimingId)
					if err != nil {
						return errors.New(err.Error() + " (Timing): " + at.TimingId)
					} else if len(tptm) == 0 {
						return fmt.Errorf("no timing with id <%s>", at.TimingId)
					}
					tm, err := MapTPTimings(tptm)
					if err != nil {
						return err
					}
					t = tm[at.TimingId]
				} else {
					t = tpr.timings[at.TimingId] // *asap
				}
				if actionPlan == nil {
					actionPlan = &ActionPlan{
						Id: accountAction.ActionPlanId,
					}
				}
				actionPlan.ActionTimings = append(actionPlan.ActionTimings, &ActionTiming{
					Uuid:   utils.GenUUID(),
					Weight: at.Weight,
					Timing: &RateInterval{
						Timing: &RITiming{
							Months:    t.Months,
							MonthDays: t.MonthDays,
							WeekDays:  t.WeekDays,
							StartTime: t.StartTime,
						},
					},
					ActionsID: at.ActionsId,
				})
				// collect action ids from timings
				actionIDs = append(actionIDs, at.ActionsId)
				exitingAccountIds[id] = true
				actionPlan.AccountIDs = exitingAccountIds
			}
			// write tasks
			for _, at := range actionPlan.ActionTimings {
				if at.IsASAP() {
					for accID := range actionPlan.AccountIDs {
						t := &Task{
							Uuid:      utils.GenUUID(),
							AccountID: accID,
							ActionsID: at.ActionsID,
						}
						if err = tpr.dm.DataDB().PushTask(t); err != nil {
							return err
						}
					}
				}
			}
			// write action plan
			if err = tpr.dm.SetActionPlan(accountAction.ActionPlanId, actionPlan, false, utils.NonTransactional); err != nil {
				return errors.New(err.Error() + " (SetActionPlan): " + accountAction.ActionPlanId)
			}
			if err = tpr.dm.SetAccountActionPlans(id, []string{accountAction.ActionPlanId}, false); err != nil {
				return err
			}
			var reply string
			if err := connMgr.Call(tpr.cacheConns, nil,
				utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithArgDispatcher{
					TenantArg: utils.TenantArg{
						Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
					},
					AttrReloadCache: utils.AttrReloadCache{
						ArgsCache: utils.ArgsCache{AccountActionPlanIDs: []string{id},
							ActionPlanIDs: []string{accountAction.ActionPlanId}},
					},
				}, &reply); err != nil {
				return err
			}

		}
		// action triggers
		var actionTriggers ActionTriggers
		//ActionTriggerPriotityList []*ActionTrigger
		if accountAction.ActionTriggersId != "" {
			tpatrs, err := tpr.lr.GetTPActionTriggers(tpr.tpid, accountAction.ActionTriggersId)
			if err != nil {
				return errors.New(err.Error() + " (ActionTriggers): " + accountAction.ActionTriggersId)
			}
			atrs := MapTPActionTriggers(tpatrs)
			atrsMap := make(map[string][]*ActionTrigger)
			for key, atrsLst := range atrs {
				atrs := make([]*ActionTrigger, len(atrsLst))
				for idx, atr := range atrsLst {
					minSleep, _ := utils.ParseDurationWithNanosecs(atr.MinSleep)
					expTime, _ := utils.ParseTimeDetectLayout(atr.ExpirationDate, tpr.timezone)
					actTime, _ := utils.ParseTimeDetectLayout(atr.ActivationDate, tpr.timezone)
					if atr.UniqueID == "" {
						atr.UniqueID = utils.GenUUID()
					}
					atrs[idx] = &ActionTrigger{
						ID:             key,
						UniqueID:       atr.UniqueID,
						ThresholdType:  atr.ThresholdType,
						ThresholdValue: atr.ThresholdValue,
						Recurrent:      atr.Recurrent,
						MinSleep:       minSleep,
						ExpirationDate: expTime,
						ActivationDate: actTime,
						Balance:        &BalanceFilter{},
						Weight:         atr.Weight,
						ActionsID:      atr.ActionsId,
					}
					if atr.BalanceId != "" && atr.BalanceId != utils.ANY {
						atrs[idx].Balance.ID = utils.StringPointer(atr.BalanceId)
					}

					if atr.BalanceType != "" && atr.BalanceType != utils.ANY {
						atrs[idx].Balance.Type = utils.StringPointer(atr.BalanceType)
					}

					if atr.BalanceWeight != "" && atr.BalanceWeight != utils.ANY {
						u, err := strconv.ParseFloat(atr.BalanceWeight, 64)
						if err != nil {
							return err
						}
						atrs[idx].Balance.Weight = utils.Float64Pointer(u)
					}
					if atr.BalanceExpirationDate != "" && atr.BalanceExpirationDate != utils.ANY && atr.ExpirationDate != utils.UNLIMITED {
						u, err := utils.ParseTimeDetectLayout(atr.BalanceExpirationDate, tpr.timezone)
						if err != nil {
							return err
						}
						atrs[idx].Balance.ExpirationDate = utils.TimePointer(u)
					}
					if atr.BalanceRatingSubject != "" && atr.BalanceRatingSubject != utils.ANY {
						atrs[idx].Balance.RatingSubject = utils.StringPointer(atr.BalanceRatingSubject)
					}

					if atr.BalanceCategories != "" && atr.BalanceCategories != utils.ANY {
						atrs[idx].Balance.Categories = utils.StringMapPointer(utils.ParseStringMap(atr.BalanceCategories))
					}
					if atr.BalanceDestinationIds != "" && atr.BalanceDestinationIds != utils.ANY {
						atrs[idx].Balance.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(atr.BalanceDestinationIds))
					}
					if atr.BalanceSharedGroups != "" && atr.BalanceSharedGroups != utils.ANY {
						atrs[idx].Balance.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(atr.BalanceSharedGroups))
					}
					if atr.BalanceTimingTags != "" && atr.BalanceTimingTags != utils.ANY {
						atrs[idx].Balance.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(atr.BalanceTimingTags))
					}
					if atr.BalanceBlocker != "" && atr.BalanceBlocker != utils.ANY {
						u, err := strconv.ParseBool(atr.BalanceBlocker)
						if err != nil {
							return err
						}
						atrs[idx].Balance.Blocker = utils.BoolPointer(u)
					}
					if atr.BalanceDisabled != "" && atr.BalanceDisabled != utils.ANY {
						u, err := strconv.ParseBool(atr.BalanceDisabled)
						if err != nil {
							return err
						}
						atrs[idx].Balance.Disabled = utils.BoolPointer(u)
					}
				}
				atrsMap[key] = atrs
			}
			actionTriggers = atrsMap[accountAction.ActionTriggersId]
			// collect action ids from triggers
			for _, atr := range actionTriggers {
				actionIDs = append(actionIDs, atr.ActionsID)
			}
			// write action triggers
			err = tpr.dm.SetActionTriggers(accountAction.ActionTriggersId, actionTriggers, utils.NonTransactional)
			if err != nil {
				return errors.New(err.Error() + " (SetActionTriggers): " + accountAction.ActionTriggersId)
			}
		}

		// actions
		facts := make(map[string][]*Action)
		for _, actId := range actionIDs {
			tpas, err := tpr.lr.GetTPActions(tpr.tpid, actId)
			if err != nil {
				return err
			}
			as := MapTPActions(tpas)
			for tag, tpacts := range as {
				acts := make([]*Action, len(tpacts))
				for idx, tpact := range tpacts {
					// check filter field
					if len(tpact.Filter) > 0 {
						if _, err := structmatcher.NewStructMatcher(tpact.Filter); err != nil {
							return fmt.Errorf("error parsing action %s filter field: %v", tag, err)
						}
					}
					acts[idx] = &Action{
						Id:         tag,
						ActionType: tpact.Identifier,
						//BalanceType:      tpact.BalanceType,
						Weight:           tpact.Weight,
						ExtraParameters:  tpact.ExtraParameters,
						ExpirationString: tpact.ExpiryTime,
						Filter:           tpact.Filter,
						Balance:          &BalanceFilter{},
					}
					if tpact.BalanceId != "" && tpact.BalanceId != utils.ANY {
						acts[idx].Balance.ID = utils.StringPointer(tpact.BalanceId)
					}
					if tpact.BalanceType != "" && tpact.BalanceType != utils.ANY {
						acts[idx].Balance.Type = utils.StringPointer(tpact.BalanceType)
					}

					if tpact.Units != "" && tpact.Units != utils.ANY {
						vf, err := utils.ParseBalanceFilterValue(tpact.BalanceType, tpact.Units)
						if err != nil {
							return err
						}
						acts[idx].Balance.Value = vf
					}

					if tpact.BalanceWeight != "" && tpact.BalanceWeight != utils.ANY {
						u, err := strconv.ParseFloat(tpact.BalanceWeight, 64)
						if err != nil {
							return err
						}
						acts[idx].Balance.Weight = utils.Float64Pointer(u)
					}
					if tpact.RatingSubject != "" && tpact.RatingSubject != utils.ANY {
						acts[idx].Balance.RatingSubject = utils.StringPointer(tpact.RatingSubject)
					}

					if tpact.Categories != "" && tpact.Categories != utils.ANY {
						acts[idx].Balance.Categories = utils.StringMapPointer(utils.ParseStringMap(tpact.Categories))
					}
					if tpact.DestinationIds != "" && tpact.DestinationIds != utils.ANY {
						acts[idx].Balance.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(tpact.DestinationIds))
					}
					if tpact.SharedGroups != "" && tpact.SharedGroups != utils.ANY {
						acts[idx].Balance.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(tpact.SharedGroups))
					}
					if tpact.TimingTags != "" && tpact.TimingTags != utils.ANY {
						acts[idx].Balance.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(tpact.TimingTags))
					}
					if tpact.BalanceBlocker != "" && tpact.BalanceBlocker != utils.ANY {
						u, err := strconv.ParseBool(tpact.BalanceBlocker)
						if err != nil {
							return err
						}
						acts[idx].Balance.Blocker = utils.BoolPointer(u)
					}
					if tpact.BalanceDisabled != "" && tpact.BalanceDisabled != utils.ANY {
						u, err := strconv.ParseBool(tpact.BalanceDisabled)
						if err != nil {
							return err
						}
						acts[idx].Balance.Disabled = utils.BoolPointer(u)
					}
					// load action timings from tags
					if tpact.TimingTags != "" {
						timingIds := strings.Split(tpact.TimingTags, utils.INFIELD_SEP)
						for _, timingID := range timingIds {
							if timing, found := tpr.timings[timingID]; found {
								acts[idx].Balance.Timings = append(acts[idx].Balance.Timings, &RITiming{
									Years:     timing.Years,
									Months:    timing.Months,
									MonthDays: timing.MonthDays,
									WeekDays:  timing.WeekDays,
									StartTime: timing.StartTime,
									EndTime:   timing.EndTime,
								})
							} else {
								return fmt.Errorf("could not find timing: %v", timingID)
							}
						}
					}
				}
				facts[tag] = acts
			}
		}
		// write actions
		for k, as := range facts {
			err = tpr.dm.SetActions(k, as, utils.NonTransactional)
			if err != nil {
				return err
			}
		}
		ub, err := tpr.dm.GetAccount(id)
		if err != nil {
			ub = &Account{
				ID: id,
			}
		}
		ub.ActionTriggers = actionTriggers
		// init counters
		ub.InitCounters()
		if err := tpr.dm.SetAccount(ub); err != nil {
			return err
		}
	}
	return nil
}

func (tpr *TpReader) LoadAccountActions() (err error) {
	tps, err := tpr.lr.GetTPAccountActions(&utils.TPAccountActions{TPid: tpr.tpid})
	if err != nil {
		return err
	}
	storAts, err := MapTPAccountActions(tps)
	if err != nil {
		return err
	}

	for _, aa := range storAts {
		aaKeyID := aa.KeyId()
		if _, alreadyDefined := tpr.accountActions[aa.KeyId()]; alreadyDefined {
			return fmt.Errorf("duplicate account action found: %s", aaKeyID)
		}
		var aTriggers ActionTriggers
		if aa.ActionTriggersId != "" {
			var exists bool
			if aTriggers, exists = tpr.actionsTriggers[aa.ActionTriggersId]; !exists {
				return fmt.Errorf("could not get action triggers for tag %s", aa.ActionTriggersId)
			}
		}
		ub := &Account{
			ID:             aaKeyID,
			ActionTriggers: aTriggers,
			AllowNegative:  aa.AllowNegative,
			Disabled:       aa.Disabled,
		}
		ub.InitCounters()
		tpr.accountActions[aaKeyID] = ub
		if aa.ActionPlanId != "" {
			actionPlan, exists := tpr.actionPlans[aa.ActionPlanId]
			if !exists {
				return fmt.Errorf("could not get action plan for tag %v", aa.ActionPlanId)
			}
			if actionPlan.AccountIDs == nil {
				actionPlan.AccountIDs = make(utils.StringMap)
			}

			actionPlan.AccountIDs[aaKeyID] = true
			if _, hasKey := tpr.acntActionPlans[aaKeyID]; !hasKey {
				tpr.acntActionPlans[aaKeyID] = make([]string, 0)
			}
			tpr.acntActionPlans[aaKeyID] = append(tpr.acntActionPlans[aaKeyID], aa.ActionPlanId)
		}
	}
	return nil
}

func (tpr *TpReader) LoadResourceProfilesFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPResources(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapRsPfls := make(map[utils.TenantID]*utils.TPResourceProfile)
	for _, rl := range rls {
		mapRsPfls[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.resProfiles = mapRsPfls
	for tntID := range mapRsPfls {
		if has, err := tpr.dm.HasData(utils.ResourcesPrefix, tntID.ID, tntID.Tenant); err != nil {
			return err
		} else if !has {
			tpr.resources = append(tpr.resources, &utils.TenantID{Tenant: tntID.Tenant, ID: tntID.ID})
		}
	}
	return nil
}

func (tpr *TpReader) LoadResourceProfiles() error {
	return tpr.LoadResourceProfilesFiltered("")
}

func (tpr *TpReader) LoadStatsFiltered(tag string) (err error) {
	tps, err := tpr.lr.GetTPStats(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapSTs := make(map[utils.TenantID]*utils.TPStatProfile)
	for _, st := range tps {
		mapSTs[utils.TenantID{Tenant: st.Tenant, ID: st.ID}] = st
	}
	tpr.sqProfiles = mapSTs
	for tntID := range mapSTs {
		if has, err := tpr.dm.HasData(utils.StatQueuePrefix, tntID.ID, tntID.Tenant); err != nil {
			return err
		} else if !has {
			tpr.statQueues = append(tpr.statQueues, &utils.TenantID{Tenant: tntID.Tenant, ID: tntID.ID})
		}
	}
	return nil
}

func (tpr *TpReader) LoadStats() error {
	return tpr.LoadStatsFiltered("")
}

func (tpr *TpReader) LoadThresholdsFiltered(tag string) (err error) {
	tps, err := tpr.lr.GetTPThresholds(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapTHs := make(map[utils.TenantID]*utils.TPThresholdProfile)
	for _, th := range tps {
		mapTHs[utils.TenantID{Tenant: th.Tenant, ID: th.ID}] = th
	}
	tpr.thProfiles = mapTHs
	for tntID := range mapTHs {
		if has, err := tpr.dm.HasData(utils.ThresholdPrefix, tntID.ID, tntID.Tenant); err != nil {
			return err
		} else if !has {
			tpr.thresholds = append(tpr.thresholds, &utils.TenantID{Tenant: tntID.Tenant, ID: tntID.ID})
		}
	}
	return nil
}

func (tpr *TpReader) LoadThresholds() error {
	return tpr.LoadThresholdsFiltered("")
}

func (tpr *TpReader) LoadFiltersFiltered(tag string) error {
	tps, err := tpr.lr.GetTPFilters(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapTHs := make(map[utils.TenantID]*utils.TPFilterProfile)
	for _, th := range tps {
		mapTHs[utils.TenantID{Tenant: th.Tenant, ID: th.ID}] = th
	}
	tpr.filters = mapTHs
	return nil
}

func (tpr *TpReader) LoadFilters() error {
	return tpr.LoadFiltersFiltered("")
}

func (tpr *TpReader) LoadSupplierProfilesFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPSuppliers(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapRsPfls := make(map[utils.TenantID]*utils.TPSupplierProfile)
	for _, rl := range rls {
		mapRsPfls[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.sppProfiles = mapRsPfls
	return nil
}

func (tpr *TpReader) LoadSupplierProfiles() error {
	return tpr.LoadSupplierProfilesFiltered("")
}

func (tpr *TpReader) LoadAttributeProfilesFiltered(tag string) (err error) {
	attrs, err := tpr.lr.GetTPAttributes(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapAttrPfls := make(map[utils.TenantID]*utils.TPAttributeProfile)
	for _, attr := range attrs {
		mapAttrPfls[utils.TenantID{Tenant: attr.Tenant, ID: attr.ID}] = attr
	}
	tpr.attributeProfiles = mapAttrPfls
	return nil
}

func (tpr *TpReader) LoadAttributeProfiles() error {
	return tpr.LoadAttributeProfilesFiltered("")
}

func (tpr *TpReader) LoadChargerProfilesFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPChargers(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapChargerProfile := make(map[utils.TenantID]*utils.TPChargerProfile)
	for _, rl := range rls {
		mapChargerProfile[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.chargerProfiles = mapChargerProfile
	return nil
}

func (tpr *TpReader) LoadChargerProfiles() error {
	return tpr.LoadChargerProfilesFiltered("")
}

func (tpr *TpReader) LoadDispatcherProfilesFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPDispatcherProfiles(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapDispatcherProfile := make(map[utils.TenantID]*utils.TPDispatcherProfile)
	for _, rl := range rls {
		mapDispatcherProfile[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.dispatcherProfiles = mapDispatcherProfile
	return nil
}

func (tpr *TpReader) LoadDispatcherProfiles() error {
	return tpr.LoadDispatcherProfilesFiltered("")
}

func (tpr *TpReader) LoadDispatcherHostsFiltered(tag string) (err error) {
	rls, err := tpr.lr.GetTPDispatcherHosts(tpr.tpid, "", tag)
	if err != nil {
		return err
	}
	mapDispatcherHost := make(map[utils.TenantID]*utils.TPDispatcherHost)
	for _, rl := range rls {
		mapDispatcherHost[utils.TenantID{Tenant: rl.Tenant, ID: rl.ID}] = rl
	}
	tpr.dispatcherHosts = mapDispatcherHost
	return nil
}

func (tpr *TpReader) LoadDispatcherHosts() error {
	return tpr.LoadDispatcherHostsFiltered("")
}

func (tpr *TpReader) LoadAll() (err error) {
	if err = tpr.LoadDestinations(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadTimings(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadRates(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadDestinationRates(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadRatingPlans(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadRatingProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadSharedGroups(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadActions(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadActionPlans(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadActionTriggers(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadAccountActions(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadFilters(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadResourceProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadStats(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadThresholds(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadSupplierProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadAttributeProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadChargerProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadDispatcherProfiles(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	if err = tpr.LoadDispatcherHosts(); err != nil && err.Error() != utils.NotFoundCaps {
		return
	}
	return nil
}

func (tpr *TpReader) IsValid() bool {
	valid := true
	for rplTag, rpl := range tpr.ratingPlans {
		if !rpl.isContinous() {
			log.Printf("The rating plan %s is not covering all weekdays", rplTag)
			valid = false
		}
		if crazyRate := rpl.getFirstUnsaneRating(); crazyRate != "" {
			log.Printf("The rate %s is invalid", crazyRate)
			valid = false
		}
		if crazyTiming := rpl.getFirstUnsaneTiming(); crazyTiming != "" {
			log.Printf("The timing %s is invalid", crazyTiming)
			valid = false
		}
	}
	return valid
}

func (tpr *TpReader) WriteToDatabase(verbose, disable_reverse bool) (err error) {
	if tpr.dm.dataDB == nil {
		return errors.New("no database connection")
	}
	//generate a loadID
	loadID := time.Now().UnixNano()
	loadIDs := make(map[string]int64)
	if verbose {
		log.Print("Destinations:")
	}
	for _, d := range tpr.destinations {
		err = tpr.dm.SetDestination(d, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", d.Id, " : ", d.Prefixes)
		}
	}
	if len(tpr.destinations) != 0 {
		loadIDs[utils.CacheDestinations] = loadID
	}
	if len(tpr.revDests) != 0 {
		loadIDs[utils.CacheReverseDestinations] = loadID
	}
	if verbose {
		log.Print("Reverse Destinations:")
		for id, vals := range tpr.revDests {
			log.Printf("\t %s : %+v", id, vals)
		}
		log.Print("Rating Plans:")
	}
	for _, rp := range tpr.ratingPlans {
		err = tpr.dm.SetRatingPlan(rp, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", rp.Id)
		}
	}
	if len(tpr.ratingPlans) != 0 {
		loadIDs[utils.CacheRatingPlans] = loadID
	}
	if verbose {
		log.Print("Rating Profiles:")
	}
	for _, rp := range tpr.ratingProfiles {
		err = tpr.dm.SetRatingProfile(rp, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", rp.Id)
		}
	}
	if len(tpr.ratingProfiles) != 0 {
		loadIDs[utils.CacheRatingProfiles] = loadID
	}
	if verbose {
		log.Print("Action Plans:")
	}
	for k, ap := range tpr.actionPlans {
		for _, at := range ap.ActionTimings {
			if at.IsASAP() {
				for accID := range ap.AccountIDs {
					t := &Task{
						Uuid:      utils.GenUUID(),
						AccountID: accID,
						ActionsID: at.ActionsID,
					}
					if verbose {
						log.Println("\tTask: ", t)
					}
					if err = tpr.dm.DataDB().PushTask(t); err != nil {
						return err
					}
				}
				if len(ap.AccountIDs) == 0 {
					t := &Task{
						Uuid:      utils.GenUUID(),
						ActionsID: at.ActionsID,
					}
					if verbose {
						log.Println("\tTask: ", t)
					}
					if err = tpr.dm.DataDB().PushTask(t); err != nil {
						return err
					}
				}
			}
		}
		err = tpr.dm.SetActionPlan(k, ap, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if len(tpr.actionPlans) != 0 {
		loadIDs[utils.CacheActionPlans] = loadID
	}
	if len(tpr.acntActionPlans) != 0 {
		loadIDs[utils.CacheAccountActionPlans] = loadID
	}
	if verbose {
		log.Print("Account Action Plans:")
		for id, vals := range tpr.acntActionPlans {
			log.Printf("\t %s : %+v", id, vals)
		}
	}
	if verbose {
		log.Print("Action Triggers:")
	}
	for k, atrs := range tpr.actionsTriggers {
		err = tpr.dm.SetActionTriggers(k, atrs, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if len(tpr.actionsTriggers) != 0 {
		loadIDs[utils.CacheActionTriggers] = loadID
	}
	if verbose {
		log.Print("Shared Groups:")
	}
	for k, sg := range tpr.sharedGroups {
		err = tpr.dm.SetSharedGroup(sg, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if len(tpr.sharedGroups) != 0 {
		loadIDs[utils.CacheSharedGroups] = loadID
	}
	if verbose {
		log.Print("Actions:")
	}
	for k, as := range tpr.actions {
		err = tpr.dm.SetActions(k, as, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if len(tpr.actions) != 0 {
		loadIDs[utils.CacheActions] = loadID
	}
	if verbose {
		log.Print("Account Actions:")
	}
	for _, ub := range tpr.accountActions {
		err = tpr.dm.SetAccount(ub)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", ub.ID)
		}
	}
	if verbose {
		log.Print("Filters:")
	}
	for _, tpTH := range tpr.filters {
		th, err := APItoFilter(tpTH, tpr.timezone)
		if err != nil {
			return err
		}
		if err = tpr.dm.SetFilter(th); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.filters) != 0 {
		loadIDs[utils.CacheFilters] = loadID
	}
	if verbose {
		log.Print("ResourceProfiles:")
	}
	for _, tpRsp := range tpr.resProfiles {
		rsp, err := APItoResource(tpRsp, tpr.timezone)
		if err != nil {
			return err
		}
		if err = tpr.dm.SetResourceProfile(rsp, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", rsp.TenantID())
		}
	}
	if len(tpr.resProfiles) != 0 {
		loadIDs[utils.CacheResourceProfiles] = loadID
	}
	if verbose {
		log.Print("Resources:")
	}
	for _, rTid := range tpr.resources {
		if err = tpr.dm.SetResource(&Resource{Tenant: rTid.Tenant, ID: rTid.ID, Usages: make(map[string]*ResourceUsage)}); err != nil {
			return
		}
		if verbose {
			log.Print("\t", rTid.TenantID())
		}
	}
	if len(tpr.resources) != 0 {
		loadIDs[utils.CacheResources] = loadID
	}
	if verbose {
		log.Print("StatQueueProfiles:")
	}
	for _, tpST := range tpr.sqProfiles {
		st, err := APItoStats(tpST, tpr.timezone)
		if err != nil {
			return err
		}
		if err = tpr.dm.SetStatQueueProfile(st, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", st.TenantID())
		}
	}
	if len(tpr.sqProfiles) != 0 {
		loadIDs[utils.CacheStatQueueProfiles] = loadID
	}
	if verbose {
		log.Print("StatQueues:")
	}
	for _, sqTntID := range tpr.statQueues {
		metrics := make(map[string]StatMetric)
		for _, metric := range tpr.sqProfiles[utils.TenantID{Tenant: sqTntID.Tenant, ID: sqTntID.ID}].Metrics {
			if stsMetric, err := NewStatMetric(metric.MetricID,
				tpr.sqProfiles[utils.TenantID{Tenant: sqTntID.Tenant, ID: sqTntID.ID}].MinItems,
				metric.FilterIDs); err != nil {
				return err
			} else {
				metrics[metric.MetricID] = stsMetric
			}
		}
		sq := &StatQueue{Tenant: sqTntID.Tenant, ID: sqTntID.ID, SQMetrics: metrics}
		if err = tpr.dm.SetStatQueue(sq); err != nil {
			return
		}
		if verbose {
			log.Print("\t", sqTntID.TenantID())
		}
	}
	if len(tpr.statQueues) != 0 {
		loadIDs[utils.CacheStatQueues] = loadID
	}
	if verbose {
		log.Print("ThresholdProfiles:")
	}
	for _, tpTH := range tpr.thProfiles {
		th, err := APItoThresholdProfile(tpTH, tpr.timezone)
		if err != nil {
			return err
		}
		if err = tpr.dm.SetThresholdProfile(th, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.thProfiles) != 0 {
		loadIDs[utils.CacheThresholdProfiles] = loadID
	}
	if verbose {
		log.Print("Thresholds:")
	}
	for _, thd := range tpr.thresholds {
		if err = tpr.dm.SetThreshold(&Threshold{Tenant: thd.Tenant, ID: thd.ID}); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", thd.TenantID())
		}
	}
	if len(tpr.thresholds) != 0 {
		loadIDs[utils.CacheThresholds] = loadID
	}
	if verbose {
		log.Print("SupplierProfiles:")
	}
	for _, tpTH := range tpr.sppProfiles {
		th, err := APItoSupplierProfile(tpTH, tpr.timezone)
		if err != nil {
			return err
		}
		if err = tpr.dm.SetSupplierProfile(th, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.sppProfiles) != 0 {
		loadIDs[utils.CacheSupplierProfiles] = loadID
	}
	if verbose {
		log.Print("AttributeProfiles:")
	}
	for _, tpTH := range tpr.attributeProfiles {
		th, err := APItoAttributeProfile(tpTH, tpr.timezone)
		if err != nil {
			return err
		}
		if err = tpr.dm.SetAttributeProfile(th, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.attributeProfiles) != 0 {
		loadIDs[utils.CacheAttributeProfiles] = loadID
	}
	if verbose {
		log.Print("ChargerProfiles:")
	}
	for _, tpTH := range tpr.chargerProfiles {

		th, err := APItoChargerProfile(tpTH, tpr.timezone)
		if err != nil {
			return err
		}
		if err = tpr.dm.SetChargerProfile(th, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.chargerProfiles) != 0 {
		loadIDs[utils.CacheChargerProfiles] = loadID
	}
	if verbose {
		log.Print("DispatcherProfiles:")
	}
	for _, tpTH := range tpr.dispatcherProfiles {
		th, err := APItoDispatcherProfile(tpTH, tpr.timezone)
		if err != nil {
			return err
		}
		if err = tpr.dm.SetDispatcherProfile(th, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.dispatcherProfiles) != 0 {
		loadIDs[utils.CacheDispatcherProfiles] = loadID
	}
	if verbose {
		log.Print("DispatcherHosts:")
	}
	for _, tpTH := range tpr.dispatcherHosts {
		th := APItoDispatcherHost(tpTH)
		if err = tpr.dm.SetDispatcherHost(th); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", th.TenantID())
		}
	}
	if len(tpr.dispatcherHosts) != 0 {
		loadIDs[utils.CacheDispatcherHosts] = loadID
	}

	if verbose {
		log.Print("Timings:")
	}
	for _, t := range tpr.timings {
		if err = tpr.dm.SetTiming(t); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", t.ID)
		}
	}
	if len(tpr.timings) != 0 {
		loadIDs[utils.CacheTimings] = loadID
	}
	if !disable_reverse {
		if len(tpr.destinations) > 0 {
			if verbose {
				log.Print("Rebuilding reverse destinations")
			}
			if err = tpr.dm.RebuildReverseForPrefix(utils.REVERSE_DESTINATION_PREFIX); err != nil {
				return err
			}
		}
		if len(tpr.acntActionPlans) > 0 {
			if verbose {
				log.Print("Rebuilding account action plans")
			}
			if err = tpr.dm.RebuildReverseForPrefix(utils.AccountActionPlansPrefix); err != nil {
				return err
			}
		}
	}
	if err = tpr.dm.SetLoadIDs(loadIDs); err != nil {
		return err
	}
	return
}

func (tpr *TpReader) ShowStatistics() {
	// destinations
	destCount := len(tpr.destinations)
	log.Print("Destinations: ", destCount)
	prefixDist := make(map[int]int, 50)
	prefixCount := 0
	for _, d := range tpr.destinations {
		prefixDist[len(d.Prefixes)] += 1
		prefixCount += len(d.Prefixes)
	}
	log.Print("Avg Prefixes: ", prefixCount/destCount)
	log.Print("Prefixes distribution:")
	for k, v := range prefixDist {
		log.Printf("%d: %d", k, v)
	}
	// rating plans
	rplCount := len(tpr.ratingPlans)
	log.Print("Rating plans: ", rplCount)
	destRatesDist := make(map[int]int, 50)
	destRatesCount := 0
	for _, rpl := range tpr.ratingPlans {
		destRatesDist[len(rpl.DestinationRates)] += 1
		destRatesCount += len(rpl.DestinationRates)
	}
	log.Print("Avg Destination Rates: ", destRatesCount/rplCount)
	log.Print("Destination Rates distribution:")
	for k, v := range destRatesDist {
		log.Printf("%d: %d", k, v)
	}
	// rating profiles
	rpfCount := len(tpr.ratingProfiles)
	log.Print("Rating profiles: ", rpfCount)
	activDist := make(map[int]int, 50)
	activCount := 0
	for _, rpf := range tpr.ratingProfiles {
		activDist[len(rpf.RatingPlanActivations)] += 1
		activCount += len(rpf.RatingPlanActivations)
	}
	log.Print("Avg Activations: ", activCount/rpfCount)
	log.Print("Activation distribution:")
	for k, v := range activDist {
		log.Printf("%d: %d", k, v)
	}
	// actions
	log.Print("Actions: ", len(tpr.actions))
	// action plans
	log.Print("Action plans: ", len(tpr.actionPlans))
	// action triggers
	log.Print("Action triggers: ", len(tpr.actionsTriggers))
	// account actions
	log.Print("Account actions: ", len(tpr.accountActions))
	// resource profiles
	log.Print("ResourceProfiles: ", len(tpr.resProfiles))
	// stats
	log.Print("Stats: ", len(tpr.sqProfiles))
	// thresholds
	log.Print("Thresholds: ", len(tpr.thProfiles))
	// filters
	log.Print("Filters: ", len(tpr.filters))
	// Supplier profiles
	log.Print("SupplierProfiles: ", len(tpr.sppProfiles))
	// Attribute profiles
	log.Print("AttributeProfiles: ", len(tpr.attributeProfiles))
	// Charger profiles
	log.Print("ChargerProfiles: ", len(tpr.chargerProfiles))
	// Dispatcher profiles
	log.Print("DispatcherProfiles: ", len(tpr.dispatcherProfiles))
	// Dispatcher Hosts
	log.Print("DispatcherHosts: ", len(tpr.dispatcherHosts))
}

// Returns the identities loaded for a specific category, useful for cache reloads
func (tpr *TpReader) GetLoadedIds(categ string) ([]string, error) {
	switch categ {
	case utils.DESTINATION_PREFIX:
		keys := make([]string, len(tpr.destinations))
		i := 0
		for k := range tpr.destinations {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.REVERSE_DESTINATION_PREFIX:
		keys := make([]string, len(tpr.revDests))
		i := 0
		for k := range tpr.revDests {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.RATING_PLAN_PREFIX:
		keys := make([]string, len(tpr.ratingPlans))
		i := 0
		for k := range tpr.ratingPlans {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.RATING_PROFILE_PREFIX:
		keys := make([]string, len(tpr.ratingProfiles))
		i := 0
		for k := range tpr.ratingProfiles {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.ACTION_PREFIX:
		keys := make([]string, len(tpr.actions))
		i := 0
		for k := range tpr.actions {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.ACTION_PLAN_PREFIX: // actionPlans
		keys := make([]string, len(tpr.actionPlans))
		i := 0
		for k := range tpr.actionPlans {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.AccountActionPlansPrefix:
		keys := make([]string, len(tpr.acntActionPlans))
		i := 0
		for k := range tpr.acntActionPlans {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.SHARED_GROUP_PREFIX:
		keys := make([]string, len(tpr.sharedGroups))
		i := 0
		for k := range tpr.sharedGroups {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.ResourceProfilesPrefix:
		keys := make([]string, len(tpr.resProfiles))
		i := 0
		for k := range tpr.resProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.ACTION_TRIGGER_PREFIX:
		keys := make([]string, len(tpr.actionsTriggers))
		i := 0
		for k := range tpr.actionsTriggers {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.StatQueueProfilePrefix:
		keys := make([]string, len(tpr.sqProfiles))
		i := 0
		for k := range tpr.sqProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.ThresholdProfilePrefix:
		keys := make([]string, len(tpr.thProfiles))
		i := 0
		for k := range tpr.thProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.FilterPrefix:
		keys := make([]string, len(tpr.filters))
		i := 0
		for k := range tpr.filters {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.SupplierProfilePrefix:
		keys := make([]string, len(tpr.sppProfiles))
		i := 0
		for k := range tpr.sppProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.AttributeProfilePrefix:
		keys := make([]string, len(tpr.attributeProfiles))
		i := 0
		for k := range tpr.attributeProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.ChargerProfilePrefix:
		keys := make([]string, len(tpr.chargerProfiles))
		i := 0
		for k := range tpr.chargerProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	case utils.DispatcherProfilePrefix:
		keys := make([]string, len(tpr.dispatcherProfiles))
		i := 0
		for k := range tpr.dispatcherProfiles {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil

	case utils.DispatcherHostPrefix:
		keys := make([]string, len(tpr.dispatcherHosts))
		i := 0
		for k := range tpr.dispatcherHosts {
			keys[i] = k.TenantID()
			i++
		}
		return keys, nil
	}
	return nil, errors.New("Unsupported load category")
}

func (tpr *TpReader) RemoveFromDatabase(verbose, disable_reverse bool) (err error) {
	loadID := time.Now().UnixNano()
	loadIDs := make(map[string]int64)
	for _, d := range tpr.destinations {
		err = tpr.dm.RemoveDestination(d.Id, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", d.Id, " : ", d.Prefixes)
		}
	}
	if verbose {
		log.Print("Reverse Destinations:")
		for id, vals := range tpr.revDests {
			log.Printf("\t %s : %+v", id, vals)
		}
	}
	if verbose {
		log.Print("Rating Plans:")
	}
	for _, rp := range tpr.ratingPlans {
		err = tpr.dm.RemoveRatingPlan(rp.Id, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", rp.Id)
		}
	}
	if verbose {
		log.Print("Rating Profiles:")
	}
	for _, rp := range tpr.ratingProfiles {
		err = tpr.dm.RemoveRatingProfile(rp.Id, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", rp.Id)
		}
	}
	if verbose {
		log.Print("Action Plans:")
	}
	for k := range tpr.actionPlans {
		err = tpr.dm.RemoveActionPlan(k, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if verbose {
		log.Print("Account Action Plans:")
		for id, vals := range tpr.acntActionPlans {
			log.Printf("\t %s : %+v", id, vals)
		}
	}
	if verbose {
		log.Print("Action Triggers:")
	}
	for k := range tpr.actionsTriggers {
		err = tpr.dm.RemoveActionTriggers(k, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if verbose {
		log.Print("Shared Groups:")
	}
	for k := range tpr.sharedGroups {
		err = tpr.dm.RemoveSharedGroup(k, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if verbose {
		log.Print("Actions:")
	}
	for k := range tpr.actions {
		err = tpr.dm.RemoveActions(k, utils.NonTransactional)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if verbose {
		log.Print("Account Actions:")
	}
	for _, ub := range tpr.accountActions {
		err = tpr.dm.RemoveAccount(ub.ID)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", ub.ID)
		}
	}
	if verbose {
		log.Print("ResourceProfiles:")
	}
	for _, tpRsp := range tpr.resProfiles {
		if err = tpr.dm.RemoveResourceProfile(tpRsp.Tenant, tpRsp.ID, utils.NonTransactional, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpRsp.Tenant, tpRsp.ID))
		}
	}
	if verbose {
		log.Print("Resources:")
	}
	for _, rTid := range tpr.resources {
		if err = tpr.dm.RemoveResource(rTid.Tenant, rTid.ID, utils.NonTransactional); err != nil {
			return
		}
		if verbose {
			log.Print("\t", rTid.TenantID())
		}
	}
	if verbose {
		log.Print("StatQueueProfiles:")
	}
	for _, tpST := range tpr.sqProfiles {
		if err = tpr.dm.RemoveStatQueueProfile(tpST.Tenant, tpST.ID, utils.NonTransactional, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpST.Tenant, tpST.ID))
		}
	}
	if verbose {
		log.Print("StatQueues:")
	}
	for _, sqTntID := range tpr.statQueues {
		if err = tpr.dm.RemoveStatQueue(sqTntID.Tenant, sqTntID.ID, utils.NonTransactional); err != nil {
			return
		}
		if verbose {
			log.Print("\t", sqTntID.TenantID())
		}
	}
	if verbose {
		log.Print("ThresholdProfiles:")
	}
	for _, tpTH := range tpr.thProfiles {
		if err = tpr.dm.RemoveThresholdProfile(tpTH.Tenant, tpTH.ID, utils.NonTransactional, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpTH.Tenant, tpTH.ID))
		}
	}
	if verbose {
		log.Print("Thresholds:")
	}
	for _, thd := range tpr.thresholds {
		if err = tpr.dm.RemoveThreshold(thd.Tenant, thd.ID, utils.NonTransactional); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", thd.TenantID())
		}
	}

	if verbose {
		log.Print("SupplierProfiles:")
	}
	for _, tpSpl := range tpr.sppProfiles {
		if err = tpr.dm.RemoveSupplierProfile(tpSpl.Tenant, tpSpl.ID, utils.NonTransactional, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpSpl.Tenant, tpSpl.ID))
		}
	}

	if verbose {
		log.Print("AttributeProfiles:")
	}
	for _, tpAttr := range tpr.attributeProfiles {
		if err = tpr.dm.RemoveAttributeProfile(tpAttr.Tenant, tpAttr.ID,
			utils.NonTransactional, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpAttr.Tenant, tpAttr.ID))
		}
	}

	if verbose {
		log.Print("ChargerProfiles:")
	}
	for _, tpChr := range tpr.chargerProfiles {
		if err = tpr.dm.RemoveChargerProfile(tpChr.Tenant, tpChr.ID,
			utils.NonTransactional, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpChr.Tenant, tpChr.ID))
		}
	}

	if verbose {
		log.Print("DispatcherProfiles:")
	}
	for _, tpDsp := range tpr.dispatcherProfiles {
		if err = tpr.dm.RemoveDispatcherProfile(tpDsp.Tenant, tpDsp.ID,
			utils.NonTransactional, true); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpDsp.Tenant, tpDsp.ID))
		}
	}
	if verbose {
		log.Print("DispatcherHosts:")
	}
	for _, tpDsh := range tpr.dispatcherHosts {
		if err = tpr.dm.RemoveDispatcherHost(tpDsh.Tenant, tpDsh.ID,
			utils.NonTransactional); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpDsh.Tenant, tpDsh.ID))
		}
	}

	if verbose {
		log.Print("Timings:")
	}
	for _, t := range tpr.timings {
		if err = tpr.dm.RemoveTiming(t.ID, utils.NonTransactional); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", t.ID)
		}
	}
	if !disable_reverse {
		if len(tpr.destinations) > 0 {
			if verbose {
				log.Print("Removing reverse destinations")
			}
			if err = tpr.dm.DataDB().RemoveKeysForPrefix(utils.REVERSE_DESTINATION_PREFIX); err != nil {
				return err
			}
		}
		if len(tpr.acntActionPlans) > 0 {
			if verbose {
				log.Print("Removing account action plans")
			}
			if err = tpr.dm.DataDB().RemoveKeysForPrefix(utils.AccountActionPlansPrefix); err != nil {
				return err
			}
		}
	}
	//We remove the filters at the end because of indexes
	if verbose {
		log.Print("Filters:")
	}
	for _, tpFltr := range tpr.filters {
		if err = tpr.dm.RemoveFilter(tpFltr.Tenant, tpFltr.ID,
			utils.NonTransactional); err != nil {
			return err
		}
		if verbose {
			log.Print("\t", utils.ConcatenatedKey(tpFltr.Tenant, tpFltr.ID))
		}
	}
	if len(tpr.destinations) != 0 {
		loadIDs[utils.CacheDestinations] = loadID
		loadIDs[utils.CacheReverseDestinations] = loadID
	}
	if len(tpr.revDests) != 0 {
		loadIDs[utils.CacheReverseDestinations] = loadID
	}
	if len(tpr.ratingPlans) != 0 {
		loadIDs[utils.CacheRatingPlans] = loadID
	}
	if len(tpr.ratingProfiles) != 0 {
		loadIDs[utils.CacheRatingProfiles] = loadID
	}
	if len(tpr.actionPlans) != 0 {
		loadIDs[utils.CacheActionPlans] = loadID
	}
	if len(tpr.acntActionPlans) != 0 {
		loadIDs[utils.CacheAccountActionPlans] = loadID
	}
	if len(tpr.actionsTriggers) != 0 {
		loadIDs[utils.CacheActionTriggers] = loadID
	}
	if len(tpr.sharedGroups) != 0 {
		loadIDs[utils.CacheSharedGroups] = loadID
	}
	if len(tpr.actions) != 0 {
		loadIDs[utils.CacheActions] = loadID
	}
	if len(tpr.filters) != 0 {
		loadIDs[utils.CacheFilters] = loadID
	}
	if len(tpr.resProfiles) != 0 {
		loadIDs[utils.CacheResourceProfiles] = loadID
	}
	if len(tpr.resources) != 0 {
		loadIDs[utils.CacheResources] = loadID
	}
	if len(tpr.sqProfiles) != 0 {
		loadIDs[utils.CacheStatQueueProfiles] = loadID
	}
	if len(tpr.statQueues) != 0 {
		loadIDs[utils.CacheStatQueues] = loadID
	}
	if len(tpr.thProfiles) != 0 {
		loadIDs[utils.CacheThresholdProfiles] = loadID
	}
	if len(tpr.thresholds) != 0 {
		loadIDs[utils.CacheThresholds] = loadID
	}
	if len(tpr.sppProfiles) != 0 {
		loadIDs[utils.CacheSupplierProfiles] = loadID
	}
	if len(tpr.attributeProfiles) != 0 {
		loadIDs[utils.CacheAttributeProfiles] = loadID
	}
	if len(tpr.chargerProfiles) != 0 {
		loadIDs[utils.CacheChargerProfiles] = loadID
	}
	if len(tpr.dispatcherProfiles) != 0 {
		loadIDs[utils.CacheDispatcherProfiles] = loadID
	}
	if len(tpr.dispatcherHosts) != 0 {
		loadIDs[utils.CacheDispatcherHosts] = loadID
	}
	if len(tpr.timings) != 0 {
		loadIDs[utils.CacheTimings] = loadID
	}
	if err = tpr.dm.SetLoadIDs(loadIDs); err != nil {
		return err
	}
	return
}

func (tpr *TpReader) ReloadCache(caching string, verbose bool, argDispatcher *utils.ArgDispatcher) (err error) {
	if len(tpr.cacheConns) == 0 {
		log.Print("Disabled automatic reload")
		return
	}
	// take IDs for each type
	dstIds, _ := tpr.GetLoadedIds(utils.DESTINATION_PREFIX)
	revDstIDs, _ := tpr.GetLoadedIds(utils.REVERSE_DESTINATION_PREFIX)
	rplIds, _ := tpr.GetLoadedIds(utils.RATING_PLAN_PREFIX)
	rpfIds, _ := tpr.GetLoadedIds(utils.RATING_PROFILE_PREFIX)
	actIds, _ := tpr.GetLoadedIds(utils.ACTION_PREFIX)
	aapIDs, _ := tpr.GetLoadedIds(utils.AccountActionPlansPrefix)
	shgIds, _ := tpr.GetLoadedIds(utils.SHARED_GROUP_PREFIX)
	rspIDs, _ := tpr.GetLoadedIds(utils.ResourceProfilesPrefix)
	resIDs, _ := tpr.GetLoadedIds(utils.ResourcesPrefix)
	aatIDs, _ := tpr.GetLoadedIds(utils.ACTION_TRIGGER_PREFIX)
	stqIDs, _ := tpr.GetLoadedIds(utils.StatQueuePrefix)
	stqpIDs, _ := tpr.GetLoadedIds(utils.StatQueueProfilePrefix)
	trsIDs, _ := tpr.GetLoadedIds(utils.ThresholdPrefix)
	trspfIDs, _ := tpr.GetLoadedIds(utils.ThresholdProfilePrefix)
	flrIDs, _ := tpr.GetLoadedIds(utils.FilterPrefix)
	spfIDs, _ := tpr.GetLoadedIds(utils.SupplierProfilePrefix)
	apfIDs, _ := tpr.GetLoadedIds(utils.AttributeProfilePrefix)
	chargerIDs, _ := tpr.GetLoadedIds(utils.ChargerProfilePrefix)
	dppIDs, _ := tpr.GetLoadedIds(utils.DispatcherProfilePrefix)
	dphIDs, _ := tpr.GetLoadedIds(utils.DispatcherHostPrefix)
	aps, _ := tpr.GetLoadedIds(utils.ACTION_PLAN_PREFIX)

	//compose Reload Cache argument
	cacheArgs := utils.AttrReloadCacheWithArgDispatcher{
		TenantArg: utils.TenantArg{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		},
		ArgDispatcher: argDispatcher,
		AttrReloadCache: utils.AttrReloadCache{
			ArgsCache: utils.ArgsCache{
				DestinationIDs:        dstIds,
				ReverseDestinationIDs: revDstIDs,
				RatingPlanIDs:         rplIds,
				RatingProfileIDs:      rpfIds,
				ActionIDs:             actIds,
				ActionPlanIDs:         aps,
				AccountActionPlanIDs:  aapIDs,
				SharedGroupIDs:        shgIds,
				ResourceProfileIDs:    rspIDs,
				ResourceIDs:           resIDs,
				ActionTriggerIDs:      aatIDs,
				StatsQueueIDs:         stqIDs,
				StatsQueueProfileIDs:  stqpIDs,
				ThresholdIDs:          trsIDs,
				ThresholdProfileIDs:   trspfIDs,
				FilterIDs:             flrIDs,
				SupplierProfileIDs:    spfIDs,
				AttributeProfileIDs:   apfIDs,
				ChargerProfileIDs:     chargerIDs,
				DispatcherProfileIDs:  dppIDs,
				DispatcherHostIDs:     dphIDs,
			},
		},
	}

	if verbose {
		log.Print("Reloading cache")
	}
	var reply string
	switch caching {
	case utils.META_NONE:
		return
	case utils.MetaReload:
		if err = connMgr.Call(tpr.cacheConns, nil, utils.CacheSv1ReloadCache, cacheArgs, &reply); err != nil {
			return
		}
	case utils.MetaLoad:
		if err = connMgr.Call(tpr.cacheConns, nil, utils.CacheSv1LoadCache, cacheArgs, &reply); err != nil {
			return
		}
	case utils.MetaRemove:
		if err = connMgr.Call(tpr.cacheConns, nil, utils.CacheSv1FlushCache, cacheArgs, &reply); err != nil {
			return
		}
	case utils.MetaClear:
		cacheArgs.FlushAll = true
		if err = connMgr.Call(tpr.cacheConns, nil, utils.CacheSv1FlushCache, cacheArgs, &reply); err != nil {
			return
		}
	}

	// verify if we need to clear indexes
	var cacheIDs []string
	if len(apfIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheAttributeFilterIndexes)
	}
	if len(spfIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheSupplierFilterIndexes)
	}
	if len(trspfIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheThresholdFilterIndexes)
	}
	if len(stqpIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheStatFilterIndexes)
	}
	if len(rspIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheResourceFilterIndexes)
	}
	if len(chargerIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheChargerFilterIndexes)
	}
	if len(dppIDs) != 0 {
		cacheIDs = append(cacheIDs, utils.CacheDispatcherFilterIndexes)
	}
	if verbose {
		log.Print("Clearing indexes")
	}
	clearArgs := &utils.AttrCacheIDsWithArgDispatcher{
		TenantArg: utils.TenantArg{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		},
		ArgDispatcher: argDispatcher,
		CacheIDs:      cacheIDs,
	}
	if err = connMgr.Call(tpr.cacheConns, nil, utils.CacheSv1Clear, clearArgs, &reply); err != nil {
		log.Printf("WARNING: Got error on cache clear: %s\n", err.Error())
	}

	//get loadIDs for all types
	loadIDs, err := tpr.dm.GetItemLoadIDs(utils.EmptyString, false)
	if err != nil {
		return err
	}
	cacheLoadIDs := populateCacheLoadIDs(loadIDs, cacheArgs.AttrReloadCache)
	for key, val := range cacheLoadIDs {
		Cache.Set(utils.CacheLoadIDs, key, val, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (tpr *TpReader) ReloadScheduler(verbose bool) (err error) {
	var reply string
	aps, _ := tpr.GetLoadedIds(utils.ACTION_PLAN_PREFIX)
	// in case we have action plans reload the scheduler
	if len(aps) != 0 {
		if verbose {
			log.Print("Reloading scheduler")
		}
		if err = connMgr.Call(tpr.schedulerConns, nil, utils.SchedulerSv1Reload,
			new(utils.CGREventWithArgDispatcher), &reply); err != nil {
			log.Printf("WARNING: Got error on scheduler reload: %s\n", err.Error())
		}
	}
	return
}
