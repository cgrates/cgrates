/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/smtp"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

/*
Structure to be filled for each tariff plan with the bonus value for received calls minutes.
*/
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

const (
	LOG             = "*log"
	RESET_TRIGGERS  = "*reset_triggers"
	SET_RECURRENT   = "*set_recurrent"
	UNSET_RECURRENT = "*unset_recurrent"
	ALLOW_NEGATIVE  = "*allow_negative"
	DENY_NEGATIVE   = "*deny_negative"
	RESET_ACCOUNT   = "*reset_account"
	REMOVE_ACCOUNT  = "*remove_account"
	TOPUP_RESET     = "*topup_reset"
	TOPUP           = "*topup"
	DEBIT_RESET     = "*debit_reset"
	DEBIT           = "*debit"
	RESET_COUNTER   = "*reset_counter"
	RESET_COUNTERS  = "*reset_counters"
	ENABLE_ACCOUNT  = "*enable_account"
	DISABLE_ACCOUNT = "*disable_account"
	CALL_URL        = "*call_url"
	CALL_URL_ASYNC  = "*call_url_async"
	MAIL_ASYNC      = "*mail_async"
	UNLIMITED       = "*unlimited"
	CDRLOG          = "*cdrlog"
)

func (a *Action) Clone() *Action {
	return &Action{
		Id:               a.Id,
		ActionType:       a.ActionType,
		BalanceType:      a.BalanceType,
		Direction:        a.Direction,
		ExtraParameters:  a.ExtraParameters,
		ExpirationString: a.ExpirationString,
		Weight:           a.Weight,
		Balance:          a.Balance.Clone(),
	}
}

type actionTypeFunc func(*Account, *StatsQueueTriggered, *Action, Actions) error

func getActionFunc(typ string) (actionTypeFunc, bool) {
	switch typ {
	case LOG:
		return logAction, true
	case CDRLOG:
		return cdrLogAction, true
	case RESET_TRIGGERS:
		return resetTriggersAction, true
	case SET_RECURRENT:
		return setRecurrentAction, true
	case UNSET_RECURRENT:
		return unsetRecurrentAction, true
	case ALLOW_NEGATIVE:
		return allowNegativeAction, true
	case DENY_NEGATIVE:
		return denyNegativeAction, true
	case RESET_ACCOUNT:
		return resetAccountAction, true
	case TOPUP_RESET:
		return topupResetAction, true
	case TOPUP:
		return topupAction, true
	case DEBIT_RESET:
		return debitResetAction, true
	case DEBIT:
		return debitAction, true
	case RESET_COUNTER:
		return resetCounterAction, true
	case RESET_COUNTERS:
		return resetCountersAction, true
	case ENABLE_ACCOUNT:
		return enableUserAction, true
	case DISABLE_ACCOUNT:
		return disableUserAction, true
	case CALL_URL:
		return callUrl, true
	case CALL_URL_ASYNC:
		return callUrlAsync, true
	case MAIL_ASYNC:
		return mailAsync, true
	}
	return nil, false
}

func logAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub != nil {
		body, _ := json.Marshal(ub)
		Logger.Info(fmt.Sprintf("Threshold hit, Balance: %s", body))
	}
	if sq != nil {
		body, _ := json.Marshal(sq)
		Logger.Info(fmt.Sprintf("Threshold hit, StatsQueue: %s", body))
	}
	return
}

