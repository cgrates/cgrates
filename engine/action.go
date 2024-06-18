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
	"net/http"
	"net/smtp"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
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
	Filters          []string
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
	var fltrs []string
	if a.Filters != nil {
		fltrs = slices.Clone(a.Filters)
	}
	return &Action{
		Id:               a.Id,
		ActionType:       a.ActionType,
		ExtraParameters:  a.ExtraParameters,
		Filters:          fltrs,
		ExpirationString: a.ExpirationString,
		Weight:           a.Weight,
		Balance:          a.Balance.Clone(),
	}
}

// SharedActionsData holds shared data for processing actions within a group.
type SharedActionsData struct {

	// idx represents the current iteration index of the action being processed.
	// It is used as a unique key to link stored data to the correct action, since actions
	// within a group do not have a unique identifier.
	idx int

	refTime     time.Time                // reference time, constant for all actions in the group
	transferBal map[int]transferInfo     // data for *transfer_balance actions
	remBal      map[int]BalanceSummaries // data for *remove_balance actions
	remExp      map[int]BalanceSummaries // data for *remove_expired actions
	cdrLog      bool                     // indicates if a *cdrlog action is in the group
}

// NewSharedActionsData initializes SharedActionsData based on the provided actions.
func NewSharedActionsData(acts Actions) SharedActionsData {
	sd := SharedActionsData{
		refTime: time.Now(),
		cdrLog:  acts.HasAction(utils.CDRLog),
	}
	if acts.HasAction(utils.MetaRemoveBalance) {
		sd.remBal = make(map[int]BalanceSummaries)
	}
	if acts.HasAction(utils.MetaRemoveExpired) {
		sd.remExp = make(map[int]BalanceSummaries)
	}
	if acts.HasAction(utils.MetaTransferBalance) {
		sd.transferBal = make(map[int]transferInfo)
	}
	return sd
}

// transferInfo holds information for *transfer_balance actions.
type transferInfo struct {
	srcAccID  string          // source account ID
	destAccID string          // destination account ID
	units     float64         // number of units to transfer
	srcBal    *BalanceSummary // source balance summary
	destBal   *BalanceSummary // destination balance summary
}

type ActionConnCfg struct {
	ConnIDs []string
}

func newActionConnCfg(source, action string, cfg *config.CGRConfig) ActionConnCfg {
	sessionActions := []string{
		utils.MetaAlterSessions,
		utils.MetaForceDisconnectSessions,
	}
	act := ActionConnCfg{}
	switch source {
	case utils.ThresholdS:
		switch {
		case slices.Contains(sessionActions, action):
			act.ConnIDs = cfg.ThresholdSCfg().SessionSConns
		}
	case utils.RALs:
		switch {
		case slices.Contains(sessionActions, action):
			act.ConnIDs = cfg.RalsCfg().SessionSConns
		}
	}
	return act
}

type actionTypeFunc func(*Account, *Action, Actions, *FilterS, any, SharedActionsData, ActionConnCfg) error

var actionFuncMap = make(map[string]actionTypeFunc)

func init() {
	actionFuncMap[utils.MetaLog] = logAction
	actionFuncMap[utils.MetaResetTriggers] = resetTriggersAction
	actionFuncMap[utils.CDRLog] = cdrLogAction
	actionFuncMap[utils.MetaSetRecurrent] = setRecurrentAction
	actionFuncMap[utils.MetaUnsetRecurrent] = unsetRecurrentAction
	actionFuncMap[utils.MetaAllowNegative] = allowNegativeAction
	actionFuncMap[utils.MetaDenyNegative] = denyNegativeAction
	actionFuncMap[utils.MetaResetAccount] = resetAccountAction
	actionFuncMap[utils.MetaTopUpReset] = topupResetAction
	actionFuncMap[utils.MetaTopUp] = topupAction
	actionFuncMap[utils.MetaDebitReset] = debitResetAction
	actionFuncMap[utils.MetaDebit] = debitAction
	actionFuncMap[utils.MetaTransferBalance] = transferBalanceAction
	actionFuncMap[utils.MetaResetCounters] = resetCountersAction
	actionFuncMap[utils.MetaEnableAccount] = enableAccountAction
	actionFuncMap[utils.MetaDisableAccount] = disableAccountAction
	actionFuncMap[utils.MetaMailAsync] = mailAsync
	actionFuncMap[utils.MetaSetDDestinations] = setddestinations
	actionFuncMap[utils.MetaRemoveAccount] = removeAccountAction
	actionFuncMap[utils.MetaRemoveBalance] = removeBalanceAction
	actionFuncMap[utils.MetaSetBalance] = setBalanceAction
	actionFuncMap[utils.MetaTransferMonetaryDefault] = transferMonetaryDefaultAction
	actionFuncMap[utils.MetaCgrRpc] = cgrRPCAction
	actionFuncMap[utils.MetaAlterSessions] = alterSessionsAction
	actionFuncMap[utils.MetaForceDisconnectSessions] = forceDisconnectSessionsAction
	actionFuncMap[utils.TopUpZeroNegative] = topupZeroNegativeAction
	actionFuncMap[utils.SetExpiry] = setExpiryAction
	actionFuncMap[utils.MetaPublishAccount] = publishAccount
	actionFuncMap[utils.MetaRemoveSessionCosts] = removeSessionCosts
	actionFuncMap[utils.MetaRemoveExpired] = removeExpired
	actionFuncMap[utils.MetaCDRAccount] = resetAccountCDR
	actionFuncMap[utils.MetaExport] = export
	actionFuncMap[utils.MetaResetThreshold] = resetThreshold
	actionFuncMap[utils.MetaResetStatQueue] = resetStatQueue
	actionFuncMap[utils.MetaRemoteSetAccount] = remoteSetAccount
}

