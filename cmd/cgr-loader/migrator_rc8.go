package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/mediocregopher/radix.v2/redis"
)

const OLD_ACCOUNT_PREFIX = "ubl_"

type MigratorRC8 struct {
	db *redis.Client
	ms engine.Marshaler
}

func NewMigratorRC8(address string, db int, pass, mrshlerStr string) (*MigratorRC8, error) {
	client, err := redis.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	if err := client.Cmd("SELECT", db).Err; err != nil {
		return nil, err
	}
	if pass != "" {
		if err := client.Cmd("AUTH", pass).Err; err != nil {
			return nil, err
		}
	}

	var mrshler engine.Marshaler
	if mrshlerStr == utils.MSGPACK {
		mrshler = engine.NewCodecMsgpackMarshaler()
	} else if mrshlerStr == utils.JSON {
		mrshler = new(engine.JSONMarshaler)
	} else {
		return nil, fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	return &MigratorRC8{db: client, ms: mrshler}, nil
}

type Account struct {
	Id             string
	BalanceMap     map[string]BalanceChain
	UnitCounters   []*UnitsCounter
	ActionTriggers ActionTriggers
	AllowNegative  bool
	Disabled       bool
}
type BalanceChain []*Balance

type Balance struct {
	Uuid           string //system wide unique
	Id             string // account wide unique
	Value          float64
	ExpirationDate time.Time
	Weight         float64
	DestinationIds string
	RatingSubject  string
	Category       string
	SharedGroup    string
	Timings        []*engine.RITiming
	TimingIDs      string
	Disabled       bool
	precision      int
	account        *Account
	dirty          bool
}

func (b *Balance) IsDefault() bool {
	return (b.DestinationIds == "" || b.DestinationIds == utils.ANY) &&
		b.RatingSubject == "" &&
		b.Category == "" &&
		b.ExpirationDate.IsZero() &&
		b.SharedGroup == "" &&
		b.Weight == 0 &&
		b.Disabled == false
}

type UnitsCounter struct {
	Direction   string
	BalanceType string
	//	Units     float64
	Balances BalanceChain // first balance is the general one (no destination)
}

type ActionTriggers []*ActionTrigger

type ActionTrigger struct {
	Id                    string
	ThresholdType         string
	ThresholdValue        float64
	Recurrent             bool
	MinSleep              time.Duration
	BalanceId             string
	BalanceType           string
	BalanceDirection      string
	BalanceDestinationIds string
	BalanceWeight         float64
	BalanceExpirationDate time.Time
	BalanceTimingTags     string
	BalanceRatingSubject  string
	BalanceCategory       string
	BalanceSharedGroup    string
	BalanceDisabled       bool
	Weight                float64
	ActionsId             string
	MinQueuedItems        int
	Executed              bool
}
type Actions []*Action

type Action struct {
	Id               string
	ActionType       string
	BalanceType      string
	Direction        string
	ExtraParameters  string
	ExpirationString string
	Weight           float64
	Balance          *Balance
}

type ActionPlan struct {
	Uuid       string // uniquely identify the timing
	Id         string // informative purpose only
	AccountIds []string
	Timing     *engine.RateInterval
	Weight     float64
	ActionsId  string
	actions    Actions
	stCache    time.Time // cached time of the next start
}

func (at *ActionPlan) IsASAP() bool {
	if at.Timing == nil {
		return false
	}
	return at.Timing.Timing.StartTime == utils.ASAP
}

type SharedGroup struct {
	Id                string
	AccountParameters map[string]*engine.SharingParameters
	MemberIds         []string
}

type ActionPlans []*ActionPlan

func (mig MigratorRC8) migrateAccounts() error {
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
		var oldAcc Account
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
		// fix id
		idElements := strings.Split(newAcc.ID, utils.CONCATENATED_KEY_SEP)
		if len(idElements) != 3 {
			log.Printf("Malformed account ID %s", oldAcc.Id)
			continue
		}
		newAcc.ID = fmt.Sprintf("%s:%s", idElements[1], idElements[2])
		// balances
		balanceErr := false
		for oldBalKey, oldBalChain := range oldAcc.BalanceMap {
			keyElements := strings.Split(oldBalKey, "*")
			if len(keyElements) != 3 {
				log.Printf("Malformed balance key in %s: %s", oldAcc.Id, oldBalKey)
				balanceErr = true
				break
			}
			newBalKey := "*" + keyElements[1]
			newBalDirection := "*" + keyElements[2]
			newAcc.BalanceMap[newBalKey] = make(engine.Balances, len(oldBalChain))
			for index, oldBal := range oldBalChain {
				// check default to set new id
				if oldBal.IsDefault() {
					oldBal.Id = utils.META_DEFAULT
				}
				newAcc.BalanceMap[newBalKey][index] = &engine.Balance{
					Uuid:           oldBal.Uuid,
					ID:             oldBal.Id,
					Value:          oldBal.Value,
					Directions:     utils.ParseStringMap(newBalDirection),
					ExpirationDate: oldBal.ExpirationDate,
					Weight:         oldBal.Weight,
					DestinationIDs: utils.ParseStringMap(oldBal.DestinationIds),
					RatingSubject:  oldBal.RatingSubject,
					Categories:     utils.ParseStringMap(oldBal.Category),
					SharedGroups:   utils.ParseStringMap(oldBal.SharedGroup),
					Timings:        oldBal.Timings,
					TimingIDs:      utils.ParseStringMap(oldBal.TimingIDs),
					Disabled:       oldBal.Disabled,
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
				bf := &engine.BalanceFilter{}
				if oldUcBal.Uuid != "" {
					bf.Uuid = utils.StringPointer(oldUcBal.Uuid)
				}
				if oldUcBal.Id != "" {
					bf.ID = utils.StringPointer(oldUcBal.Id)
				}
				if oldUc.BalanceType != "" {
					bf.Type = utils.StringPointer(oldUc.BalanceType)
				}
				// the value was used for counter value
				/*if oldUcBal.Value != 0 {
					bf.Value = utils.Float64Pointer(oldUcBal.Value)
				}*/
				if oldUc.Direction != "" {
					bf.Directions = utils.StringMapPointer(utils.ParseStringMap(oldUc.Direction))
				}
				if !oldUcBal.ExpirationDate.IsZero() {
					bf.ExpirationDate = utils.TimePointer(oldUcBal.ExpirationDate)
				}
				if oldUcBal.Weight != 0 {
					bf.Weight = utils.Float64Pointer(oldUcBal.Weight)
				}
				if oldUcBal.DestinationIds != "" {
					bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(oldUcBal.DestinationIds))
				}
				if oldUcBal.RatingSubject != "" {
					bf.RatingSubject = utils.StringPointer(oldUcBal.RatingSubject)
				}
				if oldUcBal.Category != "" {
					bf.Categories = utils.StringMapPointer(utils.ParseStringMap(oldUcBal.Category))
				}
				if oldUcBal.SharedGroup != "" {
					bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(oldUcBal.SharedGroup))
				}
				if oldUcBal.TimingIDs != "" {
					bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(oldUcBal.TimingIDs))
				}
				if oldUcBal.Disabled != false {
					bf.Disabled = utils.BoolPointer(oldUcBal.Disabled)
				}
				bf.Timings = oldUcBal.Timings
				cf := &engine.CounterFilter{
					Value:  oldUcBal.Value,
					Filter: bf,
				}
				newUc.Counters[index] = cf
			}
			newAcc.UnitCounters[oldUc.BalanceType] = append(newAcc.UnitCounters[oldUc.BalanceType], newUc)
		}
		// action triggers
		for index, oldAtr := range oldAcc.ActionTriggers {
			at := &engine.ActionTrigger{
				UniqueID:       oldAtr.Id,
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
			if oldAtr.BalanceDirection != "" {
				bf.Directions = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceDirection))
			}
			if oldAtr.BalanceDestinationIds != "" {
				bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceDestinationIds))
			}
			if oldAtr.BalanceTimingTags != "" {
				bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceTimingTags))
			}
			if oldAtr.BalanceCategory != "" {
				bf.Categories = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceCategory))
			}
			if oldAtr.BalanceSharedGroup != "" {
				bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceSharedGroup))
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
			newAcc.ActionTriggers[index] = at
			if newAcc.ActionTriggers[index].ThresholdType == "*min_counter" ||
				newAcc.ActionTriggers[index].ThresholdType == "*max_counter" {
				newAcc.ActionTriggers[index].ThresholdType = strings.Replace(newAcc.ActionTriggers[index].ThresholdType, "_", "_event_", 1)
			}
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
	// delete old data
	log.Printf("Deleting migrated accounts...")
	for _, key := range migratedKeys {
		if err := mig.db.Cmd("DEL", key).Err; err != nil {
			return err
		}
	}
	notMigrated := len(keys) - len(migratedKeys)
	if notMigrated > 0 {
		log.Printf("WARNING: there are %d accounts that failed migration!", notMigrated)
	}
	return err
}

