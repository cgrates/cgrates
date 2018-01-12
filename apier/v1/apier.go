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

package v1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	"github.com/streadway/amqp"
)

const (
	OK = utils.OK
)

type ApierV1 struct {
	StorDb      engine.LoadStorage
	DataManager *engine.DataManager
	CdrDb       engine.CdrStorage
	Config      *config.CGRConfig
	Responder   *engine.Responder
	CdrStatsSrv rpcclient.RpcClientConnection
	Users       rpcclient.RpcClientConnection
	CDRs        rpcclient.RpcClientConnection // FixMe: populate it from cgr-engine
	ServManager *servmanager.ServiceManager   // Need to have them capitalize so we can export in V2
	HTTPPoster  *utils.HTTPPoster
}

func (self *ApierV1) GetDestination(dstId string, reply *engine.Destination) error {
	if dst, err := self.DataManager.DataDB().GetDestination(dstId, false, utils.NonTransactional); err != nil {
		return utils.ErrNotFound
	} else {
		*reply = *dst
	}
	return nil
}

type AttrRemoveDestination struct {
	DestinationIDs []string
	Prefixes       []string
}

func (self *ApierV1) RemoveDestination(attr AttrRemoveDestination, reply *string) (err error) {
	for _, dstID := range attr.DestinationIDs {
		if len(attr.Prefixes) == 0 {
			if err = self.DataManager.DataDB().RemoveDestination(dstID, utils.NonTransactional); err != nil {
				*reply = err.Error()
				break
			} else {
				*reply = utils.OK
			}
			// TODO list
			// get destination
			// remove prefixes
			// handle reverse destination
			// set destinastion
		}
	}
	return err
}

// GetReverseDestination retrieves revese destination list for a prefix
func (v1 *ApierV1) GetReverseDestination(prefix string, reply *[]string) (err error) {
	if prefix == "" {
		return utils.NewErrMandatoryIeMissing("prefix")
	}
	var revLst []string
	if revLst, err = v1.DataManager.DataDB().GetReverseDestination(prefix, false, utils.NonTransactional); err != nil {
		return
	}
	*reply = revLst
	return
}

