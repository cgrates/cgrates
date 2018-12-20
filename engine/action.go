/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"html/template"
	"net"
	"net/smtp"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
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
	LOG                       = "*log"
	RESET_TRIGGERS            = "*reset_triggers"
	SET_RECURRENT             = "*set_recurrent"
	UNSET_RECURRENT           = "*unset_recurrent"
	ALLOW_NEGATIVE            = "*allow_negative"
	DENY_NEGATIVE             = "*deny_negative"
	RESET_ACCOUNT             = "*reset_account"
	REMOVE_ACCOUNT            = "*remove_account"
	SET_BALANCE               = "*set_balance"
	REMOVE_BALANCE            = "*remove_balance"
	TOPUP_RESET               = "*topup_reset"
	TOPUP                     = "*topup"
	DEBIT_RESET               = "*debit_reset"
	DEBIT                     = "*debit"
	RESET_COUNTERS            = "*reset_counters"
	ENABLE_ACCOUNT            = "*enable_account"
	DISABLE_ACCOUNT           = "*disable_account"
	CALL_URL                  = "*call_url"
	CALL_URL_ASYNC            = "*call_url_async"
	MAIL_ASYNC                = "*mail_async"
	UNLIMITED                 = "*unlimited"
	CDRLOG                    = "*cdrlog"
	SET_DDESTINATIONS         = "*set_ddestinations"
	TRANSFER_MONETARY_DEFAULT = "*transfer_monetary_default"
	CGR_RPC                   = "*cgr_rpc"
	TopUpZeroNegative         = "*topup_zero_negative"
	SetExpiry                 = "*set_expiry"
	MetaPublishAccount        = "*publish_account"
	MetaPublishBalance        = "*publish_balance"
)

func (a *Action) Clone() *Action {
	var clonedAction Action
	utils.Clone(a, &clonedAction)
	return &clonedAction
}

type actionTypeFunc func(*Account, *Action, Actions, interface{}) error

func getActionFunc(typ string) (actionTypeFunc, bool) {
	actionFuncMap := map[string]actionTypeFunc{
		LOG:                       logAction,
		CDRLOG:                    cdrLogAction,
		RESET_TRIGGERS:            resetTriggersAction,
		SET_RECURRENT:             setRecurrentAction,
		UNSET_RECURRENT:           unsetRecurrentAction,
		ALLOW_NEGATIVE:            allowNegativeAction,
		DENY_NEGATIVE:             denyNegativeAction,
		RESET_ACCOUNT:             resetAccountAction,
		TOPUP_RESET:               topupResetAction,
		TOPUP:                     topupAction,
		DEBIT_RESET:               debitResetAction,
		DEBIT:                     debitAction,
		RESET_COUNTERS:            resetCountersAction,
		ENABLE_ACCOUNT:            enableAccountAction,
		DISABLE_ACCOUNT:           disableAccountAction,
		CALL_URL:                  callUrl,
		CALL_URL_ASYNC:            callUrlAsync,
		MAIL_ASYNC:                mailAsync,
		SET_DDESTINATIONS:         setddestinations,
		REMOVE_ACCOUNT:            removeAccountAction,
		REMOVE_BALANCE:            removeBalanceAction,
		SET_BALANCE:               setBalanceAction,
		TRANSFER_MONETARY_DEFAULT: transferMonetaryDefaultAction,
		CGR_RPC:                   cgrRPCAction,
		TopUpZeroNegative:         topupZeroNegativeAction,
		SetExpiry:                 setExpiryAction,
		MetaPublishAccount:        publishAccount,
		MetaPublishBalance:        publishBalance,
	}
	f, exists := actionFuncMap[typ]
	return f, exists
}

func logAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	switch {
	case ub != nil:
		body, _ := json.Marshal(ub)
		utils.Logger.Info(fmt.Sprintf("Threshold hit, Balance: %s", body))
	case extraData != nil:
		body, _ := json.Marshal(extraData)
		utils.Logger.Info(fmt.Sprintf("<CGRLog> extraData: %s", body))
	}
	return
}

