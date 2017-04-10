package migrator

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v1ActionTrigger struct {
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

type v1ActionTriggers []*v1ActionTrigger

func (m *Migrator) migrateActionTriggers() (err error) {
	switch m.dataDBType {
	case utils.REDIS:
		var atrrs engine.ActionTriggers
		var v1atrskeys []string
		v1atrskeys, err = m.dataDB.GetKeysForPrefix(utils.ACTION_TRIGGER_PREFIX)
		if err != nil {
			return
		}
		for _, v1atrskey := range v1atrskeys {
			v1atrs, err := m.getV1ActionTriggerFromDB(v1atrskey)
			if err != nil {
				return err
			}
			v1atr := v1atrs
			if v1atrs != nil {
				atr := v1atr.AsActionTrigger()
				atrrs = append(atrrs, atr)
			}
		}
		if err := m.dataDB.SetActionTriggers(atrrs[0].ID, atrrs, utils.NonTransactional); err != nil {
			return err
		}
		// All done, update version wtih current one
		vrs := engine.Versions{utils.ACTION_TRIGGER_PREFIX: engine.CurrentStorDBVersions()[utils.ACTION_TRIGGER_PREFIX]}
		if err = m.dataDB.SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating ActionTrigger version into StorDB", err.Error()))
		}
		return
	case utils.MONGO:
		dataDB := m.dataDB.(*engine.MongoStorage)
		mgoDB := dataDB.DB()
		defer mgoDB.Session.Close()
		var atrrs engine.ActionTriggers
		var v1atr v1ActionTrigger
		iter := mgoDB.C(utils.ACTION_TRIGGER_PREFIX).Find(nil).Iter()
		for iter.Next(&v1atr) {
			atr := v1atr.AsActionTrigger()
			atrrs = append(atrrs, atr)
		}
		if err := m.dataDB.SetActionTriggers(atrrs[0].ID, atrrs, utils.NonTransactional); err != nil {
			return err
		}
		// All done, update version wtih current one
		vrs := engine.Versions{utils.ACTION_TRIGGER_PREFIX: engine.CurrentStorDBVersions()[utils.ACTION_TRIGGER_PREFIX]}
		if err = m.dataDB.SetVersions(vrs, false); err != nil {
			return utils.NewCGRError(utils.Migrator,
				utils.ServerErrorCaps,
				err.Error(),
				fmt.Sprintf("error: <%s> when updating ActionTrigger version into StorDB", err.Error()))
		}
		return
	default:
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			utils.UnsupportedDB,
			fmt.Sprintf("error: unsupported: <%s> for migrateActionTriggers method", m.dataDBType))
	}
}
func (m *Migrator) getV1ActionTriggerFromDB(key string) (v1Atr *v1ActionTrigger, err error) {
	switch m.dataDBType {
	case utils.REDIS:
		dataDB := m.dataDB.(*engine.RedisStorage)
		if strVal, err := dataDB.Cmd("GET", key).Bytes(); err != nil {
			return nil, err
		} else {
			v1Atr := &v1ActionTrigger{Id: key}
			if err := m.mrshlr.Unmarshal(strVal, &v1Atr); err != nil {
				return nil, err
			}
			return v1Atr, nil
		}
	case utils.MONGO:
		dataDB := m.dataDB.(*engine.MongoStorage)
		mgoDB := dataDB.DB()
		defer mgoDB.Session.Close()
		v1Atr := new(v1ActionTrigger)
		if err := mgoDB.C(utils.ACTION_TRIGGER_PREFIX).Find(bson.M{"id": key}).One(v1Atr); err != nil {
			return nil, err
		}
		return v1Atr, nil
	default:
		return nil, utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			utils.UnsupportedDB,
			fmt.Sprintf("error: unsupported: <%s> for getV1ActionTriggerFromDB method", m.dataDBType))
	}
}

func (v1Act v1ActionTrigger) AsActionTrigger() (at *engine.ActionTrigger) {
	at = &engine.ActionTrigger{
		ID: v1Act.Id,
		//	UniqueID:       utils.GenUUID(),
		ThresholdType:  v1Act.ThresholdType,
		ThresholdValue: v1Act.ThresholdValue,
		Recurrent:      v1Act.Recurrent,
		MinSleep:       v1Act.MinSleep,
		Weight:         v1Act.Weight,
		ActionsID:      v1Act.ActionsId,
		MinQueuedItems: v1Act.MinQueuedItems,
		Executed:       v1Act.Executed,
	}
	bf := &engine.BalanceFilter{}
	if v1Act.BalanceId != "" {
		bf.ID = utils.StringPointer(v1Act.BalanceId)
	}
	if v1Act.BalanceType != "" {
		bf.Type = utils.StringPointer(v1Act.BalanceType)
	}
	if v1Act.BalanceRatingSubject != "" {
		bf.RatingSubject = utils.StringPointer(v1Act.BalanceRatingSubject)
	}
	if v1Act.BalanceDirection != "" {
		bf.Directions = utils.StringMapPointer(utils.ParseStringMap(v1Act.BalanceDirection))
	}
	if v1Act.BalanceDestinationIds != "" {
		bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(v1Act.BalanceDestinationIds))
	}
	if v1Act.BalanceTimingTags != "" {
		bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(v1Act.BalanceTimingTags))
	}
	if v1Act.BalanceCategory != "" {
		bf.Categories = utils.StringMapPointer(utils.ParseStringMap(v1Act.BalanceCategory))
	}
	if v1Act.BalanceSharedGroup != "" {
		bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(v1Act.BalanceSharedGroup))
	}
	if v1Act.BalanceWeight != 0 {
		bf.Weight = utils.Float64Pointer(v1Act.BalanceWeight)
	}
	if v1Act.BalanceDisabled != false {
		bf.Disabled = utils.BoolPointer(v1Act.BalanceDisabled)
	}
	if !v1Act.BalanceExpirationDate.IsZero() {
		bf.ExpirationDate = utils.TimePointer(v1Act.BalanceExpirationDate)
		at.ExpirationDate = v1Act.BalanceExpirationDate
		at.LastExecutionTime = v1Act.BalanceExpirationDate
		at.ActivationDate = v1Act.BalanceExpirationDate
	}
	at.Balance = bf
	if at.ThresholdType == "*min_counter" ||
		at.ThresholdType == "*max_counter" {
		at.ThresholdType = strings.Replace(at.ThresholdType, "_", "_event_", 1)
	}
	return
}
