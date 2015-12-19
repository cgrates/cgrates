package engine

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type TpReader struct {
	tpid              string
	timezone          string
	loadHistSize      int
	ratingStorage     RatingStorage
	accountingStorage AccountingStorage
	lr                LoadReader
	actions           map[string][]*Action
	actionPlans       map[string]*ActionPlan
	actionsTriggers   map[string]ActionTriggers
	accountActions    map[string]*Account
	dirtyRpAliases    []*TenantRatingSubject // used to clean aliases that might have changed
	dirtyAccAliases   []*TenantAccount       // used to clean aliases that might have changed
	destinations      map[string]*Destination
	timings           map[string]*utils.TPTiming
	rates             map[string]*utils.TPRate
	destinationRates  map[string]*utils.TPDestinationRate
	ratingPlans       map[string]*RatingPlan
	ratingProfiles    map[string]*RatingProfile
	sharedGroups      map[string]*SharedGroup
	lcrs              map[string]*LCR
	derivedChargers   map[string]*utils.DerivedChargers
	cdrStats          map[string]*CdrStats
	users             map[string]*UserProfile
	aliases           map[string]*Alias
	loadInstance      *LoadInstance
}

func NewTpReader(rs RatingStorage, as AccountingStorage, lr LoadReader, tpid, timezone string, loadHistSize int) *TpReader {
	tpr := &TpReader{
		tpid:              tpid,
		timezone:          timezone,
		loadHistSize:      loadHistSize,
		ratingStorage:     rs,
		accountingStorage: as,
		lr:                lr,
	}
	tpr.Init()
	//add *any and *asap timing tag (in case of no timings file)
	tpr.timings[utils.ANY] = &utils.TPTiming{
		TimingId:  utils.ANY,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: "00:00:00",
		EndTime:   "",
	}
	tpr.timings[utils.ASAP] = &utils.TPTiming{
		TimingId:  utils.ASAP,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: utils.ASAP,
		EndTime:   "",
	}
	return tpr
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
	tpr.lcrs = make(map[string]*LCR)
	tpr.accountActions = make(map[string]*Account)
	tpr.cdrStats = make(map[string]*CdrStats)
	tpr.users = make(map[string]*UserProfile)
	tpr.aliases = make(map[string]*Alias)
	tpr.derivedChargers = make(map[string]*utils.DerivedChargers)
}

func (tpr *TpReader) LoadDestinationsFiltered(tag string) (bool, error) {
	tpDests, err := tpr.lr.GetTpDestinations(tpr.tpid, tag)

	dest := &Destination{Id: tag}
	for _, tpDest := range tpDests {
		dest.AddPrefix(tpDest.Prefix)
	}
	tpr.ratingStorage.SetDestination(dest)
	return len(tpDests) > 0, err
}

func (tpr *TpReader) LoadDestinations() (err error) {
	tps, err := tpr.lr.GetTpDestinations(tpr.tpid, "")
	if err != nil {
		return err
	}
	tpr.destinations, err = TpDestinations(tps).GetDestinations()
	return err
}