// Used by cdrLogAction to dynamically parse values out of account and action
func parseTemplateValue(rsrFlds utils.RSRFields, acnt *Account, action *Action) string {
	dta, err := utils.NewDTAFromAccountKey(acnt.Id) // Account information should be valid
	if err != nil {
		dta = new(utils.DirectionTenantAccount) // Init with empty values
	}
	var parsedValue string // Template values
	for _, rsrFld := range rsrFlds {
		switch rsrFld.Id {
		case "account_id":
			parsedValue += rsrFld.ParseValue(acnt.Id)
		case "direction":
			parsedValue += rsrFld.ParseValue(dta.Direction)
		case "tenant":
			parsedValue += rsrFld.ParseValue(dta.Tenant)
		case "account":
			parsedValue += rsrFld.ParseValue(dta.Account)
		case "action_id":
			parsedValue += rsrFld.ParseValue(action.Id)
		case "action_type":
			parsedValue += rsrFld.ParseValue(action.ActionType)
		case "balance_type":
			parsedValue += rsrFld.ParseValue(action.BalanceType)
		case "balance_uuid":
			parsedValue += rsrFld.ParseValue(action.Balance.Uuid)
		case "balance_id":
			parsedValue += rsrFld.ParseValue(action.Balance.Id)
		case "balance_value":
			parsedValue += rsrFld.ParseValue(strconv.FormatFloat(action.Balance.GetValue(), 'f', -1, 64))
		case "destination_id":
			parsedValue += rsrFld.ParseValue(action.Balance.DestinationIds)
		case "extra_params":
			parsedValue += rsrFld.ParseValue(action.ExtraParameters)
		case "rating_subject":
			parsedValue += rsrFld.ParseValue(action.Balance.RatingSubject)
		case "category":
			parsedValue += rsrFld.ParseValue(action.Balance.Category)
		case "shared_group":
			parsedValue += rsrFld.ParseValue(action.Balance.SharedGroup)
		default:
			parsedValue += rsrFld.ParseValue("") // Mostly for static values
		}
	}
	return parsedValue
}

func cdrLogAction(acc *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	defaultTemplate := map[string]utils.RSRFields{
		"TOR":            utils.ParseRSRFieldsMustCompile("balance_type", utils.INFIELD_SEP),
		"CdrHost":        utils.ParseRSRFieldsMustCompile("^127.0.0.1", utils.INFIELD_SEP),
		"Direction":      utils.ParseRSRFieldsMustCompile("direction", utils.INFIELD_SEP),
		"ReqType":        utils.ParseRSRFieldsMustCompile("^"+utils.META_PREPAID, utils.INFIELD_SEP),
		"Tenant":         utils.ParseRSRFieldsMustCompile("tenant", utils.INFIELD_SEP),
		"Account":        utils.ParseRSRFieldsMustCompile("account", utils.INFIELD_SEP),
		"Subject":        utils.ParseRSRFieldsMustCompile("account", utils.INFIELD_SEP),
		"Cost":           utils.ParseRSRFieldsMustCompile("balance_value", utils.INFIELD_SEP),
		"MediationRunId": utils.ParseRSRFieldsMustCompile("^"+utils.META_DEFAULT, utils.INFIELD_SEP),
	}
	template := make(map[string]string)

	// overwrite default template
	if a.ExtraParameters != "" {
		if err = json.Unmarshal([]byte(a.ExtraParameters), &template); err != nil {
			return
		}
		for field, rsr := range template {
			defaultTemplate[field], err = utils.ParseRSRFields(rsr, utils.INFIELD_SEP)
			if err != nil {
				return err
			}
		}
	}

	// set stored cdr values
	var cdrs []*StoredCdr
	for _, action := range acs {
		if !utils.IsSliceMember([]string{DEBIT, DEBIT_RESET}, action.ActionType) || action.Balance == nil {
			continue // Only log specific actions
		}
		cdr := &StoredCdr{CdrSource: CDRLOG, SetupTime: time.Now(), AnswerTime: time.Now(), AccId: utils.GenUUID(), ExtraFields: make(map[string]string)}
		cdr.CgrId = utils.Sha1(cdr.AccId, cdr.SetupTime.String())
		cdr.Usage = time.Duration(1) * time.Second
		elem := reflect.ValueOf(cdr).Elem()
		for key, rsrFlds := range defaultTemplate {
			parsedValue := parseTemplateValue(rsrFlds, acc, action)
			field := elem.FieldByName(key)
			if field.IsValid() && field.CanSet() {
				switch field.Kind() {
				case reflect.Float64:
					value, err := strconv.ParseFloat(parsedValue, 64)
					if err != nil {
						continue
					}
					field.SetFloat(value)
				case reflect.String:
					field.SetString(parsedValue)
				}
			} else { // invalid fields go in extraFields of CDR
				cdr.ExtraFields[key] = parsedValue
			}
		}
		//Logger.Debug(fmt.Sprintf("account: %+v, action: %+v, balance: %+v", acc, action, action.Balance))
		cdrs = append(cdrs, cdr)
		if cdrStorage == nil { // Only save if the cdrStorage is defined
			continue
		}
		if err := cdrStorage.SetCdr(cdr); err != nil {
			return err
		}
		if err := cdrStorage.SetRatedCdr(cdr); err != nil {
			return err
		}
		// FixMe
		//if err := cdrStorage.LogCallCost(); err != nil {
		//	return err
		//}

	}

	b, _ := json.Marshal(cdrs)
	a.ExpirationString = string(b) // testing purpose only
	return
}

func resetTriggersAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	ub.ResetActionTriggers(a)
	return
}

func setRecurrentAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	ub.SetRecurrent(a, true)
	return
}

func unsetRecurrentAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	ub.SetRecurrent(a, false)
	return
}

func allowNegativeAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	ub.AllowNegative = true
	return
}

func denyNegativeAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	ub.AllowNegative = false
	return
}

func resetAccountAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	return genericReset(ub)
}

func topupResetAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	if ub.BalanceMap == nil { // Init the map since otherwise will get error if nil
		ub.BalanceMap = make(map[string]BalanceChain, 0)
	}
	c := a.Clone()
	genericMakeNegative(c)
	return genericDebit(ub, c, true)
}

func topupAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	c := a.Clone()
	genericMakeNegative(c)
	return genericDebit(ub, c, false)
}

func debitResetAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	if ub.BalanceMap == nil { // Init the map since otherwise will get error if nil
		ub.BalanceMap = make(map[string]BalanceChain, 0)
	}
	return genericDebit(ub, a, true)
}

func debitAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	err = genericDebit(ub, a, false)
	return
}

func resetCounterAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	uc := ub.getUnitCounter(a)
	if uc == nil {
		uc = &UnitsCounter{BalanceType: a.BalanceType, Direction: a.Direction}
		ub.UnitCounters = append(ub.UnitCounters, uc)
	}
	uc.initBalances(ub.ActionTriggers)
	return
}

func resetCountersAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	ub.UnitCounters = make([]*UnitsCounter, 0)
	ub.initCounters()
	return
}

func genericMakeNegative(a *Action) {
	if a.Balance != nil && a.Balance.GetValue() >= 0 { // only apply if not allready negative
		a.Balance.SetValue(-a.Balance.GetValue())
	}
}

func genericDebit(ub *Account, a *Action, reset bool) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]BalanceChain)
	}
	ub.debitBalanceAction(a, reset)
	return
}

func enableUserAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	ub.Disabled = false
	return
}

func disableUserAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil user balance")
	}
	ub.Disabled = true
	return
}

func genericReset(ub *Account) error {
	for k, _ := range ub.BalanceMap {
		ub.BalanceMap[k] = BalanceChain{&Balance{Value: 0}}
	}
	ub.UnitCounters = make([]*UnitsCounter, 0)
	ub.ResetActionTriggers(nil)
	return nil
}

func callUrl(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	var o interface{}
	if ub != nil {
		o = ub
	}
	if sq != nil {
		o = sq
	}
	//jsn, err := json.Marshal(o)
	//if err != nil {
	//	return err
	//}
	cfg := config.CgrConfig()
	_, err = utils.HttpJsonPost(a.ExtraParameters, cfg.HttpSkipTlsVerify, o)
	return err
}