func cdrLogAction(acc *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if schedCdrsConns == nil {
		return fmt.Errorf("No connection with CDR Server")
	}
	defaultTemplate := map[string]config.RSRParsers{
		utils.ToR:         config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+"BalanceType", true, utils.INFIELD_SEP),
		utils.OriginHost:  config.NewRSRParsersMustCompile("127.0.0.1", true, utils.INFIELD_SEP),
		utils.RequestType: config.NewRSRParsersMustCompile(utils.META_PREPAID, true, utils.INFIELD_SEP),
		utils.Tenant:      config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.Tenant, true, utils.INFIELD_SEP),
		utils.Account:     config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.Account, true, utils.INFIELD_SEP),
		utils.Subject:     config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.Account, true, utils.INFIELD_SEP),
		utils.COST:        config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+"ActionValue", true, utils.INFIELD_SEP),
	}
	template := make(map[string]string)
	// overwrite default template
	if a.ExtraParameters != "" {
		if err = json.Unmarshal([]byte(a.ExtraParameters), &template); err != nil {
			return
		}
		for field, rsr := range template {
			defaultTemplate[field] = config.NewRSRParsersMustCompile(rsr,
				true, config.CgrConfig().GeneralCfg().RsrSepatarot)
		}
	}
	// set stored cdr values
	var cdrs []*CDR
	for _, action := range acs {
		if !utils.IsSliceMember([]string{DEBIT, DEBIT_RESET, TOPUP, TOPUP_RESET}, action.ActionType) ||
			action.Balance == nil {
			continue // Only log specific actions
		}
		cdr := &CDR{
			RunID:     action.ActionType,
			Source:    CDRLOG,
			SetupTime: time.Now(), AnswerTime: time.Now(),
			OriginID:    utils.GenUUID(),
			ExtraFields: make(map[string]string),
			PreRated:    true,
		}
		cdr.CGRID = utils.Sha1(cdr.OriginID, cdr.SetupTime.String())
		cdr.Usage = time.Duration(1)
		elem := reflect.ValueOf(cdr).Elem()
		for key, rsrFlds := range defaultTemplate {
			parsedValue, err := rsrFlds.ParseDataProvider(newCdrLogProvider(acc, action), utils.NestingSep)
			field := elem.FieldByName(key)
			if err != nil {
				return err
			}
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
		var rply string
		// After compute the CDR send it to CDR Server to be processed
		if err := schedCdrsConns.Call(utils.CdrsV2ProcessCDR, cdr.AsCGREvent(), &rply); err != nil {
			return err
		}
	}
	b, _ := json.Marshal(cdrs)
	a.ExpirationString = string(b) // testing purpose only
	return
}

func resetTriggersAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.ResetActionTriggers(a)
	return
}

func setRecurrentAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.SetRecurrent(a, true)
	return
}

func unsetRecurrentAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.SetRecurrent(a, false)
	return
}

func allowNegativeAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.AllowNegative = true
	return
}

func denyNegativeAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.AllowNegative = false
	return
}

func resetAccountAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	return genericReset(ub)
}

func topupResetAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
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

func topupAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	c := a.Clone()
	genericMakeNegative(c)
	err = genericDebit(ub, c, false)
	a.balanceValue = c.balanceValue
	return
}

func debitResetAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil { // Init the map since otherwise will get error if nil
		ub.BalanceMap = make(map[string]Balances, 0)
	}
	return genericDebit(ub, a, true)
}

func debitAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	err = genericDebit(ub, a, false)
	return
}

func resetCountersAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
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
	return ub.debitBalanceAction(a, reset, false)
}

func enableAccountAction(acc *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if acc == nil {
		return errors.New("nil account")
	}
	acc.Disabled = false
	return
}

func disableAccountAction(acc *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if acc == nil {
		return errors.New("nil account")
	}
	acc.Disabled = true
	return
}

/*func enableDisableBalanceAction(ub *Account, sq *CDRStatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.enableDisableBalanceAction(a)
	return
}*/

func genericReset(ub *Account) error {
	for k := range ub.BalanceMap {
		ub.BalanceMap[k] = Balances{&Balance{Value: 0}}
	}
	ub.InitCounters()
	ub.ResetActionTriggers(nil)
	return nil
}

