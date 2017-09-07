package migrator

import (
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type v1ActionTrigger struct {
	Id            string // for visual identification
	ThresholdType string //*min_counter, *max_counter, *min_balance, *max_balance
	ThresholdValue        float64
	Recurrent             bool          // reset eexcuted flag each run
	MinSleep              time.Duration // Minimum duration between two executions in case of recurrent triggers
	BalanceId             string
	BalanceType           string
	BalanceDirection      string
	BalanceDestinationIds string    // filter for balance
	BalanceWeight         float64   // filter for balance
	BalanceExpirationDate time.Time // filter for balance
	BalanceTimingTags     string    // filter for balance
	BalanceRatingSubject  string    // filter for balance
	BalanceCategory       string    // filter for balance
	BalanceSharedGroup    string    // filter for balance
	Weight                float64
	ActionsId             string
	MinQueuedItems        int // Trigger actions only if this number is hit (stats only)
	Executed              bool
	lastExecutionTime     time.Time
}

type v1ActionTriggers []*v1ActionTrigger

func (m *Migrator) migrateActionTriggers() (err error) {
	var v1ACTs *v1ActionTriggers
	var acts engine.ActionTriggers
	for {
		v1ACTs, err = m.oldDataDB.getV1ActionTriggers()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if *v1ACTs != nil {
			for _, v1ac := range *v1ACTs {
				act := v1ac.AsActionTrigger()
				acts = append(acts, act)

			}
			if err := m.dataDB.SetActionTriggers(acts[0].ID, acts, utils.NonTransactional); err != nil {
				return err
			}

		}
	}
	// All done, update version wtih current one
	vrs := engine.Versions{utils.ActionTriggers: engine.CurrentDataDBVersions()[utils.ActionTriggers]}
	if err = m.dataDB.SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating ActionTriggers version into DataDB", err.Error()))
	}
	return

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