// Does not block for posts, no error reports
func callUrlAsync(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	var o interface{}
	if ub != nil {
		o = ub
	}
	if sq != nil {
		o = sq
	}
	//jsn, err := json.Marshal(o)
	//if err != nil {
	//	return err
	//}
	cfg := config.CgrConfig()
	go func() {
		for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
			if _, err := utils.HttpJsonPost(a.ExtraParameters, cfg.HttpSkipTlsVerify, o); err == nil {
				break // Success, no need to reinterate
			} else if i == 4 { // Last iteration, syslog the warning
				Logger.Warning(fmt.Sprintf("<Triggers> WARNING: Failed calling url: [%s], error: [%s], triggered: %s", a.ExtraParameters, err.Error(), o))
				break
			}
			time.Sleep(time.Duration(i) * time.Minute)
		}

	}()
	return nil
}

// Mails the balance hitting the threshold towards predefined list of addresses
func mailAsync(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	cgrCfg := config.CgrConfig()
	params := strings.Split(a.ExtraParameters, string(utils.CSV_SEP))
	if len(params) == 0 {
		return errors.New("Unconfigured parameters for mail action")
	}
	toAddrs := strings.Split(params[0], string(utils.FALLBACK_SEP))
	toAddrStr := ""
	for idx, addr := range toAddrs {
		if idx != 0 {
			toAddrStr += ", "
		}
		toAddrStr += addr
	}
	var message []byte
	if ub != nil {
		balJsn, err := json.Marshal(ub)
		if err != nil {
			return err
		}
		message = []byte(fmt.Sprintf("To: %s\r\nSubject: [CGR Notification] Threshold hit on Balance: %s\r\n\r\nTime: \r\n\t%s\r\n\r\nBalance:\r\n\t%s\r\n\r\nYours faithfully,\r\nCGR Balance Monitor\r\n", toAddrStr, ub.Id, time.Now(), balJsn))
	} else if sq != nil {
		message = []byte(fmt.Sprintf("To: %s\r\nSubject: [CGR Notification] Threshold hit on StatsQueueId: %s\r\n\r\nTime: \r\n\t%s\r\n\r\nStatsQueueId:\r\n\t%s\r\n\r\nMetrics:\r\n\t%+v\r\n\r\nTrigger:\r\n\t%+v\r\n\r\nYours faithfully,\r\nCGR CDR Stats Monitor\r\n",
			toAddrStr, sq.Id, time.Now(), sq.Id, sq.metrics, sq.Trigger))
	}
	auth := smtp.PlainAuth("", cgrCfg.MailerAuthUser, cgrCfg.MailerAuthPass, strings.Split(cgrCfg.MailerServer, ":")[0]) // We only need host part, so ignore port
	go func() {
		for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
			if err := smtp.SendMail(cgrCfg.MailerServer, auth, cgrCfg.MailerFromAddr, toAddrs, message); err == nil {
				break
			} else if i == 4 {
				if ub != nil {
					Logger.Warning(fmt.Sprintf("<Triggers> WARNING: Failed emailing, params: [%s], error: [%s], BalanceId: %s", a.ExtraParameters, err.Error(), ub.Id))
				} else if sq != nil {
					Logger.Warning(fmt.Sprintf("<Triggers> WARNING: Failed emailing, params: [%s], error: [%s], StatsQueueTriggeredId: %s", a.ExtraParameters, err.Error(), sq.Id))
				}
				break
			}
			time.Sleep(time.Duration(i) * time.Minute)
		}
	}()
	return nil
}

// Structure to store actions according to weight
type Actions []*Action

func (apl Actions) Len() int {
	return len(apl)
}

func (apl Actions) Swap(i, j int) {
	apl[i], apl[j] = apl[j], apl[i]
}

// we need higher weights earlyer in the list
func (apl Actions) Less(j, i int) bool {
	return apl[i].Weight < apl[j].Weight
}

func (apl Actions) Sort() {
	sort.Sort(apl)
}