func callUrl(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	var o interface{}
	switch {
	case ub != nil:
		o = ub
	case extraData != nil:
		o = extraData
	}
	jsn, err := json.Marshal(o)
	if err != nil {
		return err
	}
	cfg := config.CgrConfig()
	ffn := &utils.FallbackFileName{
		Module:     fmt.Sprintf("%s>%s", utils.ActionsPoster, a.ActionType),
		Transport:  utils.MetaHTTPjson,
		Address:    a.ExtraParameters,
		RequestID:  utils.GenUUID(),
		FileSuffix: utils.JSNSuffix,
	}
	_, err = NewHTTPPoster(config.CgrConfig().GeneralCfg().HttpSkipTlsVerify,
		config.CgrConfig().GeneralCfg().ReplyTimeout).Post(a.ExtraParameters,
		utils.CONTENT_JSON, jsn, config.CgrConfig().GeneralCfg().PosterAttempts,
		path.Join(cfg.GeneralCfg().FailedPostsDir, ffn.AsString()))
	return err
}

// Does not block for posts, no error reports
func callUrlAsync(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	var o interface{}
	switch {
	case ub != nil:
		o = ub
	case extraData != nil:
		o = extraData
	}
	jsn, err := json.Marshal(o)
	if err != nil {
		return err
	}
	cfg := config.CgrConfig()
	ffn := &utils.FallbackFileName{
		Module:     fmt.Sprintf("%s>%s", utils.ActionsPoster, a.ActionType),
		Transport:  utils.MetaHTTPjson,
		Address:    a.ExtraParameters,
		RequestID:  utils.GenUUID(),
		FileSuffix: utils.JSNSuffix,
	}
	go NewHTTPPoster(config.CgrConfig().GeneralCfg().HttpSkipTlsVerify,
		config.CgrConfig().GeneralCfg().ReplyTimeout).Post(a.ExtraParameters,
		utils.CONTENT_JSON, jsn, config.CgrConfig().GeneralCfg().PosterAttempts,
		path.Join(cfg.GeneralCfg().FailedPostsDir, ffn.AsString()))
	return nil
}

// Mails the balance hitting the threshold towards predefined list of addresses
func mailAsync(ub *Account, a *Action, acs Actions, extraData interface{}) error {
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
	}
	auth := smtp.PlainAuth("", cgrCfg.MailerCfg().MailerAuthUser, cgrCfg.MailerCfg().MailerAuthPass, strings.Split(cgrCfg.MailerCfg().MailerServer, ":")[0]) // We only need host part, so ignore port
	go func() {
		for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
			if err := smtp.SendMail(cgrCfg.MailerCfg().MailerServer, auth, cgrCfg.MailerCfg().MailerFromAddr, toAddrs, message); err == nil {
				break
			} else if i == 4 {
				if ub != nil {
					utils.Logger.Warning(fmt.Sprintf("<Triggers> WARNING: Failed emailing, params: [%s], error: [%s], BalanceId: %s", a.ExtraParameters, err.Error(), ub.ID))
				}
				break
			}
			time.Sleep(time.Duration(i) * time.Minute)
		}
	}()
	return nil
}

func setddestinations(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
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
		// Review here prefixes
		// prefixes := make([]string, len(sq.Metrics))
		// i := 0
		// for p := range sq.Metrics {
		// 	prefixes[i] = p
		// 	i++
		// }
		newDest := &Destination{Id: ddcDestId}
		oldDest, err := dm.DataDB().GetDestination(ddcDestId, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		// update destid in storage
		if err = dm.DataDB().SetDestination(newDest, utils.NonTransactional); err != nil {
			return err
		}
		if err = dm.CacheDataFromDB(utils.DESTINATION_PREFIX, []string{ddcDestId}, true); err != nil {
			return err
		}

		if err == nil && oldDest != nil {
			if err = dm.DataDB().UpdateReverseDestination(oldDest, newDest, utils.NonTransactional); err != nil {
				return err
			}
		}
	} else {
		return utils.ErrNotFound
	}
	return nil
}

