/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"sort"
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
	BalanceId        string
	Direction        string
	ExtraParameters  string
	ExpirationString string
	Weight           float64
	Balance          *Balance
	SharedGroup      string
}

const (
	LOG            = "*log"
	RESET_TRIGGERS = "*reset_triggers"
	SET_POSTPAID   = "*set_postpaid"
	RESET_POSTPAID = "*reset_postpaid"
	SET_PREPAID    = "*set_prepaid"
	RESET_PREPAID  = "*reset_prepaid"
	TOPUP_RESET    = "*topup_reset"
	TOPUP          = "*topup"
	DEBIT          = "*debit"
	RESET_COUNTER  = "*reset_counter"
	RESET_COUNTERS = "*reset_counters"
	ENABLE_USER    = "*enable_user"
	DISABLE_USER   = "*disable_user"
	CALL_URL       = "*call_url"
	CALL_URL_ASYNC = "*call_url_async"
	MAIL_ASYNC     = "*mail_async"
	UNLIMITED      = "*unlimited"
)

type actionTypeFunc func(*UserBalance, *Action) error

func getActionFunc(typ string) (actionTypeFunc, bool) {
	switch typ {
	case LOG:
		return logAction, true
	case RESET_TRIGGERS:
		return resetTriggersAction, true
	case SET_POSTPAID:
		return setPostpaidAction, true
	case RESET_POSTPAID:
		return resetPostpaidAction, true
	case SET_PREPAID:
		return setPrepaidAction, true
	case RESET_PREPAID:
		return resetPrepaidAction, true
	case TOPUP_RESET:
		return topupResetAction, true
	case TOPUP:
		return topupAction, true
	case DEBIT:
		return debitAction, true
	case RESET_COUNTER:
		return resetCounterAction, true
	case RESET_COUNTERS:
		return resetCountersAction, true
	case ENABLE_USER:
		return enableUserAction, true
	case DISABLE_USER:
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

func logAction(ub *UserBalance, a *Action) (err error) {
	ubMarshal, _ := json.Marshal(ub)
	Logger.Info(fmt.Sprintf("Threshold reached, balance: %s", ubMarshal))
	return
}

func resetTriggersAction(ub *UserBalance, a *Action) (err error) {
	ub.resetActionTriggers(a)
	return
}

func setPostpaidAction(ub *UserBalance, a *Action) (err error) {
	ub.Type = UB_TYPE_POSTPAID
	return
}

func resetPostpaidAction(ub *UserBalance, a *Action) (err error) {
	genericReset(ub)
	return setPostpaidAction(ub, a)
}

func setPrepaidAction(ub *UserBalance, a *Action) (err error) {
	ub.Type = UB_TYPE_PREPAID
	return
}

func resetPrepaidAction(ub *UserBalance, a *Action) (err error) {
	genericReset(ub)
	return setPrepaidAction(ub, a)
}

func topupResetAction(ub *UserBalance, a *Action) (err error) {
	if ub.BalanceMap == nil { // Init the map since otherwise will get error if nil
		ub.BalanceMap = make(map[string]BalanceChain, 0)
	}
	ub.BalanceMap[a.BalanceId+a.Direction] = BalanceChain{}
	genericMakeNegative(a)
	genericDebit(ub, a)
	return
}

func topupAction(ub *UserBalance, a *Action) (err error) {
	genericMakeNegative(a)
	genericDebit(ub, a)
	return
}

func debitAction(ub *UserBalance, a *Action) (err error) {
	return genericDebit(ub, a)
}

func resetCounterAction(ub *UserBalance, a *Action) (err error) {
	uc := ub.getUnitCounter(a)
	if uc == nil {
		uc = &UnitsCounter{BalanceId: a.BalanceId, Direction: a.Direction}
		ub.UnitCounters = append(ub.UnitCounters, uc)
	}
	uc.initBalances(ub.ActionTriggers)
	return
}

func resetCountersAction(ub *UserBalance, a *Action) (err error) {
	ub.UnitCounters = make([]*UnitsCounter, 0)
	ub.initCounters()
	return
}

func genericMakeNegative(a *Action) {
	if a.Balance != nil && a.Balance.Value > 0 { // only apply if not allready negative
		a.Balance.Value = -a.Balance.Value
	}
}

func genericDebit(ub *UserBalance, a *Action) (err error) {
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]BalanceChain)
	}
	ub.debitBalanceAction(a)
	return
}

func enableUserAction(ub *UserBalance, a *Action) (err error) {
	ub.Disabled = false
	return
}

func disableUserAction(ub *UserBalance, a *Action) (err error) {
	ub.Disabled = true
	return
}

func genericReset(ub *UserBalance) {
	for k, _ := range ub.BalanceMap {
		ub.BalanceMap[k] = BalanceChain{&Balance{Value: 0}}
	}
	ub.UnitCounters = make([]*UnitsCounter, 0)
	ub.resetActionTriggers(nil)
}

func callUrl(ub *UserBalance, a *Action) error {
	body, err := json.Marshal(ub)
	if err != nil {
		return err
	}
	_, err = http.Post(a.ExtraParameters, "application/json", bytes.NewBuffer(body))
	return err
}

// Does not block for posts, no error reports
func callUrlAsync(ub *UserBalance, a *Action) error {
	body, err := json.Marshal(ub)
	if err != nil {
		return err
	}
	go func() {
		for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
			if _, err = http.Post(a.ExtraParameters, "application/json", bytes.NewBuffer(body)); err == nil {
				break // Success, no need to reinterate
			} else if i == 4 { // Last iteration, syslog the warning
				Logger.Warning(fmt.Sprintf("<Triggers> WARNING: Failed calling url: [%s], error: [%s], balance: %s", a.ExtraParameters, err.Error(), body))
				break
			}
			time.Sleep(time.Duration(i) * time.Minute)
		}

	}()
	return nil
}

// Mails the balance hitting the threshold towards predefined list of addresses
func mailAsync(ub *UserBalance, a *Action) error {
	ubJson, err := json.Marshal(ub)
	if err != nil {
		return err
	}
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
	message := []byte(fmt.Sprintf("To: %s\r\nSubject: [CGR Notification] Threshold hit on balance: %s\r\n\r\nTime: \r\n\t%s\r\n\r\nBalance:\r\n\t%s\r\n\r\nYours faithfully,\r\nCGR Balance Monitor\r\n", toAddrStr, ub.Id, time.Now(), ubJson))
	auth := smtp.PlainAuth("", cgrCfg.MailerAuthUser, cgrCfg.MailerAuthPass, strings.Split(cgrCfg.MailerServer, ":")[0]) // We only need host part, so ignore port
	go func() {
		for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
			if err := smtp.SendMail(cgrCfg.MailerServer, auth, cgrCfg.MailerFromAddr, toAddrs, message); err == nil {
				break
			} else if i == 4 {
				Logger.Warning(fmt.Sprintf("<Triggers> WARNING: Failed emailing, params: [%s], error: [%s], balance: %s", a.ExtraParameters, err.Error(), ubJson))
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

func (apl Actions) Less(i, j int) bool {
	return apl[i].Weight < apl[j].Weight
}

func (apl Actions) Sort() {
	sort.Sort(apl)
}
