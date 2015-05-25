package engine

import (
	"errors"
	"log"

	"github.com/cgrates/cgrates/utils"
)

type TPData struct {
	actions          map[string][]*Action
	actionsTimings   map[string][]*ActionTiming
	actionsTriggers  map[string][]*ActionTrigger
	accountActions   map[string]*Account
	dirtyRpAliases   []*TenantRatingSubject // used to clean aliases that might have changed
	dirtyAccAliases  []*TenantAccount       // used to clean aliases that might have changed
	destinations     map[string]*Destination
	rpAliases        map[string]string
	accAliases       map[string]string
	timings          map[string]*utils.TPTiming
	rates            map[string]*utils.TPRate
	destinationRates map[string]*utils.TPDestinationRate
	ratingPlans      map[string]*RatingPlan
	ratingProfiles   map[string]*RatingProfile
	sharedGroups     map[string]*SharedGroup
	lcrs             map[string]*LCR
	derivedChargers  map[string]utils.DerivedChargers
	cdrStats         map[string]*CdrStats
}

func NewTPData() *TPData {
	tp := &TPData{}
	tp.actions = make(map[string][]*Action)
	tp.actionsTimings = make(map[string][]*ActionTiming)
	tp.actionsTriggers = make(map[string][]*ActionTrigger)
	tp.rates = make(map[string]*utils.TPRate)
	tp.destinations = make(map[string]*Destination)
	tp.destinationRates = make(map[string]*utils.TPDestinationRate)
	tp.timings = make(map[string]*utils.TPTiming)
	tp.ratingPlans = make(map[string]*RatingPlan)
	tp.ratingProfiles = make(map[string]*RatingProfile)
	tp.sharedGroups = make(map[string]*SharedGroup)
	tp.lcrs = make(map[string]*LCR)
	tp.rpAliases = make(map[string]string)
	tp.accAliases = make(map[string]string)
	tp.timings = make(map[string]*utils.TPTiming)
	tp.accountActions = make(map[string]*Account)
	tp.destinations = make(map[string]*Destination)
	tp.cdrStats = make(map[string]*CdrStats)
	tp.derivedChargers = make(map[string]utils.DerivedChargers)
	return tp
}

func (tp *TPData) LoadDestinations(tpDests []*TpDestination) error {
	for _, tpDest := range tpDests {
		var dest *Destination
		var found bool
		if dest, found = tp.destinations[tpDest.Tag]; !found {
			dest = &Destination{Id: tpDest.Tag}
			tp.destinations[tpDest.Tag] = dest
		}
		dest.AddPrefix(tpDest.Prefix)
	}
	return nil
}