func removeAccountAction(ub *Account, a *Action, acs Actions, extraData interface{}) error {
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

	if err := dm.DataDB().RemoveAccount(accID); err != nil {
		utils.Logger.Err(fmt.Sprintf("Could not remove account Id: %s: %v", accID, err))
		return err
	}

	_, err := guardian.Guardian.Guard(func() (interface{}, error) {
		acntAPids, err := dm.DataDB().GetAccountActionPlans(accID, false, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			utils.Logger.Err(fmt.Sprintf("Could not get action plans: %s: %v", accID, err))
			return 0, err
		}
		for _, apID := range acntAPids {
			ap, err := dm.DataDB().GetActionPlan(apID, false, utils.NonTransactional)
			if err != nil {
				utils.Logger.Err(fmt.Sprintf("Could not retrieve action plan: %s: %v", apID, err))
				return 0, err
			}
			delete(ap.AccountIDs, accID)
			if err := dm.DataDB().SetActionPlan(apID, ap, true, utils.NonTransactional); err != nil {
				utils.Logger.Err(fmt.Sprintf("Could not save action plan: %s: %v", apID, err))
				return 0, err
			}
		}
		if err = dm.CacheDataFromDB(utils.ACTION_PLAN_PREFIX, acntAPids, true); err != nil {
			return 0, err
		}
		if err = dm.DataDB().RemAccountActionPlans(accID, nil); err != nil {
			return 0, err
		}
		if err = dm.CacheDataFromDB(utils.AccountActionPlansPrefix, []string{accID}, true); err != nil && err.Error() != utils.ErrNotFound.Error() {
			return 0, err
		}
		return 0, nil

	}, 0, utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return err
	}
	return nil
}