// ComputeReverseDestinations will rebuild complete reverse destinations data
func (v1 *ApierV1) ComputeReverseDestinations(ignr string, reply *string) (err error) {
	if err = v1.DataManager.DataDB().RebuildReverseForPrefix(utils.REVERSE_DESTINATION_PREFIX); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// ComputeReverseAliases will rebuild complete reverse aliases data
func (v1 *ApierV1) ComputeReverseAliases(ignr string, reply *string) (err error) {
	if err = v1.DataManager.DataDB().RebuildReverseForPrefix(utils.REVERSE_ALIASES_PREFIX); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// ComputeReverseAliases will rebuild complete reverse aliases data
func (v1 *ApierV1) ComputeAccountActionPlans(ignr string, reply *string) (err error) {
	if err = v1.DataManager.DataDB().RebuildReverseForPrefix(utils.AccountActionPlansPrefix); err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (apier *ApierV1) GetSharedGroup(sgId string, reply *engine.SharedGroup) error {
	if sg, err := apier.DataManager.GetSharedGroup(sgId, false, utils.NonTransactional); err != nil && err != utils.ErrNotFound { // Not found is not an error here
		return err
	} else {
		if sg != nil {
			*reply = *sg
		}
	}
	return nil
}

func (self *ApierV1) SetDestination(attrs utils.AttrSetDestination, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"Id", "Prefixes"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	dest := &engine.Destination{Id: attrs.Id, Prefixes: attrs.Prefixes}
	var oldDest *engine.Destination
	if oldDest, err = self.DataManager.DataDB().GetDestination(attrs.Id, false, utils.NonTransactional); err != nil {
		if err != utils.ErrNotFound {
			return utils.NewErrServerError(err)
		}
	} else if !attrs.Overwrite {
		return utils.ErrExists
	}
	if err := self.DataManager.DataDB().SetDestination(dest, utils.NonTransactional); err != nil {
		return utils.NewErrServerError(err)
	}
	if err = self.DataManager.CacheDataFromDB(utils.DESTINATION_PREFIX, []string{attrs.Id}, true); err != nil {
		return
	}
	if err = self.DataManager.DataDB().UpdateReverseDestination(oldDest, dest, utils.NonTransactional); err != nil {
		return
	}
	if err = self.DataManager.CacheDataFromDB(utils.REVERSE_DESTINATION_PREFIX, dest.Prefixes, true); err != nil {
		return
	}
	*reply = OK
	return nil
}

func (self *ApierV1) GetRatingPlan(rplnId string, reply *engine.RatingPlan) error {
	if rpln, err := self.DataManager.GetRatingPlan(rplnId, false, utils.NonTransactional); err != nil {
		return utils.ErrNotFound
	} else {
		*reply = *rpln
	}
	return nil
}

func (self *ApierV1) ExecuteAction(attr *utils.AttrExecuteAction, reply *string) error {
	at := &engine.ActionTiming{
		ActionsID: attr.ActionsId,
	}
	if attr.Tenant != "" && attr.Account != "" {
		at.SetAccountIDs(utils.StringMap{utils.AccountKey(attr.Tenant, attr.Account): true})
	}
	if err := at.Execute(nil, nil); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

type AttrLoadDestination struct {
	TPid string
	ID   string
}

// Load destinations from storDb into dataDb.
func (self *ApierV1) LoadDestination(attrs AttrLoadDestination, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	if loaded, err := dbReader.LoadDestinationsFiltered(attrs.ID); err != nil {
		return utils.NewErrServerError(err)
	} else if !loaded {
		return utils.ErrNotFound
	}
	if err := self.DataManager.CacheDataFromDB(utils.DESTINATION_PREFIX, []string{attrs.ID}, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}

// Load derived chargers from storDb into dataDb.
func (self *ApierV1) LoadDerivedChargers(attrs utils.TPDerivedChargers, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	if err := dbReader.LoadDerivedChargersFiltered(&attrs, true); err != nil {
		return utils.NewErrServerError(err)
	}
	if err := self.DataManager.CacheDataFromDB(utils.DERIVEDCHARGERS_PREFIX, []string{attrs.GetDerivedChargersKey()}, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}

type AttrLoadRatingPlan struct {
	TPid         string
	RatingPlanId string
}

// Process dependencies and load a specific rating plan from storDb into dataDb.
func (self *ApierV1) LoadRatingPlan(attrs AttrLoadRatingPlan, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	if loaded, err := dbReader.LoadRatingPlansFiltered(attrs.RatingPlanId); err != nil {
		return utils.NewErrServerError(err)
	} else if !loaded {
		return utils.ErrNotFound
	}
	*reply = OK
	return nil
}

// Process dependencies and load a specific rating profile from storDb into dataDb.
func (self *ApierV1) LoadRatingProfile(attrs utils.TPRatingProfile, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	if err := dbReader.LoadRatingProfilesFiltered(&attrs); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}

type AttrLoadSharedGroup struct {
	TPid          string
	SharedGroupId string
}

// Load destinations from storDb into dataDb.
func (self *ApierV1) LoadSharedGroup(attrs AttrLoadSharedGroup, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	if err := dbReader.LoadSharedGroupsFiltered(attrs.SharedGroupId, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}

type AttrLoadCdrStats struct {
	TPid       string
	CdrStatsId string
}

// Load destinations from storDb into dataDb.
func (self *ApierV1) LoadCdrStats(attrs AttrLoadCdrStats, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	if err := dbReader.LoadCdrStatsFiltered(attrs.CdrStatsId, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}

type AttrLoadTpFromStorDb struct {
	TPid     string
	FlushDb  bool // Flush dataDB before loading
	DryRun   bool // Only simulate, no write
	Validate bool // Run structural checks
}

// Loads complete data in a TP from storDb
func (self *ApierV1) LoadTariffPlanFromStorDb(attrs AttrLoadTpFromStorDb, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	if err := dbReader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.Validate {
		if !dbReader.IsValid() {
			*reply = OK
			return errors.New("invalid data")
		}
	}
	if attrs.DryRun {
		*reply = OK
		return nil // Mission complete, no errors
	}
	if err := dbReader.WriteToDatabase(attrs.FlushDb, false, false); err != nil {
		return utils.NewErrServerError(err)
	}
	utils.Logger.Info("ApierV1.LoadTariffPlanFromStorDb, reloading cache.")
	for _, prfx := range []string{
		utils.DESTINATION_PREFIX,
		utils.REVERSE_DESTINATION_PREFIX,
		utils.ACTION_PLAN_PREFIX,
		utils.AccountActionPlansPrefix,
		utils.DERIVEDCHARGERS_PREFIX,
		utils.ALIASES_PREFIX,
		utils.REVERSE_ALIASES_PREFIX} {
		loadedIDs, _ := dbReader.GetLoadedIds(prfx)
		if err := self.DataManager.CacheDataFromDB(prfx, loadedIDs, true); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	aps, _ := dbReader.GetLoadedIds(utils.ACTION_PLAN_PREFIX)
	cstKeys, _ := dbReader.GetLoadedIds(utils.CDR_STATS_PREFIX)
	userKeys, _ := dbReader.GetLoadedIds(utils.USERS_PREFIX)

	// relase tp data
	dbReader.Init()

	if len(aps) != 0 {
		sched := self.ServManager.GetScheduler()
		if sched != nil {
			sched.Reload()
		}
	}

	if len(cstKeys) != 0 && self.CdrStatsSrv != nil {
		var out int
		if err := self.CdrStatsSrv.Call("CDRStatsV1.ReloadQueues", cstKeys, &out); err != nil {
			return err
		}
	}
	if len(userKeys) != 0 && self.Users != nil {
		var r string
		if err := self.Users.Call("AliasV1.ReloadUsers", "", &r); err != nil {
			return err
		}
	}
	*reply = OK
	return nil
}

func (self *ApierV1) ImportTariffPlanFromFolder(attrs utils.AttrImportTPFromFolder, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "FolderPath"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if len(attrs.CsvSeparator) == 0 {
		attrs.CsvSeparator = ","
	}
	if fi, err := os.Stat(attrs.FolderPath); err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return utils.ErrInvalidPath
		}
		return utils.NewErrServerError(err)
	} else if !fi.IsDir() {
		return utils.ErrInvalidPath
	}
	csvImporter := engine.TPCSVImporter{
		TPid:     attrs.TPid,
		StorDb:   self.StorDb,
		DirPath:  attrs.FolderPath,
		Sep:      rune(attrs.CsvSeparator[0]),
		Verbose:  false,
		ImportId: attrs.RunId,
	}
	if err := csvImporter.Run(); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrSetRatingProfile struct {
	Tenant                string                      // Tenant's Id
	Category              string                      // TypeOfRecord
	Direction             string                      // Traffic direction, OUT is the only one supported for now
	Subject               string                      // Rating subject, usually the same as account
	Overwrite             bool                        // Overwrite if exists
	RatingPlanActivations []*utils.TPRatingActivation // Activate rating plans at specific time
}

// Sets a specific rating profile working with data directly in the DataDB without involving storDb
func (self *ApierV1) SetRatingProfile(attrs AttrSetRatingProfile, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "TOR", "Direction", "Subject", "RatingPlanActivations"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, rpa := range attrs.RatingPlanActivations {
		if missing := utils.MissingStructFields(rpa, []string{"ActivationTime", "RatingPlanId"}); len(missing) != 0 {
			return fmt.Errorf("%s:RatingPlanActivation:%v", utils.ErrMandatoryIeMissing.Error(), missing)
		}
	}
	tpRpf := utils.TPRatingProfile{Tenant: attrs.Tenant, Category: attrs.Category, Direction: attrs.Direction, Subject: attrs.Subject}
	keyId := tpRpf.KeyId()
	var rpfl *engine.RatingProfile
	if !attrs.Overwrite {
		if rpfl, err = self.DataManager.GetRatingProfile(keyId, false, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return utils.NewErrServerError(err)
		}
	}
	if rpfl == nil {
		rpfl = &engine.RatingProfile{Id: keyId, RatingPlanActivations: make(engine.RatingPlanActivations, 0)}
	}
	for _, ra := range attrs.RatingPlanActivations {
		at, err := utils.ParseTimeDetectLayout(ra.ActivationTime, self.Config.DefaultTimezone)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("%s:Cannot parse activation time from %v", utils.ErrServerError.Error(), ra.ActivationTime))
		}
		if exists, err := self.DataManager.HasData(utils.RATING_PLAN_PREFIX, ra.RatingPlanId); err != nil {
			return utils.NewErrServerError(err)
		} else if !exists {
			return fmt.Errorf(fmt.Sprintf("%s:RatingPlanId:%s", utils.ErrNotFound.Error(), ra.RatingPlanId))
		}
		rpfl.RatingPlanActivations = append(rpfl.RatingPlanActivations, &engine.RatingPlanActivation{ActivationTime: at, RatingPlanId: ra.RatingPlanId,
			FallbackKeys: utils.FallbackSubjKeys(tpRpf.Direction, tpRpf.Tenant, tpRpf.Category, ra.FallbackSubjects)})
	}
	if err := self.DataManager.SetRatingProfile(rpfl, utils.NonTransactional); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}

// Deprecated attrs
type V1AttrSetActions struct {
	ActionsId string        // Actions id
	Overwrite bool          // If previously defined, will be overwritten
	Actions   []*V1TPAction // Set of actions this Actions profile will perform
}
type V1TPActions struct {
	TPid      string        // Tariff plan id
	ActionsId string        // Actions id
	Actions   []*V1TPAction // Set of actions this Actions profile will perform
}

type V1TPAction struct {
	Identifier      string   // Identifier mapped in the code
	BalanceId       string   // Balance identification string (account scope)
	BalanceUuid     string   // Balance identification string (global scope)
	BalanceType     string   // Type of balance the action will operate on
	Directions      string   // Balance direction
	Units           float64  // Number of units to add/deduct
	ExpiryTime      string   // Time when the units will expire
	Filter          string   // The condition on balances that is checked before the action
	TimingTags      string   // Timing when balance is active
	DestinationIds  string   // Destination profile id
	RatingSubject   string   // Reference a rate subject defined in RatingProfiles
	Categories      string   // category filter for balances
	SharedGroups    string   // Reference to a shared group
	BalanceWeight   *float64 // Balance weight
	ExtraParameters string
	BalanceBlocker  string
	BalanceDisabled string
	Weight          float64 // Action's weight
}

func (self *ApierV1) SetActions(attrs V1AttrSetActions, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"ActionsId", "Actions"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, action := range attrs.Actions {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), action.Identifier, missing)
		}
	}
	if !attrs.Overwrite {
		if exists, err := self.DataManager.HasData(utils.ACTION_PREFIX, attrs.ActionsId); err != nil {
			return utils.NewErrServerError(err)
		} else if exists {
			return utils.ErrExists
		}
	}
	storeActions := make(engine.Actions, len(attrs.Actions))
	for idx, apiAct := range attrs.Actions {
		a := &engine.Action{
			Id:               attrs.ActionsId,
			ActionType:       apiAct.Identifier,
			Weight:           apiAct.Weight,
			ExpirationString: apiAct.ExpiryTime,
			ExtraParameters:  apiAct.ExtraParameters,
			Filter:           apiAct.Filter,
			Balance: &engine.BalanceFilter{ // TODO: update this part
				Uuid:           utils.StringPointer(apiAct.BalanceUuid),
				ID:             utils.StringPointer(apiAct.BalanceId),
				Type:           utils.StringPointer(apiAct.BalanceType),
				Value:          &utils.ValueFormula{Static: apiAct.Units},
				Weight:         apiAct.BalanceWeight,
				Directions:     utils.StringMapPointer(utils.ParseStringMap(apiAct.Directions)),
				DestinationIDs: utils.StringMapPointer(utils.ParseStringMap(apiAct.DestinationIds)),
				RatingSubject:  utils.StringPointer(apiAct.RatingSubject),
				SharedGroups:   utils.StringMapPointer(utils.ParseStringMap(apiAct.SharedGroups)),
			},
		}
		storeActions[idx] = a
	}
	if err := self.DataManager.SetActions(attrs.ActionsId, storeActions, utils.NonTransactional); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}

// Retrieves actions attached to specific ActionsId within cache
func (self *ApierV1) GetActions(actsId string, reply *[]*utils.TPAction) error {
	if len(actsId) == 0 {
		return fmt.Errorf("%s ActionsId: %s", utils.ErrMandatoryIeMissing.Error(), actsId)
	}
	acts := make([]*utils.TPAction, 0)
	engActs, err := self.DataManager.GetActions(actsId, false, utils.NonTransactional)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	for _, engAct := range engActs {
		act := &utils.TPAction{
			Identifier:      engAct.ActionType,
			ExpiryTime:      engAct.ExpirationString,
			ExtraParameters: engAct.ExtraParameters,
			Filter:          engAct.Filter,
			Weight:          engAct.Weight,
		}
		bf := engAct.Balance
		if bf != nil {
			act.BalanceType = bf.GetType()
			act.Units = strconv.FormatFloat(bf.GetValue(), 'f', -1, 64)
			act.Directions = bf.GetDirections().String()
			act.DestinationIds = bf.GetDestinationIDs().String()
			act.RatingSubject = bf.GetRatingSubject()
			act.SharedGroups = bf.GetSharedGroups().String()
			act.BalanceWeight = strconv.FormatFloat(bf.GetWeight(), 'f', -1, 64)
			act.TimingTags = bf.GetTimingIDs().String()
			act.BalanceId = bf.GetID()
			act.Categories = bf.GetCategories().String()
			act.BalanceBlocker = strconv.FormatBool(bf.GetBlocker())
			act.BalanceDisabled = strconv.FormatBool(bf.GetDisabled())
		}
		acts = append(acts, act)
	}
	*reply = acts
	return nil
}

type AttrSetActionPlan struct {
	Id              string            // Profile id
	ActionPlan      []*AttrActionPlan // Set of actions this Actions profile will perform
	Overwrite       bool              // If previously defined, will be overwritten
	ReloadScheduler bool              // Enables automatic reload of the scheduler (eg: useful when adding a single action timing)
}

type AttrActionPlan struct {
	ActionsId string  // Actions id
	Years     string  // semicolon separated list of years this timing is valid on, *any or empty supported
	Months    string  // semicolon separated list of months this timing is valid on, *any or empty supported
	MonthDays string  // semicolon separated list of month's days this timing is valid on, *any or empty supported
	WeekDays  string  // semicolon separated list of week day names this timing is valid on *any or empty supported
	Time      string  // String representing the time this timing starts on, *asap supported
	Weight    float64 // Binding's weight
}

func (self *ApierV1) SetActionPlan(attrs AttrSetActionPlan, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"Id", "ActionPlan"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, at := range attrs.ActionPlan {
		requiredFields := []string{"ActionsId", "Time", "Weight"}
		if missing := utils.MissingStructFields(at, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), at.ActionsId, missing)
		}
	}
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		var prevAccountIDs utils.StringMap
		if prevAP, err := self.DataManager.DataDB().GetActionPlan(attrs.Id, false, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return 0, utils.NewErrServerError(err)
		} else if err == nil && !attrs.Overwrite {
			return 0, utils.ErrExists
		} else if prevAP != nil {
			prevAccountIDs = prevAP.AccountIDs
		}
		ap := &engine.ActionPlan{
			Id: attrs.Id,
		}
		for _, apiAtm := range attrs.ActionPlan {
			if exists, err := self.DataManager.HasData(utils.ACTION_PREFIX, apiAtm.ActionsId); err != nil {
				return 0, utils.NewErrServerError(err)
			} else if !exists {
				return 0, fmt.Errorf("%s:%s", utils.ErrBrokenReference.Error(), apiAtm.ActionsId)
			}
			timing := new(engine.RITiming)
			timing.Years.Parse(apiAtm.Years, ";")
			timing.Months.Parse(apiAtm.Months, ";")
			timing.MonthDays.Parse(apiAtm.MonthDays, ";")
			timing.WeekDays.Parse(apiAtm.WeekDays, ";")
			timing.StartTime = apiAtm.Time
			ap.ActionTimings = append(ap.ActionTimings, &engine.ActionTiming{
				Uuid:      utils.GenUUID(),
				Weight:    apiAtm.Weight,
				Timing:    &engine.RateInterval{Timing: timing},
				ActionsID: apiAtm.ActionsId,
			})
		}
		if err := self.DataManager.DataDB().SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
			return 0, utils.NewErrServerError(err)
		}
		if err = self.DataManager.CacheDataFromDB(utils.ACTION_PLAN_PREFIX, []string{ap.Id}, true); err != nil {
			return 0, utils.NewErrServerError(err)
		}
		for acntID := range prevAccountIDs {
			if err := self.DataManager.DataDB().RemAccountActionPlans(acntID, []string{attrs.Id}); err != nil {
				return 0, utils.NewErrServerError(err)
			}
		}
		if len(prevAccountIDs) != 0 {
			if err = self.DataManager.CacheDataFromDB(utils.AccountActionPlansPrefix, prevAccountIDs.Slice(), true); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return 0, utils.NewErrServerError(err)
			}
		}
		return 0, nil
	}, 0, utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return err
	}
	if attrs.ReloadScheduler {
		sched := self.ServManager.GetScheduler()
		if sched == nil {
			return errors.New(utils.SchedulerNotRunningCaps)
		}
		sched.Reload()
	}
	*reply = OK
	return nil
}

