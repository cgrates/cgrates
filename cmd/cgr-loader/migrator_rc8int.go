package main

import (
	"log"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type Account1 struct {
	Id             string
	BalanceMap     map[string]BalanceChain1
	UnitCounters   UnitCounters1
	ActionTriggers ActionTriggers1
	AllowNegative  bool
	Disabled       bool
}

type Balance1 struct {
	Uuid           string //system wide unique
	Id             string // account wide unique
	Value          float64
	Directions     utils.StringMap
	ExpirationDate time.Time
	Weight         float64
	DestinationIds utils.StringMap
	RatingSubject  string
	Categories     utils.StringMap
	SharedGroups   utils.StringMap
	Timings        []*engine.RITiming
	TimingIDs      utils.StringMap
	Disabled       bool
	Factor         engine.ValueFactor
	Blocker        bool
	precision      int
	account        *Account // used to store ub reference for shared balances
	dirty          bool
}

type BalanceChain1 []*Balance1

type UnitCounter1 struct {
	BalanceType string        // *monetary/*voice/*sms/etc
	CounterType string        // *event or *balance
	Balances    BalanceChain1 // first balance is the general one (no destination)
}

type UnitCounters1 []*UnitCounter1

type Action1 struct {
	Id               string
	ActionType       string
	BalanceType      string
	ExtraParameters  string
	Filter           string
	ExpirationString string // must stay as string because it can have relative values like 1month
	Weight           float64
	Balance          *Balance1
}

type Actions1 []*Action1

type ActionTrigger1 struct {
	ID                    string // original csv tag
	UniqueID              string // individual id
	ThresholdType         string
	ThresholdValue        float64
	Recurrent             bool          // reset eexcuted flag each run
	MinSleep              time.Duration // Minimum duration between two executions in case of recurrent triggers
	BalanceId             string
	BalanceType           string          // *monetary/*voice etc
	BalanceDirections     utils.StringMap // filter for balance
	BalanceDestinationIds utils.StringMap // filter for balance
	BalanceWeight         float64         // filter for balance
	BalanceExpirationDate time.Time       // filter for balance
	BalanceTimingTags     utils.StringMap // filter for balance
	BalanceRatingSubject  string          // filter for balance
	BalanceCategories     utils.StringMap // filter for balance
	BalanceSharedGroups   utils.StringMap // filter for balance
	BalanceBlocker        bool
	BalanceDisabled       bool // filter for balance
	Weight                float64
	ActionsId             string
	MinQueuedItems        int // Trigger actions only if this number is hit (stats only)
	Executed              bool
}
type ActionTriggers1 []*ActionTrigger1

func (mig MigratorRC8) migrateAccountsInt() error {
	keys, err := mig.db.Cmd("KEYS", OLD_ACCOUNT_PREFIX+"*").List()
	if err != nil {
		return err
	}
	newAccounts := make([]*engine.Account, 0)
	var migratedKeys []string
	// get existing accounts
	for _, key := range keys {
		log.Printf("Migrating account: %s...", key)
		values, err := mig.db.Cmd("GET", key).Bytes()
		if err != nil {
			continue
		}
		var oldAcc Account1
		if err = mig.ms.Unmarshal(values, &oldAcc); err != nil {
			return err
		}
		// transfer data into new structurse
		newAcc := &engine.Account{
			ID:             oldAcc.Id,
			BalanceMap:     make(map[string]engine.Balances, len(oldAcc.BalanceMap)),
			UnitCounters:   make(engine.UnitCounters, len(oldAcc.UnitCounters)),
			ActionTriggers: make(engine.ActionTriggers, len(oldAcc.ActionTriggers)),
			AllowNegative:  oldAcc.AllowNegative,
			Disabled:       oldAcc.Disabled,
		}
		// balances
		balanceErr := false
		for key, oldBalChain := range oldAcc.BalanceMap {
			newAcc.BalanceMap[key] = make(engine.Balances, len(oldBalChain))
			for index, oldBal := range oldBalChain {
				newAcc.BalanceMap[key][index] = &engine.Balance{
					Uuid:           oldBal.Uuid,
					ID:             oldBal.Id,
					Value:          oldBal.Value,
					Directions:     oldBal.Directions,
					ExpirationDate: oldBal.ExpirationDate,
					Weight:         oldBal.Weight,
					DestinationIDs: oldBal.DestinationIds,
					RatingSubject:  oldBal.RatingSubject,
					Categories:     oldBal.Categories,
					SharedGroups:   oldBal.SharedGroups,
					Timings:        oldBal.Timings,
					TimingIDs:      oldBal.TimingIDs,
					Disabled:       oldBal.Disabled,
					Factor:         oldBal.Factor,
					Blocker:        oldBal.Blocker,
				}
			}
		}
		if balanceErr {
			continue
		}
		// unit counters
		for _, oldUc := range oldAcc.UnitCounters {
			newUc := &engine.UnitCounter{
				Counters: make(engine.CounterFilters, len(oldUc.Balances)),
			}
			for index, oldUcBal := range oldUc.Balances {
				b := &engine.Balance{
					Uuid:           oldUcBal.Uuid,
					ID:             oldUcBal.Id,
					Value:          oldUcBal.Value,
					Directions:     oldUcBal.Directions,
					ExpirationDate: oldUcBal.ExpirationDate,
					Weight:         oldUcBal.Weight,
					DestinationIDs: oldUcBal.DestinationIds,
					RatingSubject:  oldUcBal.RatingSubject,
					Categories:     oldUcBal.Categories,
					SharedGroups:   oldUcBal.SharedGroups,
					Timings:        oldUcBal.Timings,
					TimingIDs:      oldUcBal.TimingIDs,
					Disabled:       oldUcBal.Disabled,
					Factor:         oldUcBal.Factor,
					Blocker:        oldUcBal.Blocker,
				}
				bf := &engine.BalanceFilter{}
				bf.LoadFromBalance(b)
				cf := &engine.CounterFilter{
					Value:  oldUcBal.Value,
					Filter: bf,
				}
				newUc.Counters[index] = cf
			}
			newAcc.UnitCounters[oldUc.BalanceType] = append(newAcc.UnitCounters[oldUc.BalanceType], newUc)
		}
		// action triggers
		for _, oldAtr := range oldAcc.ActionTriggers {
			at := &engine.ActionTrigger{
				ID:             oldAtr.ID,
				UniqueID:       oldAtr.UniqueID,
				ThresholdType:  oldAtr.ThresholdType,
				ThresholdValue: oldAtr.ThresholdValue,
				Recurrent:      oldAtr.Recurrent,
				MinSleep:       oldAtr.MinSleep,
				Weight:         oldAtr.Weight,
				ActionsID:      oldAtr.ActionsId,
				MinQueuedItems: oldAtr.MinQueuedItems,
				Executed:       oldAtr.Executed,
			}
			bf := &engine.BalanceFilter{}
			if oldAtr.BalanceId != "" {
				bf.ID = utils.StringPointer(oldAtr.BalanceId)
			}
			if oldAtr.BalanceType != "" {
				bf.Type = utils.StringPointer(oldAtr.BalanceType)
			}
			if oldAtr.BalanceRatingSubject != "" {
				bf.RatingSubject = utils.StringPointer(oldAtr.BalanceRatingSubject)
			}
			if !oldAtr.BalanceDirections.IsEmpty() {
				bf.Directions = utils.StringMapPointer(oldAtr.BalanceDirections)
			}
			if !oldAtr.BalanceDestinationIds.IsEmpty() {
				bf.DestinationIDs = utils.StringMapPointer(oldAtr.BalanceDestinationIds)
			}
			if !oldAtr.BalanceTimingTags.IsEmpty() {
				bf.TimingIDs = utils.StringMapPointer(oldAtr.BalanceTimingTags)
			}
			if !oldAtr.BalanceCategories.IsEmpty() {
				bf.Categories = utils.StringMapPointer(oldAtr.BalanceCategories)
			}
			if !oldAtr.BalanceSharedGroups.IsEmpty() {
				bf.SharedGroups = utils.StringMapPointer(oldAtr.BalanceSharedGroups)
			}
			if oldAtr.BalanceWeight != 0 {
				bf.Weight = utils.Float64Pointer(oldAtr.BalanceWeight)
			}
			if oldAtr.BalanceDisabled != false {
				bf.Disabled = utils.BoolPointer(oldAtr.BalanceDisabled)
			}
			if !oldAtr.BalanceExpirationDate.IsZero() {
				bf.ExpirationDate = utils.TimePointer(oldAtr.BalanceExpirationDate)
			}
			at.Balance = bf
		}
		newAcc.InitCounters()
		newAccounts = append(newAccounts, newAcc)
		migratedKeys = append(migratedKeys, key)
	}
	// write data back
	for _, newAcc := range newAccounts {
		result, err := mig.ms.Marshal(newAcc)
		if err != nil {
			return err
		}
		if err := mig.db.Cmd("SET", utils.ACCOUNT_PREFIX+newAcc.ID, result).Err; err != nil {
			return err
		}
	}
	notMigrated := len(keys) - len(migratedKeys)
	if notMigrated > 0 {
		log.Printf("WARNING: there are %d accounts that failed migration!", notMigrated)
	}
	return err
}

func (mig MigratorRC8) migrateActionTriggersInt() error {
	keys, err := mig.db.Cmd("KEYS", utils.ACTION_TRIGGER_PREFIX+"*").List()
	if err != nil {
		return err
	}
	newAtrsMap := make(map[string]engine.ActionTriggers, len(keys))
	for _, key := range keys {
		log.Printf("Migrating action trigger: %s...", key)
		var oldAtrs ActionTriggers1
		var values []byte
		if values, err = mig.db.Cmd("GET", key).Bytes(); err == nil {
			if err := mig.ms.Unmarshal(values, &oldAtrs); err != nil {
				return err
			}
		}
		newAtrs := make(engine.ActionTriggers, len(oldAtrs))
		for index, oldAtr := range oldAtrs {
			at := &engine.ActionTrigger{
				ID:             oldAtr.ID,
				UniqueID:       oldAtr.UniqueID,
				ThresholdType:  oldAtr.ThresholdType,
				ThresholdValue: oldAtr.ThresholdValue,
				Recurrent:      oldAtr.Recurrent,
				MinSleep:       oldAtr.MinSleep,
				Weight:         oldAtr.Weight,
				ActionsID:      oldAtr.ActionsId,
				MinQueuedItems: oldAtr.MinQueuedItems,
				Executed:       oldAtr.Executed,
			}
			bf := &engine.BalanceFilter{}
			if oldAtr.BalanceId != "" {
				bf.ID = utils.StringPointer(oldAtr.BalanceId)
			}
			if oldAtr.BalanceType != "" {
				bf.Type = utils.StringPointer(oldAtr.BalanceType)
			}
			if oldAtr.BalanceRatingSubject != "" {
				bf.RatingSubject = utils.StringPointer(oldAtr.BalanceRatingSubject)
			}
			if !oldAtr.BalanceDirections.IsEmpty() {
				bf.Directions = utils.StringMapPointer(oldAtr.BalanceDirections)
			}
			if !oldAtr.BalanceDestinationIds.IsEmpty() {
				bf.DestinationIDs = utils.StringMapPointer(oldAtr.BalanceDestinationIds)
			}
			if !oldAtr.BalanceTimingTags.IsEmpty() {
				bf.TimingIDs = utils.StringMapPointer(oldAtr.BalanceTimingTags)
			}
			if !oldAtr.BalanceCategories.IsEmpty() {
				bf.Categories = utils.StringMapPointer(oldAtr.BalanceCategories)
			}
			if !oldAtr.BalanceSharedGroups.IsEmpty() {
				bf.SharedGroups = utils.StringMapPointer(oldAtr.BalanceSharedGroups)
			}
			if oldAtr.BalanceWeight != 0 {
				bf.Weight = utils.Float64Pointer(oldAtr.BalanceWeight)
			}
			if oldAtr.BalanceDisabled != false {
				bf.Disabled = utils.BoolPointer(oldAtr.BalanceDisabled)
			}
			if !oldAtr.BalanceExpirationDate.IsZero() {
				bf.ExpirationDate = utils.TimePointer(oldAtr.BalanceExpirationDate)
			}
			at.Balance = bf
			newAtrs[index] = at
		}
		newAtrsMap[key] = newAtrs
	}
	// write data back
	for key, atrs := range newAtrsMap {
		result, err := mig.ms.Marshal(&atrs)
		if err != nil {
			return err
		}
		if err = mig.db.Cmd("SET", key, result).Err; err != nil {
			return err
		}
	}
	return nil
}

func (mig MigratorRC8) migrateActionsInt() error {
	keys, err := mig.db.Cmd("KEYS", utils.ACTION_PREFIX+"*").List()
	if err != nil {
		return err
	}
	newAcsMap := make(map[string]engine.Actions, len(keys))
	for _, key := range keys {
		log.Printf("Migrating action: %s...", key)
		var oldAcs Actions1
		var values []byte
		if values, err = mig.db.Cmd("GET", key).Bytes(); err == nil {
			if err := mig.ms.Unmarshal(values, &oldAcs); err != nil {
				return err
			}
		}
		newAcs := make(engine.Actions, len(oldAcs))
		for index, oldAc := range oldAcs {
			a := &engine.Action{
				Id:               oldAc.Id,
				ActionType:       oldAc.ActionType,
				ExtraParameters:  oldAc.ExtraParameters,
				ExpirationString: oldAc.ExpirationString,
				Filter:           oldAc.Filter,
				Weight:           oldAc.Weight,
				Balance:          &engine.BalanceFilter{},
			}
			bf := a.Balance
			if oldAc.Balance.Uuid != "" {
				bf.Uuid = utils.StringPointer(oldAc.Balance.Uuid)
			}
			if oldAc.Balance.Id != "" {
				bf.ID = utils.StringPointer(oldAc.Balance.Id)
			}
			if oldAc.BalanceType != "" {
				bf.Type = utils.StringPointer(oldAc.BalanceType)
			}
			if oldAc.Balance.Value != 0 {
				bf.Value = utils.Float64Pointer(oldAc.Balance.Value)
			}
			if oldAc.Balance.RatingSubject != "" {
				bf.RatingSubject = utils.StringPointer(oldAc.Balance.RatingSubject)
			}
			if !oldAc.Balance.DestinationIds.IsEmpty() {
				bf.DestinationIDs = utils.StringMapPointer(oldAc.Balance.DestinationIds)
			}
			if !oldAc.Balance.TimingIDs.IsEmpty() {
				bf.TimingIDs = utils.StringMapPointer(oldAc.Balance.TimingIDs)
			}
			if !oldAc.Balance.Categories.IsEmpty() {
				bf.Categories = utils.StringMapPointer(oldAc.Balance.Categories)
			}
			if !oldAc.Balance.SharedGroups.IsEmpty() {
				bf.SharedGroups = utils.StringMapPointer(oldAc.Balance.SharedGroups)
			}
			if oldAc.Balance.Weight != 0 {
				bf.Weight = utils.Float64Pointer(oldAc.Balance.Weight)
			}
			if oldAc.Balance.Disabled != false {
				bf.Disabled = utils.BoolPointer(oldAc.Balance.Disabled)
			}
			if !oldAc.Balance.ExpirationDate.IsZero() {
				bf.ExpirationDate = utils.TimePointer(oldAc.Balance.ExpirationDate)
			}
			bf.Timings = oldAc.Balance.Timings
			newAcs[index] = a
		}
		newAcsMap[key] = newAcs
	}
	// write data back
	for key, acs := range newAcsMap {
		result, err := mig.ms.Marshal(&acs)
		if err != nil {
			return err
		}
		if err = mig.db.Cmd("SET", key, result).Err; err != nil {
			return err
		}
	}
	return nil
}
