package v1

import (
	"log"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Returns a list of ActionTriggers on an account
func (self *ApierV1) GetAccountActionTriggers(attrs AttrAcntAction, reply *engine.ActionTriggers) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if account, err := self.AccountDb.GetAccount(utils.AccountKey(attrs.Tenant, attrs.Account)); err != nil {
		return utils.NewErrServerError(err)
	} else {
		ats := account.ActionTriggers
		if ats == nil {
			ats = engine.ActionTriggers{}
		}
		*reply = ats
	}
	return nil
}

type AttrAddAccountActionTriggers struct {
	Tenant                 string
	Account                string
	ActionTriggerIDs       *[]string
	ActionTriggerOverwrite bool
	ActivationDate         string
	Executed               bool
}

func (self *ApierV1) AddAccountActionTriggers(attr AttrAddAccountActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	actTime, err := utils.ParseTimeDetectLayout(attr.ActivationDate, self.Config.DefaultTimezone)
	if err != nil {
		*reply = err.Error()
		return err
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	var account *engine.Account
	_, err = engine.Guardian.Guard(func() (interface{}, error) {
		if acc, err := self.AccountDb.GetAccount(accID); err == nil {
			account = acc
		} else {
			return 0, err
		}
		if attr.ActionTriggerIDs != nil {
			if attr.ActionTriggerOverwrite {
				account.ActionTriggers = make(engine.ActionTriggers, 0)
			}
			for _, actionTriggerID := range *attr.ActionTriggerIDs {
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
					at.ActivationDate = actTime
					at.Executed = attr.Executed
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

type AttrResetAccountActionTriggers struct {
	Tenant   string
	Account  string
	GroupID  string
	UniqueID string
	Executed bool
}

func (self *ApierV1) ResetAccountActionTriggers(attr AttrResetAccountActionTriggers, reply *string) error {

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
				at.Executed = attr.Executed
			}

		}
		if attr.Executed == false {
			account.ExecuteActionTriggers(nil)
		}
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

type AttrSetAccountActionTriggers struct {
	Tenant                string
	Account               string
	GroupID               string
	UniqueID              string
	ThresholdType         *string
	ThresholdValue        *float64
	Recurrent             *bool
	Executed              *bool
	MinSleep              *string
	ExpirationDate        *string
	ActivationDate        *string
	BalanceID             *string
	BalanceType           *string
	BalanceDirections     *[]string
	BalanceDestinationIds *[]string
	BalanceWeight         *float64
	BalanceExpirationDate *string
	BalanceTimingTags     *[]string
	BalanceRatingSubject  *string
	BalanceCategories     *[]string
	BalanceSharedGroups   *[]string
	BalanceBlocker        *bool
	BalanceDisabled       *bool
	MinQueuedItems        *int
	ActionsID             *string
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
		log.Print("HERE: ", account.ActionTriggers)
		for _, at := range account.ActionTriggers {
			if (attr.UniqueID == "" || at.UniqueID == attr.UniqueID) &&
				(attr.GroupID == "" || at.ID == attr.GroupID) {
				// we have a winner
				if attr.ThresholdType != nil {
					at.ThresholdType = *attr.ThresholdType
				}
				if attr.ThresholdValue != nil {
					at.ThresholdValue = *attr.ThresholdValue
				}
				if attr.Recurrent != nil {
					at.Recurrent = *attr.Recurrent
				}
				if attr.Executed != nil {
					at.Executed = *attr.Executed
				}
				if attr.MinSleep != nil {
					minSleep, err := utils.ParseDurationWithSecs(*attr.MinSleep)
					if err != nil {
						return 0, err
					}
					at.MinSleep = minSleep
				}
				if attr.ExpirationDate != nil {
					expTime, err := utils.ParseTimeDetectLayout(*attr.ExpirationDate, self.Config.DefaultTimezone)
					if err != nil {
						return 0, err
					}
					at.ExpirationDate = expTime
				}
				if attr.ActivationDate != nil {
					actTime, err := utils.ParseTimeDetectLayout(*attr.ActivationDate, self.Config.DefaultTimezone)
					if err != nil {
						return 0, err
					}
					at.ActivationDate = actTime
				}
				at.Balance = &engine.BalanceFilter{}
				if attr.BalanceID != nil {
					at.Balance.ID = attr.BalanceID
				}
				if attr.BalanceType != nil {
					at.Balance.Type = attr.BalanceType
				}
				if attr.BalanceDirections != nil {
					at.Balance.Directions = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceDirections...))
				}
				if attr.BalanceDestinationIds != nil {
					at.Balance.DestinationIDs = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceDestinationIds...))
				}
				if attr.BalanceWeight != nil {
					at.Balance.Weight = attr.BalanceWeight
				}
				if attr.BalanceExpirationDate != nil {
					balanceExpTime, err := utils.ParseDate(*attr.BalanceExpirationDate)
					if err != nil {
						return 0, err
					}
					at.Balance.ExpirationDate = &balanceExpTime
				}
				if attr.BalanceTimingTags != nil {
					at.Balance.TimingIDs = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceTimingTags...))
				}
				if attr.BalanceRatingSubject != nil {
					at.Balance.RatingSubject = attr.BalanceRatingSubject
				}
				if attr.BalanceCategories != nil {
					at.Balance.Categories = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceCategories...))
				}
				if attr.BalanceSharedGroups != nil {
					at.Balance.SharedGroups = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceSharedGroups...))
				}
				if attr.BalanceBlocker != nil {
					at.Balance.Blocker = attr.BalanceBlocker
				}
				if attr.BalanceDisabled != nil {
					at.Balance.Disabled = attr.BalanceDisabled
				}
				if attr.MinQueuedItems != nil {
					at.MinQueuedItems = *attr.MinQueuedItems
				}
				if attr.ActionsID != nil {
					at.ActionsID = *attr.ActionsID
				}
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