func (tp *TPData) IsValid() bool {
	valid := true
	for rplTag, rpl := range tp.ratingPlans {
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

func (tp *TPData) WriteToDatabase(dataStorage RatingStorage, accountingStorage AccountingStorage, flush, verbose bool) (err error) {
	if dataStorage == nil {
		return errors.New("No database connection!")
	}
	if flush {
		dataStorage.Flush("")
	}
	if verbose {
		log.Print("Destinations:")
	}
	for _, d := range tp.destinations {
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
	for _, rp := range tp.ratingPlans {
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
	for _, rp := range tp.ratingProfiles {
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
	for k, ats := range tp.actionsTimings {
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
	for k, sg := range tp.sharedGroups {
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
	for k, lcr := range tp.lcrs {
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
	for k, as := range tp.actions {
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
	for _, ub := range tp.accountActions {
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
	if err := dataStorage.RemoveRpAliases(tp.dirtyRpAliases); err != nil {
		return err
	}
	for key, alias := range tp.rpAliases {
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
	if err := accountingStorage.RemoveAccAliases(tp.dirtyAccAliases); err != nil {
		return err
	}
	for key, alias := range tp.accAliases {
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
	for key, dcs := range tp.derivedChargers {
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
	for _, sq := range tp.cdrStats {
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

func (tp *TPData) ShowStatistics() {
	// destinations
	destCount := len(tp.destinations)
	log.Print("Destinations: ", destCount)
	prefixDist := make(map[int]int, 50)
	prefixCount := 0
	for _, d := range tp.destinations {
		prefixDist[len(d.Prefixes)] += 1
		prefixCount += len(d.Prefixes)
	}
	log.Print("Avg Prefixes: ", prefixCount/destCount)
	log.Print("Prefixes distribution:")
	for k, v := range prefixDist {
		log.Printf("%d: %d", k, v)
	}
	// rating plans
	rplCount := len(tp.ratingPlans)
	log.Print("Rating plans: ", rplCount)
	destRatesDist := make(map[int]int, 50)
	destRatesCount := 0
	for _, rpl := range tp.ratingPlans {
		destRatesDist[len(rpl.DestinationRates)] += 1
		destRatesCount += len(rpl.DestinationRates)
	}
	log.Print("Avg Destination Rates: ", destRatesCount/rplCount)
	log.Print("Destination Rates distribution:")
	for k, v := range destRatesDist {
		log.Printf("%d: %d", k, v)
	}
	// rating profiles
	rpfCount := len(tp.ratingProfiles)
	log.Print("Rating profiles: ", rpfCount)
	activDist := make(map[int]int, 50)
	activCount := 0
	for _, rpf := range tp.ratingProfiles {
		activDist[len(rpf.RatingPlanActivations)] += 1
		activCount += len(rpf.RatingPlanActivations)
	}
	log.Print("Avg Activations: ", activCount/rpfCount)
	log.Print("Activation distribution:")
	for k, v := range activDist {
		log.Printf("%d: %d", k, v)
	}
	// actions
	log.Print("Actions: ", len(tp.actions))
	// action plans
	log.Print("Action plans: ", len(tp.actionsTimings))
	// action trigers
	log.Print("Action trigers: ", len(tp.actionsTriggers))
	// account actions
	log.Print("Account actions: ", len(tp.accountActions))
	// derivedChargers
	log.Print("Derived Chargers: ", len(tp.derivedChargers))
	// lcr rules
	log.Print("LCR rules: ", len(tp.lcrs))
	// cdr stats
	log.Print("CDR stats: ", len(tp.cdrStats))
}

// Returns the identities loaded for a specific category, useful for cache reloads
func (tp *TPData) GetLoadedIds(categ string) ([]string, error) {
	switch categ {
	case DESTINATION_PREFIX:
		keys := make([]string, len(tp.destinations))
		i := 0
		for k := range tp.destinations {
			keys[i] = k
			i++
		}
		return keys, nil
	case RATING_PLAN_PREFIX:
		keys := make([]string, len(tp.ratingPlans))
		i := 0
		for k := range tp.ratingPlans {
			keys[i] = k
			i++
		}
		return keys, nil
	case RATING_PROFILE_PREFIX:
		keys := make([]string, len(tp.ratingProfiles))
		i := 0
		for k := range tp.ratingProfiles {
			keys[i] = k
			i++
		}
		return keys, nil
	case ACTION_PREFIX: // actionsTimings
		keys := make([]string, len(tp.actions))
		i := 0
		for k := range tp.actions {
			keys[i] = k
			i++
		}
		return keys, nil
	case ACTION_TIMING_PREFIX: // actionsTimings
		keys := make([]string, len(tp.actionsTimings))
		i := 0
		for k := range tp.actionsTimings {
			keys[i] = k
			i++
		}
		return keys, nil
	case RP_ALIAS_PREFIX: // aliases
		keys := make([]string, len(tp.rpAliases))
		i := 0
		for k := range tp.rpAliases {
			keys[i] = k
			i++
		}
		return keys, nil
	case ACC_ALIAS_PREFIX: // aliases
		keys := make([]string, len(tp.accAliases))
		i := 0
		for k := range tp.accAliases {
			keys[i] = k
			i++
		}
		return keys, nil
	case DERIVEDCHARGERS_PREFIX: // derived chargers
		keys := make([]string, len(tp.derivedChargers))
		i := 0
		for k := range tp.derivedChargers {
			keys[i] = k
			i++
		}
		return keys, nil
	case CDR_STATS_PREFIX: // cdr stats
		keys := make([]string, len(tp.cdrStats))
		i := 0
		for k := range tp.cdrStats {
			keys[i] = k
			i++
		}
		return keys, nil
	case SHARED_GROUP_PREFIX:
		keys := make([]string, len(tp.sharedGroups))
		i := 0
		for k := range tp.sharedGroups {
			keys[i] = k
			i++
		}
		return keys, nil
	}
	return nil, errors.New("Unsupported category")
}