func (mig MigratorRC8) migrateActionTriggers() error {
	keys, err := mig.db.Cmd("KEYS", utils.ACTION_TRIGGER_PREFIX+"*").List()
	if err != nil {
		return err
	}
	newAtrsMap := make(map[string]engine.ActionTriggers, len(keys))
	for _, key := range keys {
		log.Printf("Migrating action trigger: %s...", key)
		var oldAtrs ActionTriggers
		var values []byte
		if values, err = mig.db.Cmd("GET", key).Bytes(); err == nil {
			if err := mig.ms.Unmarshal(values, &oldAtrs); err != nil {
				return err
			}
		}
		newAtrs := make(engine.ActionTriggers, len(oldAtrs))
		for index, oldAtr := range oldAtrs {
			at := &engine.ActionTrigger{
				UniqueID:       oldAtr.Id,
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
			if oldAtr.BalanceDirection != "" {
				bf.Directions = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceDirection))
			}
			if oldAtr.BalanceDestinationIds != "" {
				bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceDestinationIds))
			}
			if oldAtr.BalanceTimingTags != "" {
				bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceTimingTags))
			}
			if oldAtr.BalanceCategory != "" {
				bf.Categories = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceCategory))
			}
			if oldAtr.BalanceSharedGroup != "" {
				bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(oldAtr.BalanceSharedGroup))
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
			if newAtrs[index].ThresholdType == "*min_counter" ||
				newAtrs[index].ThresholdType == "*max_counter" {
				newAtrs[index].ThresholdType = strings.Replace(newAtrs[index].ThresholdType, "_", "_event_", 1)
			}
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

