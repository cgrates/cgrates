package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrAddActionTrigger struct {
	ActionTriggersId       string
	ActionTriggersUniqueId string
	Tenant                 string
	Account                string
	ThresholdType          string
	ThresholdValue         float64
	BalanceId              string
	BalanceType            string
	BalanceDirection       string
	BalanceDestinationIds  string
	BalanceRatingSubject   string //ToDo
	BalanceWeight          float64
	BalanceExpiryTime      string
	BalanceSharedGroup     string //ToDo
	Weight                 float64
	ActionsId              string
}

func (self *ApierV1) AddTriggeredAction(attr AttrAddActionTrigger, reply *string) error {
	if attr.BalanceDirection == "" {
		attr.BalanceDirection = utils.OUT
	}
	balExpiryTime, err := utils.ParseTimeDetectLayout(attr.BalanceExpiryTime, self.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	at := &engine.ActionTrigger{
		ID:                    attr.ActionTriggersId,
		UniqueID:              attr.ActionTriggersUniqueId,
		ThresholdType:         attr.ThresholdType,
		ThresholdValue:        attr.ThresholdValue,
		BalanceId:             attr.BalanceId,
		BalanceType:           attr.BalanceType,
		BalanceDirections:     utils.ParseStringMap(attr.BalanceDirection),
		BalanceDestinationIds: utils.ParseStringMap(attr.BalanceDestinationIds),
		BalanceWeight:         attr.BalanceWeight,
		BalanceExpirationDate: balExpiryTime,
		Weight:                attr.Weight,
		ActionsId:             attr.ActionsId,
		Executed:              false,
	}

	tag := utils.AccountKey(attr.Tenant, attr.Account)
	_, err = engine.Guardian.Guard(func() (interface{}, error) {
		userBalance, err := self.AccountDb.GetAccount(tag)
		if err != nil {
			return 0, err
		}

		userBalance.ActionTriggers = append(userBalance.ActionTriggers, at)

		if err = self.AccountDb.SetAccount(userBalance); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, tag)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

type AttrSetAccountActionTriggers struct {
	Tenant                 string
	Account                string
	ActionTriggersIDs      *[]string
	ActionTriggerOverwrite bool
}

func (self *ApierV1) SetAccountActionTriggers(attr AttrSetAccountActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	var account *engine.Account
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if acc, err := self.AccountDb.GetAccount(accID); err == nil {
			account = acc
		} else {
			return 0, err
		}
		if attr.ActionTriggersIDs != nil {
			if attr.ActionTriggerOverwrite {
				account.ActionTriggers = make(engine.ActionTriggers, 0)
			}
			for _, actionTriggerID := range *attr.ActionTriggersIDs {
				atrs, err := self.RatingDb.GetActionTriggers(actionTriggerID)
				if err != nil {

					return 0, err
				}
				for _, at := range atrs {
					var found bool
					for _, existingAt := range account.ActionTriggers {
						if existingAt.Equals(at) {
							found = true
							break
						}
					}
					if !found {
						account.ActionTriggers = append(account.ActionTriggers, at)
					}
				}
			}
		}
		account.InitCounters()
		if err := self.AccountDb.SetAccount(account); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrRemoveAccountActionTriggers struct {
	Tenant   string
	Account  string
	GroupID  string
	UniqueID string
}

func (self *ApierV1) RemoveAccountActionTriggers(attr AttrRemoveAccountActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		var account *engine.Account
		if acc, err := self.AccountDb.GetAccount(accID); err == nil {
			account = acc
		} else {
			return 0, err
		}
		var newActionTriggers engine.ActionTriggers
		for _, at := range account.ActionTriggers {
			if (attr.UniqueID == "" || at.UniqueID == attr.UniqueID) &&
				(attr.GroupID == "" || at.ID == attr.GroupID) {
				// remove action trigger
				continue
			}
			newActionTriggers = append(newActionTriggers, at)
		}
		account.ActionTriggers = newActionTriggers
		account.InitCounters()
		if err := self.AccountDb.SetAccount(account); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

func (self *ApierV1) ResetAccountActionTriggers(attr AttrRemoveAccountActionTriggers, reply *string) error {

	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	var account *engine.Account
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if acc, err := self.AccountDb.GetAccount(accID); err == nil {
			account = acc
		} else {
			return 0, err
		}
		for _, at := range account.ActionTriggers {
			if (attr.UniqueID == "" || at.UniqueID == attr.UniqueID) &&
				(attr.GroupID == "" || at.ID == attr.GroupID) {
				// reset action trigger
				at.Executed = false
			}

		}
		account.ExecuteActionTriggers(nil)
		if err := self.AccountDb.SetAccount(account); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}
