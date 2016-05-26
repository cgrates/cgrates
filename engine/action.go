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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/mitchellh/mapstructure"
)

/*
Structure to be filled for each tariff plan with the bonus value for received calls minutes.
*/
type Action struct {
	Id               string
	ActionType       string
	ExtraParameters  string
	Filter           string
	ExpirationString string // must stay as string because it can have relative values like 1month
	Weight           float64
	Balance          *BalanceFilter
	balanceValue     float64 // balance value after action execution, used with cdrlog
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
	SET_BALANCE     = "*set_balance"
	REMOVE_BALANCE  = "*remove_balance"
	TOPUP_RESET     = "*topup_reset"
	TOPUP           = "*topup"
	DEBIT_RESET     = "*debit_reset"
	DEBIT           = "*debit"
	RESET_COUNTERS  = "*reset_counters"
	ENABLE_ACCOUNT  = "*enable_account"
	DISABLE_ACCOUNT = "*disable_account"
	//ENABLE_DISABLE_BALANCE    = "*enable_disable_balance"
	CALL_URL                  = "*call_url"
	CALL_URL_ASYNC            = "*call_url_async"
	MAIL_ASYNC                = "*mail_async"
	UNLIMITED                 = "*unlimited"
	CDRLOG                    = "*cdrlog"
	SET_DDESTINATIONS         = "*set_ddestinations"
	TRANSFER_MONETARY_DEFAULT = "*transfer_monetary_default"
	CGR_RPC                   = "*cgr_rpc"
)

func (a *Action) Clone() *Action {
	return &Action{
		Id:         a.Id,
		ActionType: a.ActionType,
		//BalanceType:      a.BalanceType,
		ExtraParameters:  a.ExtraParameters,
		ExpirationString: a.ExpirationString,
		Weight:           a.Weight,
		Balance:          a.Balance,
	}
}

type actionTypeFunc func(*Account, *StatsQueueTriggered, *Action, Actions) error

func getActionFunc(typ string) (actionTypeFunc, bool) {
	actionFuncMap := map[string]actionTypeFunc{
		LOG:             logAction,
		CDRLOG:          cdrLogAction,
		RESET_TRIGGERS:  resetTriggersAction,
		SET_RECURRENT:   setRecurrentAction,
		UNSET_RECURRENT: unsetRecurrentAction,
		ALLOW_NEGATIVE:  allowNegativeAction,
		DENY_NEGATIVE:   denyNegativeAction,
		RESET_ACCOUNT:   resetAccountAction,
		TOPUP_RESET:     topupResetAction,
		TOPUP:           topupAction,
		DEBIT_RESET:     debitResetAction,
		DEBIT:           debitAction,
		RESET_COUNTERS:  resetCountersAction,
		ENABLE_ACCOUNT:  enableUserAction,
		DISABLE_ACCOUNT: disableUserAction,
		//case ENABLE_DISABLE_BALANCE:
		//	return enableDisableBalanceAction, true
		CALL_URL:                  callUrl,
		CALL_URL_ASYNC:            callUrlAsync,
		MAIL_ASYNC:                mailAsync,
		SET_DDESTINATIONS:         setddestinations,
		REMOVE_ACCOUNT:            removeAccountAction,
		REMOVE_BALANCE:            removeBalanceAction,
		SET_BALANCE:               setBalanceAction,
		TRANSFER_MONETARY_DEFAULT: transferMonetaryDefaultAction,
		CGR_RPC:                   cgrRPCAction,
	}
	f, exists := actionFuncMap[typ]
	return f, exists
}

func logAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub != nil {
		body, _ := json.Marshal(ub)
		utils.Logger.Info(fmt.Sprintf("Threshold hit, Balance: %s", body))
	}
	if sq != nil {
		body, _ := json.Marshal(sq)
		utils.Logger.Info(fmt.Sprintf("Threshold hit, StatsQueue: %s", body))
	}
	return
}

