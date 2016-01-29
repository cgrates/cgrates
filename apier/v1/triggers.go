package v1

import (
	"regexp"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Returns a list of ActionTriggers on an account
func (self *ApierV1) GetAccountActionTriggers(attrs AttrAcntAction, reply *engine.ActionTriggers) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if balance, err := self.AccountDb.GetAccount(utils.AccountKey(attrs.Tenant, attrs.Account)); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = balance.ActionTriggers
	}
	return nil
}

type AttrRemAcntActionTriggers struct {
	Tenant                 string // Tenant he account belongs to
	Account                string // Account name
	ActionTriggersId       string // Id filtering only specific id to remove (can be regexp pattern)
	ActionTriggersUniqueId string
}

// Returns a list of ActionTriggers on an account
func (self *ApierV1) RemAccountActionTriggers(attrs AttrRemAcntActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attrs.Tenant, attrs.Account)
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		ub, err := self.AccountDb.GetAccount(accID)
		if err != nil {
			return 0, err
		}
		nactrs := make(engine.ActionTriggers, 0)
		for _, actr := range ub.ActionTriggers {
			match, _ := regexp.MatchString(attrs.ActionTriggersId, actr.ID)
			if len(attrs.ActionTriggersId) != 0 && match {
				continue
			}
			if len(attrs.ActionTriggersUniqueId) != 0 && attrs.ActionTriggersUniqueId == actr.UniqueID {
				continue
			}
			nactrs = append(nactrs, actr)
		}
		ub.ActionTriggers = nactrs
		if err := self.AccountDb.SetAccount(ub); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}

type AttrSetAccountActionTriggers struct {
	Tenant                 string
	Account                string
	ActionTriggerIDs       *[]string
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