func getActionFunc(typ string) (f actionTypeFunc, exists bool) {
	f, exists = actionFuncMap[typ]
	return
}

func RegisterActionFunc(action string, f actionTypeFunc) {
	actionFuncMap[action] = f
}

// transferBalanceAction transfers units between accounts' balances.
// It ensures both source and destination balances are of the same type and non-expired.
// Destination account and balance IDs, and optionally a reference value, are obtained from Action's ExtraParameters.
// If a reference value is specified, the transfer ensures the destination balance reaches this value.
// If the destination account is different from the source, it is locked during the transfer.
func transferBalanceAction(srcAcc *Account, act *Action, acts Actions, fltrS *FilterS, _ any, sd SharedActionsData, _ ActionConnCfg) error {
	if srcAcc == nil {
		return errors.New("source account is nil")
	}
	if act.Balance.ID == nil {
		return errors.New("source balance ID is missing")
	}
	if act.ExtraParameters == "" {
		return errors.New("ExtraParameters used to identify the destination balance are missing")
	}
	if len(srcAcc.BalanceMap) == 0 {
		return fmt.Errorf("account %s has no balances to transfer from", srcAcc.ID)
	}

	srcBalance, srcBalanceType := srcAcc.FindBalanceByID(*act.Balance.ID)
	if srcBalance == nil || srcBalance.IsExpiredAt(time.Now()) {
		return errors.New("source balance not found or expired")
	}

	destInfo := struct {
		AccID  string   `json:"DestinationAccountID"`
		BalID  string   `json:"DestinationBalanceID"`
		RefVal *float64 `json:"DestinationReferenceValue"`
	}{}
	if err := json.Unmarshal([]byte(act.ExtraParameters), &destInfo); err != nil {
		return err
	}

	// Lock the destination account if different from source, otherwise
	// pass without lock key and timeout.
	diffAcnts := srcAcc.ID != destInfo.AccID
	var lockTimeout time.Duration
	lockKeys := make([]string, 0, 1)
	if diffAcnts {
		lockTimeout = config.CgrConfig().GeneralCfg().LockingTimeout
		lockKeys = append(lockKeys, utils.AccountPrefix+destInfo.AccID)
	}

	// This guard is meant to lock the destination account as we are making changes
	// to it. It is needed for the source account due to it being locked from outside
	// this function.
	guardErr := guardian.Guardian.Guard(func() error {

		var destAcc *Account
		switch diffAcnts {
		case true:
			var err error
			if destAcc, err = dm.GetAccount(destInfo.AccID); err != nil {
				return fmt.Errorf("retrieving destination account failed: %w", err)
			}
		case false:
			destAcc = srcAcc
		}

		if destAcc.BalanceMap == nil {
			destAcc.BalanceMap = make(map[string]Balances)
		}

		// Look for the destination balance only through balances of the same type as the source balance.
		destBalance := destAcc.GetBalanceWithID(srcBalanceType, destInfo.BalID)
		if destBalance != nil && destBalance.IsExpiredAt(time.Now()) {
			return errors.New("destination balance expired")
		}

		if destBalance == nil {
			// Destination Balance was not found. Create it and add it to the balance map.
			destBalance = &Balance{
				ID:   destInfo.BalID,
				Uuid: utils.GenUUID(),
			}
			destAcc.BalanceMap[srcBalanceType] = append(destAcc.BalanceMap[srcBalanceType], destBalance)
		}

		// If DestinationReferenceValue is specified adjust transferUnits to make the
		// destination balance match the DestinationReferenceValue.
		transferUnits := act.Balance.GetValue()
		if destInfo.RefVal != nil {
			transferUnits = *destInfo.RefVal - destBalance.Value
		}
		if transferUnits == 0 {
			return errors.New("transfer amount is missing or 0")
		}
		if srcBalance.ID != utils.MetaDefault && transferUnits > srcBalance.Value {
			return fmt.Errorf("insufficient credits in source balance %q (account %q) for transfer of %.2f units",
				srcBalance.ID, srcAcc.ID, transferUnits)
		}
		if destBalance.ID != utils.MetaDefault && -transferUnits > destBalance.Value {
			return fmt.Errorf("insufficient credits in destination balance %q (account %q) for transfer of %.2f units",
				destBalance.ID, destAcc.ID, transferUnits)
		}

		srcBalance.SubtractValue(transferUnits)
		srcBalance.dirty = true
		destBalance.AddValue(transferUnits)
		destBalance.dirty = true

		if sd.cdrLog {
			sd.transferBal[sd.idx] = transferInfo{
				srcAccID:  utils.NewTenantID(srcAcc.ID).ID,
				destAccID: utils.NewTenantID(destAcc.ID).ID,
				units:     transferUnits,
				srcBal:    srcBalance.AsBalanceSummary(srcBalanceType),
				destBal:   destBalance.AsBalanceSummary(srcBalanceType),
			}
			sd.transferBal[sd.idx].srcBal.Initial = sd.transferBal[sd.idx].srcBal.Value + transferUnits
			sd.transferBal[sd.idx].destBal.Initial = sd.transferBal[sd.idx].destBal.Value - transferUnits
		}

		if diffAcnts {
			destAcc.InitCounters()
			destAcc.ExecuteActionTriggers(act, fltrS)
			if err := dm.SetAccount(destAcc); err != nil {
				return fmt.Errorf("updating destination account failed: %w", err)
			}
		}
		return nil
	}, lockTimeout, lockKeys...)
	if guardErr != nil {
		return guardErr
	}

	// Execute action triggers for the source account.
	// This account will be updated in the parent function.
	srcAcc.InitCounters()
	srcAcc.ExecuteActionTriggers(act, fltrS)
	return nil
}