// Used by cdrLogAction to dynamically parse values out of account and action
func parseTemplateValue(rsrFlds utils.RSRFields, acnt *Account, action *Action) string {
	var err error
	var dta *utils.TenantAccount
	if acnt != nil {
		dta, err = utils.NewTAFromAccountKey(acnt.ID) // Account information should be valid
	}
	if err != nil || acnt == nil {
		dta = new(utils.TenantAccount) // Init with empty values
	}
	var parsedValue string // Template values
	b := action.Balance.CreateBalance()
	for _, rsrFld := range rsrFlds {
		switch rsrFld.Id {
		case "AccountID":
			parsedValue += rsrFld.ParseValue(acnt.ID)
		case "Directions":
			parsedValue += rsrFld.ParseValue(b.Directions.String())
		case utils.TENANT:
			parsedValue += rsrFld.ParseValue(dta.Tenant)
		case utils.ACCOUNT:
			parsedValue += rsrFld.ParseValue(dta.Account)
		case "ActionID":
			parsedValue += rsrFld.ParseValue(action.Id)
		case "ActionType":
			parsedValue += rsrFld.ParseValue(action.ActionType)
		case "ActionValue":
			parsedValue += rsrFld.ParseValue(strconv.FormatFloat(b.GetValue(), 'f', -1, 64))
		case "BalanceType":
			parsedValue += rsrFld.ParseValue(action.Balance.GetType())
		case "BalanceUUID":
			parsedValue += rsrFld.ParseValue(b.Uuid)
		case "BalanceID":
			parsedValue += rsrFld.ParseValue(b.ID)
		case "BalanceValue":
			parsedValue += rsrFld.ParseValue(strconv.FormatFloat(action.balanceValue, 'f', -1, 64))
		case "DestinationIDs":
			parsedValue += rsrFld.ParseValue(b.DestinationIDs.String())
		case "ExtraParameters":
			parsedValue += rsrFld.ParseValue(action.ExtraParameters)
		case "RatingSubject":
			parsedValue += rsrFld.ParseValue(b.RatingSubject)
		case utils.CATEGORY:
			parsedValue += rsrFld.ParseValue(action.Balance.Categories.String())
		case "SharedGroups":
			parsedValue += rsrFld.ParseValue(action.Balance.SharedGroups.String())
		default:
			parsedValue += rsrFld.ParseValue("") // Mostly for static values
		}
	}
	return parsedValue
}

func cdrLogAction(acc *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	defaultTemplate := map[string]utils.RSRFields{
		utils.TOR:       utils.ParseRSRFieldsMustCompile("BalanceType", utils.INFIELD_SEP),
		utils.CDRHOST:   utils.ParseRSRFieldsMustCompile("^127.0.0.1", utils.INFIELD_SEP),
		utils.DIRECTION: utils.ParseRSRFieldsMustCompile("Directions", utils.INFIELD_SEP),
		utils.REQTYPE:   utils.ParseRSRFieldsMustCompile("^"+utils.META_PREPAID, utils.INFIELD_SEP),
		utils.TENANT:    utils.ParseRSRFieldsMustCompile(utils.TENANT, utils.INFIELD_SEP),
		utils.ACCOUNT:   utils.ParseRSRFieldsMustCompile(utils.ACCOUNT, utils.INFIELD_SEP),
		utils.SUBJECT:   utils.ParseRSRFieldsMustCompile(utils.ACCOUNT, utils.INFIELD_SEP),
		utils.COST:      utils.ParseRSRFieldsMustCompile("ActionValue", utils.INFIELD_SEP),
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
	var cdrs []*CDR
	for _, action := range acs {
		if !utils.IsSliceMember([]string{DEBIT, DEBIT_RESET, TOPUP, TOPUP_RESET}, action.ActionType) || action.Balance == nil {
			continue // Only log specific actions
		}
		cdr := &CDR{RunID: action.ActionType, Source: CDRLOG, SetupTime: time.Now(), AnswerTime: time.Now(), OriginID: utils.GenUUID(), ExtraFields: make(map[string]string)}
		cdr.CGRID = utils.Sha1(cdr.OriginID, cdr.SetupTime.String())
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
		cdrs = append(cdrs, cdr)
		if cdrStorage == nil { // Only save if the cdrStorage is defined
			continue
		}
		if err := cdrStorage.SetCDR(cdr, true); err != nil {
			return err
		}
	}
	b, _ := json.Marshal(cdrs)
	a.ExpirationString = string(b) // testing purpose only
	return
}

func resetTriggersAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.ResetActionTriggers(a)
	return
}

func setRecurrentAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.SetRecurrent(a, true)
	return
}

func unsetRecurrentAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.SetRecurrent(a, false)
	return
}

func allowNegativeAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.AllowNegative = true
	return
}

func denyNegativeAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.AllowNegative = false
	return
}

func resetAccountAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	return genericReset(ub)
}

func topupResetAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil { // Init the map since otherwise will get error if nil
		ub.BalanceMap = make(map[string]Balances, 0)
	}
	c := a.Clone()
	genericMakeNegative(c)
	err = genericDebit(ub, c, true)
	a.balanceValue = c.balanceValue
	return
}

func topupAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	c := a.Clone()
	genericMakeNegative(c)
	err = genericDebit(ub, c, false)
	a.balanceValue = c.balanceValue
	return
}

func debitResetAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil { // Init the map since otherwise will get error if nil
		ub.BalanceMap = make(map[string]Balances, 0)
	}
	return genericDebit(ub, a, true)
}

func debitAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	err = genericDebit(ub, a, false)
	return
}

func resetCountersAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.UnitCounters != nil {
		ub.UnitCounters.resetCounters(a)
	}
	return
}

func genericMakeNegative(a *Action) {
	if a.Balance != nil && a.Balance.GetValue() > 0 { // only apply if not allready negative
		a.Balance.SetValue(-a.Balance.GetValue())
	}
}

func genericDebit(ub *Account, a *Action, reset bool) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]Balances)
	}
	return ub.debitBalanceAction(a, reset)
}

func enableUserAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.Disabled = false
	return
}

func disableUserAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.Disabled = true
	return
}

/*func enableDisableBalanceAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.enableDisableBalanceAction(a)
	return
}*/

func genericReset(ub *Account) error {
	for k, _ := range ub.BalanceMap {
		ub.BalanceMap[k] = Balances{&Balance{Value: 0}}
	}
	ub.InitCounters()
	ub.ResetActionTriggers(nil)
	return nil
}

