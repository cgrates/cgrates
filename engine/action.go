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

// Action will be filled for each tariff plan with the bonus value for received calls minutes.
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

// Clone returns a clone of the action
func (a *Action) Clone() (cln *Action) {
	if a == nil {
		return
	}
	return &Action{
		Id:               a.Id,
		ActionType:       a.ActionType,
		ExtraParameters:  a.ExtraParameters,
		Filter:           a.Filter,
		ExpirationString: a.ExpirationString,
		Weight:           a.Weight,
		Balance:          a.Balance.Clone(),
	}
}

type actionTypeFunc func(*Account, *Action, Actions, interface{}) error

func getActionFunc(typ string) (actionTypeFunc, bool) {
	actionFuncMap := map[string]actionTypeFunc{
		utils.LOG:                       logAction,
		utils.RESET_TRIGGERS:            resetTriggersAction,
		utils.CDRLOG:                    cdrLogAction,
		utils.SET_RECURRENT:             setRecurrentAction,
		utils.UNSET_RECURRENT:           unsetRecurrentAction,
		utils.ALLOW_NEGATIVE:            allowNegativeAction,
		utils.DENY_NEGATIVE:             denyNegativeAction,
		utils.RESET_ACCOUNT:             resetAccountAction,
		utils.TOPUP_RESET:               topupResetAction,
		utils.TOPUP:                     topupAction,
		utils.DEBIT_RESET:               debitResetAction,
		utils.DEBIT:                     debitAction,
		utils.RESET_COUNTERS:            resetCountersAction,
		utils.ENABLE_ACCOUNT:            enableAccountAction,
		utils.DISABLE_ACCOUNT:           disableAccountAction,
		utils.HttpPost:                  callURL,
		utils.HttpPostAsync:             callURLAsync,
		utils.MAIL_ASYNC:                mailAsync,
		utils.SET_DDESTINATIONS:         setddestinations,
		utils.REMOVE_ACCOUNT:            removeAccountAction,
		utils.REMOVE_BALANCE:            removeBalanceAction,
		utils.SET_BALANCE:               setBalanceAction,
		utils.TRANSFER_MONETARY_DEFAULT: transferMonetaryDefaultAction,
		utils.CGR_RPC:                   cgrRPCAction,
		utils.TopUpZeroNegative:         topupZeroNegativeAction,
		utils.SetExpiry:                 setExpiryAction,
		utils.MetaPublishAccount:        publishAccount,
		utils.MetaPublishBalance:        publishBalance,
		utils.MetaAMQPjsonMap:           sendAMQP,
		utils.MetaAMQPV1jsonMap:         sendAWS,
		utils.MetaSQSjsonMap:            sendSQS,
		utils.MetaKafkajsonMap:          sendKafka,
		utils.MetaS3jsonMap:             sendS3,
		utils.MetaRemoveSessionCosts:    removeSessionCosts,
		utils.MetaRemoveExpired:         removeExpired,
		utils.MetaPostEvent:             postEvent,
		utils.MetaCDRAccount:            resetAccount,
	}
	f, exists := actionFuncMap[typ]
	return f, exists
}

func logAction(ub *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	switch {
	case ub != nil:
		body, _ := json.Marshal(ub)
		utils.Logger.Info(fmt.Sprintf("LOG Account: %s", body))
	case extraData != nil:
		body, _ := json.Marshal(extraData)
		utils.Logger.Info(fmt.Sprintf("LOG ExtraData: %s", body))
	}
	return
}