func logAction(ub *Account, _ *Action, _ Actions, _ *FilterS, extraData any, _ SharedActionsData, _ ActionConnCfg) (err error) {
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

func cdrLogAction(acc *Account, a *Action, acs Actions, _ *FilterS, extraData any,
	sd SharedActionsData, _ ActionConnCfg) (err error) {
	if len(config.CgrConfig().SchedulerCfg().CDRsConns) == 0 {
		return errors.New("No connection with CDR Server")
	}
	defaultTemplate := map[string]config.RSRParsers{
		utils.ToR:          config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaAcnt+utils.NestingSep+utils.BalanceType, utils.InfieldSep),
		utils.OriginHost:   config.NewRSRParsersMustCompile("127.0.0.1", utils.InfieldSep),
		utils.RequestType:  config.NewRSRParsersMustCompile(utils.MetaNone, utils.InfieldSep),
		utils.Tenant:       config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaAcnt+utils.NestingSep+utils.Tenant, utils.InfieldSep),
		utils.AccountField: config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaAcnt+utils.NestingSep+utils.AccountField, utils.InfieldSep),
		utils.Subject:      config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaAcnt+utils.NestingSep+utils.AccountField, utils.InfieldSep),
		utils.Cost:         config.NewRSRParsersMustCompile(utils.DynamicDataPrefix+utils.MetaAct+utils.NestingSep+utils.ActionValue, utils.InfieldSep),
	}
	template := make(map[string]string)
	// overwrite default template
	if a.ExtraParameters != "" {
		if err = json.Unmarshal([]byte(a.ExtraParameters), &template); err != nil {
			return err
		}
		for field, rsr := range template {
			if defaultTemplate[field], err = config.NewRSRParsers(rsr,
				config.CgrConfig().GeneralCfg().RSRSep); err != nil {
				return err
			}
		}
	}
	//In case that we have extra data we populate default templates
	mapExtraData, _ := extraData.(map[string]any)
	for key, val := range mapExtraData {
		if defaultTemplate[key], err = config.NewRSRParsers(utils.IfaceAsString(val),
			config.CgrConfig().GeneralCfg().RSRSep); err != nil {
			return err
		}
	}

	// set stored cdr values
	var cdrs []*CDR
	for i, action := range acs {
		if !slices.Contains(
			[]string{utils.MetaDebit, utils.MetaDebitReset,
				utils.MetaTopUp, utils.MetaTopUpReset,
				utils.MetaSetBalance, utils.MetaRemoveBalance,
				utils.MetaRemoveExpired, utils.MetaTransferBalance,
			}, action.ActionType) || action.Balance == nil {
			continue // Only log specific actions
		}

		cdr := &CDR{
			RunID:       action.ActionType,
			Source:      utils.CDRLog,
			SetupTime:   sd.refTime,
			AnswerTime:  sd.refTime,
			OriginID:    utils.GenUUID(),
			ExtraFields: make(map[string]string),
			PreRated:    true,
			Usage:       time.Duration(1),
		}
		cdr.CGRID = utils.Sha1(cdr.OriginID, cdr.OriginHost)

		cdrLogProvider := newCdrLogProvider(acc, action)
		elem := reflect.ValueOf(cdr).Elem()
		for key, rsrFlds := range defaultTemplate {
			parsedValue, err := rsrFlds.ParseDataProvider(cdrLogProvider)
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

		// Function to create (and append) CDRs based on each BalanceSummary element.
		processBalances := func(balances BalanceSummaries) error {
			if len(balances) == 0 {
				return utils.ErrNotFound
			}
			for _, b := range balances {
				// Create a new CDR instance for each balance that meets the condition.
				newCDR := *cdr // Copy CDR's values to a new CDR instance.
				newCDR.Cost = b.Value
				newCDR.OriginID = utils.GenUUID() // OriginID must be unique for every CDR.
				newCDR.CGRID = utils.Sha1(newCDR.OriginID, newCDR.OriginHost)
				newCDR.ToR = b.Type

				// Clone the ExtraFields map to avoid changing its value in
				// CDRs appended previously.
				newCDR.ExtraFields = make(map[string]string, len(cdr.ExtraFields)+1)
				for key, val := range cdr.ExtraFields {
					newCDR.ExtraFields[key] = val
				}
				newCDR.ExtraFields[utils.BalanceID] = b.ID

				cdrs = append(cdrs, &newCDR) // Append the address of the new instance.
			}
			return nil
		}

		// If the action is of type *remove_balance or *remove_expired, for each matched balance,
		// assign the balance values to the CDR cost and append to the list of CDRs.
		switch action.ActionType {
		case utils.MetaRemoveBalance:
			if err = processBalances(sd.remBal[i]); err != nil {
				return err
			}
			continue
		case utils.MetaRemoveExpired:
			if err = processBalances(sd.remExp[i]); err != nil {
				return err
			}
			continue
		case utils.MetaTransferBalance:
			cdr.Account = sd.transferBal[i].srcAccID
			cdr.Destination = sd.transferBal[i].destAccID
			cdr.Cost = sd.transferBal[i].units
			cdr.ExtraFields[utils.SourceBalanceSummary] = utils.ToJSON(sd.transferBal[i].srcBal)
			cdr.ExtraFields[utils.DestinationBalanceSummary] = utils.ToJSON(sd.transferBal[i].destBal)
		}
		cdrs = append(cdrs, cdr)
	}

	events := make([]*utils.CGREvent, 0, len(cdrs))
	for _, cdr := range cdrs {
		events = append(events, cdr.AsCGREvent())
	}
	var reply string
	if err := connMgr.Call(context.TODO(), config.CgrConfig().SchedulerCfg().CDRsConns,
		utils.CDRsV1ProcessEvents,
		&ArgV1ProcessEvents{
			Flags:     []string{utils.ConcatenatedKey(utils.MetaChargers, "false")},
			CGREvents: events,
		}, &reply); err != nil {
		return err
	}
	b, _ := json.Marshal(cdrs)
	a.ExpirationString = string(b) // testing purpose only
	return nil
}