type AttrGetActionPlan struct {
	ID string
}

func (self *ApierV1) GetActionPlan(attr AttrGetActionPlan, reply *[]*engine.ActionPlan) error {
	var result []*engine.ActionPlan
	if attr.ID == "" || attr.ID == "*" {
		aplsMap, err := self.DataManager.DataDB().GetAllActionPlans()
		if err != nil {
			return err
		}
		for _, apls := range aplsMap {
			result = append(result, apls)
		}
	} else {
		apls, err := self.DataManager.DataDB().GetActionPlan(attr.ID, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		result = append(result, apls)
	}
	*reply = result
	return nil
}

// Process dependencies and load a specific AccountActions profile from storDb into dataDb.
func (self *ApierV1) LoadAccountActions(attrs utils.TPAccountActions, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader := engine.NewTpReader(self.DataManager.DataDB(), self.StorDb, attrs.TPid, self.Config.DefaultTimezone)
	if _, err := guardian.Guardian.Guard(func() (interface{}, error) {
		if err := dbReader.LoadAccountActionsFiltered(&attrs); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, attrs.LoadId); err != nil {
		return utils.NewErrServerError(err)
	}
	// ToDo: Get the action keys loaded by dbReader so we reload only these in cache
	// Need to do it before scheduler otherwise actions to run will be unknown
	sched := self.ServManager.GetScheduler()
	if sched != nil {
		sched.Reload()
	}
	*reply = OK
	return nil
}

func (self *ApierV1) ReloadScheduler(ignore string, reply *string) error {
	sched := self.ServManager.GetScheduler()
	if sched == nil {
		return errors.New(utils.SchedulerNotRunningCaps)
	}
	sched.Reload()
	*reply = utils.OK
	return nil
}

func (self *ApierV1) ReloadCache(attrs utils.AttrReloadCache, reply *string) (err error) {
	if attrs.FlushAll {
		cache.Flush()
		return
	}
	// Reload Destinations
	dataIDs := make([]string, 0)
	if attrs.DestinationIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.DestinationIDs) > 0 {
		dataIDs = make([]string, len(*attrs.DestinationIDs))
		for idx, dId := range *attrs.DestinationIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.DESTINATION_PREFIX, dataIDs, true); err != nil {
		return
	}
	// Reload ReverseDestinations
	dataIDs = make([]string, 0)
	if attrs.ReverseDestinationIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.ReverseDestinationIDs) > 0 {
		dataIDs = make([]string, len(*attrs.ReverseDestinationIDs))
		for idx, dId := range *attrs.ReverseDestinationIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.REVERSE_DESTINATION_PREFIX, dataIDs, true); err != nil {
		return
	}
	// RatingPlans
	dataIDs = make([]string, 0)
	if attrs.RatingPlanIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.RatingPlanIDs) > 0 {
		dataIDs = make([]string, len(*attrs.RatingPlanIDs))
		for idx, dId := range *attrs.RatingPlanIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.RATING_PLAN_PREFIX, dataIDs, true); err != nil {
		return
	}
	// RatingProfiles
	dataIDs = make([]string, 0)
	if attrs.RatingProfileIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.RatingProfileIDs) > 0 {
		dataIDs = make([]string, len(*attrs.RatingProfileIDs))
		for idx, dId := range *attrs.RatingProfileIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.RATING_PROFILE_PREFIX, dataIDs, true); err != nil {
		return
	}
	// Actions
	dataIDs = make([]string, 0)
	if attrs.ActionIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.ActionIDs) > 0 {
		dataIDs = make([]string, len(*attrs.ActionIDs))
		for idx, dId := range *attrs.ActionIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.ACTION_PREFIX, dataIDs, true); err != nil {
		return
	}
	// ActionPlans
	dataIDs = make([]string, 0)
	if attrs.ActionPlanIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.ActionPlanIDs) > 0 {
		dataIDs = make([]string, len(*attrs.ActionPlanIDs))
		for idx, dId := range *attrs.ActionPlanIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.ACTION_PLAN_PREFIX, dataIDs, true); err != nil {
		return
	}
	// AccountActionPlans
	dataIDs = make([]string, 0)
	if attrs.AccountActionPlanIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.AccountActionPlanIDs) > 0 {
		dataIDs = make([]string, len(*attrs.AccountActionPlanIDs))
		for idx, dId := range *attrs.AccountActionPlanIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.AccountActionPlansPrefix, dataIDs, true); err != nil {
		return
	}
	// ActionTriggers
	dataIDs = make([]string, 0)
	if attrs.ActionTriggerIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.ActionTriggerIDs) > 0 {
		dataIDs = make([]string, len(*attrs.ActionTriggerIDs))
		for idx, dId := range *attrs.ActionTriggerIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.ACTION_TRIGGER_PREFIX, dataIDs, true); err != nil {
		return
	}
	// SharedGroups
	dataIDs = make([]string, 0)
	if attrs.SharedGroupIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.SharedGroupIDs) > 0 {
		dataIDs = make([]string, len(*attrs.SharedGroupIDs))
		for idx, dId := range *attrs.SharedGroupIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.SHARED_GROUP_PREFIX, dataIDs, true); err != nil {
		return
	}
	// LCR Profiles
	dataIDs = make([]string, 0)
	if attrs.LCRids == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.LCRids) > 0 {
		dataIDs = make([]string, len(*attrs.LCRids))
		for idx, dId := range *attrs.LCRids {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.LCR_PREFIX, dataIDs, true); err != nil {
		return
	}
	// DerivedChargers
	dataIDs = make([]string, 0)
	if attrs.DerivedChargerIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.DerivedChargerIDs) > 0 {
		dataIDs = make([]string, len(*attrs.DerivedChargerIDs))
		for idx, dId := range *attrs.DerivedChargerIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.DERIVEDCHARGERS_PREFIX, dataIDs, true); err != nil {
		return
	}
	// Aliases
	dataIDs = make([]string, 0)
	if attrs.AliasIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.AliasIDs) > 0 {
		dataIDs = make([]string, len(*attrs.AliasIDs))
		for idx, dId := range *attrs.AliasIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.ALIASES_PREFIX, dataIDs, true); err != nil {
		return
	}
	// ReverseAliases
	dataIDs = make([]string, 0)
	if attrs.ReverseAliasIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.ReverseAliasIDs) > 0 {
		dataIDs = make([]string, len(*attrs.ReverseAliasIDs))
		for idx, dId := range *attrs.ReverseAliasIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.REVERSE_ALIASES_PREFIX, dataIDs, true); err != nil {
		return
	}
	// ResourceProfiles
	dataIDs = make([]string, 0)
	if attrs.ResourceProfileIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.ResourceProfileIDs) > 0 {
		dataIDs = make([]string, len(*attrs.ResourceProfileIDs))
		for idx, dId := range *attrs.ResourceProfileIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.ResourceProfilesPrefix, dataIDs, true); err != nil {
		return
	}
	// Resources
	dataIDs = make([]string, 0)
	if attrs.ResourceIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.ResourceIDs) > 0 {
		dataIDs = make([]string, len(*attrs.ResourceIDs))
		for idx, dId := range *attrs.ResourceIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.ResourcesPrefix, dataIDs, true); err != nil {
		return
	}
	// StatQueues
	dataIDs = make([]string, 0)
	if attrs.StatsQueueIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.StatsQueueIDs) > 0 {
		dataIDs = make([]string, len(*attrs.StatsQueueIDs))
		for idx, dId := range *attrs.StatsQueueIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.StatQueuePrefix, dataIDs, true); err != nil {
		return
	}
	// StatQueueProfiles
	dataIDs = make([]string, 0)
	if attrs.StatsQueueProfileIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.StatsQueueProfileIDs) > 0 {
		dataIDs = make([]string, len(*attrs.StatsQueueProfileIDs))
		for idx, dId := range *attrs.StatsQueueProfileIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.StatQueueProfilePrefix, dataIDs, true); err != nil {
		return
	}
	// Thresholds
	dataIDs = make([]string, 0)
	if attrs.ThresholdIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.ThresholdIDs) > 0 {
		dataIDs = make([]string, len(*attrs.ThresholdIDs))
		for idx, dId := range *attrs.ThresholdIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.ThresholdPrefix, dataIDs, true); err != nil {
		return
	}
	// ThresholdProfiles
	dataIDs = make([]string, 0)
	if attrs.ThresholdProfileIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.ThresholdProfileIDs) > 0 {
		dataIDs = make([]string, len(*attrs.ThresholdProfileIDs))
		for idx, dId := range *attrs.ThresholdProfileIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.ThresholdProfilePrefix, dataIDs, true); err != nil {
		return
	}
	// Filters
	dataIDs = make([]string, 0)
	if attrs.FilterIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.FilterIDs) > 0 {
		dataIDs = make([]string, len(*attrs.FilterIDs))
		for idx, dId := range *attrs.FilterIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.FilterPrefix, dataIDs, true); err != nil {
		return
	}
	// SupplierProfile
	dataIDs = make([]string, 0)
	if attrs.SupplierProfileIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.SupplierProfileIDs) > 0 {
		dataIDs = make([]string, len(*attrs.SupplierProfileIDs))
		for idx, dId := range *attrs.SupplierProfileIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.SupplierProfilePrefix, dataIDs, true); err != nil {
		return
	}
	// AttributeProfile
	dataIDs = make([]string, 0)
	if attrs.AttributeProfileIDs == nil {
		dataIDs = nil // Reload all
	} else if len(*attrs.AttributeProfileIDs) > 0 {
		dataIDs = make([]string, len(*attrs.AttributeProfileIDs))
		for idx, dId := range *attrs.AttributeProfileIDs {
			dataIDs[idx] = dId
		}
	}
	if err = self.DataManager.CacheDataFromDB(utils.AttributeProfilePrefix, dataIDs, true); err != nil {
		return
	}

	*reply = utils.OK
	return nil
}