func cdrLogAction(acc *Account, a *Action, acs Actions, extraData interface{}) (err error) {
	if len(config.CgrConfig().SchedulerCfg().CDRsConns) == 0 {
		return fmt.Errorf("No connection with CDR Server")
	}
	defaultTemplate := map[string]config.RSRParsers{
		utils.ToR:         config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+"BalanceType", true, utils.INFIELD_SEP),
		utils.OriginHost:  config.NewRSRParsersMustCompile("127.0.0.1", true, utils.INFIELD_SEP),
		utils.RequestType: config.NewRSRParsersMustCompile(utils.META_NONE, true, utils.INFIELD_SEP),
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
			if defaultTemplate[field], err = config.NewRSRParsers(rsr,
				true, config.CgrConfig().GeneralCfg().RSRSep); err != nil {
				return
			}
		}
	}
	//In case that we have extra data we populate default templates
	mapExtraData, _ := extraData.(map[string]interface{})
	for key, val := range mapExtraData {
		if defaultTemplate[key], err = config.NewRSRParsers(utils.IfaceAsString(val),
			true, config.CgrConfig().GeneralCfg().RSRSep); err != nil {
			return
		}
	}

	// set stored cdr values
	var cdrs []*CDR
	for _, action := range acs {
		if !utils.SliceHasMember([]string{utils.DEBIT, utils.DEBIT_RESET, utils.SET_BALANCE, utils.TOPUP, utils.TOPUP_RESET}, action.ActionType) ||
			action.Balance == nil {
			continue // Only log specific actions
		}
		cdrLogProvider := newCdrLogProvider(acc, action)
		cdr := &CDR{
			RunID:     action.ActionType,
			Source:    utils.CDRLOG,
			SetupTime: time.Now(), AnswerTime: time.Now(),
			OriginID:    utils.GenUUID(),
			ExtraFields: make(map[string]string),
			PreRated:    true,
			Usage:       time.Duration(1),
		}
		cdr.CGRID = utils.Sha1(cdr.OriginID, cdr.OriginHost)
		elem := reflect.ValueOf(cdr).Elem()
		for key, rsrFlds := range defaultTemplate {
			parsedValue, err := rsrFlds.ParseDataProvider(cdrLogProvider, utils.NestingSep)
			if err != nil {
				return err
			}
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
				case reflect.Int64:
					value, err := strconv.ParseInt(parsedValue, 10, 64)
					if err != nil {
						continue
					}
					field.SetInt(value)
				}
			} else { // invalid fields go in extraFields of CDR
				cdr.ExtraFields[key] = parsedValue
			}
		}
		cdrs = append(cdrs, cdr)
		var rply string
		// After compute the CDR send it to CDR Server to be processed
		if err := connMgr.Call(config.CgrConfig().SchedulerCfg().CDRsConns, nil,
			utils.CDRsV1ProcessEvent,
			&ArgV1ProcessEvent{
				Flags:    []string{utils.ConcatenatedKey(utils.MetaChargers, "false")}, // do not try to get the chargers for cdrlog
				CGREvent: *cdr.AsCGREvent()}, &rply); err != nil {
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

func getOneData(ub *Account, extraData interface{}) ([]byte, error) {
	switch {
	case ub != nil:
		return json.Marshal(ub)
	case extraData != nil:
		return json.Marshal(extraData)
	}
	return nil, nil
}

func sendAMQP(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	err = PostersCache.PostAMQP(a.ExtraParameters, config.CgrConfig().GeneralCfg().PosterAttempts, body)
	if err != nil && config.CgrConfig().GeneralCfg().FailedPostsDir != utils.META_NONE {
		addFailedPost(a.ExtraParameters, utils.MetaAMQPjsonMap, utils.ActionsPoster+utils.HIERARCHY_SEP+a.ActionType, body)
		err = nil
	}
	return err
}

func sendAWS(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	err = PostersCache.PostAMQPv1(a.ExtraParameters, config.CgrConfig().GeneralCfg().PosterAttempts, body)
	if err != nil && config.CgrConfig().GeneralCfg().FailedPostsDir != utils.META_NONE {
		addFailedPost(a.ExtraParameters, utils.MetaAMQPV1jsonMap, utils.ActionsPoster+utils.HIERARCHY_SEP+a.ActionType, body)
		err = nil
	}
	return err
}

func sendSQS(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	err = PostersCache.PostSQS(a.ExtraParameters, config.CgrConfig().GeneralCfg().PosterAttempts, body)
	if err != nil && config.CgrConfig().GeneralCfg().FailedPostsDir != utils.META_NONE {
		addFailedPost(a.ExtraParameters, utils.MetaSQSjsonMap, utils.ActionsPoster+utils.HIERARCHY_SEP+a.ActionType, body)
		err = nil
	}
	return err
}

func sendKafka(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	err = PostersCache.PostKafka(a.ExtraParameters, config.CgrConfig().GeneralCfg().PosterAttempts, body, utils.UUIDSha1Prefix())
	if err != nil && config.CgrConfig().GeneralCfg().FailedPostsDir != utils.META_NONE {
		addFailedPost(a.ExtraParameters, utils.MetaKafkajsonMap, utils.ActionsPoster+utils.HIERARCHY_SEP+a.ActionType, body)
		err = nil
	}
	return err
}

func sendS3(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	err = PostersCache.PostS3(a.ExtraParameters, config.CgrConfig().GeneralCfg().PosterAttempts, body, utils.UUIDSha1Prefix())
	if err != nil && config.CgrConfig().GeneralCfg().FailedPostsDir != utils.META_NONE {
		addFailedPost(a.ExtraParameters, utils.MetaS3jsonMap, utils.ActionsPoster+utils.HIERARCHY_SEP+a.ActionType, body)
		err = nil
	}
	return err
}

func callURL(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	pstr, err := NewHTTPPoster(config.CgrConfig().GeneralCfg().HttpSkipTlsVerify,
		config.CgrConfig().GeneralCfg().ReplyTimeout, a.ExtraParameters,
		utils.CONTENT_JSON, config.CgrConfig().GeneralCfg().PosterAttempts)
	if err != nil {
		return err
	}
	err = pstr.Post(body, utils.EmptyString)
	if err != nil && config.CgrConfig().GeneralCfg().FailedPostsDir != utils.META_NONE {
		addFailedPost(a.ExtraParameters, utils.MetaHTTPjson, utils.ActionsPoster+utils.HIERARCHY_SEP+a.ActionType, body)
		err = nil
	}
	return err
}

// Does not block for posts, no error reports
func callURLAsync(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	pstr, err := NewHTTPPoster(config.CgrConfig().GeneralCfg().HttpSkipTlsVerify,
		config.CgrConfig().GeneralCfg().ReplyTimeout, a.ExtraParameters,
		utils.CONTENT_JSON, config.CgrConfig().GeneralCfg().PosterAttempts)
	if err != nil {
		return err
	}
	go func() {
		err := pstr.Post(body, utils.EmptyString)
		if err != nil && config.CgrConfig().GeneralCfg().FailedPostsDir != utils.META_NONE {
			addFailedPost(a.ExtraParameters, utils.MetaHTTPjson, utils.ActionsPoster+utils.HIERARCHY_SEP+a.ActionType, body)
		}
	}()
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
	var ddcDestID string
	for _, bchain := range ub.BalanceMap {
		for _, b := range bchain {
			for destID := range b.DestinationIDs {
				if strings.HasPrefix(destID, utils.MetaDDC) {
					ddcDestID = destID
					break
				}
			}
			if ddcDestID != "" {
				break
			}
		}
		if ddcDestID != "" {
			break
		}
	}
	if ddcDestID != "" {
		destinations := utils.NewStringSet(nil)
		for _, statID := range strings.Split(a.ExtraParameters, utils.INFIELD_SEP) {
			if statID == utils.EmptyString {
				continue
			}
			var sts StatQueue
			if err = connMgr.Call(config.CgrConfig().RalsCfg().StatSConns, nil, utils.StatSv1GetStatQueue,
				&utils.TenantIDWithArgDispatcher{
					TenantID: &utils.TenantID{
						Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
						ID:     statID,
					},
				}, &sts); err != nil {
				return
			}
			ddcIface, has := sts.SQMetrics[utils.MetaDDC]
			if !has {
				continue
			}
			ddcMetric := ddcIface.(*StatDDC)

			// make slice from prefixes
			// Review here prefixes
			for p := range ddcMetric.FieldValues {
				destinations.Add(p)
			}
		}

		newDest := &Destination{Id: ddcDestID, Prefixes: destinations.AsSlice()}
		oldDest, err := dm.GetDestination(ddcDestID, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		// update destid in storage
		if err = dm.SetDestination(newDest, utils.NonTransactional); err != nil {
			return err
		}
		if err = dm.CacheDataFromDB(utils.DESTINATION_PREFIX, []string{ddcDestID}, true); err != nil {
			return err
		}

		if err == nil && oldDest != nil {
			if err = dm.UpdateReverseDestination(oldDest, newDest, utils.NonTransactional); err != nil {
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
		accID = utils.ConcatenatedKey(accountInfo.Tenant, accountInfo.Account)
	}
	if accID == "" {
		return utils.ErrInvalidKey
	}

	if err := dm.RemoveAccount(accID); err != nil {
		utils.Logger.Err(fmt.Sprintf("Could not remove account Id: %s: %v", accID, err))
		return err
	}

	_, err := guardian.Guardian.Guard(func() (interface{}, error) {
		acntAPids, err := dm.GetAccountActionPlans(accID, true, true, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			utils.Logger.Err(fmt.Sprintf("Could not get action plans: %s: %v", accID, err))
			return 0, err
		}
		for _, apID := range acntAPids {
			ap, err := dm.GetActionPlan(apID, true, true, utils.NonTransactional)
			if err != nil {
				utils.Logger.Err(fmt.Sprintf("Could not retrieve action plan: %s: %v", apID, err))
				return 0, err
			}
			delete(ap.AccountIDs, accID)
			if err := dm.SetActionPlan(apID, ap, true, utils.NonTransactional); err != nil {
				utils.Logger.Err(fmt.Sprintf("Could not save action plan: %s: %v", apID, err))
				return 0, err
			}
		}
		if err = dm.CacheDataFromDB(utils.ACTION_PLAN_PREFIX, acntAPids, true); err != nil {
			return 0, err
		}
		if err = dm.RemAccountActionPlans(accID, nil); err != nil {
			return 0, err
		}
		if err = dm.CacheDataFromDB(utils.AccountActionPlansPrefix, []string{accID}, true); err != nil && err.Error() != utils.ErrNotFound.Error() {
			return 0, err
		}
		return 0, nil

	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ACTION_PLAN_PREFIX)
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
			i--
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

// RPCRequest used by rpc action
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
	var client rpcclient.ClientConnector
	if req.Address != utils.MetaInternal {
		if client, err = rpcclient.NewRPCClient(utils.TCP, req.Address, false, "", "", "",
			req.Attempts, 0, config.CgrConfig().GeneralCfg().ConnectTimeout,
			config.CgrConfig().GeneralCfg().ReplyTimeout, req.Transport,
			nil, false); err != nil {
			return err
		}
	} else {
		client = params.Object.(rpcclient.ClientConnector)
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

// Actions used to store actions according to weight
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

// Sort used to implement sort interface
func (apl Actions) Sort() {
	sort.Sort(apl)
}

// Clone returns a clone from object
func (apl Actions) Clone() (interface{}, error) {
	if apl == nil {
		return nil, nil
	}
	cln := make(Actions, len(apl))
	for i, action := range apl {
		cln[i] = action.Clone()
	}
	return cln, nil
}

// newCdrLogProvider constructs a DataProvider
func newCdrLogProvider(acnt *Account, action *Action) (dP utils.DataProvider) {
	dP = &cdrLogProvider{acnt: acnt, action: action, cache: utils.MapStorage{}}
	return
}

// cdrLogProvider implements engine.DataProvider so we can pass it to filters
type cdrLogProvider struct {
	acnt   *Account
	action *Action
	cache  utils.MapStorage
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
	cdrP.cache.Set(fldPath, data)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (cdrP *cdrLogProvider) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = cdrP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// RemoteHost is part of engine.DataProvider interface
func (cdrP *cdrLogProvider) RemoteHost() net.Addr {
	return utils.LocalAddr()
}

func removeSessionCosts(_ *Account, action *Action, _ Actions, _ interface{}) error { // FiltersID;inlineFilter
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	smcFilter := new(utils.SMCostFilter)
	for _, fltrID := range strings.Split(action.ExtraParameters, utils.INFIELD_SEP) {
		if len(fltrID) == 0 {
			continue
		}
		fltr, err := GetFilter(dm, tenant, fltrID, true, true, utils.NonTransactional)
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s>  Error: %s for filter: %s in action: <%s>",
				utils.Actions, err.Error(), fltrID, utils.MetaRemoveSessionCosts))
			continue
		}
		for _, rule := range fltr.Rules {
			smcFilter, err = utils.AppendToSMCostFilter(smcFilter, rule.Type, rule.Element, rule.Values, config.CgrConfig().GeneralCfg().DefaultTimezone)
			if err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> %s in action: <%s>", utils.Actions, err.Error(), utils.MetaRemoveSessionCosts))
			}
		}
	}
	return cdrStorage.RemoveSMCosts(smcFilter)
}

func removeExpired(acc *Account, action *Action, _ Actions, extraData interface{}) error {
	if acc == nil {
		return fmt.Errorf("nil account for %s action", utils.ToJSON(action))
	}

	bChain, exists := acc.BalanceMap[action.Balance.GetType()]
	if !exists {
		return utils.ErrNotFound
	}

	found := false
	for i := 0; i < len(bChain); i++ {
		if bChain[i].IsExpiredAt(time.Now()) {
			// delete without preserving order
			bChain[i] = bChain[len(bChain)-1]
			bChain = bChain[:len(bChain)-1]
			i--
			found = true
		}
	}
	acc.BalanceMap[action.Balance.GetType()] = bChain
	if !found {
		return utils.ErrNotFound
	}
	return nil
}

func postEvent(ub *Account, a *Action, acs Actions, extraData interface{}) error {
	body, err := json.Marshal(extraData)
	if err != nil {
		return err
	}
	pstr, err := NewHTTPPoster(config.CgrConfig().GeneralCfg().HttpSkipTlsVerify,
		config.CgrConfig().GeneralCfg().ReplyTimeout, a.ExtraParameters,
		utils.CONTENT_JSON, config.CgrConfig().GeneralCfg().PosterAttempts)
	if err != nil {
		return err
	}
	err = pstr.Post(body, utils.EmptyString)
	if err != nil && config.CgrConfig().GeneralCfg().FailedPostsDir != utils.META_NONE {
		addFailedPost(a.ExtraParameters, utils.MetaHTTPjson, utils.ActionsPoster+utils.HIERARCHY_SEP+a.ActionType, body)
		err = nil
	}
	return err
}

func resetAccount(ub *Account, action *Action, acts Actions, _ interface{}) error {
	if ub == nil {
		return errors.New("nil account")
	}
	if cdrStorage == nil {
		return fmt.Errorf("nil cdrStorage for %s action", utils.ToJSON(action))
	}
	account := ub.GetID()
	filter := &utils.CDRsFilter{
		Accounts:  []string{account},
		NotCosts:  []float64{-1},
		OrderBy:   fmt.Sprintf("%s%sdesc", utils.OrderID, utils.INFIELD_SEP),
		Paginator: utils.Paginator{Limit: utils.IntPointer(1)},
	}
	cdrs, _, err := cdrStorage.GetCDRs(filter, false)
	if err != nil {
		return err
	}
	cd := cdrs[0].CostDetails
	if cd == nil {
		return errors.New("nil CostDetails")
	}
	acs := cd.AccountSummary
	if acs == nil {
		return errors.New("nil AccountSummary")
	}
	for _, bsum := range acs.BalanceSummaries {
		if bsum == nil {
			continue
		}
		if err := ub.setBalanceAction(&Action{
			Balance: &BalanceFilter{
				Uuid:     &bsum.UUID,
				ID:       &bsum.ID,
				Type:     &bsum.Type,
				Value:    &utils.ValueFormula{Static: bsum.Value},
				Disabled: &bsum.Disabled,
			},
		}); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> Error %s setting balance %s for account: %s", utils.Actions, err, bsum.UUID, account))
		}
	}
	return nil
}
