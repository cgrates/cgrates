package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrSetActionTriggers struct {
	Tenant                 string
	Account                string
	ActionTriggersIds      *[]string
	ActionTriggerOverwrite bool
}

func (self *ApierV1) SetActionTriggers(attr AttrSetActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	var account *engine.Account
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if acc, err := self.AccountDb.GetAccount(accID); err != nil {
			account = acc
		} else {
			return 0, err
		}
		if attr.ActionTriggersIds != nil {
			if attr.ActionTriggerOverwrite {
				account.ActionTriggers = make(engine.ActionTriggers, 0)
			}
			for _, actionTriggerID := range *attr.ActionTriggersIds {
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
		return 0, nil
	}, 0, accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrRemoveActionTriggers struct {
	Tenant   string
	Account  string
	GroupID  string
	UniqueID string
}

func (self *ApierV1) RemoveActionTriggers(attr AttrRemoveActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	var account *engine.Account
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if acc, err := self.AccountDb.GetAccount(accID); err != nil {
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
		return 0, nil
	}, 0, accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrResetActionTriggers struct {
	Tenant   string
	Account  string
	GroupID  string
	UniqueID string
}

func (self *ApierV1) ResetActionTriggers(attr AttrResetActionTriggers, reply *string) error {

	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	var account *engine.Account
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if acc, err := self.AccountDb.GetAccount(accID); err != nil {
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
		return 0, nil
	}, 0, accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}