func (tpr *TpReader) LoadTimings() (err error) {
	tps, err := tpr.lr.GetTpTimings(tpr.tpid, "")
	if err != nil {
		return err
	}

	tpr.timings, err = TpTimings(tps).GetTimings()
	// add *any timing tag
	tpr.timings[utils.ANY] = &utils.TPTiming{
		TimingId:  utils.ANY,
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		StartTime: "00:00:00",
		EndTime:   "",
	}
	tpr.timings[utils.ASAP] = &utils.TPTiming{
		TimingId:  utils.ASAP,
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
	tps, err := tpr.lr.GetTpRates(tpr.tpid, "")
	if err != nil {
		return err
	}
	tpr.rates, err = TpRates(tps).GetRates()
	return err
}

func (tpr *TpReader) LoadDestinationRates() (err error) {
	tps, err := tpr.lr.GetTpDestinationRates(tpr.tpid, "", nil)
	if err != nil {
		return err
	}
	tpr.destinationRates, err = TpDestinationRates(tps).GetDestinationRates()
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
			if !destinationExists && tpr.ratingStorage != nil {
				if destinationExists, err = tpr.ratingStorage.HasData(utils.DESTINATION_PREFIX, dr.DestinationId); err != nil {
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

// Returns true, nil in case of load success, false, nil in case of RatingPlan  not found ratingStorage
func (tpr *TpReader) LoadRatingPlansFiltered(tag string) (bool, error) {
	mpRpls, err := tpr.lr.GetTpRatingPlans(tpr.tpid, tag, nil)
	if err != nil {
		return false, err
	} else if len(mpRpls) == 0 {
		return false, nil
	}

	bindings, err := TpRatingPlans(mpRpls).GetRatingPlans()
	if err != nil {
		return false, err
	}

	for tag, rplBnds := range bindings {
		ratingPlan := &RatingPlan{Id: tag}
		for _, rp := range rplBnds {
			tptm, err := tpr.lr.GetTpTimings(tpr.tpid, rp.TimingId)
			if err != nil || len(tptm) == 0 {
				return false, fmt.Errorf("no timing with id %s: %v", rp.TimingId, err)
			}
			tm, err := TpTimings(tptm).GetTimings()
			if err != nil {
				return false, err
			}

			rp.SetTiming(tm[rp.TimingId])
			tpdrm, err := tpr.lr.GetTpDestinationRates(tpr.tpid, rp.DestinationRatesId, nil)
			if err != nil || len(tpdrm) == 0 {
				return false, fmt.Errorf("no DestinationRates profile with id %s: %v", rp.DestinationRatesId, err)
			}
			drm, err := TpDestinationRates(tpdrm).GetDestinationRates()
			if err != nil {
				return false, err
			}
			for _, drate := range drm[rp.DestinationRatesId].DestinationRates {
				tprt, err := tpr.lr.GetTpRates(tpr.tpid, drate.RateId)
				if err != nil || len(tprt) == 0 {
					return false, fmt.Errorf("no Rates profile with id %s: %v", drate.RateId, err)
				}
				rt, err := TpRates(tprt).GetRates()
				if err != nil {
					return false, err
				}

				drate.Rate = rt[drate.RateId]
				ratingPlan.AddRateInterval(drate.DestinationId, GetRateInterval(rp, drate))
				if drate.DestinationId == utils.ANY {
					continue // no need of loading the destinations in this case
				}
				tpDests, err := tpr.lr.GetTpDestinations(tpr.tpid, drate.DestinationId)
				dms, err := TpDestinations(tpDests).GetDestinations()
				if err != nil {
					return false, err
				}
				destsExist := len(dms) != 0
				if !destsExist && tpr.ratingStorage != nil {
					if dbExists, err := tpr.ratingStorage.HasData(utils.DESTINATION_PREFIX, drate.DestinationId); err != nil {
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
					tpr.ratingStorage.SetDestination(destination)
				}
			}
		}
		if err := tpr.ratingStorage.SetRatingPlan(ratingPlan); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (tpr *TpReader) LoadRatingPlans() (err error) {
	tps, err := tpr.lr.GetTpRatingPlans(tpr.tpid, "", nil)
	if err != nil {
		return err
	}
	bindings, err := TpRatingPlans(tps).GetRatingPlans()

	if err != nil {
		return err
	}

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

func (tpr *TpReader) LoadRatingProfilesFiltered(qriedRpf *TpRatingProfile) error {
	var resultRatingProfile *RatingProfile
	mpTpRpfs, err := tpr.lr.GetTpRatingProfiles(qriedRpf)
	if err != nil {
		return fmt.Errorf("no RateProfile for filter %v, error: %v", qriedRpf, err)
	}

	rpfs, err := TpRatingProfiles(mpTpRpfs).GetRatingProfiles()
	if err != nil {
		return err
	}
	for _, tpRpf := range rpfs {
		resultRatingProfile = &RatingProfile{Id: tpRpf.KeyId()}
		for _, tpRa := range tpRpf.RatingPlanActivations {
			at, err := utils.ParseDate(tpRa.ActivationTime)
			if err != nil {
				return fmt.Errorf("cannot parse activation time from %v", tpRa.ActivationTime)
			}
			_, exists := tpr.ratingPlans[tpRa.RatingPlanId]
			if !exists && tpr.ratingStorage != nil {
				if exists, err = tpr.ratingStorage.HasData(utils.RATING_PLAN_PREFIX, tpRa.RatingPlanId); err != nil {
					return err
				}
			}
			if !exists {
				return fmt.Errorf("could not load rating plans for tag: %v", tpRa.RatingPlanId)
			}
			resultRatingProfile.RatingPlanActivations = append(resultRatingProfile.RatingPlanActivations,
				&RatingPlanActivation{
					ActivationTime:  at,
					RatingPlanId:    tpRa.RatingPlanId,
					FallbackKeys:    utils.FallbackSubjKeys(tpRpf.Direction, tpRpf.Tenant, tpRpf.Category, tpRa.FallbackSubjects),
					CdrStatQueueIds: strings.Split(tpRa.CdrStatQueueIds, utils.INFIELD_SEP),
				})
		}
		if err := tpr.ratingStorage.SetRatingProfile(resultRatingProfile); err != nil {
			return err
		}
	}
	return nil
}

func (tpr *TpReader) LoadRatingProfiles() (err error) {
	tps, err := tpr.lr.GetTpRatingProfiles(&TpRatingProfile{Tpid: tpr.tpid})
	if err != nil {
		return err
	}
	mpTpRpfs, err := TpRatingProfiles(tps).GetRatingProfiles()

	if err != nil {
		return err
	}
	for _, tpRpf := range mpTpRpfs {
		rpf := &RatingProfile{Id: tpRpf.KeyId()}
		for _, tpRa := range tpRpf.RatingPlanActivations {
			at, err := utils.ParseDate(tpRa.ActivationTime)
			if err != nil {
				return fmt.Errorf("cannot parse activation time from %v", tpRa.ActivationTime)
			}
			_, exists := tpr.ratingPlans[tpRa.RatingPlanId]
			if !exists && tpr.ratingStorage != nil { // Only query if there is a connection, eg on dry run there is none
				if exists, err = tpr.ratingStorage.HasData(utils.RATING_PLAN_PREFIX, tpRa.RatingPlanId); err != nil {
					return err
				}
			}
			if !exists {
				return fmt.Errorf("could not load rating plans for tag: %v", tpRa.RatingPlanId)
			}
			rpf.RatingPlanActivations = append(rpf.RatingPlanActivations,
				&RatingPlanActivation{
					ActivationTime:  at,
					RatingPlanId:    tpRa.RatingPlanId,
					FallbackKeys:    utils.FallbackSubjKeys(tpRpf.Direction, tpRpf.Tenant, tpRpf.Category, tpRa.FallbackSubjects),
					CdrStatQueueIds: strings.Split(tpRa.CdrStatQueueIds, utils.INFIELD_SEP),
				})
		}
		tpr.ratingProfiles[tpRpf.KeyId()] = rpf
	}
	return nil
}

func (tpr *TpReader) LoadSharedGroupsFiltered(tag string, save bool) (err error) {
	tps, err := tpr.lr.GetTpSharedGroups(tpr.tpid, "")
	if err != nil {
		return err
	}
	storSgs, err := TpSharedGroups(tps).GetSharedGroups()
	if err != nil {
		return err
	}
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
			if err := tpr.ratingStorage.SetSharedGroup(sg); err != nil {
				return err
			}
		}
	}
	return nil
}

func (tpr *TpReader) LoadSharedGroups() error {
	return tpr.LoadSharedGroupsFiltered(tpr.tpid, false)
}

func (tpr *TpReader) LoadLCRs() (err error) {
	tps, err := tpr.lr.GetTpLCRs(&TpLcrRule{Tpid: tpr.tpid})
	if err != nil {
		return err
	}

	for _, tpLcr := range tps {
		// check the rating profiles
		ratingProfileSearchKey := utils.ConcatenatedKey(tpLcr.Direction, tpLcr.Tenant, tpLcr.RpCategory)
		found := false
		for rpfKey := range tpr.ratingProfiles {
			if strings.HasPrefix(rpfKey, ratingProfileSearchKey) {
				found = true
				break
			}
		}
		if !found && tpr.ratingStorage != nil {
			if keys, err := tpr.ratingStorage.GetKeysForPrefix(utils.RATING_PROFILE_PREFIX + ratingProfileSearchKey); err != nil {
				return fmt.Errorf("[LCR] error querying ratingDb %s", err.Error())
			} else if len(keys) != 0 {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("[LCR] could not find ratingProfiles with prefix %s", ratingProfileSearchKey)
		}

		// check destination tags
		if tpLcr.DestinationTag != "" && tpLcr.DestinationTag != utils.ANY {
			_, found := tpr.destinations[tpLcr.DestinationTag]
			if !found && tpr.ratingStorage != nil {
				if found, err = tpr.ratingStorage.HasData(utils.DESTINATION_PREFIX, tpLcr.DestinationTag); err != nil {
					return fmt.Errorf("[LCR] error querying ratingDb %s", err.Error())
				}
			}
			if !found {
				return fmt.Errorf("[LCR] could not find destination with tag %s", tpLcr.DestinationTag)
			}
		}
		tag := utils.LCRKey(tpLcr.Direction, tpLcr.Tenant, tpLcr.Category, tpLcr.Account, tpLcr.Subject)
		activationTime, _ := utils.ParseTimeDetectLayout(tpLcr.ActivationTime, tpr.timezone)

		lcr, found := tpr.lcrs[tag]
		if !found {
			lcr = &LCR{
				Direction: tpLcr.Direction,
				Tenant:    tpLcr.Tenant,
				Category:  tpLcr.Category,
				Account:   tpLcr.Account,
				Subject:   tpLcr.Subject,
			}
		}
		var act *LCRActivation
		for _, existingAct := range lcr.Activations {
			if existingAct.ActivationTime.Equal(activationTime) {
				act = existingAct
				break
			}
		}
		if act == nil {
			act = &LCRActivation{
				ActivationTime: activationTime,
			}
			lcr.Activations = append(lcr.Activations, act)
		}
		act.Entries = append(act.Entries, &LCREntry{
			DestinationId:  tpLcr.DestinationTag,
			RPCategory:     tpLcr.RpCategory,
			Strategy:       tpLcr.Strategy,
			StrategyParams: tpLcr.StrategyParams,
			Weight:         tpLcr.Weight,
		})
		tpr.lcrs[tag] = lcr
	}
	return nil
}

func (tpr *TpReader) LoadActions() (err error) {
	tps, err := tpr.lr.GetTpActions(tpr.tpid, "")
	if err != nil {
		return err
	}

	storActs, err := TpActions(tps).GetActions()
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
				Weight:           tpact.Weight,
				ExtraParameters:  tpact.ExtraParameters,
				ExpirationString: tpact.ExpiryTime,
				Balance: &Balance{
					Id:             tpact.BalanceId,
					Value:          tpact.Units,
					Weight:         tpact.BalanceWeight,
					RatingSubject:  tpact.RatingSubject,
					Categories:     utils.ParseStringMap(tpact.Categories),
					Directions:     utils.ParseStringMap(tpact.Directions),
					DestinationIds: utils.ParseStringMap(tpact.DestinationIds),
					SharedGroups:   utils.ParseStringMap(tpact.SharedGroups),
					TimingIDs:      utils.ParseStringMap(tpact.TimingTags),
				},
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
		tpr.actions[tag] = acts
	}
	return nil
}

func (tpr *TpReader) LoadActionPlans() (err error) {
	tps, err := tpr.lr.GetTpActionPlans(tpr.tpid, "")
	if err != nil {
		return err
	}

	storAps, err := TpActionPlans(tps).GetActionPlans()
	if err != nil {
		return err
	}
	for atId, ats := range storAps {
		for _, at := range ats {

			_, exists := tpr.actions[at.ActionsId]
			if !exists && tpr.ratingStorage != nil {
				if exists, err = tpr.ratingStorage.HasData(utils.ACTION_PREFIX, at.ActionsId); err != nil {
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
	tps, err := tpr.lr.GetTpActionTriggers(tpr.tpid, "")
	if err != nil {
		return err
	}
	storAts, err := TpActionTriggers(tps).GetActionTriggers()
	if err != nil {
		return err
	}
	for key, atrsLst := range storAts {
		atrs := make([]*ActionTrigger, len(atrsLst))
		for idx, atr := range atrsLst {
			balanceExpirationDate, _ := utils.ParseTimeDetectLayout(atr.BalanceExpirationDate, tpr.timezone)
			minSleep, err := utils.ParseDurationWithSecs(atr.MinSleep)
			if err != nil {
				return err
			}
			if atr.UniqueID == "" {
				atr.UniqueID = utils.GenUUID()
			}
			atrs[idx] = &ActionTrigger{
				ID:                    key,
				UniqueID:              atr.UniqueID,
				ThresholdType:         atr.ThresholdType,
				ThresholdValue:        atr.ThresholdValue,
				Recurrent:             atr.Recurrent,
				MinSleep:              minSleep,
				BalanceId:             atr.BalanceId,
				BalanceType:           atr.BalanceType,
				BalanceDirections:     utils.ParseStringMap(atr.BalanceDirections),
				BalanceDestinationIds: utils.ParseStringMap(atr.BalanceDestinationIds),
				BalanceWeight:         atr.BalanceWeight,
				BalanceExpirationDate: balanceExpirationDate,
				BalanceTimingTags:     utils.ParseStringMap(atr.BalanceTimingTags),
				BalanceRatingSubject:  atr.BalanceRatingSubject,
				BalanceCategories:     utils.ParseStringMap(atr.BalanceCategories),
				BalanceSharedGroups:   utils.ParseStringMap(atr.BalanceSharedGroups),
				Weight:                atr.Weight,
				ActionsId:             atr.ActionsId,
				MinQueuedItems:        atr.MinQueuedItems,
			}
		}
		tpr.actionsTriggers[key] = atrs
	}

	return nil
}

func (tpr *TpReader) LoadAccountActionsFiltered(qriedAA *TpAccountAction) error {
	accountActions, err := tpr.lr.GetTpAccountActions(qriedAA)
	if err != nil {
		return errors.New(err.Error() + ": " + fmt.Sprintf("%+v", qriedAA))
	}
	storAas, err := TpAccountActions(accountActions).GetAccountActions()
	if err != nil {
		return err
	}
	for _, accountAction := range storAas {
		id := accountAction.KeyId()
		var actionsIds []string // collects action ids
		// action timings
		if accountAction.ActionPlanId != "" {
			// get old userBalanceIds
			exitingAccountIds := make(map[string]struct{})
			existingActionPlan, err := tpr.ratingStorage.GetActionPlan(accountAction.ActionPlanId, true)
			if err == nil && existingActionPlan != nil {
				exitingAccountIds = existingActionPlan.AccountIDs
			}

			tpap, err := tpr.lr.GetTpActionPlans(tpr.tpid, accountAction.ActionPlanId)
			if err != nil {
				return errors.New(err.Error() + " (ActionPlan): " + accountAction.ActionPlanId)
			} else if len(tpap) == 0 {
				return fmt.Errorf("no action plan with id <%s>", accountAction.ActionPlanId)
			}
			aps, err := TpActionPlans(tpap).GetActionPlans()
			if err != nil {
				return err
			}
			var actionPlan *ActionPlan
			ats := aps[accountAction.ActionPlanId]
			for _, at := range ats {
				// Check action exists before saving it inside actionTiming key
				// ToDo: try saving the key after the actions was retrieved in order to save one query here.
				if actions, err := tpr.lr.GetTpActions(tpr.tpid, at.ActionsId); err != nil {
					return errors.New(err.Error() + " (Actions): " + at.ActionsId)
				} else if len(actions) == 0 {
					return fmt.Errorf("no action with id <%s>", at.ActionsId)
				}
				var t *utils.TPTiming
				if at.TimingId != utils.ASAP {
					tptm, err := tpr.lr.GetTpTimings(tpr.tpid, at.TimingId)
					if err != nil {
						return errors.New(err.Error() + " (Timing): " + at.TimingId)
					} else if len(tptm) == 0 {
						return fmt.Errorf("no timing with id <%s>", at.TimingId)
					}
					tm, err := TpTimings(tptm).GetTimings()
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
				actionsIds = append(actionsIds, at.ActionsId)
				exitingAccountIds[id] = struct{}{}
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
						if err = tpr.ratingStorage.PushTask(t); err != nil {
							return err
						}
					}
				}
			}
			// write action plan
			err = tpr.ratingStorage.SetActionPlan(accountAction.ActionPlanId, actionPlan)
			if err != nil {
				return errors.New(err.Error() + " (SetActionPlan): " + accountAction.ActionPlanId)
			}
		}
		// action triggers
		var actionTriggers ActionTriggers
		//ActionTriggerPriotityList []*ActionTrigger
		if accountAction.ActionTriggersId != "" {
			tpatrs, err := tpr.lr.GetTpActionTriggers(tpr.tpid, accountAction.ActionTriggersId)
			if err != nil {
				return errors.New(err.Error() + " (ActionTriggers): " + accountAction.ActionTriggersId)
			}
			atrs, err := TpActionTriggers(tpatrs).GetActionTriggers()
			if err != nil {
				return err
			}

			atrsMap := make(map[string][]*ActionTrigger)
			for key, atrsLst := range atrs {
				atrs := make([]*ActionTrigger, len(atrsLst))
				for idx, apiAtr := range atrsLst {
					minSleep, _ := utils.ParseDurationWithSecs(apiAtr.MinSleep)
					expTime, _ := utils.ParseDate(apiAtr.BalanceExpirationDate)
					if apiAtr.UniqueID == "" {
						apiAtr.UniqueID = utils.GenUUID()
					}
					atrs[idx] = &ActionTrigger{
						ID:                    key,
						UniqueID:              apiAtr.UniqueID,
						ThresholdType:         apiAtr.ThresholdType,
						ThresholdValue:        apiAtr.ThresholdValue,
						Recurrent:             apiAtr.Recurrent,
						MinSleep:              minSleep,
						BalanceId:             apiAtr.BalanceId,
						BalanceType:           apiAtr.BalanceType,
						BalanceDirections:     utils.ParseStringMap(apiAtr.BalanceDirections),
						BalanceDestinationIds: utils.ParseStringMap(apiAtr.BalanceDestinationIds),
						BalanceWeight:         apiAtr.BalanceWeight,
						BalanceExpirationDate: expTime,
						BalanceRatingSubject:  apiAtr.BalanceRatingSubject,
						BalanceCategories:     utils.ParseStringMap(apiAtr.BalanceCategories),
						BalanceSharedGroups:   utils.ParseStringMap(apiAtr.BalanceSharedGroups),
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
			// write action triggers
			err = tpr.ratingStorage.SetActionTriggers(accountAction.ActionTriggersId, actionTriggers)
			if err != nil {
				return errors.New(err.Error() + " (SetActionTriggers): " + accountAction.ActionTriggersId)
			}
		}

		// actions
		acts := make(map[string][]*Action)
		for _, actId := range actionsIds {
			tpas, err := tpr.lr.GetTpActions(tpr.tpid, actId)
			if err != nil {
				return err
			}
			as, err := TpActions(tpas).GetActions()
			if err != nil {
				return err
			}
			for tag, tpacts := range as {
				enacts := make([]*Action, len(tpacts))
				for idx, tpact := range tpacts {
					enacts[idx] = &Action{
						Id:               tag + strconv.Itoa(idx),
						ActionType:       tpact.Identifier,
						BalanceType:      tpact.BalanceType,
						Weight:           tpact.Weight,
						ExtraParameters:  tpact.ExtraParameters,
						ExpirationString: tpact.ExpiryTime,
						Balance: &Balance{
							Value:          tpact.Units,
							Weight:         tpact.BalanceWeight,
							RatingSubject:  tpact.RatingSubject,
							Directions:     utils.ParseStringMap(tpact.Directions),
							DestinationIds: utils.ParseStringMap(tpact.DestinationIds),
							SharedGroups:   utils.ParseStringMap(tpact.SharedGroups),
							TimingIDs:      utils.ParseStringMap(tpact.TimingTags),
						},
					}
				}
				acts[tag] = enacts
			}
		}
		// write actions
		for k, as := range acts {
			err = tpr.ratingStorage.SetActions(k, as)
			if err != nil {
				return err
			}
		}
		ub, err := tpr.accountingStorage.GetAccount(id)
		if err != nil {
			ub = &Account{
				Id: id,
			}
		}
		ub.ActionTriggers = actionTriggers
		// init counters
		ub.InitCounters()
		if err := tpr.accountingStorage.SetAccount(ub); err != nil {
			return err
		}
	}
	return nil
}

func (tpr *TpReader) LoadAccountActions() (err error) {
	tps, err := tpr.lr.GetTpAccountActions(&TpAccountAction{Tpid: tpr.tpid})
	if err != nil {
		return err
	}
	storAts, err := TpAccountActions(tps).GetAccountActions()
	if err != nil {
		return err
	}

	for _, aa := range storAts {
		if _, alreadyDefined := tpr.accountActions[aa.KeyId()]; alreadyDefined {
			return fmt.Errorf("duplicate account action found: %s", aa.KeyId())
		}
		var aTriggers ActionTriggers
		if aa.ActionTriggersId != "" {
			var exists bool
			if aTriggers, exists = tpr.actionsTriggers[aa.ActionTriggersId]; !exists {
				return fmt.Errorf("could not get action triggers for tag %s", aa.ActionTriggersId)
			}
		}
		ub := &Account{
			Id:             aa.KeyId(),
			ActionTriggers: aTriggers,
			AllowNegative:  aa.AllowNegative,
			Disabled:       aa.Disabled,
		}
		ub.InitCounters()
		tpr.accountActions[aa.KeyId()] = ub
		actionPlan, exists := tpr.actionPlans[aa.ActionPlanId]
		if !exists {
			log.Printf("could not get action plan for tag %v", aa.ActionPlanId)
			// must not continue here
		}
		if actionPlan.AccountIDs == nil {
			actionPlan.AccountIDs = make(map[string]struct{})
		}
		actionPlan.AccountIDs[aa.KeyId()] = struct{}{}
	}
	return nil
}

func (tpr *TpReader) LoadDerivedChargersFiltered(filter *TpDerivedCharger, save bool) (err error) {
	tps, err := tpr.lr.GetTpDerivedChargers(filter)
	if err != nil {
		return err
	}
	storDcs, err := TpDerivedChargers(tps).GetDerivedChargers()
	if err != nil {
		return err
	}
	for _, tpDcs := range storDcs {
		tag := tpDcs.GetDerivedChargersKey()
		if _, hasIt := tpr.derivedChargers[tag]; !hasIt {
			tpr.derivedChargers[tag] = &utils.DerivedChargers{
				DestinationIds: make(utils.StringMap),
				Chargers:       make([]*utils.DerivedCharger, 0),
			} // Load object map since we use this method also from LoadDerivedChargers
		}
		for _, tpDc := range tpDcs.DerivedChargers {
			dc, err := utils.NewDerivedCharger(tpDc.RunId, tpDc.RunFilters, tpDc.ReqTypeField, tpDc.DirectionField, tpDc.TenantField, tpDc.CategoryField,
				tpDc.AccountField, tpDc.SubjectField, tpDc.DestinationField, tpDc.SetupTimeField, tpDc.PddField, tpDc.AnswerTimeField, tpDc.UsageField, tpDc.SupplierField,
				tpDc.DisconnectCauseField, tpDc.RatedField, tpDc.CostField)
			if err != nil {
				return err
			}
			tpr.derivedChargers[tag].DestinationIds.Copy(utils.ParseStringMap(tpDcs.DestinationIds))
			tpr.derivedChargers[tag].Chargers = append(tpr.derivedChargers[tag].Chargers, dc)
		}
	}
	if save {
		for dcsKey, dcs := range tpr.derivedChargers {
			if err := tpr.ratingStorage.SetDerivedChargers(dcsKey, dcs); err != nil {
				return err
			}
		}
	}
	return nil
}

func (tpr *TpReader) LoadDerivedChargers() (err error) {
	return tpr.LoadDerivedChargersFiltered(&TpDerivedCharger{Tpid: tpr.tpid}, false)
}

func (tpr *TpReader) LoadCdrStatsFiltered(tag string, save bool) (err error) {
	tps, err := tpr.lr.GetTpCdrStats(tpr.tpid, tag)
	if err != nil {
		return err
	}
	storStats, err := TpCdrStats(tps).GetCdrStats()
	if err != nil {
		return err
	}
	var actionsIds []string // collect action ids
	for tag, tpStats := range storStats {
		for _, tpStat := range tpStats {
			var cs *CdrStats
			var exists bool
			if cs, exists = tpr.cdrStats[tag]; !exists {
				cs = &CdrStats{Id: tag}
			}
			// action triggers
			triggerTag := tpStat.ActionTriggers
			if triggerTag != "" {
				_, exists := tpr.actionsTriggers[triggerTag]
				if !exists {
					tpatrs, err := tpr.lr.GetTpActionTriggers(tpr.tpid, triggerTag)
					if err != nil {
						return errors.New(err.Error() + " (ActionTriggers): " + triggerTag)
					}
					atrsM, err := TpActionTriggers(tpatrs).GetActionTriggers()
					if err != nil {
						return err
					}

					for _, atrsLst := range atrsM {
						atrs := make([]*ActionTrigger, len(atrsLst))
						for idx, apiAtr := range atrsLst {
							minSleep, _ := utils.ParseDurationWithSecs(apiAtr.MinSleep)
							expTime, _ := utils.ParseDate(apiAtr.BalanceExpirationDate)
							if apiAtr.UniqueID == "" {
								apiAtr.UniqueID = utils.GenUUID()
							}
							atrs[idx] = &ActionTrigger{
								ID:                    triggerTag,
								UniqueID:              apiAtr.UniqueID,
								ThresholdType:         apiAtr.ThresholdType,
								ThresholdValue:        apiAtr.ThresholdValue,
								Recurrent:             apiAtr.Recurrent,
								MinSleep:              minSleep,
								BalanceId:             apiAtr.BalanceId,
								BalanceType:           apiAtr.BalanceType,
								BalanceDirections:     utils.ParseStringMap(apiAtr.BalanceDirections),
								BalanceDestinationIds: utils.ParseStringMap(apiAtr.BalanceDestinationIds),
								BalanceWeight:         apiAtr.BalanceWeight,
								BalanceExpirationDate: expTime,
								BalanceRatingSubject:  apiAtr.BalanceRatingSubject,
								BalanceCategories:     utils.ParseStringMap(apiAtr.BalanceCategories),
								BalanceSharedGroups:   utils.ParseStringMap(apiAtr.BalanceSharedGroups),
								BalanceTimingTags:     utils.ParseStringMap(apiAtr.BalanceTimingTags),
								Weight:                apiAtr.Weight,
								ActionsId:             apiAtr.ActionsId,
							}
						}
						tpr.actionsTriggers[triggerTag] = atrs
					}
				}
				// collect action ids from triggers
				for _, atr := range tpr.actionsTriggers[triggerTag] {
					actionsIds = append(actionsIds, atr.ActionsId)
				}
			}
			triggers, exists := tpr.actionsTriggers[triggerTag]
			if triggerTag != "" && !exists {
				// only return error if there was something there for the tag
				return fmt.Errorf("could not get action triggers for cdr stats id %s: %s", cs.Id, triggerTag)
			}
			// write action triggers
			err = tpr.ratingStorage.SetActionTriggers(triggerTag, triggers)
			if err != nil {
				return errors.New(err.Error() + " (SetActionTriggers): " + triggerTag)
			}
			UpdateCdrStats(cs, triggers, tpStat, tpr.timezone)
			tpr.cdrStats[tag] = cs
		}
	}
	// actions
	for _, actId := range actionsIds {
		_, exists := tpr.actions[actId]
		if !exists {
			tpas, err := tpr.lr.GetTpActions(tpr.tpid, actId)
			if err != nil {
				return err
			}
			as, err := TpActions(tpas).GetActions()
			if err != nil {
				return err
			}
			for tag, tpacts := range as {
				enacts := make([]*Action, len(tpacts))
				for idx, tpact := range tpacts {
					enacts[idx] = &Action{
						Id:               tag + strconv.Itoa(idx),
						ActionType:       tpact.Identifier,
						BalanceType:      tpact.BalanceType,
						Weight:           tpact.Weight,
						ExtraParameters:  tpact.ExtraParameters,
						ExpirationString: tpact.ExpiryTime,
						Balance: &Balance{
							Value:          tpact.Units,
							Weight:         tpact.BalanceWeight,
							RatingSubject:  tpact.RatingSubject,
							Directions:     utils.ParseStringMap(tpact.Directions),
							DestinationIds: utils.ParseStringMap(tpact.DestinationIds),
							SharedGroups:   utils.ParseStringMap(tpact.SharedGroups),
							TimingIDs:      utils.ParseStringMap(tpact.TimingTags),
						},
					}
				}
				tpr.actions[tag] = enacts
			}
		}
	}

	if save {
		// write actions
		for k, as := range tpr.actions {
			err = tpr.ratingStorage.SetActions(k, as)
			if err != nil {
				return err
			}
		}
		for _, stat := range tpr.cdrStats {
			if err := tpr.ratingStorage.SetCdrStats(stat); err != nil {
				return err
			}
		}
	}
	return nil
}

func (tpr *TpReader) LoadCdrStats() error {
	return tpr.LoadCdrStatsFiltered("", false)
}

func (tpr *TpReader) LoadUsersFiltered(filter *TpUser) (bool, error) {
	tpUsers, err := tpr.lr.GetTpUsers(filter)

	user := &UserProfile{
		Tenant:   filter.Tenant,
		UserName: filter.UserName,
		Profile:  make(map[string]string),
	}
	for _, tpUser := range tpUsers {
		user.Profile[tpUser.AttributeName] = tpUser.AttributeValue
	}
	tpr.accountingStorage.SetUser(user)
	return len(tpUsers) > 0, err
}

func (tpr *TpReader) LoadUsers() error {
	tps, err := tpr.lr.GetTpUsers(&TpUser{Tpid: tpr.tpid})
	if err != nil {
		return err
	}
	userMap, err := TpUsers(tps).GetUsers()
	if err != nil {
		return err
	}
	for key, usr := range userMap {
		up, found := tpr.users[key]
		if !found {
			up = &UserProfile{
				Tenant:   usr.Tenant,
				UserName: usr.UserName,
				Profile:  make(map[string]string),
			}
			tpr.users[key] = up
		}
		for _, p := range usr.Profile {
			up.Profile[p.AttrName] = p.AttrValue
		}
	}
	return err
}

func (tpr *TpReader) LoadAliasesFiltered(filter *TpAlias) (bool, error) {
	tpAliases, err := tpr.lr.GetTpAliases(filter)

	alias := &Alias{
		Direction: filter.Direction,
		Tenant:    filter.Tenant,
		Category:  filter.Category,
		Account:   filter.Account,
		Subject:   filter.Subject,
		Context:   filter.Context,
		Values:    make(AliasValues, 0),
	}
	for _, tpAlias := range tpAliases {
		av := alias.Values.GetValueByDestId(tpAlias.DestinationId)
		if av == nil {
			av = &AliasValue{
				DestinationId: tpAlias.DestinationId,
				Pairs:         make(AliasPairs),
				Weight:        tpAlias.Weight,
			}
			alias.Values = append(alias.Values, av)
		}
		if av.Pairs[tpAlias.Target] == nil {
			av.Pairs[tpAlias.Target] = make(map[string]string)
		}
		av.Pairs[tpAlias.Target][tpAlias.Original] = tpAlias.Alias

	}
	tpr.accountingStorage.SetAlias(alias)
	return len(tpAliases) > 0, err
}

func (tpr *TpReader) LoadAliases() error {
	tps, err := tpr.lr.GetTpAliases(&TpAlias{Tpid: tpr.tpid})
	if err != nil {
		return err
	}
	alMap, err := TpAliases(tps).GetAliases()
	if err != nil {
		return err
	}
	for key, tal := range alMap {
		al, found := tpr.aliases[key]
		if !found {
			al = &Alias{
				Direction: tal.Direction,
				Tenant:    tal.Tenant,
				Category:  tal.Category,
				Account:   tal.Account,
				Subject:   tal.Subject,
				Context:   tal.Context,
				Values:    make(AliasValues, 0),
			}
			tpr.aliases[key] = al
		}
		for _, v := range tal.Values {
			av := al.Values.GetValueByDestId(v.DestinationId)
			if av == nil {
				av = &AliasValue{
					DestinationId: v.DestinationId,
					Pairs:         make(AliasPairs),
					Weight:        v.Weight,
				}
				al.Values = append(al.Values, av)
			}
			if av.Pairs[v.Target] == nil {
				av.Pairs[v.Target] = make(map[string]string)
			}
			av.Pairs[v.Target][v.Original] = v.Alias
		}
	}
	return err
}

func (tpr *TpReader) LoadAll() error {
	var err error
	if err = tpr.LoadDestinations(); err != nil {
		return err
	}
	if err = tpr.LoadTimings(); err != nil {
		return err
	}
	if err = tpr.LoadRates(); err != nil {
		return err
	}
	if err = tpr.LoadDestinationRates(); err != nil {
		return err
	}
	if err = tpr.LoadRatingPlans(); err != nil {
		return err
	}
	if err = tpr.LoadRatingProfiles(); err != nil {
		return err
	}
	if err = tpr.LoadSharedGroups(); err != nil {
		return err
	}
	if err = tpr.LoadLCRs(); err != nil {
		return err
	}
	if err = tpr.LoadActions(); err != nil {
		return err
	}
	if err = tpr.LoadActionPlans(); err != nil {
		return err
	}
	if err = tpr.LoadActionTriggers(); err != nil {
		return err
	}
	if err = tpr.LoadAccountActions(); err != nil {
		return err
	}
	if err = tpr.LoadDerivedChargers(); err != nil {
		return err
	}
	if err = tpr.LoadCdrStats(); err != nil {
		return err
	}
	if err = tpr.LoadUsers(); err != nil {
		return err
	}
	if err = tpr.LoadAliases(); err != nil {
		return err
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

func (tpr *TpReader) GetLoadInstance() *LoadInstance {
	if tpr.loadInstance == nil {
		tpr.loadInstance = &LoadInstance{LoadId: utils.GenUUID(), TariffPlanId: tpr.tpid, LoadTime: time.Now()}
	}
	return tpr.loadInstance
}

func (tpr *TpReader) WriteToDatabase(flush, verbose bool) (err error) {
	if tpr.ratingStorage == nil || tpr.accountingStorage == nil {
		return errors.New("no database connection")
	}
	if flush {
		tpr.ratingStorage.Flush("")
	}
	if verbose {
		log.Print("Destinations:")
	}
	for _, d := range tpr.destinations {
		err = tpr.ratingStorage.SetDestination(d)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", d.Id, " : ", d.Prefixes)
		}
	}
	if verbose {
		log.Print("Rating Plans:")
	}
	for _, rp := range tpr.ratingPlans {
		err = tpr.ratingStorage.SetRatingPlan(rp)
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
		err = tpr.ratingStorage.SetRatingProfile(rp)
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
					if err = tpr.ratingStorage.PushTask(t); err != nil {
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
					if err = tpr.ratingStorage.PushTask(t); err != nil {
						return err
					}
				}
			}
		}
		err = tpr.ratingStorage.SetActionPlan(k, ap)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if verbose {
		log.Print("Action Triggers:")
	}
	for k, atrs := range tpr.actionsTriggers {
		err = tpr.ratingStorage.SetActionTriggers(k, atrs)
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
	for k, sg := range tpr.sharedGroups {
		err = tpr.ratingStorage.SetSharedGroup(sg)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", k)
		}
	}
	if verbose {
		log.Print("LCR Rules:")
	}
	for k, lcr := range tpr.lcrs {
		err = tpr.ratingStorage.SetLCR(lcr)
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
	for k, as := range tpr.actions {
		err = tpr.ratingStorage.SetActions(k, as)
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
		err = tpr.accountingStorage.SetAccount(ub)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", ub.Id)
		}
	}
	if verbose {
		log.Print("Derived Chargers:")
	}
	for key, dcs := range tpr.derivedChargers {
		err = tpr.ratingStorage.SetDerivedChargers(key, dcs)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", key)
		}
	}
	if verbose {
		log.Print("CDR Stats Queues:")
	}
	for _, sq := range tpr.cdrStats {
		err = tpr.ratingStorage.SetCdrStats(sq)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", sq.Id)
		}
	}
	if verbose {
		log.Print("Users:")
	}
	for _, u := range tpr.users {
		err = tpr.accountingStorage.SetUser(u)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", u.GetId())
		}
	}
	if verbose {
		log.Print("Aliases:")
	}
	for _, al := range tpr.aliases {
		err = tpr.accountingStorage.SetAlias(al)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", al.GetId())
		}
	}
	ldInst := tpr.GetLoadInstance()
	if verbose {
		log.Printf("LoadHistory, instance: %+v\n", ldInst)
	}
	if err = tpr.accountingStorage.AddLoadHistory(ldInst, tpr.loadHistSize); err != nil {
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
	// action trigers
	log.Print("Action trigers: ", len(tpr.actionsTriggers))
	// account actions
	log.Print("Account actions: ", len(tpr.accountActions))
	// derivedChargers
	log.Print("Derived Chargers: ", len(tpr.derivedChargers))
	// lcr rules
	log.Print("LCR rules: ", len(tpr.lcrs))
	// cdr stats
	log.Print("CDR stats: ", len(tpr.cdrStats))
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
	case utils.DERIVEDCHARGERS_PREFIX: // derived chargers
		keys := make([]string, len(tpr.derivedChargers))
		i := 0
		for k := range tpr.derivedChargers {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.CDR_STATS_PREFIX: // cdr stats
		keys := make([]string, len(tpr.cdrStats))
		i := 0
		for k := range tpr.cdrStats {
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
	case utils.USERS_PREFIX:
		keys := make([]string, len(tpr.users))
		i := 0
		for k := range tpr.users {
			keys[i] = k
			i++
		}
		return keys, nil
	case utils.ALIASES_PREFIX:
		keys := make([]string, len(tpr.aliases))
		i := 0
		for k := range tpr.aliases {
			keys[i] = k
			i++
		}
		return keys, nil
	}
	return nil, errors.New("Unsupported load category")
}
