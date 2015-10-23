package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/hoisie/redis"
)

const OLD_ACCOUNT_PREFIX = "ubl_"

type MigratorRC8 struct {
	dbNb int
	db   *redis.Client
	ms   engine.Marshaler
}

func NewMigratorRC8(address string, db int, pass, mrshlerStr string) (*MigratorRC8, error) {
	ndb := &redis.Client{Addr: address, Db: db, Password: pass}

	var mrshler engine.Marshaler
	if mrshlerStr == utils.MSGPACK {
		mrshler = engine.NewCodecMsgpackMarshaler()
	} else if mrshlerStr == utils.JSON {
		mrshler = new(engine.JSONMarshaler)
	} else {
		return nil, fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	return &MigratorRC8{db: ndb, dbNb: db, ms: mrshler}, nil
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

func (mig MigratorRC8) migrateAccounts() error {
	keys, err := mig.db.Keys(OLD_ACCOUNT_PREFIX + "*")
	if err != nil {
		return err
	}
	newAccounts := make([]*engine.Account, len(keys))
	// get existing accounts
	for keyIndex, key := range keys {
		values, err := mig.db.Get(key)
		if err != nil {
			continue
		}
		var oldAcc Account
		if err = mig.ms.Unmarshal(values, &oldAcc); err != nil {
			return err
		}
		// transfer data into new structurse
		newAcc := &engine.Account{
			Id:             oldAcc.Id,
			BalanceMap:     make(map[string]engine.BalanceChain, len(oldAcc.BalanceMap)),
			UnitCounters:   make([]*engine.UnitsCounter, len(oldAcc.UnitCounters)),
			ActionTriggers: make(engine.ActionTriggers, len(oldAcc.ActionTriggers)),
			AllowNegative:  oldAcc.AllowNegative,
			Disabled:       oldAcc.Disabled,
		}
		// balances
		for oldBalKey, oldBalChain := range oldAcc.BalanceMap {
			keyElements := strings.Split(oldBalKey, "*")
			if len(key) != 4 {
				return fmt.Errorf("Malformed balance key in %s: %s", oldAcc.Id, oldBalKey)
			}
			newBalKey := "*" + keyElements[1]
			newBalDirection := "*" + keyElements[3]
			newAcc.BalanceMap[newBalKey] = make(engine.BalanceChain, len(oldBalChain))
			for index, oldBal := range oldBalChain {
				newAcc.BalanceMap[newBalKey][index] = &engine.Balance{
					Uuid:           oldBal.Uuid,
					Id:             oldBal.Id,
					Value:          oldBal.Value,
					Directions:     utils.ParseStringMap(newBalDirection),
					ExpirationDate: oldBal.ExpirationDate,
					Weight:         oldBal.Weight,
					DestinationIds: utils.ParseStringMap(oldBal.DestinationIds),
					RatingSubject:  oldBal.RatingSubject,
					Categories:     utils.ParseStringMap(oldBal.Category),
					SharedGroup:    oldBal.SharedGroup,
					Timings:        oldBal.Timings,
					TimingIDs:      oldBal.TimingIDs,
					Disabled:       oldBal.Disabled,
				}
			}
		}
		// unit counters
		for _, oldUc := range oldAcc.UnitCounters {
			newUc := &engine.UnitsCounter{
				BalanceType: oldUc.BalanceType,
				Balances:    make(engine.BalanceChain, len(oldUc.Balances)),
			}
			for index, oldUcBal := range oldUc.Balances {
				newUc.Balances[index] = &engine.Balance{
					Uuid:           oldUcBal.Uuid,
					Id:             oldUcBal.Id,
					Value:          oldUcBal.Value,
					Directions:     utils.ParseStringMap(oldUc.Direction),
					ExpirationDate: oldUcBal.ExpirationDate,
					Weight:         oldUcBal.Weight,
					DestinationIds: utils.ParseStringMap(oldUcBal.DestinationIds),
					RatingSubject:  oldUcBal.RatingSubject,
					Categories:     utils.ParseStringMap(oldUcBal.Category),
					SharedGroup:    oldUcBal.SharedGroup,
					Timings:        oldUcBal.Timings,
					TimingIDs:      oldUcBal.TimingIDs,
					Disabled:       oldUcBal.Disabled,
				}
			}
		}
		// action triggers

		for index, oldAtr := range oldAcc.ActionTriggers {
			newAcc.ActionTriggers[index] = &engine.ActionTrigger{
				Id:                    oldAtr.Id,
				ThresholdType:         oldAtr.ThresholdType,
				ThresholdValue:        oldAtr.ThresholdValue,
				Recurrent:             oldAtr.Recurrent,
				MinSleep:              oldAtr.MinSleep,
				BalanceId:             oldAtr.BalanceId,
				BalanceType:           oldAtr.BalanceType,
				BalanceDirections:     utils.ParseStringMap(oldAtr.BalanceDirection),
				BalanceDestinationIds: utils.ParseStringMap(oldAtr.BalanceDestinationIds),
				BalanceWeight:         oldAtr.BalanceWeight,
				BalanceExpirationDate: oldAtr.BalanceExpirationDate,
				BalanceTimingTags:     oldAtr.BalanceTimingTags,
				BalanceRatingSubject:  oldAtr.BalanceRatingSubject,
				BalanceCategories:     utils.ParseStringMap(oldAtr.BalanceCategory),
				BalanceSharedGroup:    oldAtr.BalanceSharedGroup,
				BalanceDisabled:       oldAtr.BalanceDisabled,
				Weight:                oldAtr.Weight,
				ActionsId:             oldAtr.ActionsId,
				MinQueuedItems:        oldAtr.MinQueuedItems,
				Executed:              oldAtr.Executed,
			}
		}
		newAccounts[keyIndex] = newAcc
	}
	// write data back
	for _, newAcc := range newAccounts {
		result, err := mig.ms.Marshal(newAcc)
		if err != nil {
			return err
		}
		if err := mig.db.Set(utils.ACCOUNT_PREFIX+newAcc.Id, result); err != nil {
			return err
		}
	}
	// delete old data
	_, err = mig.db.Del(OLD_ACCOUNT_PREFIX + "*")
	return err
}

func (mig MigratorRC8) migrateActionTriggers() error {
	keys, err := mig.db.Keys(utils.ACTION_TRIGGER_PREFIX + "*")
	if err != nil {
		return err
	}
	newAtrsMap := make(map[string]engine.ActionTriggers, len(keys))
	for _, key := range keys {
		var oldAtrs ActionTriggers
		var values []byte
		if values, err = mig.db.Get(key); err == nil {
			if err := mig.ms.Unmarshal(values, &oldAtrs); err != nil {
				return err
			}
		}
		newAtrs := make(engine.ActionTriggers, len(oldAtrs))
		for index, oldAtr := range oldAtrs {
			newAtrs[index] = &engine.ActionTrigger{
				Id:                    oldAtr.Id,
				ThresholdType:         oldAtr.ThresholdType,
				ThresholdValue:        oldAtr.ThresholdValue,
				Recurrent:             oldAtr.Recurrent,
				MinSleep:              oldAtr.MinSleep,
				BalanceId:             oldAtr.BalanceId,
				BalanceType:           oldAtr.BalanceType,
				BalanceDirections:     utils.ParseStringMap(oldAtr.BalanceDirection),
				BalanceDestinationIds: utils.ParseStringMap(oldAtr.BalanceDestinationIds),
				BalanceWeight:         oldAtr.BalanceWeight,
				BalanceExpirationDate: oldAtr.BalanceExpirationDate,
				BalanceTimingTags:     oldAtr.BalanceTimingTags,
				BalanceRatingSubject:  oldAtr.BalanceRatingSubject,
				BalanceCategories:     utils.ParseStringMap(oldAtr.BalanceCategory),
				BalanceSharedGroup:    oldAtr.BalanceSharedGroup,
				BalanceDisabled:       oldAtr.BalanceDisabled,
				Weight:                oldAtr.Weight,
				ActionsId:             oldAtr.ActionsId,
				MinQueuedItems:        oldAtr.MinQueuedItems,
				Executed:              oldAtr.Executed,
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
		if err = mig.db.Set(key, result); err != nil {
			return err
		}
	}
	return nil
}