func callUrl(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	var o interface{}
	if ub != nil {
		o = ub
	}
	if sq != nil {
		o = sq
	}
	jsn, err := json.Marshal(o)
	if err != nil {
		return err
	}
	cfg := config.CgrConfig()
	fallbackPath := path.Join(cfg.HttpFailedDir, fmt.Sprintf("act_%s_%s_%s.json", a.ActionType, a.ExtraParameters, utils.GenUUID()))
	_, _, err = utils.HttpPoster(a.ExtraParameters, cfg.HttpSkipTlsVerify, jsn, utils.CONTENT_JSON, 1, fallbackPath, false)
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
	jsn, err := json.Marshal(o)
	if err != nil {
		return err
	}
	cfg := config.CgrConfig()
	fallbackPath := path.Join(cfg.HttpFailedDir, fmt.Sprintf("act_%s_%s_%s.json", a.ActionType, a.ExtraParameters, utils.GenUUID()))
	go utils.HttpPoster(a.ExtraParameters, cfg.HttpSkipTlsVerify, jsn, utils.CONTENT_JSON, 3, fallbackPath, false)
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
		message = []byte(fmt.Sprintf("To: %s\r\nSubject: [CGR Notification] Threshold hit on Balance: %s\r\n\r\nTime: \r\n\t%s\r\n\r\nBalance:\r\n\t%s\r\n\r\nYours faithfully,\r\nCGR Balance Monitor\r\n", toAddrStr, ub.ID, time.Now(), balJsn))
	} else if sq != nil {
		message = []byte(fmt.Sprintf("To: %s\r\nSubject: [CGR Notification] Threshold hit on StatsQueueId: %s\r\n\r\nTime: \r\n\t%s\r\n\r\nStatsQueueId:\r\n\t%s\r\n\r\nMetrics:\r\n\t%+v\r\n\r\nTrigger:\r\n\t%+v\r\n\r\nYours faithfully,\r\nCGR CDR Stats Monitor\r\n",
			toAddrStr, sq.Id, time.Now(), sq.Id, sq.Metrics, sq.Trigger))
	}
	auth := smtp.PlainAuth("", cgrCfg.MailerAuthUser, cgrCfg.MailerAuthPass, strings.Split(cgrCfg.MailerServer, ":")[0]) // We only need host part, so ignore port
	go func() {
		for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
			if err := smtp.SendMail(cgrCfg.MailerServer, auth, cgrCfg.MailerFromAddr, toAddrs, message); err == nil {
				break
			} else if i == 4 {
				if ub != nil {
					utils.Logger.Warning(fmt.Sprintf("<Triggers> WARNING: Failed emailing, params: [%s], error: [%s], BalanceId: %s", a.ExtraParameters, err.Error(), ub.ID))
				} else if sq != nil {
					utils.Logger.Warning(fmt.Sprintf("<Triggers> WARNING: Failed emailing, params: [%s], error: [%s], StatsQueueTriggeredId: %s", a.ExtraParameters, err.Error(), sq.Id))
				}
				break
			}
			time.Sleep(time.Duration(i) * time.Minute)
		}
	}()
	return nil
}

func setddestinations(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	var ddcDestId string
	for _, bchain := range ub.BalanceMap {
		for _, b := range bchain {
			for destId := range b.DestinationIDs {
				if strings.HasPrefix(destId, "*ddc") {
					ddcDestId = destId
					break
				}
			}
			if ddcDestId != "" {
				break
			}
		}
		if ddcDestId != "" {
			break
		}
	}
	if ddcDestId != "" {
		// make slice from prefixes
		prefixes := make([]string, len(sq.Metrics))
		i := 0
		for p := range sq.Metrics {
			prefixes[i] = p
			i++
		}
		// update destid in storage
		ratingStorage.SetDestination(&Destination{Id: ddcDestId, Prefixes: prefixes})
		// remove existing from cache
		CleanStalePrefixes([]string{ddcDestId})
		// update new values from redis
		ratingStorage.CacheRatingPrefixValues(map[string][]string{utils.DESTINATION_PREFIX: []string{utils.DESTINATION_PREFIX + ddcDestId}})
	} else {
		return utils.ErrNotFound
	}
	return nil
}

func removeAccountAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	var accID string
	if ub != nil {
		accID = ub.ID
	} else {
		accountInfo := struct {
			Tenant  string
			Account string
		}{}
		if a.ExtraParameters != "" {
			if err := json.Unmarshal([]byte(a.ExtraParameters), &accountInfo); err != nil {
				return err
			}
		}
		accID = utils.AccountKey(accountInfo.Tenant, accountInfo.Account)
	}
	if accID == "" {
		return utils.ErrInvalidKey
	}
	if err := accountingStorage.RemoveAccount(accID); err != nil {
		utils.Logger.Err(fmt.Sprintf("Could not remove account Id: %s: %v", accID, err))
		return err
	}
	_, err := Guardian.Guard(func() (interface{}, error) {
		// clean the account id from all action plans
		allAPs, err := ratingStorage.GetAllActionPlans()
		if err != nil && err != utils.ErrNotFound {
			utils.Logger.Err(fmt.Sprintf("Could not get action plans: %s: %v", accID, err))
			return 0, err
		}
		var dirtyAps []string
		for key, ap := range allAPs {
			if _, exists := ap.AccountIDs[accID]; !exists {
				// save action plan
				delete(ap.AccountIDs, key)
				ratingStorage.SetActionPlan(key, ap, true)
				dirtyAps = append(dirtyAps, utils.ACTION_PLAN_PREFIX+key)
			}
		}
		if len(dirtyAps) > 0 {
			// cache
			ratingStorage.CacheRatingPrefixValues(map[string][]string{
				utils.ACTION_PLAN_PREFIX: dirtyAps})
		}
		return 0, nil

	}, 0, utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return err
	}
	return nil
}

func removeBalanceAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	if _, exists := ub.BalanceMap[a.Balance.GetType()]; !exists {
		return utils.ErrNotFound
	}
	bChain := ub.BalanceMap[a.Balance.GetType()]
	found := false
	for i := 0; i < len(bChain); i++ {
		if bChain[i].MatchFilter(a.Balance, false) {
			// delete without preserving order
			bChain[i] = bChain[len(bChain)-1]
			bChain = bChain[:len(bChain)-1]
			i -= 1
			found = true
		}
	}
	ub.BalanceMap[a.Balance.GetType()] = bChain
	if !found {
		return utils.ErrNotFound
	}
	return nil
}

func setBalanceAction(acc *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	return acc.setBalanceAction(a)
}

func transferMonetaryDefaultAction(acc *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	if acc == nil {
		utils.Logger.Err("*transfer_monetary_default called without account")
		return utils.ErrAccountNotFound
	}
	if _, exists := acc.BalanceMap[utils.MONETARY]; !exists {
		return utils.ErrNotFound
	}
	defaultBalance := acc.GetDefaultMoneyBalance()
	bChain := acc.BalanceMap[utils.MONETARY]
	for _, balance := range bChain {
		if balance.Uuid != defaultBalance.Uuid &&
			balance.ID != defaultBalance.ID && // extra caution
			balance.MatchFilter(a.Balance, false) {
			if balance.Value > 0 {
				defaultBalance.Value += balance.Value
				balance.Value = 0
			}
		}
	}
	return nil
}

type RPCRequest struct {
	Address   string
	Transport string
	Method    string
	Attempts  int
	Async     bool
	Params    map[string]interface{}
}

func cgrRPCAction(account *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	// parse template
	tmpl := template.New("extra_params")
	tmpl.Delims("<<", ">>")
	t, err := tmpl.Parse(a.ExtraParameters)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("error parsing *cgr_rpc template: %s", err.Error()))
		return err
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, struct {
		Account *Account
		Sq      *StatsQueueTriggered
		Action  *Action
		Actions Actions
	}{account, sq, a, acs}); err != nil {
		utils.Logger.Err(fmt.Sprintf("error executing *cgr_rpc template %s:", err.Error()))
		return err
	}
	a.ExtraParameters = buf.String()
	//utils.Logger.Info("ExtraParameters: " + a.ExtraParameters)
	req := RPCRequest{}
	if err := json.Unmarshal([]byte(a.ExtraParameters), &req); err != nil {
		return err
	}
	params, err := utils.GetRpcParams(req.Method)
	if err != nil {
		return err
	}
	var client rpcclient.RpcClientConnection
	if req.Address != utils.MetaInternal {
		if client, err = rpcclient.NewRpcClient("tcp", req.Address, req.Attempts, 0, req.Transport, nil); err != nil {
			return err
		}
	} else {
		client = params.Object.(rpcclient.RpcClientConnection)
	}
	in, out := params.InParam, params.OutParam
	//utils.Logger.Info("Params: " + utils.ToJSON(req.Params))
	//p, err := utils.FromMapStringInterfaceValue(req.Params, in)
	mapstructure.Decode(req.Params, in)
	if err != nil {
		utils.Logger.Info("err3: " + err.Error())
		return err
	}
	utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> calling: %s with: %s", req.Method, utils.ToJSON(in)))
	if !req.Async {
		err = client.Call(req.Method, in, out)
		utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> result: %s err: %v", utils.ToJSON(out), err))
		return err
	}
	go func() {
		err := client.Call(req.Method, in, out)
		utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> result: %s err: %v", utils.ToJSON(out), err))
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
