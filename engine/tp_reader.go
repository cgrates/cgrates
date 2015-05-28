package engine

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

type TpReader struct {
	tpid              string
	ratingStorage     RatingStorage
	accountingStorage AccountingStorage
	lr                LoadReader
	actions           map[string][]*Action
	actionsTimings    map[string][]*ActionTiming
	actionsTriggers   map[string][]*ActionTrigger
	accountActions    map[string]*Account
	dirtyRpAliases    []*TenantRatingSubject // used to clean aliases that might have changed
	dirtyAccAliases   []*TenantAccount       // used to clean aliases that might have changed
	destinations      map[string]*Destination
	rpAliases         map[string]string
	accAliases        map[string]string
	timings           map[string]*utils.TPTiming
	rates             map[string]*utils.TPRate
	destinationRates  map[string]*utils.TPDestinationRate
	ratingPlans       map[string]*RatingPlan
	ratingProfiles    map[string]*RatingProfile
	sharedGroups      map[string]*SharedGroup
	lcrs              map[string]*LCR
	derivedChargers   map[string]utils.DerivedChargers
	cdrStats          map[string]*CdrStats
}

func NewTpReader(rs RatingStorage, as AccountingStorage, lr LoadReader, tpid string) *TpReader {
	return &TpReader{
		tpid:              tpid,
		ratingStorage:     rs,
		accountingStorage: as,
		lr:                lr,
		actions:           make(map[string][]*Action),
		actionsTimings:    make(map[string][]*ActionTiming),
		actionsTriggers:   make(map[string][]*ActionTrigger),
		rates:             make(map[string]*utils.TPRate),
		destinations:      make(map[string]*Destination),
		destinationRates:  make(map[string]*utils.TPDestinationRate),
		timings:           make(map[string]*utils.TPTiming),
		ratingPlans:       make(map[string]*RatingPlan),
		ratingProfiles:    make(map[string]*RatingProfile),
		sharedGroups:      make(map[string]*SharedGroup),
		lcrs:              make(map[string]*LCR),
		rpAliases:         make(map[string]string),
		accAliases:        make(map[string]string),
		accountActions:    make(map[string]*Account),
		cdrStats:          make(map[string]*CdrStats),
		derivedChargers:   make(map[string]utils.DerivedChargers),
	}
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
			if !destinationExists {
				if dbExists, err := tpr.ratingStorage.HasData(DESTINATION_PREFIX, dr.DestinationId); err != nil {
					return err
				} else if !dbExists {
					return fmt.Errorf("could not get destination for tag %v", dr.DestinationId)
				}
			}
		}
	}
	return nil
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