func resetTriggersAction(ub *Account, a *Action, _ Actions, fltrS *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.ResetActionTriggers(a, fltrS)
	return
}

func setRecurrentAction(ub *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.SetRecurrent(a, true)
	return
}

func unsetRecurrentAction(ub *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.SetRecurrent(a, false)
	return
}

func allowNegativeAction(ub *Account, _ *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.AllowNegative = true
	return
}

func denyNegativeAction(ub *Account, _ *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.AllowNegative = false
	return
}

func resetAccountAction(ub *Account, _ *Action, _ Actions, fltrS *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	return genericReset(ub, fltrS)
}

func topupResetAction(ub *Account, a *Action, _ Actions, fltrS *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil { // Init the map since otherwise will get error if nil
		ub.BalanceMap = make(map[string]Balances)
	}
	c := a.Clone()
	genericMakeNegative(c)
	err = genericDebit(ub, c, true, fltrS)
	a.balanceValue = c.balanceValue
	return
}

func topupAction(ub *Account, a *Action, _ Actions, fltrS *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	c := a.Clone()
	genericMakeNegative(c)
	err = genericDebit(ub, c, false, fltrS)
	a.balanceValue = c.balanceValue
	return
}

func debitResetAction(ub *Account, a *Action, _ Actions, fltrS *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil { // Init the map since otherwise will get error if nil
		ub.BalanceMap = make(map[string]Balances)
	}
	return genericDebit(ub, a, true, fltrS)
}

func debitAction(ub *Account, a *Action, _ Actions, fltrS *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	err = genericDebit(ub, a, false, fltrS)
	return
}

func resetCountersAction(ub *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
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

func genericDebit(ub *Account, a *Action, reset bool, fltrS *FilterS) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]Balances)
	}
	return ub.debitBalanceAction(a, reset, false, fltrS)
}

func enableAccountAction(acc *Account, _ *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	if acc == nil {
		return errors.New("nil account")
	}
	acc.Disabled = false
	return
}