func (self *ApierV1) LoadCache(args utils.AttrReloadCache, reply *string) (err error) {
	if args.FlushAll {
		cache.Flush()
	}
	var dstIDs, rvDstIDs, rplIDs, rpfIDs, actIDs, aplIDs, aapIDs, atrgIDs, sgIDs, lcrIDs, dcIDs, alsIDs, rvAlsIDs, rspIDs, resIDs, stqIDs, stqpIDs, thIDs, thpIDs, fltrIDs, splpIDs, alsPrfIDs []string
	if args.DestinationIDs == nil {
		dstIDs = nil
	} else {
		dstIDs = *args.DestinationIDs
	}
	if args.ReverseDestinationIDs == nil {
		rvDstIDs = nil
	} else {
		rvDstIDs = *args.ReverseDestinationIDs
	}
	if args.RatingPlanIDs == nil {
		rplIDs = nil
	} else {
		rplIDs = *args.RatingPlanIDs
	}
	if args.RatingProfileIDs == nil {
		rpfIDs = nil
	} else {
		rpfIDs = *args.RatingProfileIDs
	}
	if args.ActionIDs == nil {
		actIDs = nil
	} else {
		actIDs = *args.ActionIDs
	}
	if args.ActionPlanIDs == nil {
		aplIDs = nil
	} else {
		aplIDs = *args.ActionPlanIDs
	}
	if args.AccountActionPlanIDs == nil {
		aapIDs = nil
	} else {
		aapIDs = *args.AccountActionPlanIDs
	}
	if args.ActionTriggerIDs == nil {
		atrgIDs = nil
	} else {
		atrgIDs = *args.ActionTriggerIDs
	}
	if args.SharedGroupIDs == nil {
		sgIDs = nil
	} else {
		sgIDs = *args.SharedGroupIDs
	}
	if args.LCRids == nil {
		lcrIDs = nil
	} else {
		lcrIDs = *args.LCRids
	}
	if args.DerivedChargerIDs == nil {
		dcIDs = nil
	} else {
		dcIDs = *args.DerivedChargerIDs
	}
	if args.AliasIDs == nil {
		alsIDs = nil
	} else {
		alsIDs = *args.AliasIDs
	}
	if args.ReverseAliasIDs == nil {
		rvAlsIDs = nil
	} else {
		rvAlsIDs = *args.ReverseAliasIDs
	}
	if args.ResourceProfileIDs == nil {
		rspIDs = nil
	} else {
		rspIDs = *args.ResourceProfileIDs
	}
	if args.ResourceIDs == nil {
		resIDs = nil
	} else {
		resIDs = *args.ResourceIDs
	}
	if args.StatsQueueIDs == nil {
		stqIDs = nil
	} else {
		stqIDs = *args.StatsQueueIDs
	}
	if args.StatsQueueProfileIDs == nil {
		stqpIDs = nil
	} else {
		stqpIDs = *args.StatsQueueProfileIDs
	}
	if args.ThresholdIDs == nil {
		thIDs = nil
	} else {
		thIDs = *args.ThresholdIDs
	}
	if args.ThresholdProfileIDs == nil {
		thpIDs = nil
	} else {
		thpIDs = *args.ThresholdProfileIDs
	}
	if args.FilterIDs == nil {
		fltrIDs = nil
	} else {
		fltrIDs = *args.FilterIDs
	}
	if args.SupplierProfileIDs == nil {
		splpIDs = nil
	} else {
		splpIDs = *args.SupplierProfileIDs
	}
	if args.AttributeProfileIDs == nil {
		alsPrfIDs = nil
	} else {
		alsPrfIDs = *args.AttributeProfileIDs
	}
	if err := self.DataManager.LoadDataDBCache(dstIDs, rvDstIDs, rplIDs, rpfIDs, actIDs, aplIDs, aapIDs, atrgIDs, sgIDs, lcrIDs, dcIDs, alsIDs, rvAlsIDs, rspIDs, resIDs, stqIDs, stqpIDs, thIDs, thpIDs, fltrIDs, splpIDs, alsPrfIDs); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// FlushCache wipes out cache for a prefix or completely
func (self *ApierV1) FlushCache(args utils.AttrReloadCache, reply *string) (err error) {
	if args.FlushAll {
		cache.Flush()
		*reply = utils.OK
		return
	}
	if args.DestinationIDs == nil {
		cache.RemPrefixKey(utils.DESTINATION_PREFIX, true, utils.NonTransactional)
	} else if len(*args.DestinationIDs) != 0 {
		for _, key := range *args.DestinationIDs {
			cache.RemKey(utils.DESTINATION_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.ReverseDestinationIDs == nil {
		cache.RemPrefixKey(utils.REVERSE_DESTINATION_PREFIX, true, utils.NonTransactional)
	} else if len(*args.ReverseDestinationIDs) != 0 {
		for _, key := range *args.ReverseDestinationIDs {
			cache.RemKey(utils.REVERSE_DESTINATION_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.RatingPlanIDs == nil {
		cache.RemPrefixKey(utils.RATING_PLAN_PREFIX, true, utils.NonTransactional)
	} else if len(*args.RatingPlanIDs) != 0 {
		for _, key := range *args.RatingPlanIDs {
			cache.RemKey(utils.RATING_PLAN_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.RatingProfileIDs == nil {
		cache.RemPrefixKey(utils.RATING_PROFILE_PREFIX, true, utils.NonTransactional)
	} else if len(*args.RatingProfileIDs) != 0 {
		for _, key := range *args.RatingProfileIDs {
			cache.RemKey(utils.RATING_PROFILE_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.ActionIDs == nil {
		cache.RemPrefixKey(utils.ACTION_PREFIX, true, utils.NonTransactional)
	} else if len(*args.ActionIDs) != 0 {
		for _, key := range *args.ActionIDs {
			cache.RemKey(utils.ACTION_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.ActionPlanIDs == nil {
		cache.RemPrefixKey(utils.ACTION_PLAN_PREFIX, true, utils.NonTransactional)
	} else if len(*args.ActionPlanIDs) != 0 {
		for _, key := range *args.ActionPlanIDs {
			cache.RemKey(utils.ACTION_PLAN_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.ActionTriggerIDs == nil {
		cache.RemPrefixKey(utils.ACTION_TRIGGER_PREFIX, true, utils.NonTransactional)
	} else if len(*args.ActionTriggerIDs) != 0 {
		for _, key := range *args.ActionTriggerIDs {
			cache.RemKey(utils.ACTION_TRIGGER_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.SharedGroupIDs == nil {
		cache.RemPrefixKey(utils.SHARED_GROUP_PREFIX, true, utils.NonTransactional)
	} else if len(*args.SharedGroupIDs) != 0 {
		for _, key := range *args.SharedGroupIDs {
			cache.RemKey(utils.SHARED_GROUP_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.LCRids == nil {
		cache.RemPrefixKey(utils.LCR_PREFIX, true, utils.NonTransactional)
	} else if len(*args.LCRids) != 0 {
		for _, key := range *args.LCRids {
			cache.RemKey(utils.LCR_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.DerivedChargerIDs == nil {
		cache.RemPrefixKey(utils.DERIVEDCHARGERS_PREFIX, true, utils.NonTransactional)
	} else if len(*args.DerivedChargerIDs) != 0 {
		for _, key := range *args.DerivedChargerIDs {
			cache.RemKey(utils.DERIVEDCHARGERS_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.AliasIDs == nil {
		cache.RemPrefixKey(utils.ALIASES_PREFIX, true, utils.NonTransactional)
	} else if len(*args.AliasIDs) != 0 {
		for _, key := range *args.AliasIDs {
			cache.RemKey(utils.ALIASES_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.ReverseAliasIDs == nil {
		cache.RemPrefixKey(utils.REVERSE_ALIASES_PREFIX, true, utils.NonTransactional)
	} else if len(*args.ReverseAliasIDs) != 0 {
		for _, key := range *args.ReverseAliasIDs {
			cache.RemKey(utils.REVERSE_ALIASES_PREFIX+key, true, utils.NonTransactional)
		}
	}
	if args.ResourceProfileIDs == nil {
		cache.RemPrefixKey(utils.ResourceProfilesPrefix, true, utils.NonTransactional)
	} else if len(*args.ResourceProfileIDs) != 0 {
		for _, key := range *args.ResourceProfileIDs {
			cache.RemKey(utils.ResourceProfilesPrefix+key, true, utils.NonTransactional)
		}
	}
	if args.ResourceIDs == nil {
		cache.RemPrefixKey(utils.ResourcesPrefix, true, utils.NonTransactional)
	} else if len(*args.ResourceIDs) != 0 {
		for _, key := range *args.ResourceIDs {
			cache.RemKey(utils.ResourcesPrefix+key, true, utils.NonTransactional)
		}
	}
	if args.StatsQueueIDs == nil {
		cache.RemPrefixKey(utils.StatQueuePrefix, true, utils.NonTransactional)
	} else if len(*args.StatsQueueIDs) != 0 {
		for _, key := range *args.StatsQueueIDs {
			cache.RemKey(utils.StatQueuePrefix+key, true, utils.NonTransactional)
		}
	}
	if args.StatsQueueProfileIDs == nil {
		cache.RemPrefixKey(utils.StatQueueProfilePrefix, true, utils.NonTransactional)
	} else if len(*args.StatsQueueProfileIDs) != 0 {
		for _, key := range *args.StatsQueueProfileIDs {
			cache.RemKey(utils.StatQueueProfilePrefix+key, true, utils.NonTransactional)
		}
	}
	if args.ThresholdIDs == nil {
		cache.RemPrefixKey(utils.ThresholdPrefix, true, utils.NonTransactional)
	} else if len(*args.ThresholdIDs) != 0 {
		for _, key := range *args.ThresholdProfileIDs {
			cache.RemKey(utils.ThresholdPrefix+key, true, utils.NonTransactional)
		}
	}
	if args.ThresholdProfileIDs == nil {
		cache.RemPrefixKey(utils.ThresholdProfilePrefix, true, utils.NonTransactional)
	} else if len(*args.ThresholdProfileIDs) != 0 {
		for _, key := range *args.ThresholdProfileIDs {
			cache.RemKey(utils.ThresholdProfilePrefix+key, true, utils.NonTransactional)
		}
	}
	if args.FilterIDs == nil {
		cache.RemPrefixKey(utils.FilterPrefix, true, utils.NonTransactional)
	} else if len(*args.FilterIDs) != 0 {
		for _, key := range *args.FilterIDs {
			cache.RemKey(utils.FilterPrefix+key, true, utils.NonTransactional)
		}
	}
	if args.SupplierProfileIDs == nil {
		cache.RemPrefixKey(utils.SupplierProfilePrefix, true, utils.NonTransactional)
	} else if len(*args.SupplierProfileIDs) != 0 {
		for _, key := range *args.SupplierProfileIDs {
			cache.RemKey(utils.SupplierProfilePrefix+key, true, utils.NonTransactional)
		}
	}
	if args.AttributeProfileIDs == nil {
		cache.RemPrefixKey(utils.AttributeProfilePrefix, true, utils.NonTransactional)
	} else if len(*args.AttributeProfileIDs) != 0 {
		for _, key := range *args.AttributeProfileIDs {
			cache.RemKey(utils.AttributeProfilePrefix+key, true, utils.NonTransactional)
		}
	}

	*reply = utils.OK
	return
}

func (self *ApierV1) GetCacheStats(attrs utils.AttrCacheStats, reply *utils.CacheStats) error {
	cs := new(utils.CacheStats)
	cs.Destinations = cache.CountEntries(utils.DESTINATION_PREFIX)
	cs.ReverseDestinations = cache.CountEntries(utils.REVERSE_DESTINATION_PREFIX)
	cs.RatingPlans = cache.CountEntries(utils.RATING_PLAN_PREFIX)
	cs.RatingProfiles = cache.CountEntries(utils.RATING_PROFILE_PREFIX)
	cs.Actions = cache.CountEntries(utils.ACTION_PREFIX)
	cs.ActionPlans = cache.CountEntries(utils.ACTION_PLAN_PREFIX)
	cs.AccountActionPlans = cache.CountEntries(utils.AccountActionPlansPrefix)
	cs.SharedGroups = cache.CountEntries(utils.SHARED_GROUP_PREFIX)
	cs.DerivedChargers = cache.CountEntries(utils.DERIVEDCHARGERS_PREFIX)
	cs.LcrProfiles = cache.CountEntries(utils.LCR_PREFIX)
	cs.Aliases = cache.CountEntries(utils.ALIASES_PREFIX)
	cs.ReverseAliases = cache.CountEntries(utils.REVERSE_ALIASES_PREFIX)
	cs.ResourceProfiles = cache.CountEntries(utils.ResourceProfilesPrefix)
	cs.Resources = cache.CountEntries(utils.ResourcesPrefix)
	cs.StatQueues = cache.CountEntries(utils.StatQueuePrefix)
	cs.StatQueueProfiles = cache.CountEntries(utils.StatQueueProfilePrefix)
	cs.Thresholds = cache.CountEntries(utils.ThresholdPrefix)
	cs.ThresholdProfiles = cache.CountEntries(utils.ThresholdProfilePrefix)
	cs.Filters = cache.CountEntries(utils.FilterPrefix)
	cs.SupplierProfiles = cache.CountEntries(utils.SupplierProfilePrefix)
	cs.AttributeProfiles = cache.CountEntries(utils.AttributeProfilePrefix)

	if self.CdrStatsSrv != nil {
		var queueIds []string
		if err := self.CdrStatsSrv.Call("CDRStatsV1.GetQueueIds", 0, &queueIds); err != nil {
			return utils.NewErrServerError(err)
		}
		cs.CdrStats = len(queueIds)
	}
	if self.Users != nil {
		var ups engine.UserProfiles
		if err := self.Users.Call("UsersV1.GetUsers", &engine.UserProfile{}, &ups); err != nil {
			return utils.NewErrServerError(err)
		}
		cs.Users = len(ups)
	}
	*reply = *cs
	return nil
}

// GetCacheKeys returns a list of keys available in cache based on query arguments
// If keys are provided in arguments, they will be checked for existence
func (v1 *ApierV1) GetCacheKeys(args utils.ArgsCacheKeys, reply *utils.ArgsCache) (err error) {
	if args.DestinationIDs != nil {
		var ids []string
		if len(*args.DestinationIDs) != 0 {
			for _, id := range *args.DestinationIDs {
				if _, hasIt := cache.Get(utils.DESTINATION_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.DESTINATION_PREFIX) {
				ids = append(ids, id[len(utils.DESTINATION_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.DestinationIDs = &ids
		}
	}
	if args.ReverseDestinationIDs != nil {
		var ids []string
		if len(*args.ReverseDestinationIDs) != 0 {
			for _, id := range *args.ReverseDestinationIDs {
				if _, hasIt := cache.Get(utils.REVERSE_DESTINATION_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.REVERSE_DESTINATION_PREFIX) {
				ids = append(ids, id[len(utils.REVERSE_DESTINATION_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.ReverseDestinationIDs = &ids
		}
	}
	if args.RatingPlanIDs != nil {
		var ids []string
		if len(*args.RatingPlanIDs) != 0 {
			for _, id := range *args.RatingPlanIDs {
				if _, hasIt := cache.Get(utils.RATING_PLAN_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.RATING_PLAN_PREFIX) {
				ids = append(ids, id[len(utils.RATING_PLAN_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.RatingPlanIDs = &ids
		}
	}
	if args.RatingProfileIDs != nil {
		var ids []string
		if len(*args.RatingProfileIDs) != 0 {
			for _, id := range *args.RatingProfileIDs {
				if _, hasIt := cache.Get(utils.RATING_PROFILE_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.RATING_PROFILE_PREFIX) {
				ids = append(ids, id[len(utils.RATING_PROFILE_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.RatingProfileIDs = &ids
		}
	}
	if args.ActionIDs != nil {
		var ids []string
		if len(*args.ActionIDs) != 0 {
			for _, id := range *args.ActionIDs {
				if _, hasIt := cache.Get(utils.ACTION_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.ACTION_PREFIX) {
				ids = append(ids, id[len(utils.ACTION_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.ActionIDs = &ids
		}
	}
	if args.ActionPlanIDs != nil {
		var ids []string
		if len(*args.ActionPlanIDs) != 0 {
			for _, id := range *args.ActionPlanIDs {
				if _, hasIt := cache.Get(utils.ACTION_PLAN_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.ACTION_PLAN_PREFIX) {
				ids = append(ids, id[len(utils.ACTION_PLAN_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.ActionPlanIDs = &ids
		}
	}

	if args.AccountActionPlanIDs != nil {
		var ids []string
		if len(*args.AccountActionPlanIDs) != 0 {
			for _, id := range *args.AccountActionPlanIDs {
				if _, hasIt := cache.Get(utils.AccountActionPlansPrefix + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.AccountActionPlansPrefix) {
				ids = append(ids, id[len(utils.AccountActionPlansPrefix):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.AccountActionPlanIDs = &ids
		}
	}
	if args.ActionTriggerIDs != nil {
		var ids []string
		if len(*args.ActionTriggerIDs) != 0 {
			for _, id := range *args.ActionTriggerIDs {
				if _, hasIt := cache.Get(utils.ACTION_TRIGGER_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.ACTION_TRIGGER_PREFIX) {
				ids = append(ids, id[len(utils.ACTION_TRIGGER_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.ActionTriggerIDs = &ids
		}
	}
	if args.SharedGroupIDs != nil {
		var ids []string
		if len(*args.SharedGroupIDs) != 0 {
			for _, id := range *args.SharedGroupIDs {
				if _, hasIt := cache.Get(utils.SHARED_GROUP_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.SHARED_GROUP_PREFIX) {
				ids = append(ids, id[len(utils.SHARED_GROUP_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.SharedGroupIDs = &ids
		}
	}
	if args.LCRids != nil {
		var ids []string
		if len(*args.LCRids) != 0 {
			for _, id := range *args.LCRids {
				if _, hasIt := cache.Get(utils.LCR_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.LCR_PREFIX) {
				ids = append(ids, id[len(utils.LCR_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.LCRids = &ids
		}
	}
	if args.DerivedChargerIDs != nil {
		var ids []string
		if len(*args.DerivedChargerIDs) != 0 {
			for _, id := range *args.DerivedChargerIDs {
				if _, hasIt := cache.Get(utils.DERIVEDCHARGERS_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.DERIVEDCHARGERS_PREFIX) {
				ids = append(ids, id[len(utils.DERIVEDCHARGERS_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.DerivedChargerIDs = &ids
		}
	}
	if args.AliasIDs != nil {
		var ids []string
		if len(*args.AliasIDs) != 0 {
			for _, id := range *args.AliasIDs {
				if _, hasIt := cache.Get(utils.ALIASES_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.ALIASES_PREFIX) {
				ids = append(ids, id[len(utils.ALIASES_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.AliasIDs = &ids
		}
	}
	if args.ReverseAliasIDs != nil {
		var ids []string
		if len(*args.ReverseAliasIDs) != 0 {
			for _, id := range *args.ReverseAliasIDs {
				if _, hasIt := cache.Get(utils.REVERSE_ALIASES_PREFIX + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.REVERSE_ALIASES_PREFIX) {
				ids = append(ids, id[len(utils.REVERSE_ALIASES_PREFIX):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.ReverseAliasIDs = &ids
		}
	}
	if args.ResourceProfileIDs != nil {
		var ids []string
		if len(*args.ResourceProfileIDs) != 0 {
			for _, id := range *args.ResourceProfileIDs {
				if _, hasIt := cache.Get(utils.ResourceProfilesPrefix + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.ResourceProfilesPrefix) {
				ids = append(ids, id[len(utils.ResourceProfilesPrefix):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.ResourceProfileIDs = &ids
		}
	}
	if args.ResourceIDs != nil {
		var ids []string
		if len(*args.ResourceIDs) != 0 {
			for _, id := range *args.ResourceIDs {
				if _, hasIt := cache.Get(utils.ResourcesPrefix + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.ResourcesPrefix) {
				ids = append(ids, id[len(utils.ResourcesPrefix):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.ResourceIDs = &ids
		}
	}

	if args.StatsQueueIDs != nil {
		var ids []string
		if len(*args.StatsQueueIDs) != 0 {
			for _, id := range *args.StatsQueueIDs {
				if _, hasIt := cache.Get(utils.StatQueuePrefix + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.StatQueuePrefix) {
				ids = append(ids, id[len(utils.StatQueuePrefix):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.StatsQueueIDs = &ids
		}
	}

	if args.StatsQueueProfileIDs != nil {
		var ids []string
		if len(*args.StatsQueueProfileIDs) != 0 {
			for _, id := range *args.StatsQueueProfileIDs {
				if _, hasIt := cache.Get(utils.StatQueueProfilePrefix + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.StatQueueProfilePrefix) {
				ids = append(ids, id[len(utils.StatQueueProfilePrefix):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.StatsQueueProfileIDs = &ids
		}
	}

	if args.ThresholdIDs != nil {
		var ids []string
		if len(*args.ThresholdIDs) != 0 {
			for _, id := range *args.ThresholdIDs {
				if _, hasIt := cache.Get(utils.ThresholdPrefix + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.ThresholdPrefix) {
				ids = append(ids, id[len(utils.ThresholdPrefix):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.ThresholdIDs = &ids
		}
	}

	if args.ThresholdProfileIDs != nil {
		var ids []string
		if len(*args.ThresholdProfileIDs) != 0 {
			for _, id := range *args.ThresholdProfileIDs {
				if _, hasIt := cache.Get(utils.ThresholdProfilePrefix + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.ThresholdProfilePrefix) {
				ids = append(ids, id[len(utils.ThresholdProfilePrefix):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.ThresholdProfileIDs = &ids
		}
	}

	if args.FilterIDs != nil {
		var ids []string
		if len(*args.FilterIDs) != 0 {
			for _, id := range *args.FilterIDs {
				if _, hasIt := cache.Get(utils.FilterPrefix + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.FilterPrefix) {
				ids = append(ids, id[len(utils.FilterPrefix):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.FilterIDs = &ids
		}
	}

	if args.SupplierProfileIDs != nil {
		var ids []string
		if len(*args.SupplierProfileIDs) != 0 {
			for _, id := range *args.SupplierProfileIDs {
				if _, hasIt := cache.Get(utils.SupplierProfilePrefix + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.SupplierProfilePrefix) {
				ids = append(ids, id[len(utils.SupplierProfilePrefix):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.SupplierProfileIDs = &ids
		}
	}

	if args.AttributeProfileIDs != nil {
		var ids []string
		if len(*args.AttributeProfileIDs) != 0 {
			for _, id := range *args.AttributeProfileIDs {
				if _, hasIt := cache.Get(utils.AttributeProfilePrefix + id); hasIt {
					ids = append(ids, id)
				}
			}
		} else {
			for _, id := range cache.GetEntryKeys(utils.AttributeProfilePrefix) {
				ids = append(ids, id[len(utils.AttributeProfilePrefix):])
			}
		}
		ids = args.Paginator.PaginateStringSlice(ids)
		if len(ids) != 0 {
			reply.AttributeProfileIDs = &ids
		}
	}

	return
}

func (self *ApierV1) LoadTariffPlanFromFolder(attrs utils.AttrLoadTpFromFolder, reply *string) error {
	if len(attrs.FolderPath) == 0 {
		return fmt.Errorf("%s:%s", utils.ErrMandatoryIeMissing.Error(), "FolderPath")
	}
	if fi, err := os.Stat(attrs.FolderPath); err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return utils.ErrInvalidPath
		}
		return utils.NewErrServerError(err)
	} else if !fi.IsDir() {
		return utils.ErrInvalidPath
	}
	loader := engine.NewTpReader(self.DataManager.DataDB(),
		engine.NewFileCSVStorage(utils.CSV_SEP,
			path.Join(attrs.FolderPath, utils.DESTINATIONS_CSV),
			path.Join(attrs.FolderPath, utils.TIMINGS_CSV),
			path.Join(attrs.FolderPath, utils.RATES_CSV),
			path.Join(attrs.FolderPath, utils.DESTINATION_RATES_CSV),
			path.Join(attrs.FolderPath, utils.RATING_PLANS_CSV),
			path.Join(attrs.FolderPath, utils.RATING_PROFILES_CSV),
			path.Join(attrs.FolderPath, utils.SHARED_GROUPS_CSV),
			path.Join(attrs.FolderPath, utils.LCRS_CSV),
			path.Join(attrs.FolderPath, utils.ACTIONS_CSV),
			path.Join(attrs.FolderPath, utils.ACTION_PLANS_CSV),
			path.Join(attrs.FolderPath, utils.ACTION_TRIGGERS_CSV),
			path.Join(attrs.FolderPath, utils.ACCOUNT_ACTIONS_CSV),
			path.Join(attrs.FolderPath, utils.DERIVED_CHARGERS_CSV),
			path.Join(attrs.FolderPath, utils.CDR_STATS_CSV),
			path.Join(attrs.FolderPath, utils.USERS_CSV),
			path.Join(attrs.FolderPath, utils.ALIASES_CSV),
			path.Join(attrs.FolderPath, utils.ResourcesCsv),
			path.Join(attrs.FolderPath, utils.StatsCsv),
			path.Join(attrs.FolderPath, utils.ThresholdsCsv),
			path.Join(attrs.FolderPath, utils.FiltersCsv),
			path.Join(attrs.FolderPath, utils.SuppliersCsv),
			path.Join(attrs.FolderPath, utils.AttributesCsv),
		), "", self.Config.DefaultTimezone)
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.DryRun {
		*reply = OK
		return nil // Mission complete, no errors
	}

	if attrs.Validate {
		if !loader.IsValid() {
			return errors.New("invalid data")
		}
	}

	if err := loader.WriteToDatabase(attrs.FlushDb, false, false); err != nil {
		return utils.NewErrServerError(err)
	}
	utils.Logger.Info("ApierV1.LoadTariffPlanFromFolder, reloading cache.")
	for _, prfx := range []string{
		utils.DESTINATION_PREFIX,
		utils.REVERSE_DESTINATION_PREFIX,
		utils.ACTION_PLAN_PREFIX,
		utils.AccountActionPlansPrefix,
		utils.DERIVEDCHARGERS_PREFIX,
		utils.ALIASES_PREFIX,
		utils.REVERSE_ALIASES_PREFIX} {
		loadedIDs, _ := loader.GetLoadedIds(prfx)
		if err := self.DataManager.CacheDataFromDB(prfx, loadedIDs, true); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	aps, _ := loader.GetLoadedIds(utils.ACTION_PLAN_PREFIX)
	cstKeys, _ := loader.GetLoadedIds(utils.CDR_STATS_PREFIX)
	userKeys, _ := loader.GetLoadedIds(utils.USERS_PREFIX)

	// relase tp data
	loader.Init()

	if len(aps) != 0 {
		sched := self.ServManager.GetScheduler()
		if sched != nil {
			utils.Logger.Info("ApierV1.LoadTariffPlanFromFolder, reloading scheduler.")
			sched.Reload()
		}
	}
	if len(cstKeys) != 0 && self.CdrStatsSrv != nil {
		var out int
		if err := self.CdrStatsSrv.Call("CDRStatsV1.ReloadQueues", cstKeys, &out); err != nil {
			return err
		}
	}
	if len(userKeys) != 0 && self.Users != nil {
		var r string
		if err := self.Users.Call("UsersV1.ReloadUsers", "", &r); err != nil {
			return err
		}
	}
	*reply = utils.OK
	return nil
}

type AttrRemoveRatingProfile struct {
	Direction string
	Tenant    string
	Category  string
	Subject   string
}

func (arrp *AttrRemoveRatingProfile) GetId() (result string) {
	if arrp.Direction != "" && arrp.Direction != utils.ANY {
		result += arrp.Direction
		result += utils.CONCATENATED_KEY_SEP
	} else {
		return
	}
	if arrp.Tenant != "" && arrp.Tenant != utils.ANY {
		result += arrp.Tenant
		result += utils.CONCATENATED_KEY_SEP
	} else {
		return
	}

	if arrp.Category != "" && arrp.Category != utils.ANY {
		result += arrp.Category
		result += utils.CONCATENATED_KEY_SEP
	} else {
		return
	}
	if arrp.Subject != "" && arrp.Subject != utils.ANY {
		result += arrp.Subject
	}
	return
}

func (self *ApierV1) RemoveRatingProfile(attr AttrRemoveRatingProfile, reply *string) error {
	if attr.Direction == "" {
		attr.Direction = utils.OUT
	}
	if (attr.Subject != "" && utils.IsSliceMember([]string{attr.Direction, attr.Tenant, attr.Category}, "")) ||
		(attr.Category != "" && utils.IsSliceMember([]string{attr.Direction, attr.Tenant}, "")) ||
		attr.Tenant != "" && attr.Direction == "" {
		return utils.ErrMandatoryIeMissing
	}
	_, err := guardian.Guardian.Guard(func() (interface{}, error) {
		err := self.DataManager.RemoveRatingProfile(attr.GetId(), utils.NonTransactional)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, "RemoveRatingProfile")
	if err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (self *ApierV1) GetLoadHistory(attrs utils.Paginator, reply *[]*utils.LoadInstance) error {
	nrItems := -1
	offset := 0
	if attrs.Offset != nil { // For offset we need full data
		offset = *attrs.Offset
	} else if attrs.Limit != nil {
		nrItems = *attrs.Limit
	}
	loadHist, err := self.DataManager.DataDB().GetLoadHistory(nrItems, true, utils.NonTransactional)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.Offset != nil && attrs.Limit != nil { // Limit back to original
		nrItems = *attrs.Limit
	}
	if len(loadHist) == 0 || len(loadHist) <= offset || nrItems == 0 {
		return utils.ErrNotFound
	}
	if offset != 0 {
		nrItems = offset + nrItems
	}
	if nrItems == -1 || nrItems > len(loadHist) { // So we can use it in indexing bellow
		nrItems = len(loadHist)
	}
	*reply = loadHist[offset:nrItems]
	return nil
}

type AttrRemoveActions struct {
	ActionIDs []string
}

func (self *ApierV1) RemoveActions(attr AttrRemoveActions, reply *string) error {
	if attr.ActionIDs == nil {
		err := utils.ErrNotFound
		*reply = err.Error()
		return err
	}
	// The check could lead to very long execution time. So we decided to leave it at the user's risck.'
	/*
		stringMap := utils.NewStringMap(attr.ActionIDs...)
		keys, err := self.DataManager.DataDB().GetKeysForPrefix(utils.ACTION_TRIGGER_PREFIX, true)
		if err != nil {
			*reply = err.Error()
			return err
		}
		for _, key := range keys {
			getAttrs, err := self.DataManager.DataDB().GetActionTriggers(key[len(utils.ACTION_TRIGGER_PREFIX):])
			if err != nil {
				*reply = err.Error()
				return err
			}
			for _, atr := range getAttrs {
				if _, found := stringMap[atr.ActionsID]; found {
					// found action trigger referencing action; abort
					err := fmt.Errorf("action %s refenced by action trigger %s", atr.ActionsID, atr.ID)
					*reply = err.Error()
					return err
				}
			}
		}
		allAplsMap, err := self.DataManager.DataDB().GetAllActionPlans()
		if err != nil && err != utils.ErrNotFound {
			*reply = err.Error()
			return err
		}
		for _, apl := range allAplsMap {
			for _, atm := range apl.ActionTimings {
				if _, found := stringMap[atm.ActionsID]; found {
					err := fmt.Errorf("action %s refenced by action plan %s", atm.ActionsID, apl.Id)
					*reply = err.Error()
					return err
				}
			}
		}
	*/
	for _, aID := range attr.ActionIDs {
		if err := self.DataManager.RemoveActions(aID, utils.NonTransactional); err != nil {
			*reply = err.Error()
			return err
		}
	}
	*reply = utils.OK
	return nil
}

type AttrRemoteLock struct {
	LockIDs []string      // List of IDs to obtain lock for
	Timeout time.Duration // Automatically unlock on timeout
}

func (self *ApierV1) RemoteLock(attr AttrRemoteLock, reply *string) error {
	guardian.Guardian.GuardIDs(attr.Timeout, attr.LockIDs...)
	*reply = utils.OK
	return nil
}

func (self *ApierV1) RemoteUnlock(lockIDs []string, reply *string) error {
	guardian.Guardian.UnguardIDs(lockIDs...)
	*reply = utils.OK
	return nil
}

func (v1 *ApierV1) StartService(args servmanager.ArgStartService, reply *string) (err error) {
	return v1.ServManager.V1StartService(args, reply)
}

func (v1 *ApierV1) StopService(args servmanager.ArgStartService, reply *string) (err error) {
	return v1.ServManager.V1StopService(args, reply)
}

func (v1 *ApierV1) ServiceStatus(args servmanager.ArgStartService, reply *string) (err error) {
	return v1.ServManager.V1ServiceStatus(args, reply)
}

type ArgsReplyFailedPosts struct {
	FailedRequestsInDir  *string  // if defined it will be our source of requests to be replayed
	FailedRequestsOutDir *string  // if defined it will become our destination for files failing to be replayed, *none to be discarded
	Modules              []string // list of modules for which replay the requests, nil for all
	Transports           []string // list of transports
}

// ReplayFailedPosts will repost failed requests found in the FailedRequestsInDir
func (v1 *ApierV1) ReplayFailedPosts(args ArgsReplyFailedPosts, reply *string) (err error) {
	failedReqsInDir := v1.Config.FailedPostsDir
	if args.FailedRequestsInDir != nil && *args.FailedRequestsInDir != "" {
		failedReqsInDir = *args.FailedRequestsInDir
	}
	failedReqsOutDir := failedReqsInDir
	if args.FailedRequestsOutDir != nil && *args.FailedRequestsOutDir != "" {
		failedReqsOutDir = *args.FailedRequestsOutDir
	}
	filesInDir, _ := ioutil.ReadDir(failedReqsInDir)
	if len(filesInDir) == 0 {
		return utils.ErrNotFound
	}
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		filePath := path.Join(failedReqsInDir, file.Name())
		ffn, err := utils.NewFallbackFileNameFronString(file.Name())
		if err != nil {
			return utils.NewErrServerError(err)
		}
		if len(args.Modules) != 0 {
			var allowedModule bool
			for _, mod := range args.Modules {
				if strings.HasPrefix(ffn.Module, mod) {
					allowedModule = true
					break
				}
			}
			if !allowedModule {
				continue // this file is not to be processed due to Modules ACL
			}
		}
		if len(args.Transports) != 0 && !utils.IsSliceMember(args.Transports, ffn.Transport) {
			continue // this file is not to be processed due to Transports ACL
		}
		var fileContent []byte
		_, err = guardian.Guardian.Guard(func() (interface{}, error) {
			if fileContent, err = ioutil.ReadFile(filePath); err != nil {
				return 0, err
			}
			if err := os.Remove(filePath); err != nil {
				return 0, err
			}
			return 0, nil
		}, v1.Config.LockingTimeout, utils.FileLockPrefix+filePath)
		if err != nil {
			return utils.NewErrServerError(err)
		}
		failoverPath := utils.META_NONE
		if failedReqsOutDir != utils.META_NONE {
			failoverPath = path.Join(failedReqsOutDir, file.Name())
		}
		switch ffn.Transport {
		case utils.MetaHTTPjsonCDR, utils.MetaHTTPjsonMap, utils.MetaHTTPjson, utils.META_HTTP_POST:
			_, err = utils.NewHTTPPoster(v1.Config.HttpSkipTlsVerify,
				v1.Config.ReplyTimeout).Post(ffn.Address, utils.PosterTransportContentTypes[ffn.Transport], fileContent,
				v1.Config.PosterAttempts, failoverPath)
		case utils.MetaAMQPjsonCDR, utils.MetaAMQPjsonMap:
			var amqpPoster *utils.AMQPPoster
			amqpPoster, err = utils.AMQPPostersCache.GetAMQPPoster(ffn.Address, v1.Config.PosterAttempts, failedReqsOutDir)
			if err == nil { // error will be checked bellow
				var chn *amqp.Channel
				chn, err = amqpPoster.Post(
					nil, utils.PosterTransportContentTypes[ffn.Transport], fileContent, file.Name())
				if chn != nil {
					chn.Close()
				}
			}
		default:
			err = fmt.Errorf("unsupported replication transport: %s", ffn.Transport)
		}
		if err != nil && failedReqsOutDir != utils.META_NONE { // Got error from HTTPPoster could be that content was not written, we need to write it ourselves
			_, err := guardian.Guardian.Guard(func() (interface{}, error) {
				if _, err := os.Stat(failoverPath); err == nil || !os.IsNotExist(err) {
					return 0, err
				}
				fileOut, err := os.Create(failoverPath)
				if err != nil {
					return 0, err
				}
				defer fileOut.Close()
				if _, err := fileOut.Write(fileContent); err != nil {
					return 0, err
				}
				return 0, nil
			}, v1.Config.LockingTimeout, utils.FileLockPrefix+failoverPath)
			if err != nil {
				return utils.NewErrServerError(err)
			}
		}
	}
	*reply = utils.OK
	return nil
}