func (mig MigratorRC8) migrateActions() error {
	keys, err := mig.db.Cmd("KEYS", utils.ACTION_PREFIX+"*").List()
	if err != nil {
		return err
	}
	newAcsMap := make(map[string]engine.Actions, len(keys))
	for _, key := range keys {
		log.Printf("Migrating action: %s...", key)
		var oldAcs Actions
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
				bf.Value = &utils.ValueFormula{Static: oldAc.Balance.Value}
			}
			if oldAc.Balance.RatingSubject != "" {
				bf.RatingSubject = utils.StringPointer(oldAc.Balance.RatingSubject)
			}
			if oldAc.Balance.DestinationIds != "" {
				bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(oldAc.Balance.DestinationIds))
			}
			if oldAc.Balance.TimingIDs != "" {
				bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(oldAc.Balance.TimingIDs))
			}
			if oldAc.Balance.Category != "" {
				bf.Categories = utils.StringMapPointer(utils.ParseStringMap(oldAc.Balance.Category))
			}
			if oldAc.Balance.SharedGroup != "" {
				bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(oldAc.Balance.SharedGroup))
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

func (mig MigratorRC8) migrateDerivedChargers() error {
	keys, err := mig.db.Cmd("KEYS", utils.DERIVEDCHARGERS_PREFIX+"*").List()
	if err != nil {
		return err
	}
	newDcsMap := make(map[string]*utils.DerivedChargers, len(keys))
	for _, key := range keys {
		log.Printf("Migrating derived charger: %s...", key)
		var oldDcs []*utils.DerivedCharger
		var values []byte
		if values, err = mig.db.Cmd("GET", key).Bytes(); err == nil {
			if err := mig.ms.Unmarshal(values, &oldDcs); err != nil {
				return err
			}
		}
		newDcs := &utils.DerivedChargers{
			DestinationIDs: make(utils.StringMap),
			Chargers:       oldDcs,
		}
		newDcsMap[key] = newDcs
	}
	// write data back
	for key, dcs := range newDcsMap {
		result, err := mig.ms.Marshal(&dcs)
		if err != nil {
			return err
		}
		if err = mig.db.Cmd("SET", key, result).Err; err != nil {
			return err
		}
	}
	return nil
}