type AttrRemoveActionTrigger struct {
	GroupID  string
	UniqueID string
}

func (self *ApierV1) RemoveActionTrigger(attr AttrRemoveActionTrigger, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"GroupID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attr.UniqueID == "" {
		err := self.RatingDb.RemoveActionTriggers(attr.GroupID)
		if err != nil {
			*reply = err.Error()
		} else {
			*reply = utils.OK
		}
		return err
	} else {
		atrs, err := self.RatingDb.GetActionTriggers(attr.GroupID)
		if err != nil {
			*reply = err.Error()
			return err
		}
		var remainingAtrs engine.ActionTriggers
		for _, atr := range atrs {
			if atr.UniqueID == attr.UniqueID {
				continue
			}
			remainingAtrs = append(remainingAtrs, atr)
		}
		// set the cleared list back
		err = self.RatingDb.SetActionTriggers(attr.GroupID, remainingAtrs)
		if err != nil {
			*reply = err.Error()
		} else {
			*reply = utils.OK
		}
		return err
	}
}

type AttrSetActionTrigger struct {
	GroupID               string
	UniqueID              string
	ThresholdType         *string
	ThresholdValue        *float64
	Recurrent             *bool
	MinSleep              *string
	ExpirationDate        *string
	ActivationDate        *string
	BalanceID             *string
	BalanceType           *string
	BalanceDirections     *[]string
	BalanceDestinationIds *[]string
	BalanceWeight         *float64
	BalanceExpirationDate *string
	BalanceTimingTags     *[]string
	BalanceRatingSubject  *string
	BalanceCategories     *[]string
	BalanceSharedGroups   *[]string
	BalanceBlocker        *bool
	BalanceDisabled       *bool
	MinQueuedItems        *int
	ActionsID             *string
}

