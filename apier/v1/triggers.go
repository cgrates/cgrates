package v1

import (
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
}

func (self *ApierV1) AddAccountActionTriggers(attr AttrAddAccountActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	actTime, err := utils.ParseTimeDetectLayout(attr.ActivationDate, self.Config.DefaultTimezone)
	if err != nil {
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

type AttrSetAccountActionTriggers struct {
	Tenant                string
	Account               string
	GroupID               string
	UniqueID              string
	ThresholdType         *string
	ThresholdValue        *float64
	Recurrent             *bool
	MinSleep              *string
	ExpirationDate        *string
	ActivationDate        *string
	BalanceId             *string
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
	ActionsId             *string
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
				if attr.BalanceId != nil {
					at.BalanceId = *attr.BalanceId
				}
				if attr.BalanceType != nil {
					at.BalanceType = *attr.BalanceType
				}
				if attr.BalanceDirections != nil {
					at.BalanceDirections = utils.NewStringMap(*attr.BalanceDirections...)
				}
				if attr.BalanceDestinationIds != nil {
					at.BalanceDestinationIds = utils.NewStringMap(*attr.BalanceDestinationIds...)
				}
				if attr.BalanceWeight != nil {
					at.BalanceWeight = *attr.BalanceWeight
				}
				if attr.BalanceExpirationDate != nil {
					balanceExpTime, err := utils.ParseDate(*attr.BalanceExpirationDate)
					if err != nil {
						return 0, err
					}
					at.BalanceExpirationDate = balanceExpTime
				}
				if attr.BalanceTimingTags != nil {
					at.BalanceTimingTags = utils.NewStringMap(*attr.BalanceTimingTags...)
				}
				if attr.BalanceRatingSubject != nil {
					at.BalanceRatingSubject = *attr.BalanceRatingSubject
				}
				if attr.BalanceCategories != nil {
					at.BalanceCategories = utils.NewStringMap(*attr.BalanceCategories...)
				}
				if attr.BalanceSharedGroups != nil {
					at.BalanceSharedGroups = utils.NewStringMap(*attr.BalanceSharedGroups...)
				}
				if attr.BalanceBlocker != nil {
					at.BalanceBlocker = *attr.BalanceBlocker
				}
				if attr.BalanceDisabled != nil {
					at.BalanceDisabled = *attr.BalanceDisabled
				}
				if attr.MinQueuedItems != nil {
					at.MinQueuedItems = *attr.MinQueuedItems
				}
				if attr.ActionsId != nil {
					at.ActionsId = *attr.ActionsId
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