func (tpr *TpReader) LoadRatingProfiles() (err error) {
	tps, err := tpr.lr.GetTpRatingProfiles(nil)
	if err != nil {
		return err
	}
	mpTpRpfs, err := TpRatingProfiles(tps).GetRatingProfiles()

	if err != nil {
		return err
	}
	for _, tpRpf := range mpTpRpfs {
		// extract aliases from subject
		aliases := strings.Split(tpRpf.Subject, ";")
		tpr.dirtyRpAliases = append(tpr.dirtyRpAliases, &TenantRatingSubject{Tenant: tpRpf.Tenant, Subject: aliases[0]})
		if len(aliases) > 1 {
			tpRpf.Subject = aliases[0]
			for _, alias := range aliases[1:] {
				tpr.rpAliases[utils.RatingSubjectAliasKey(tpRpf.Tenant, alias)] = tpRpf.Subject
			}
		}
		rpf := &RatingProfile{Id: tpRpf.KeyId()}
		for _, tpRa := range tpRpf.RatingPlanActivations {
			at, err := utils.ParseDate(tpRa.ActivationTime)
			if err != nil {
				return fmt.Errorf("cannot parse activation time from %v", tpRa.ActivationTime)
			}
			_, exists := tpr.ratingPlans[tpRa.RatingPlanId]
			if !exists {
				if dbExists, err := tpr.ratingStorage.HasData(RATING_PLAN_PREFIX, tpRa.RatingPlanId); err != nil {
					return err
				} else if !dbExists {
					return fmt.Errorf("could not load rating plans for tag: %v", tpRa.RatingPlanId)
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
		tpr.ratingProfiles[tpRpf.KeyId()] = rpf
	}
	return nil
}

func (tpr *TpReader) LoadSharedGroups() (err error) {
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
	return nil
}

func (tpr *TpReader) LoadLCRs() (err error) {
	tps, err := tpr.lr.GetTpLCRs(tpr.tpid, "")
	if err != nil {
		return err
	}

	for _, tpLcr := range tps {
		tag := utils.LCRKey(tpLcr.Direction, tpLcr.Tenant, tpLcr.Category, tpLcr.Account, tpLcr.Subject)
		activationTime, _ := utils.ParseTimeDetectLayout(tpLcr.ActivationTime)

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
			RPCategory:     tpLcr.Category,
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
			if !exists {
				return fmt.Errorf("actionTiming: Could not load the action for tag: %v", at.ActionsId)
			}
			t, exists := tpr.timings[at.TimingId]
			if !exists {
				return fmt.Errorf("actionTiming: Could not load the timing for tag: %v", at.TimingId)
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
			tpr.actionsTimings[atId] = append(tpr.actionsTimings[atId], actTmg)
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
			balanceExpirationDate, _ := utils.ParseTimeDetectLayout(atr.BalanceExpirationDate)
			id := atr.Id
			if id == "" {
				id = utils.GenUUID()
			}
			minSleep, err := utils.ParseDurationWithSecs(atr.MinSleep)
			if err != nil {
				return err
			}
			atrs[idx] = &ActionTrigger{
				Id:                    id,
				ThresholdType:         atr.ThresholdType,
				ThresholdValue:        atr.ThresholdValue,
				Recurrent:             atr.Recurrent,
				MinSleep:              minSleep,
				BalanceId:             atr.BalanceId,
				BalanceType:           atr.BalanceType,
				BalanceDirection:      atr.BalanceDirection,
				BalanceDestinationIds: atr.BalanceDestinationIds,
				BalanceWeight:         atr.BalanceWeight,
				BalanceExpirationDate: balanceExpirationDate,
				BalanceTimingTags:     atr.BalanceTimingTags,
				BalanceRatingSubject:  atr.BalanceRatingSubject,
				BalanceCategory:       atr.BalanceCategory,
				BalanceSharedGroup:    atr.BalanceSharedGroup,
				Weight:                atr.Weight,
				ActionsId:             atr.ActionsId,
				MinQueuedItems:        atr.MinQueuedItems,
			}
			if atrs[idx].Id == "" {
				atrs[idx].Id = utils.GenUUID()
			}
		}
		tpr.actionsTriggers[key] = atrs
	}

	return nil
}

func (tpr *TpReader) LoadAccountActions() (err error) {
	tps, err := tpr.lr.GetTpAccountActions(nil)
	if err != nil {
		return err
	}
	storAts, err := TpAccountActions(tps).GetAccountActions()
	if err != nil {
		return err
	}

	for _, aa := range storAts {
		if _, alreadyDefined := tpr.accountActions[aa.KeyId()]; alreadyDefined {
			return fmt.Errorf("Duplicate account action found: %s", aa.KeyId())
		}

		// extract aliases from subject
		aliases := strings.Split(aa.Account, ";")
		tpr.dirtyAccAliases = append(tpr.dirtyAccAliases, &TenantAccount{Tenant: aa.Tenant, Account: aliases[0]})
		if len(aliases) > 1 {
			aa.Account = aliases[0]
			for _, alias := range aliases[1:] {
				tpr.accAliases[utils.AccountAliasKey(aa.Tenant, alias)] = aa.Account
			}
		}
		aTriggers, exists := tpr.actionsTriggers[aa.ActionTriggersId]
		if !exists {
			return fmt.Errorf("Could not get action triggers for tag %v", aa.ActionTriggersId)
		}
		ub := &Account{
			Id:             aa.KeyId(),
			ActionTriggers: aTriggers,
		}
		tpr.accountActions[aa.KeyId()] = ub
		aTimings, exists := tpr.actionsTimings[aa.ActionPlanId]
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

func (tpr *TpReader) LoadDerivedChargers() (err error) {
	tps, err := tpr.lr.GetTpDerivedChargers(nil)
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
			tpr.derivedChargers[tag] = make(utils.DerivedChargers, 0) // Load object map since we use this method also from LoadDerivedChargers
		}
		for _, tpDc := range tpDcs.DerivedChargers {
			if dc, err := utils.NewDerivedCharger(tpDc.RunId, tpDc.RunFilters, tpDc.ReqTypeField, tpDc.DirectionField, tpDc.TenantField, tpDc.CategoryField,
				tpDc.AccountField, tpDc.SubjectField, tpDc.DestinationField, tpDc.SetupTimeField, tpDc.AnswerTimeField, tpDc.UsageField, tpDc.SupplierField,
				tpDc.DisconnectCauseField); err != nil {
				return err
			} else {
				tpr.derivedChargers[tag] = append(tpr.derivedChargers[tag], dc)
			}
		}
	}
	return nil
}

func (tpr *TpReader) LoadCdrStats() (err error) {
	tps, err := tpr.lr.GetTpCdrStats(tpr.tpid, "")
	if err != nil {
		return err
	}
	storStats, err := TpCdrStats(tps).GetCdrStats()
	if err != nil {
		return err
	}
	for tag, tpStats := range storStats {
		for _, tpStat := range tpStats {
			var cs *CdrStats
			var exists bool
			if cs, exists = tpr.cdrStats[tag]; !exists {
				cs = &CdrStats{Id: tag}
			}
			triggerTag := tpStat.ActionTriggers
			triggers, exists := tpr.actionsTriggers[triggerTag]
			if triggerTag != "" && !exists {
				// only return error if there was something there for the tag
				return fmt.Errorf("Could not get action triggers for cdr stats id %s: %s", cs.Id, triggerTag)
			}
			UpdateCdrStats(cs, triggers, tpStat)
			tpr.cdrStats[tag] = cs
		}
	}
	return nil
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
	return nil
}

func (tpr *TpReader) IsValid() bool {
	valid := true
	for rplTag, rpl := range tpr.ratingPlans {
		if !rpl.isContinous() {
			log.Printf("The rating plan %s is not covering all weekdays", rplTag)
			valid = false
		}
		if !rpl.areRatesSane() {
			log.Printf("The rating plan %s contains invalid rate groups", rplTag)
			valid = false
		}
		if !rpl.areTimingsSane() {
			log.Printf("The rating plan %s contains invalid timings", rplTag)
			valid = false
		}
	}
	return valid
}

func (tpr *TpReader) WriteToDatabase(dataStorage RatingStorage, accountingStorage AccountingStorage, flush, verbose bool) (err error) {
	if dataStorage == nil {
		return errors.New("no database connection!")
	}
	if flush {
		dataStorage.Flush("")
	}
	if verbose {
		log.Print("Destinations:")
	}
	for _, d := range tpr.destinations {
		err = dataStorage.SetDestination(d)
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
		err = dataStorage.SetRatingPlan(rp)
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
		err = dataStorage.SetRatingProfile(rp)
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
	for k, ats := range tpr.actionsTimings {
		err = accountingStorage.SetActionTimings(k, ats)
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
		err = accountingStorage.SetSharedGroup(sg)
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
		err = dataStorage.SetLCR(lcr)
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
		err = accountingStorage.SetActions(k, as)
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
		err = accountingStorage.SetAccount(ub)
		if err != nil {
			return err
		}
		if verbose {
			log.Println("\t", ub.Id)
		}
	}
	if verbose {
		log.Print("Rating Profile Aliases:")
	}
	if err := dataStorage.RemoveRpAliases(tpr.dirtyRpAliases); err != nil {
		return err
	}
	for key, alias := range tpr.rpAliases {
		err = dataStorage.SetRpAlias(key, alias)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", key)
		}
	}
	if verbose {
		log.Print("Account Aliases:")
	}
	if err := accountingStorage.RemoveAccAliases(tpr.dirtyAccAliases); err != nil {
		return err
	}
	for key, alias := range tpr.accAliases {
		err = accountingStorage.SetAccAlias(key, alias)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", key)
		}
	}
	if verbose {
		log.Print("Derived Chargers:")
	}
	for key, dcs := range tpr.derivedChargers {
		err = accountingStorage.SetDerivedChargers(key, dcs)
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
		err = dataStorage.SetCdrStats(sq)
		if err != nil {
			return err
		}
		if verbose {
			log.Print("\t", sq.Id)
		}
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
	log.Print("Action plans: ", len(tpr.actionsTimings))
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
	case DESTINATION_PREFIX:
		keys := make([]string, len(tpr.destinations))
		i := 0
		for k := range tpr.destinations {
			keys[i] = k
			i++
		}
		return keys, nil
	case RATING_PLAN_PREFIX:
		keys := make([]string, len(tpr.ratingPlans))
		i := 0
		for k := range tpr.ratingPlans {
			keys[i] = k
			i++
		}
		return keys, nil
	case RATING_PROFILE_PREFIX:
		keys := make([]string, len(tpr.ratingProfiles))
		i := 0
		for k := range tpr.ratingProfiles {
			keys[i] = k
			i++
		}
		return keys, nil
	case ACTION_PREFIX: // actionsTimings
		keys := make([]string, len(tpr.actions))
		i := 0
		for k := range tpr.actions {
			keys[i] = k
			i++
		}
		return keys, nil
	case ACTION_TIMING_PREFIX: // actionsTimings
		keys := make([]string, len(tpr.actionsTimings))
		i := 0
		for k := range tpr.actionsTimings {
			keys[i] = k
			i++
		}
		return keys, nil
	case RP_ALIAS_PREFIX: // aliases
		keys := make([]string, len(tpr.rpAliases))
		i := 0
		for k := range tpr.rpAliases {
			keys[i] = k
			i++
		}
		return keys, nil
	case ACC_ALIAS_PREFIX: // aliases
		keys := make([]string, len(tpr.accAliases))
		i := 0
		for k := range tpr.accAliases {
			keys[i] = k
			i++
		}
		return keys, nil
	case DERIVEDCHARGERS_PREFIX: // derived chargers
		keys := make([]string, len(tpr.derivedChargers))
		i := 0
		for k := range tpr.derivedChargers {
			keys[i] = k
			i++
		}
		return keys, nil
	case CDR_STATS_PREFIX: // cdr stats
		keys := make([]string, len(tpr.cdrStats))
		i := 0
		for k := range tpr.cdrStats {
			keys[i] = k
			i++
		}
		return keys, nil
	case SHARED_GROUP_PREFIX:
		keys := make([]string, len(tpr.sharedGroups))
		i := 0
		for k := range tpr.sharedGroups {
			keys[i] = k
			i++
		}
		return keys, nil
	}
	return nil, errors.New("Unsupported category")
}
