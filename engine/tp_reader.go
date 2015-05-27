package engine

import (
	"errors"
	"fmt"
	"log"
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
				return fmt.Errorf("Could not find rate for tag %v", dr.RateId)
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
					return fmt.Errorf("Could not get destination for tag %v", dr.DestinationId)
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
				return fmt.Errorf("Could not get timing for tag %v", rplBnd.TimingId)
			}
			rplBnd.SetTiming(t)
			drs, exists := tpr.destinationRates[rplBnd.DestinationRatesId]
			if !exists {
				return fmt.Errorf("Could not find destination rate for tag %v", rplBnd.DestinationRatesId)
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
				return fmt.Errorf("Cannot parse activation time from %v", tpRa.ActivationTime)
			}
			_, exists := tpr.ratingPlans[tpRa.RatingPlanId]
			if !exists {
				if dbExists, err := tpr.ratingStorage.HasData(RATING_PLAN_PREFIX, tpRa.RatingPlanId); err != nil {
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
		tpr.ratingProfiles[tpRpf.KeyId()] = rpf
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
	if err = tpr.LoadActionTimings(); err != nil {
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
		return errors.New("No database connection!")
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