func disableAccountAction(acc *Account, _ *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
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

func genericReset(ub *Account, fltrS *FilterS) error {
	for k := range ub.BalanceMap {
		ub.BalanceMap[k] = Balances{&Balance{Value: 0}}
	}
	ub.InitCounters()
	ub.ResetActionTriggers(nil, fltrS)
	return nil
}

// Mails the balance hitting the threshold towards predefined list of addresses
func mailAsync(ub *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) error {
	cgrCfg := config.CgrConfig()
	params := strings.Split(a.ExtraParameters, string(utils.CSVSep))
	if len(params) == 0 {
		return errors.New("Unconfigured parameters for mail action")
	}
	toAddrs := strings.Split(params[0], string(utils.FallbackSep))
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
	var auth smtp.Auth
	if len(cgrCfg.MailerCfg().MailerAuthUser) > 0 || len(cgrCfg.MailerCfg().MailerAuthPass) > 0 { //use auth if user/pass not empty in config
		auth = smtp.PlainAuth("", cgrCfg.MailerCfg().MailerAuthUser, cgrCfg.MailerCfg().MailerAuthPass, strings.Split(cgrCfg.MailerCfg().MailerServer, ":")[0]) // We only need host part, so ignore port
	}
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

func setddestinations(ub *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
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
		for _, statID := range strings.Split(a.ExtraParameters, utils.InfieldSep) {
			if statID == utils.EmptyString {
				continue
			}
			var sts StatQueue
			if err = connMgr.Call(context.TODO(), config.CgrConfig().RalsCfg().StatSConns,
				utils.StatSv1GetStatQueue,
				&utils.TenantIDWithAPIOpts{
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
		oldDest, err := dm.GetDestination(ddcDestID, true, true, utils.NonTransactional)
		if err != nil {
			return err
		}
		// update destid in storage
		if err = dm.SetDestination(newDest, utils.NonTransactional); err != nil {
			return err
		}
		if err = dm.CacheDataFromDB(utils.DestinationPrefix, []string{ddcDestID}, true); err != nil {
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

func removeAccountAction(ub *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) error {
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

	return guardian.Guardian.Guard(func() error {
		acntAPids, err := dm.GetAccountActionPlans(accID, true, true, utils.NonTransactional)
		if err != nil && err != utils.ErrNotFound {
			utils.Logger.Err(fmt.Sprintf("Could not get action plans: %s: %v", accID, err))
			return err
		}
		for _, apID := range acntAPids {
			ap, err := dm.GetActionPlan(apID, true, true, utils.NonTransactional)
			if err != nil {
				utils.Logger.Err(fmt.Sprintf("Could not retrieve action plan: %s: %v", apID, err))
				return err
			}
			delete(ap.AccountIDs, accID)
			if err := dm.SetActionPlan(apID, ap, true, utils.NonTransactional); err != nil {
				utils.Logger.Err(fmt.Sprintf("Could not save action plan: %s: %v", apID, err))
				return err
			}
		}
		if err = dm.CacheDataFromDB(utils.ActionPlanPrefix, acntAPids, true); err != nil {
			return err
		}
		if err = dm.RemAccountActionPlans(accID, nil); err != nil {
			return err
		}
		if err = dm.CacheDataFromDB(utils.AccountActionPlansPrefix, []string{accID}, true); err != nil && err.Error() != utils.ErrNotFound.Error() {
			return err
		}
		return nil

	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ActionPlanPrefix)
}

func removeBalanceAction(acc *Account, a *Action, acts Actions, _ *FilterS, _ any,
	sd SharedActionsData, _ ActionConnCfg) error {
	if acc == nil {
		return fmt.Errorf("nil account for %s action", utils.ToJSON(a))
	}
	found := false
	for bType, bChain := range acc.BalanceMap {
		for i := 0; i < len(bChain); i++ {
			if bChain[i].MatchFilter(a.Balance, bType, false, false) {
				if sd.cdrLog {
					// If *cdrlog action is present, add the balance summary to remBal in SharedActionsData
					// for CDR creation.
					sd.remBal[sd.idx] = append(sd.remBal[sd.idx], bChain[i].AsBalanceSummary(bType))
				}

				// Remove balance without preserving order.
				bChain[i] = bChain[len(bChain)-1]
				bChain = bChain[:len(bChain)-1]
				i--
				found = true
			}
		}
		acc.BalanceMap[bType] = bChain
	}
	if !found {
		return utils.ErrNotFound
	}
	return nil
}

func setBalanceAction(ub *Account, a *Action, _ Actions, fltrS *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) error {
	if ub == nil {
		return fmt.Errorf("nil account for %s action", utils.ToJSON(a))
	}
	return ub.setBalanceAction(a, fltrS)
}

func transferMonetaryDefaultAction(ub *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) error {
	if ub == nil {
		utils.Logger.Err("*transfer_monetary_default called without account")
		return utils.ErrAccountNotFound
	}
	if _, exists := ub.BalanceMap[utils.MetaMonetary]; !exists {
		return utils.ErrNotFound
	}
	defaultBalance := ub.GetDefaultMoneyBalance()
	bChain := ub.BalanceMap[utils.MetaMonetary]
	for _, balance := range bChain {
		if balance.Uuid != defaultBalance.Uuid &&
			balance.ID != defaultBalance.ID && // extra caution
			balance.MatchFilter(a.Balance, "", false, false) {
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
	Params    map[string]any
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
func cgrRPCAction(ub *Account, a *Action, acs Actions, _ *FilterS, extraData any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	// parse template
	tmpl := template.New("extra_params")
	tmpl.Delims("<<", ">>")
	if tmpl, err = tmpl.Parse(a.ExtraParameters); err != nil {
		utils.Logger.Err(fmt.Sprintf("error parsing *cgr_rpc template: %s", err.Error()))
		return
	}
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, struct {
		Account   *Account
		Action    *Action
		Actions   Actions
		ExtraData any
	}{ub, a, acs, extraData}); err != nil {
		utils.Logger.Err(fmt.Sprintf("error executing *cgr_rpc template %s:", err.Error()))
		return
	}
	var req RPCRequest
	if err = json.Unmarshal(buf.Bytes(), &req); err != nil {
		return
	}
	var params *utils.RpcParams
	if params, err = utils.GetRpcParams(req.Method); err != nil {
		return
	}
	var client birpc.ClientConnector
	if req.Address == utils.MetaInternal {
		client = params.Object.(birpc.ClientConnector)
	} else if client, err = rpcclient.NewRPCClient(context.TODO(), utils.TCP, req.Address, false, "", "", "",
		req.Attempts, 0,
		config.CgrConfig().GeneralCfg().MaxReconnectInterval,
		utils.FibDuration,
		config.CgrConfig().GeneralCfg().ConnectTimeout,
		config.CgrConfig().GeneralCfg().ReplyTimeout,
		req.Transport, nil, false, nil); err != nil {
		return
	}

	// Decode's output parameter requires a pointer.
	if reflect.TypeOf(params.InParam).Kind() == reflect.Pointer {
		err = mapstructure.Decode(req.Params, params.InParam)
	} else {
		err = mapstructure.Decode(req.Params, &params.InParam)

	}
	if err != nil {
		utils.Logger.Info("<*cgr_rpc> err: " + err.Error())
		return
	}
	if params.InParam == nil {
		utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> nil params err: req.Params: %+v params: %+v", req.Params, params))
		return utils.ErrParserError
	}
	utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> calling: %s with: %s and result %v", req.Method, utils.ToJSON(params.InParam), params.OutParam))
	if !req.Async {
		err = client.Call(context.TODO(), req.Method, params.InParam, params.OutParam)
		utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> result: %s err: %v", utils.ToJSON(params.OutParam), err))
		return
	}
	go func() {
		err := client.Call(context.TODO(), req.Method, params.InParam, params.OutParam)
		utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> result: %s err: %v", utils.ToJSON(params.OutParam), err))
	}()
	return
}

// alterSessionsAction processes the `ExtraParameters` field from the action to construct a request
// for the `SessionSv1.AlterSessions` API call.
//
// The ExtraParameters field format is expected as follows:
//   - tenant: string
//   - filters: strings separated by "&".
//   - limit: integer, specifying the maximum number of sessions to alter.
//   - APIOpts: set of key-value pairs (separated by "&").
//   - Event: set of key-value pairs (separated by "&").
//
// Parameters are separated by ";" and must be provided in the specified order.
func alterSessionsAction(_ *Account, act *Action, _ Actions, _ *FilterS, _ any,
	_ SharedActionsData, connCfg ActionConnCfg) (err error) {

	// Parse action parameters based on the predefined format.
	params := strings.Split(act.ExtraParameters, ";")
	if len(params) != 5 {
		return errors.New("invalid number of parameters; expected 5")
	}

	// If conversion fails, limit will default to 0.
	limit, _ := strconv.Atoi(params[2])

	// Prepare request arguments based on provided parameters.
	attr := utils.SessionFilterWithEvent{
		SessionFilter: &utils.SessionFilter{
			Limit:   &limit,
			Tenant:  params[0],
			Filters: strings.Split(params[1], "&"),
			APIOpts: make(map[string]any),
		},
		Event: make(map[string]any),
	}

	if err := parseParamStringToMap(params[3], attr.APIOpts); err != nil {
		return err
	}
	if err := parseParamStringToMap(params[4], attr.Event); err != nil {
		return err
	}

	var reply string
	return connMgr.Call(context.Background(), connCfg.ConnIDs, utils.SessionSv1AlterSessions, attr, &reply)
}

// forceDisconnectSessionsAction processes the `ExtraParameters` field from the action to construct a request
// for the `SessionSv1.ForceDisconnect` API call.
//
// The ExtraParameters field format is expected as follows:
//   - tenant: string
//   - filters: strings separated by "&".
//   - limit: integer, specifying the maximum number of sessions to disconnect.
//   - APIOpts: set of key-value pairs (separated by "&").
//   - Event: set of key-value pairs (separated by "&").
//
// Parameters are separated by ";" and must be provided in the specified order.
func forceDisconnectSessionsAction(_ *Account, act *Action, _ Actions, _ *FilterS, _ any,
	_ SharedActionsData, connCfg ActionConnCfg) (err error) {

	// Parse action parameters based on the predefined format.
	params := strings.Split(act.ExtraParameters, ";")
	if len(params) != 5 {
		return errors.New("invalid number of parameters; expected 5")
	}

	// If conversion fails, limit will default to 0.
	limit, _ := strconv.Atoi(params[2])

	// Prepare request arguments based on provided parameters.
	attr := utils.SessionFilterWithEvent{
		SessionFilter: &utils.SessionFilter{
			Limit:   &limit,
			Tenant:  params[0],
			Filters: strings.Split(params[1], "&"),
			APIOpts: make(map[string]any),
		},
		Event: make(map[string]any),
	}

	if err := parseParamStringToMap(params[3], attr.APIOpts); err != nil {
		return err
	}
	if err := parseParamStringToMap(params[4], attr.Event); err != nil {
		return err
	}

	var reply string
	return connMgr.Call(context.Background(), connCfg.ConnIDs, utils.SessionSv1ForceDisconnect, attr, &reply)
}

// parseParamStringToMap parses a string containing key-value pairs separated by "&" and assigns
// these pairs to a given map. Each pair is expected to be in the format "key:value".
func parseParamStringToMap(paramStr string, targetMap map[string]any) error {
	for _, tuple := range strings.Split(paramStr, "&") {
		// Use strings.Cut to split 'tuple' into key-value pairs at the first occurrence of ':'.
		// This ensures that additional ':' characters within the value do not affect parsing.
		key, value, found := strings.Cut(tuple, ":")
		if !found {
			return fmt.Errorf("invalid key-value pair: %s", tuple)
		}
		targetMap[key] = value
	}
	return nil
}

func topupZeroNegativeAction(ub *Account, a *Action, _ Actions, fltrS *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) error {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]Balances)
	}
	return ub.debitBalanceAction(a, false, true, fltrS)
}

func setExpiryAction(ub *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) error {
	if ub == nil {
		return errors.New("nil account")
	}
	balanceType := a.Balance.GetType()
	for _, b := range ub.BalanceMap[balanceType] {
		if b.MatchFilter(a.Balance, "", false, true) {
			b.ExpirationDate = a.Balance.GetExpirationDate()
		}
	}
	return nil
}

// publishAccount will publish the account as well as each balance received to ThresholdS
func publishAccount(ub *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) error {
	if ub == nil {
		return errors.New("nil account")
	}
	initBal := make(map[string]float64)
	for _, bals := range ub.BalanceMap {
		for _, bal := range bals {
			initBal[bal.Uuid] = bal.Value
		}
	}
	ub.Publish(initBal)
	return nil
}

// Actions used to store actions according to weight
type Actions []*Action

func (a Actions) Len() int {
	return len(a)
}

func (a Actions) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// we need higher weights earlyer in the list
func (a Actions) Less(j, i int) bool {
	return a[i].Weight < a[j].Weight
}

// Sort used to implement sort interface
func (a Actions) Sort() {
	sort.Sort(a)
}

// Clone returns a clone from object
func (a Actions) Clone() (any, error) {
	if a == nil {
		return nil, nil
	}
	cln := make(Actions, len(a))
	for i, action := range a {
		cln[i] = action.Clone()
	}
	return cln, nil
}

// HasAction checks if the action list contains an action of the given type.
func (a Actions) HasAction(typ string) bool {
	return slices.ContainsFunc(a, func(act *Action) bool {
		return act.ActionType == typ
	})
}

// newCdrLogProvider constructs a DataProvider
func newCdrLogProvider(acnt *Account, action *Action) (dP utils.DataProvider) {
	dP = &cdrLogProvider{acnt: acnt, action: action, cache: utils.MapStorage{}}
	return
}

// cdrLogProvider implements utils.DataProvider so we can pass it to filters
type cdrLogProvider struct {
	acnt   *Account
	action *Action
	cache  utils.MapStorage
}

// String is part of utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (cdrP *cdrLogProvider) String() string {
	return utils.ToJSON(cdrP)
}

// FieldAsInterface is part of utils.DataProvider interface
func (cdrP *cdrLogProvider) FieldAsInterface(fldPath []string) (data any, err error) {
	if data, err = cdrP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	if len(fldPath) == 2 {
		switch fldPath[0] {
		case utils.MetaAcnt:
			switch fldPath[1] {
			case utils.AccountID:
				data = cdrP.acnt.ID
			case utils.Tenant:
				tntAcnt := new(utils.TenantAccount) // Init with empty values
				if cdrP.acnt != nil {
					if tntAcnt, err = utils.NewTAFromAccountKey(cdrP.acnt.ID); err != nil {
						return
					}
				}
				data = tntAcnt.Tenant
			case utils.AccountField:
				tntAcnt := new(utils.TenantAccount) // Init with empty values
				if cdrP.acnt != nil {
					if tntAcnt, err = utils.NewTAFromAccountKey(cdrP.acnt.ID); err != nil {
						return
					}
				}
				data = tntAcnt.Account
			case utils.BalanceType:
				data = cdrP.action.Balance.GetType()
			case utils.BalanceUUID:
				data = cdrP.action.Balance.CreateBalance().Uuid
			case utils.BalanceID:
				data = cdrP.action.Balance.CreateBalance().ID
			case utils.BalanceValue:
				data = strconv.FormatFloat(cdrP.action.balanceValue, 'f', -1, 64)
			case utils.DestinationIDs:
				data = cdrP.action.Balance.CreateBalance().DestinationIDs.String()
			case utils.ExtraParameters:
				data = cdrP.action.ExtraParameters
			case utils.RatingSubject:
				data = cdrP.action.Balance.CreateBalance().RatingSubject
			case utils.Category:
				data = cdrP.action.Balance.Categories.String()
			case utils.SharedGroups:
				data = cdrP.action.Balance.SharedGroups.String()
			case utils.Factors:
				data = cdrP.action.Balance.Factors.String()
			}
		case utils.MetaAct:
			switch fldPath[1] {
			case utils.ActionID:
				data = cdrP.action.Id
			case utils.ActionType:
				data = cdrP.action.ActionType
			case utils.ActionValue:
				data = strconv.FormatFloat(cdrP.action.Balance.CreateBalance().GetValue(), 'f', -1, 64)
			}
		}
	} else {
		data = fldPath[0]
	}
	cdrP.cache.Set(fldPath, data)
	return
}

// FieldAsString is part of utils.DataProvider interface
func (cdrP *cdrLogProvider) FieldAsString(fldPath []string) (data string, err error) {
	var valIface any
	valIface, err = cdrP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

func removeSessionCosts(_ *Account, action *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) error { // FiltersID;inlineFilter
	tenant := config.CgrConfig().GeneralCfg().DefaultTenant
	smcFilter := new(utils.SMCostFilter)
	for _, fltrID := range strings.Split(action.ExtraParameters, utils.InfieldSep) {
		if len(fltrID) == 0 {
			continue
		}
		fltr, err := dm.GetFilter(tenant, fltrID, true, true, utils.NonTransactional)
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

func removeExpired(acc *Account, action *Action, acts Actions, _ *FilterS, _ any,
	sd SharedActionsData, _ ActionConnCfg) error {
	if acc == nil {
		return fmt.Errorf("nil account for %s action", utils.ToJSON(action))
	}
	found := false
	for bType, bChain := range acc.BalanceMap {
		for i := 0; i < len(bChain); i++ {
			if bChain[i].IsExpiredAt(sd.refTime) &&
				bChain[i].MatchFilter(action.Balance, bType, false, false) {
				if sd.cdrLog {
					// If *cdrlog action is present, add the balance summary to remExp in SharedActionsData
					// for CDR creation.
					sd.remExp[sd.idx] = append(sd.remExp[sd.idx], bChain[i].AsBalanceSummary(bType))
				}

				// Remove balance without maintaining order.
				bChain[i] = bChain[len(bChain)-1]
				bChain = bChain[:len(bChain)-1]
				i--
				found = true
			}
		}
		acc.BalanceMap[bType] = bChain
	}

	if !found {
		return utils.ErrNotFound
	}
	return nil
}

// resetAccountCDR resets the account out of values from CDR
func resetAccountCDR(ub *Account, action *Action, _ Actions, fltrS *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) error {
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
		OrderBy:   fmt.Sprintf("%s%sdesc", utils.OrderID, utils.InfieldSep),
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
				Weight:   &bsum.Weight,
				Disabled: &bsum.Disabled,
				Factors:  &bsum.Factors,
			},
		}, fltrS); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> Error %s setting balance %s for account: %s", utils.Actions, err, bsum.UUID, account))
		}
	}
	return nil
}

func export(ub *Account, a *Action, _ Actions, _ *FilterS, extraData any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	var cgrEv *utils.CGREvent
	switch {
	case ub != nil:
		cgrEv = &utils.CGREvent{
			Tenant: utils.NewTenantID(ub.ID).Tenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.AccountField:   ub.ID,
				utils.EventType:      utils.AccountUpdate,
				utils.EventSource:    utils.AccountService,
				utils.AllowNegative:  ub.AllowNegative,
				utils.Disabled:       ub.Disabled,
				utils.BalanceMap:     ub.BalanceMap,
				utils.UnitCounters:   ub.UnitCounters,
				utils.ActionTriggers: ub.ActionTriggers,
				utils.UpdateTime:     ub.UpdateTime,
			},

			APIOpts: map[string]any{
				utils.MetaEventType: utils.AccountUpdate,
			},
		}
	case extraData != nil:
		ev, canCast := extraData.(*utils.CGREvent)
		if !canCast {
			return
		}
		cgrEv = ev // only export  CGREvents
	default:
		return // nothing to post
	}
	args := &CGREventWithEeIDs{
		EeIDs:    strings.Split(a.ExtraParameters, utils.InfieldSep),
		CGREvent: cgrEv,
	}
	var rply map[string]map[string]any
	return connMgr.Call(context.TODO(), config.CgrConfig().ApierCfg().EEsConns,
		utils.EeSv1ProcessEvent, args, &rply)
}

func resetThreshold(_ *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID(a.ExtraParameters),
	}
	var rply string
	return connMgr.Call(context.TODO(), config.CgrConfig().SchedulerCfg().ThreshSConns,
		utils.ThresholdSv1ResetThreshold, args, &rply)
}

func resetStatQueue(_ *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID(a.ExtraParameters),
	}
	var rply string
	return connMgr.Call(context.TODO(), config.CgrConfig().SchedulerCfg().StatSConns,
		utils.StatSv1ResetStatQueue, args, &rply)
}

func remoteSetAccount(ub *Account, a *Action, _ Actions, _ *FilterS, _ any, _ SharedActionsData, _ ActionConnCfg) (err error) {
	client := &http.Client{Transport: httpPstrTransport}
	var resp *http.Response
	req := new(bytes.Buffer)
	if err = json.NewEncoder(req).Encode(ub); err != nil {
		return
	}
	if resp, err = client.Post(a.ExtraParameters, "application/json", req); err != nil {
		return
	}
	acc := new(Account)
	err = json.NewDecoder(resp.Body).Decode(acc)
	if err != nil {
		return
	}
	if len(acc.BalanceMap) != 0 {
		*ub = *acc
	}
	return
}