func (self *ApierV1) SetActionTrigger(attr AttrSetActionTrigger, reply *string) error {

	if missing := utils.MissingStructFields(&attr, []string{"GroupID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	atrs, _ := self.RatingDb.GetActionTriggers(attr.GroupID)
	var newAtr *engine.ActionTrigger
	if attr.UniqueID != "" {
		//search for exiting one
		for _, atr := range atrs {
			if atr.UniqueID == attr.UniqueID {
				newAtr = atr
				break
			}
		}
	}

	if newAtr == nil {
		newAtr = &engine.ActionTrigger{}
		atrs = append(atrs, newAtr)
	}
	newAtr.ID = attr.GroupID
	if attr.UniqueID != "" {
		newAtr.UniqueID = attr.UniqueID
	} else {
		newAtr.UniqueID = utils.GenUUID()
	}

	if attr.ThresholdType != nil {
		newAtr.ThresholdType = *attr.ThresholdType
	}
	if attr.ThresholdValue != nil {
		newAtr.ThresholdValue = *attr.ThresholdValue
	}
	if attr.Recurrent != nil {
		newAtr.Recurrent = *attr.Recurrent
	}
	if attr.MinSleep != nil {
		minSleep, err := utils.ParseDurationWithSecs(*attr.MinSleep)
		if err != nil {
			*reply = err.Error()
			return err
		}
		newAtr.MinSleep = minSleep
	}
	if attr.ExpirationDate != nil {
		expTime, err := utils.ParseTimeDetectLayout(*attr.ExpirationDate, self.Config.DefaultTimezone)
		if err != nil {
			*reply = err.Error()
			return err
		}
		newAtr.ExpirationDate = expTime
	}
	if attr.ActivationDate != nil {
		actTime, err := utils.ParseTimeDetectLayout(*attr.ActivationDate, self.Config.DefaultTimezone)
		if err != nil {
			*reply = err.Error()
			return err
		}
		newAtr.ActivationDate = actTime
	}
	newAtr.Balance = &engine.BalanceFilter{}
	if attr.BalanceID != nil {
		newAtr.Balance.ID = attr.BalanceID
	}
	if attr.BalanceType != nil {
		newAtr.Balance.Type = attr.BalanceType
	}
	if attr.BalanceDirections != nil {
		newAtr.Balance.Directions = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceDirections...))
	}
	if attr.BalanceDestinationIds != nil {
		newAtr.Balance.DestinationIDs = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceDestinationIds...))
	}
	if attr.BalanceWeight != nil {
		newAtr.Balance.Weight = attr.BalanceWeight
	}
	if attr.BalanceExpirationDate != nil {
		balanceExpTime, err := utils.ParseDate(*attr.BalanceExpirationDate)
		if err != nil {
			*reply = err.Error()
			return err
		}
		newAtr.Balance.ExpirationDate = &balanceExpTime
	}
	if attr.BalanceTimingTags != nil {
		newAtr.Balance.TimingIDs = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceTimingTags...))
	}
	if attr.BalanceRatingSubject != nil {
		newAtr.Balance.RatingSubject = attr.BalanceRatingSubject
	}
	if attr.BalanceCategories != nil {
		newAtr.Balance.Categories = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceCategories...))
	}
	if attr.BalanceSharedGroups != nil {
		newAtr.Balance.SharedGroups = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceSharedGroups...))
	}
	if attr.BalanceBlocker != nil {
		newAtr.Balance.Blocker = attr.BalanceBlocker
	}
	if attr.BalanceDisabled != nil {
		newAtr.Balance.Disabled = attr.BalanceDisabled
	}
	if attr.MinQueuedItems != nil {
		newAtr.MinQueuedItems = *attr.MinQueuedItems
	}
	if attr.ActionsID != nil {
		newAtr.ActionsID = *attr.ActionsID
	}

	if err := self.RatingDb.SetActionTriggers(attr.GroupID, atrs); err != nil {
		*reply = err.Error()
		return err
	}
	//no cache for action triggers
	*reply = utils.OK
	return nil
}

type AttrGetActionTriggers struct {
	GroupIDs []string
}

func (self *ApierV1) GetActionTriggers(attr AttrGetActionTriggers, atrs *engine.ActionTriggers) error {
	var allAttrs engine.ActionTriggers
	if len(attr.GroupIDs) > 0 {
		for _, key := range attr.GroupIDs {
			getAttrs, err := self.RatingDb.GetActionTriggers(key)
			if err != nil {
				return err
			}
			allAttrs = append(allAttrs, getAttrs...)
		}

	} else {
		keys, err := self.RatingDb.GetKeysForPrefix(utils.ACTION_TRIGGER_PREFIX, true)
		if err != nil {
			return err
		}
		for _, key := range keys {
			getAttrs, err := self.RatingDb.GetActionTriggers(key[len(utils.ACTION_TRIGGER_PREFIX):])
			if err != nil {
				return err
			}
			allAttrs = append(allAttrs, getAttrs...)
		}
	}
	*atrs = allAttrs
	return nil
}