func removeBalanceAction(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	if ub == nil {
		return fmt.Errorf("nil account for %s action", utils.ToJSON(a))
	}
	if _, exists := ub.BalanceMap[a.Balance.GetType()]; !exists {
		return utils.ErrNotFound
	}
	bChain := ub.BalanceMap[a.Balance.GetType()]
	found := false
	for i := 0; i < len(bChain); i++ {
		if bChain[i].MatchFilter(a.Balance, false, false) {
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

func setBalanceAction(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	if ub == nil {
		return fmt.Errorf("nil account for %s action", utils.ToJSON(a))
	}
	return ub.setBalanceAction(a)
}

func transferMonetaryDefaultAction(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	if ub == nil {
		utils.Logger.Err("*transfer_monetary_default called without account")
		return utils.ErrAccountNotFound
	}
	if _, exists := ub.BalanceMap[utils.MONETARY]; !exists {
		return utils.ErrNotFound
	}
	defaultBalance := ub.GetDefaultMoneyBalance()
	bChain := ub.BalanceMap[utils.MONETARY]
	for _, balance := range bChain {
		if balance.Uuid != defaultBalance.Uuid &&
			balance.ID != defaultBalance.ID && // extra caution
			balance.MatchFilter(a.Balance, false, false) {
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

/*
<< .Object.Property >>

Property can be a attribute or a method both used without ()
Please also note the initial dot .

Currently there are following objects that can be used:

Account -  the account that this action is called on
Action - the action with all it's attributs
Actions - the list of actions in the current action set
Sq - CDRStatsQueueTriggered object

We can actually use everythiong that go templates offer. You can read more here: https://golang.org/pkg/text/template/
*/
func cgrRPCAction(ub *Account, a *Action, acs Actions, extraData interface{}) error {
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
		Account   *Account
		Action    *Action
		Actions   Actions
		ExtraData interface{}
	}{ub, a, acs, extraData}); err != nil {
		utils.Logger.Err(fmt.Sprintf("error executing *cgr_rpc template %s:", err.Error()))
		return err
	}
	processedExtraParam := buf.String()
	//utils.Logger.Info("ExtraParameters: " + parsedExtraParameters)
	req := RPCRequest{}
	if err := json.Unmarshal([]byte(processedExtraParam), &req); err != nil {
		return err
	}
	params, err := utils.GetRpcParams(req.Method)
	if err != nil {
		return err
	}
	var client rpcclient.RpcClientConnection
	if req.Address != utils.MetaInternal {
		if client, err = rpcclient.NewRpcClient("tcp", req.Address, false, "", "", "",
			req.Attempts, 0, config.CgrConfig().GeneralCfg().ConnectTimeout,
			config.CgrConfig().GeneralCfg().ReplyTimeout, req.Transport,
			nil, false); err != nil {
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
		utils.Logger.Info("<*cgr_rpc> err: " + err.Error())
		return err
	}
	if in == nil {
		utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> nil params err: req.Params: %+v params: %+v", req.Params, params))
		return utils.ErrParserError
	}
	utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> calling: %s with: %s and result %v", req.Method, utils.ToJSON(in), out))
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

func topupZeroNegativeAction(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]Balances)
	}
	return ub.debitBalanceAction(a, false, true)
}

func setExpiryAction(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	if ub == nil {
		return errors.New("nil account")
	}
	balanceType := a.Balance.GetType()
	for _, b := range ub.BalanceMap[balanceType] {
		if b.MatchFilter(a.Balance, false, true) {
			b.ExpirationDate = a.Balance.GetExpirationDate()
		}
	}
	return nil
}

// publishAccount will publish the account as well as each balance received to ThresholdS
func publishAccount(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.Publish()
	for bType := range ub.BalanceMap {
		for _, b := range ub.BalanceMap[bType] {
			if b.account == nil {
				b.account = ub
			}
			b.Publish()
		}
	}
	return nil
}

// publishAccount will publish the account as well as each balance received to ThresholdS
func publishBalance(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	if ub == nil {
		return errors.New("nil account")
	}
	for bType := range ub.BalanceMap {
		for _, b := range ub.BalanceMap[bType] {
			if b.account == nil {
				b.account = ub
			}
			b.Publish()
		}
	}
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

func (apl Actions) Clone() (interface{}, error) {
	var cln Actions
	if err := utils.Clone(apl, &cln); err != nil {
		return nil, err
	}
	for i, act := range apl { // Fix issues with gob cloning nil pointer towards false value
		if act.Balance != nil {
			if act.Balance.Disabled != nil && !*act.Balance.Disabled {
				cln[i].Balance.Disabled = utils.BoolPointer(*act.Balance.Disabled)
			}
			if act.Balance.Blocker != nil && !*act.Balance.Blocker {
				cln[i].Balance.Blocker = utils.BoolPointer(*act.Balance.Blocker)
			}
		}
	}
	return cln, nil
}

// newCdrLogProvider constructs a DataProvider
func newCdrLogProvider(acnt *Account, action *Action) (dP config.DataProvider) {
	dP = &cdrLogProvider{acnt: acnt, action: action, cache: config.NewNavigableMap(nil)}
	return
}

// cdrLogProvider implements engine.DataProvider so we can pass it to filters
type cdrLogProvider struct {
	acnt   *Account
	action *Action
	cache  *config.NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (cdrP *cdrLogProvider) String() string {
	return utils.ToJSON(cdrP)
}

// FieldAsInterface is part of engine.DataProvider interface
func (cdrP *cdrLogProvider) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	if data, err = cdrP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	var dta *utils.TenantAccount
	if cdrP.acnt != nil {
		dta, err = utils.NewTAFromAccountKey(cdrP.acnt.ID) // Account information should be valid
	}
	if err != nil || cdrP.acnt == nil {
		dta = new(utils.TenantAccount) // Init with empty values
	}
	b := cdrP.action.Balance.CreateBalance()
	switch fldPath[0] {
	case "AccountID":
		data = cdrP.acnt.ID
	case "Directions":
		data = b.Directions.String()
	case utils.Tenant:
		data = dta.Tenant
	case utils.Account:
		data = dta.Account
	case "ActionID":
		data = cdrP.action.Id
	case "ActionType":
		data = cdrP.action.ActionType
	case "ActionValue":
		data = strconv.FormatFloat(b.GetValue(), 'f', -1, 64)
	case "BalanceType":
		data = cdrP.action.Balance.GetType()
	case "BalanceUUID":
		data = b.Uuid
	case "BalanceID":
		data = b.ID
	case "BalanceValue":
		data = strconv.FormatFloat(cdrP.action.balanceValue, 'f', -1, 64)
	case "DestinationIDs":
		data = b.DestinationIDs.String()
	case "ExtraParameters":
		data = cdrP.action.ExtraParameters
	case "RatingSubject":
		data = b.RatingSubject
	case utils.Category:
		data = cdrP.action.Balance.Categories.String()
	case "SharedGroups":
		data = cdrP.action.Balance.SharedGroups.String()
	default:
		data = fldPath[0]
	}
	cdrP.cache.Set(fldPath, data, false, false)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (cdrP *cdrLogProvider) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = cdrP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	data, err = utils.IfaceAsString(valIface)
	return
}

// AsNavigableMap is part of engine.DataProvider interface
func (cdrP *cdrLogProvider) AsNavigableMap([]*config.FCTemplate) (
	nm *config.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// RemoteHost is part of engine.DataProvider interface
func (cdrP *cdrLogProvider) RemoteHost() net.Addr {
	return utils.LocalAddr()
}