func (mig MigratorRC8) migrateActionPlans() error {
	keys, err := mig.db.Cmd("KEYS", utils.ACTION_PLAN_PREFIX+"*").List()
	if err != nil {
		return err
	}
	aplsMap := make(map[string]ActionPlans, len(keys))
	for _, key := range keys {
		log.Printf("Migrating action plans: %s...", key)
		var apls ActionPlans
		var values []byte
		if values, err = mig.db.Cmd("GET", key).Bytes(); err == nil {
			if err := mig.ms.Unmarshal(values, &apls); err != nil {
				return err
			}
		}
		// change all AccountIds
		for _, apl := range apls {
			for idx, actionId := range apl.AccountIds {
				// fix id
				idElements := strings.Split(actionId, utils.CONCATENATED_KEY_SEP)
				if len(idElements) != 3 {
					//log.Printf("Malformed account ID %s", actionId)
					continue
				}
				apl.AccountIds[idx] = fmt.Sprintf("%s:%s", idElements[1], idElements[2])
			}
		}
		aplsMap[key] = apls
	}
	// write data back
	newAplMap := make(map[string]*engine.ActionPlan)
	for key, apls := range aplsMap {
		for _, apl := range apls {
			newApl, exists := newAplMap[key]
			if !exists {
				newApl = &engine.ActionPlan{
					Id:         apl.Id,
					AccountIDs: make(utils.StringMap),
				}
				newAplMap[key] = newApl
			}
			if !apl.IsASAP() {
				for _, accID := range apl.AccountIds {
					if _, exists := newApl.AccountIDs[accID]; !exists {
						newApl.AccountIDs[accID] = true
					}
				}
			}
			newApl.ActionTimings = append(newApl.ActionTimings, &engine.ActionTiming{
				Uuid:      utils.GenUUID(),
				Timing:    apl.Timing,
				ActionsID: apl.ActionsId,
				Weight:    apl.Weight,
			})
		}
	}
	for key, apl := range newAplMap {
		result, err := mig.ms.Marshal(apl)
		if err != nil {
			return err
		}
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		w.Write(result)
		w.Close()
		if err = mig.db.Cmd("SET", key, b.Bytes()).Err; err != nil {
			return err
		}
	}
	return nil
}

func (mig MigratorRC8) migrateSharedGroups() error {
	keys, err := mig.db.Cmd("KEYS", utils.SHARED_GROUP_PREFIX+"*").List()
	if err != nil {
		return err
	}
	newShgMap := make(map[string]*engine.SharedGroup, len(keys))
	for _, key := range keys {
		log.Printf("Migrating shared groups: %s...", key)
		oldShg := SharedGroup{}
		var values []byte
		if values, err = mig.db.Cmd("GET", key).Bytes(); err == nil {
			if err := mig.ms.Unmarshal(values, &oldShg); err != nil {
				return err
			}
		}
		newShg := &engine.SharedGroup{
			Id:                oldShg.Id,
			AccountParameters: oldShg.AccountParameters,
			MemberIds:         make(utils.StringMap),
		}
		for _, accID := range oldShg.MemberIds {
			newShg.MemberIds[accID] = true
		}
		newShgMap[key] = newShg
	}
	// write data back
	for key, shg := range newShgMap {
		result, err := mig.ms.Marshal(&shg)
		if err != nil {
			return err
		}
		if err = mig.db.Cmd("SET", key, result).Err; err != nil {
			return err
		}
	}
	return nil
}

func (mig MigratorRC8) writeVersion() error {
	result, err := mig.ms.Marshal(engine.CurrentVersion)
	if err != nil {
		return err
	}
	return mig.db.Cmd("SET", utils.VERSION_PREFIX+"struct", result).Err
}
