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
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
)

// SchedulerGeter used to avoid ciclic dependency
type SchedulerGeter interface {
	GetScheduler() *scheduler.Scheduler
}

type APIerSv1 struct {
	StorDb           engine.LoadStorage // we should consider keeping only one of StorDB type
	CdrDb            engine.CdrStorage
	DataManager      *engine.DataManager
	Config           *config.CGRConfig
	Responder        *engine.Responder
	SchedulerService SchedulerGeter  // Need to have them capitalize so we can export in V2
	FilterS          *engine.FilterS //Used for CDR Exporter
	ConnMgr          *engine.ConnManager

	StorDBChan chan engine.StorDB
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (apierSv1 *APIerSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(apierSv1, serviceMethod, args, reply)
}

func (apierSv1 *APIerSv1) GetDestination(dstId *string, reply *engine.Destination) error {
	if dst, err := apierSv1.DataManager.GetDestination(*dstId, true, true, utils.NonTransactional); err != nil {
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

func (apierSv1 *APIerSv1) RemoveDestination(attr *AttrRemoveDestination, reply *string) (err error) {
	for _, dstID := range attr.DestinationIDs {
		var oldDst *engine.Destination
		if oldDst, err = apierSv1.DataManager.GetDestination(dstID, true, true,
			utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return
		}
		if len(attr.Prefixes) != 0 {
			newDst := &engine.Destination{
				Id:       dstID,
				Prefixes: make([]string, 0, len(oldDst.Prefixes)),
			}
			toRemove := utils.NewStringSet(attr.Prefixes)
			for _, prfx := range oldDst.Prefixes {
				if !toRemove.Has(prfx) {
					newDst.Prefixes = append(newDst.Prefixes, prfx)
				}
			}
			if len(newDst.Prefixes) != 0 { // only update the current destination
				if err = apierSv1.DataManager.SetDestination(newDst, utils.NonTransactional); err != nil {
					return
				}
				if err = apierSv1.DataManager.UpdateReverseDestination(oldDst, newDst, utils.NonTransactional); err != nil {
					return
				}
				if err = apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
					utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
						ArgsCache: map[string][]string{utils.ReverseDestinationIDs: oldDst.Prefixes,
							utils.DestinationIDs: {dstID}},
					}, reply); err != nil {
					return
				}
				continue
			}
		}
		if err = apierSv1.DataManager.RemoveDestination(dstID, utils.NonTransactional); err != nil {
			return
		}
		if err = apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
			utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
				ArgsCache: map[string][]string{utils.ReverseDestinationIDs: oldDst.Prefixes,
					utils.DestinationIDs: {dstID}},
			}, reply); err != nil {
			return
		}
	}
	*reply = utils.OK
	return
}

// GetReverseDestination retrieves revese destination list for a prefix
func (apierSv1 *APIerSv1) GetReverseDestination(prefix *string, reply *[]string) (err error) {
	if *prefix == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing("prefix")
	}
	var revLst []string
	if revLst, err = apierSv1.DataManager.GetReverseDestination(*prefix, true, true, utils.NonTransactional); err != nil {
		return
	}
	*reply = revLst
	return
}

// ComputeReverseDestinations will rebuild complete reverse destinations data
func (apierSv1 *APIerSv1) ComputeReverseDestinations(ignr *string, reply *string) (err error) {
	if err = apierSv1.DataManager.RebuildReverseForPrefix(utils.REVERSE_DESTINATION_PREFIX); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// ComputeAccountActionPlans will rebuild complete reverse accountActions data
func (apierSv1 *APIerSv1) ComputeAccountActionPlans(ignr *string, reply *string) (err error) {
	if err = apierSv1.DataManager.RebuildReverseForPrefix(utils.AccountActionPlansPrefix); err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (apierSv1 *APIerSv1) GetSharedGroup(sgId *string, reply *engine.SharedGroup) error {
	if sg, err := apierSv1.DataManager.GetSharedGroup(*sgId, false, utils.NonTransactional); err != nil && err != utils.ErrNotFound { // Not found is not an error here
		return err
	} else {
		if sg != nil {
			*reply = *sg
		}
	}
	return nil
}

func (apierSv1 *APIerSv1) SetDestination(attrs *utils.AttrSetDestination, reply *string) (err error) {
	if missing := utils.MissingStructFields(attrs, []string{"Id", "Prefixes"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	dest := &engine.Destination{Id: attrs.Id, Prefixes: attrs.Prefixes}
	var oldDest *engine.Destination
	if oldDest, err = apierSv1.DataManager.GetDestination(attrs.Id, true, true, utils.NonTransactional); err != nil {
		if err != utils.ErrNotFound {
			return utils.NewErrServerError(err)
		}
	} else if !attrs.Overwrite {
		return utils.ErrExists
	}
	if err := apierSv1.DataManager.SetDestination(dest, utils.NonTransactional); err != nil {
		return utils.NewErrServerError(err)
	}
	if err = apierSv1.DataManager.UpdateReverseDestination(oldDest, dest, utils.NonTransactional); err != nil {
		return
	}
	if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
			ArgsCache: map[string][]string{utils.ReverseDestinationIDs: dest.Prefixes,
				utils.DestinationIDs: {attrs.Id}},
		}, reply); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (apierSv1 *APIerSv1) GetRatingPlan(rplnId *string, reply *engine.RatingPlan) error {
	rpln, err := apierSv1.DataManager.GetRatingPlan(*rplnId, false, utils.NonTransactional)
	if err != nil {
		if err.Error() == utils.ErrNotFound.Error() {
			return err
		}
		return utils.NewErrServerError(err)
	}
	*reply = *rpln
	return nil
}

func (apierSv1 *APIerSv1) RemoveRatingPlan(ID *string, reply *string) error {
	if len(*ID) == 0 {
		return utils.NewErrMandatoryIeMissing("ID")
	}
	err := apierSv1.DataManager.RemoveRatingPlan(*ID, utils.NonTransactional)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
			ArgsCache: map[string][]string{utils.RatingPlanIDs: {*ID}},
		}, reply); err != nil {
		return err
	}
	//generate a loadID for CacheRatingPlans and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRatingPlans: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func (apierSv1 *APIerSv1) ExecuteAction(attr *utils.AttrExecuteAction, reply *string) error {
	at := &engine.ActionTiming{
		ActionsID: attr.ActionsId,
	}
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if attr.Account != "" {
		at.SetAccountIDs(utils.StringMap{utils.ConcatenatedKey(tnt, attr.Account): true})
	}
	if err := at.Execute(nil, nil); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrLoadDestination struct {
	TPid string
	ID   string
}

// Load destinations from storDb into dataDb.
func (apierSv1 *APIerSv1) LoadDestination(attrs *AttrLoadDestination, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader, err := engine.NewTpReader(apierSv1.DataManager.DataDB(), apierSv1.StorDb,
		attrs.TPid, apierSv1.Config.GeneralCfg().DefaultTimezone, apierSv1.Config.ApierCfg().CachesConns,
		apierSv1.Config.ApierCfg().SchedulerConns,
		apierSv1.Config.DataDbCfg().DataDbType == utils.INTERNAL)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if loaded, err := dbReader.LoadDestinationsFiltered(attrs.ID); err != nil {
		return utils.NewErrServerError(err)
	} else if !loaded {
		return utils.ErrNotFound
	}
	if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
			ArgsCache: map[string][]string{utils.DestinationIDs: {attrs.ID}},
		}, reply); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrLoadRatingPlan struct {
	TPid         string
	RatingPlanId string
}

// Process dependencies and load a specific rating plan from storDb into dataDb.
func (apierSv1 *APIerSv1) LoadRatingPlan(attrs *AttrLoadRatingPlan, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader, err := engine.NewTpReader(apierSv1.DataManager.DataDB(), apierSv1.StorDb,
		attrs.TPid, apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().SchedulerConns,
		apierSv1.Config.DataDbCfg().DataDbType == utils.INTERNAL)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if loaded, err := dbReader.LoadRatingPlansFiltered(attrs.RatingPlanId); err != nil {
		return utils.NewErrServerError(err)
	} else if !loaded {
		return utils.ErrNotFound
	}
	*reply = utils.OK
	return nil
}

// Process dependencies and load a specific rating profile from storDb into dataDb.
func (apierSv1 *APIerSv1) LoadRatingProfile(attrs *utils.TPRatingProfile, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	dbReader, err := engine.NewTpReader(apierSv1.DataManager.DataDB(), apierSv1.StorDb,
		attrs.TPid, apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().SchedulerConns,
		apierSv1.Config.DataDbCfg().DataDbType == utils.INTERNAL)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := dbReader.LoadRatingProfilesFiltered(attrs); err != nil {
		return utils.NewErrServerError(err)
	}
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRatingProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err = dbReader.ReloadCache(config.CgrConfig().GeneralCfg().DefaultCaching, true, make(map[string]interface{})); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrLoadSharedGroup struct {
	TPid          string
	SharedGroupId string
}

// Load destinations from storDb into dataDb.
func (apierSv1 *APIerSv1) LoadSharedGroup(attrs *AttrLoadSharedGroup, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader, err := engine.NewTpReader(apierSv1.DataManager.DataDB(), apierSv1.StorDb,
		attrs.TPid, apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().SchedulerConns,
		apierSv1.Config.DataDbCfg().DataDbType == utils.INTERNAL)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := dbReader.LoadSharedGroupsFiltered(attrs.SharedGroupId, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrLoadTpFromStorDb struct {
	TPid     string
	DryRun   bool // Only simulate, no write
	Validate bool // Run structural checks
	Opts     map[string]interface{}
	Caching  *string // Caching strategy
}

// Loads complete data in a TP from storDb
func (apierSv1 *APIerSv1) LoadTariffPlanFromStorDb(attrs *AttrLoadTpFromStorDb, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader, err := engine.NewTpReader(apierSv1.DataManager.DataDB(), apierSv1.StorDb,
		attrs.TPid, apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().SchedulerConns,
		apierSv1.Config.DataDbCfg().DataDbType == utils.INTERNAL)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := dbReader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.Validate {
		if !dbReader.IsValid() {
			*reply = utils.OK
			return errors.New("invalid data")
		}
	}
	if attrs.DryRun {
		*reply = utils.OK
		return nil // Mission complete, no errors
	}
	if err := dbReader.WriteToDatabase(false, false); err != nil {
		return utils.NewErrServerError(err)
	}
	//verify If Caching is present in arguments
	caching := config.CgrConfig().GeneralCfg().DefaultCaching
	if attrs.Caching != nil {
		caching = *attrs.Caching
	}
	// reload cache
	utils.Logger.Info("APIerSv1.LoadTariffPlanFromStorDb, reloading cache.")
	if err := dbReader.ReloadCache(caching, true, attrs.Opts); err != nil {
		return utils.NewErrServerError(err)
	}
	if len(apierSv1.Config.ApierCfg().SchedulerConns) != 0 {
		utils.Logger.Info("APIerSv1.LoadTariffPlanFromStorDb, reloading scheduler.")
		if err := dbReader.ReloadScheduler(true); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	// release the reader with it's structures
	dbReader.Init()
	*reply = utils.OK
	return nil
}

func (apierSv1 *APIerSv1) ImportTariffPlanFromFolder(attrs *utils.AttrImportTPFromFolder, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "FolderPath"}); len(missing) != 0 {
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
		StorDb:   apierSv1.StorDb,
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

// SetRatingProfile sets a specific rating profile working with data directly in the DataDB without involving storDb
func (apierSv1 *APIerSv1) SetRatingProfile(attrs *utils.AttrSetRatingProfile, reply *string) (err error) {
	if missing := utils.MissingStructFields(attrs, []string{"ToR", "Subject", "RatingPlanActivations"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, rpa := range attrs.RatingPlanActivations {
		if missing := utils.MissingStructFields(rpa, []string{"ActivationTimes", "RatingPlanId"}); len(missing) != 0 {
			return fmt.Errorf("%s:RatingPlanActivation:%v", utils.ErrMandatoryIeMissing.Error(), missing)
		}
	}
	tnt := attrs.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	keyID := utils.ConcatenatedKey(utils.META_OUT,
		tnt, attrs.Category, attrs.Subject)
	var rpfl *engine.RatingProfile
	if !attrs.Overwrite {
		if rpfl, err = apierSv1.DataManager.GetRatingProfile(keyID, false, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return utils.NewErrServerError(err)
		}
	}
	if rpfl == nil {
		rpfl = &engine.RatingProfile{Id: keyID, RatingPlanActivations: make(engine.RatingPlanActivations, 0)}
	}
	for _, ra := range attrs.RatingPlanActivations {
		at, err := utils.ParseTimeDetectLayout(ra.ActivationTime,
			apierSv1.Config.GeneralCfg().DefaultTimezone)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("%s:Cannot parse activation time from %v", utils.ErrServerError.Error(), ra.ActivationTime))
		}
		if exists, err := apierSv1.DataManager.HasData(utils.RATING_PLAN_PREFIX,
			ra.RatingPlanId, ""); err != nil {
			return utils.NewErrServerError(err)
		} else if !exists {
			return fmt.Errorf(fmt.Sprintf("%s:RatingPlanId:%s", utils.ErrNotFound.Error(), ra.RatingPlanId))
		}
		rpfl.RatingPlanActivations = append(rpfl.RatingPlanActivations,
			&engine.RatingPlanActivation{
				ActivationTime: at,
				RatingPlanId:   ra.RatingPlanId,
				FallbackKeys: utils.FallbackSubjKeys(tnt,
					attrs.Category, ra.FallbackSubjects)})
	}
	if err := apierSv1.DataManager.SetRatingProfile(rpfl, utils.NonTransactional); err != nil {
		return utils.NewErrServerError(err)
	}
	//CacheReload
	if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
			ArgsCache: map[string][]string{utils.RatingProfileIDs: {rpfl.Id}},
		}, reply); err != nil {
		return err
	}
	//generate a loadID for CacheRatingProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRatingProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetRatingProfileIDs returns list of resourceProfile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetRatingProfileIDs(args *utils.PaginatorWithTenant, rsPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.RATING_PROFILE_PREFIX + "*out:" + tnt + utils.CONCATENATED_KEY_SEP
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*rsPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

func (apierSv1 *APIerSv1) GetRatingProfile(attrs *utils.AttrGetRatingProfile, reply *engine.RatingProfile) (err error) {
	if missing := utils.MissingStructFields(attrs, []string{utils.Category, utils.Subject}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if rpPrf, err := apierSv1.DataManager.GetRatingProfile(attrs.GetID(),
		false, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *rpPrf
	}
	return
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

func (apierSv1 *APIerSv1) SetActions(attrs *V1AttrSetActions, reply *string) (err error) {
	if missing := utils.MissingStructFields(attrs, []string{"ActionsId", "Actions"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, action := range attrs.Actions {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != utils.EmptyString { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), action.Identifier, missing)
		}
	}
	if !attrs.Overwrite {
		if exists, err := apierSv1.DataManager.HasData(utils.ACTION_PREFIX, attrs.ActionsId, ""); err != nil {
			return utils.NewErrServerError(err)
		} else if exists {
			return utils.ErrExists
		}
	}
	storeActions := make(engine.Actions, len(attrs.Actions))
	for idx, apiAct := range attrs.Actions {
		var blocker *bool
		if apiAct.BalanceBlocker != utils.EmptyString {
			if x, err := strconv.ParseBool(apiAct.BalanceBlocker); err == nil {
				blocker = &x
			} else {
				return err
			}
		}

		var disabled *bool
		if apiAct.BalanceDisabled != utils.EmptyString {
			if x, err := strconv.ParseBool(apiAct.BalanceDisabled); err == nil {
				disabled = &x
			} else {
				return err
			}
		}
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
				DestinationIDs: utils.StringMapPointer(utils.ParseStringMap(apiAct.DestinationIds)),
				RatingSubject:  utils.StringPointer(apiAct.RatingSubject),
				SharedGroups:   utils.StringMapPointer(utils.ParseStringMap(apiAct.SharedGroups)),
				Categories:     utils.StringMapPointer(utils.ParseStringMap(apiAct.Categories)),
				TimingIDs:      utils.StringMapPointer(utils.ParseStringMap(apiAct.TimingTags)),
				Blocker:        blocker,
				Disabled:       disabled,
			},
		}
		storeActions[idx] = a
	}
	if err := apierSv1.DataManager.SetActions(attrs.ActionsId, storeActions, utils.NonTransactional); err != nil {
		return utils.NewErrServerError(err)
	}
	//CacheReload
	if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
			ArgsCache: map[string][]string{utils.ActionIDs: {attrs.ActionsId}},
		}, reply); err != nil {
		return err
	}
	//generate a loadID for CacheActions and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheActions: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// Retrieves actions attached to specific ActionsId within cache
func (apierSv1 *APIerSv1) GetActions(actsId *string, reply *[]*utils.TPAction) error {
	if len(*actsId) == 0 {
		return fmt.Errorf("%s ActionsId: %s", utils.ErrMandatoryIeMissing.Error(), *actsId)
	}
	acts := make([]*utils.TPAction, 0)
	engActs, err := apierSv1.DataManager.GetActions(*actsId, false, utils.NonTransactional)
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
	TimingID  string  // timingID is used to specify the ID of the timing for a corner case ( e.g. *monthly_estimated )
	Years     string  // semicolon separated list of years this timing is valid on, *any or empty supported
	Months    string  // semicolon separated list of months this timing is valid on, *any or empty supported
	MonthDays string  // semicolon separated list of month's days this timing is valid on, *any or empty supported
	WeekDays  string  // semicolon separated list of week day names this timing is valid on *any or empty supported
	Time      string  // String representing the time this timing starts on, *asap supported
	Weight    float64 // Binding's weight
}

func (attr *AttrActionPlan) getRITiming(dm *engine.DataManager) (timing *engine.RITiming, err error) {
	if dfltTiming, isDefault := checkDefaultTiming(attr.Time); isDefault {
		return dfltTiming, nil
	}
	timing = new(engine.RITiming)

	if attr.TimingID != utils.EmptyString &&
		!strings.HasPrefix(attr.TimingID, utils.Meta) { // in case of dynamic timing
		if dbTiming, err := dm.GetTiming(attr.TimingID, false, utils.NonTransactional); err != nil {
			if err != utils.ErrNotFound { // if not found let the user to populate all the timings values
				return nil, err
			}
		} else {
			timing.ID = dbTiming.ID
			timing.Years = dbTiming.Years
			timing.Months = dbTiming.Months
			timing.MonthDays = dbTiming.MonthDays
			timing.WeekDays = dbTiming.WeekDays
			timing.StartTime = dbTiming.StartTime
			timing.EndTime = dbTiming.EndTime
		}
	}
	timing.ID = attr.TimingID
	timing.Years.Parse(attr.Years, ";")
	timing.Months.Parse(attr.Months, ";")
	timing.MonthDays.Parse(attr.MonthDays, ";")
	timing.WeekDays.Parse(attr.WeekDays, ";")
	if !verifyFormat(attr.Time) {
		err = fmt.Errorf("%s:%s", utils.ErrUnsupportedFormat.Error(), attr.Time)
		return
	}
	timing.StartTime = attr.Time
	return
}

func (apierSv1 *APIerSv1) SetActionPlan(attrs *AttrSetActionPlan, reply *string) (err error) {
	if missing := utils.MissingStructFields(attrs, []string{"Id", "ActionPlan"}); len(missing) != 0 {
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
		if prevAP, err := apierSv1.DataManager.GetActionPlan(attrs.Id, false, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
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
			if exists, err := apierSv1.DataManager.HasData(utils.ACTION_PREFIX, apiAtm.ActionsId, ""); err != nil {
				return 0, utils.NewErrServerError(err)
			} else if !exists {
				return 0, fmt.Errorf("%s:%s", utils.ErrBrokenReference.Error(), apiAtm.ActionsId)
			}
			timing, err := apiAtm.getRITiming(apierSv1.DataManager)
			if err != nil {
				return 0, err
			}
			ap.ActionTimings = append(ap.ActionTimings, &engine.ActionTiming{
				Uuid:      utils.GenUUID(),
				Weight:    apiAtm.Weight,
				Timing:    &engine.RateInterval{Timing: timing},
				ActionsID: apiAtm.ActionsId,
			})
		}
		if err := apierSv1.DataManager.SetActionPlan(ap.Id, ap, true, utils.NonTransactional); err != nil {
			return 0, utils.NewErrServerError(err)
		}
		if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
			utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
				ArgsCache: map[string][]string{utils.ActionPlanIDs: {ap.Id}},
			}, reply); err != nil {
			return 0, err
		}
		for acntID := range prevAccountIDs {
			if err := apierSv1.DataManager.RemAccountActionPlans(acntID, []string{attrs.Id}); err != nil {
				return 0, utils.NewErrServerError(err)
			}
		}
		if len(prevAccountIDs) != 0 {
			sl := prevAccountIDs.Slice()
			if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
				utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
					ArgsCache: map[string][]string{utils.AccountActionPlanIDs: sl},
				}, reply); err != nil {
				return 0, err
			}
		}
		return 0, nil
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return err
	}
	if attrs.ReloadScheduler {
		sched := apierSv1.SchedulerService.GetScheduler()
		if sched == nil {
			return errors.New(utils.SchedulerNotRunningCaps)
		}
		sched.Reload()
	}
	//generate a loadID for CacheActionPlans and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheActionPlans: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func verifyFormat(tStr string) bool {
	if tStr == utils.EmptyString ||
		tStr == utils.ASAP {
		return true
	}

	if len(tStr) > 8 { // hh:mm:ss
		return false
	}
	if a := strings.Split(tStr, utils.InInFieldSep); len(a) != 3 {
		return false
	} else {
		if _, err := strconv.Atoi(a[0]); err != nil {
			return false
		} else if _, err := strconv.Atoi(a[1]); err != nil {
			return false
		} else if _, err := strconv.Atoi(a[2]); err != nil {
			return false
		}
	}
	return true
}

// checkDefaultTiming will check the tStr if it's of the the default timings ( the same as in TPReader )
// and will compute it properly
func checkDefaultTiming(tStr string) (rTm *engine.RITiming, isDefault bool) {
	startTime := time.Now().Format("15:04:05")
	switch tStr {
	case utils.MetaEveryMinute:
		return &engine.RITiming{
			ID:        utils.MetaEveryMinute,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
			StartTime: utils.ConcatenatedKey(utils.Meta, utils.Meta, strconv.Itoa(time.Now().Second())),
			EndTime:   "",
		}, true
	case utils.MetaHourly:
		return &engine.RITiming{
			ID:        utils.MetaHourly,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
			StartTime: utils.ConcatenatedKey(utils.Meta, strconv.Itoa(time.Now().Minute()), strconv.Itoa(time.Now().Second())),
			EndTime:   "",
		}, true
	case utils.MetaDaily:
		return &engine.RITiming{
			ID:        utils.MetaDaily,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
			StartTime: startTime,
			EndTime:   ""}, true
	case utils.MetaWeekly:
		return &engine.RITiming{
			ID:        utils.MetaWeekly,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{time.Now().Weekday()},
			StartTime: startTime,
			EndTime:   "",
		}, true
	case utils.MetaMonthly:
		return &engine.RITiming{
			ID:        utils.MetaMonthly,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{time.Now().Day()},
			WeekDays:  utils.WeekDays{},
			StartTime: startTime,
			EndTime:   "",
		}, true
	case utils.MetaMonthlyEstimated:
		return &engine.RITiming{
			ID:        utils.MetaMonthlyEstimated,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{time.Now().Day()},
			WeekDays:  utils.WeekDays{},
			StartTime: startTime,
			EndTime:   "",
		}, true
	case utils.MetaMonthEnd:
		return &engine.RITiming{
			ID:        utils.MetaMonthEnd,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{-1},
			WeekDays:  utils.WeekDays{},
			StartTime: startTime,
			EndTime:   "",
		}, true
	case utils.MetaYearly:
		return &engine.RITiming{
			ID:        utils.MetaYearly,
			Years:     utils.Years{},
			Months:    utils.Months{time.Now().Month()},
			MonthDays: utils.MonthDays{time.Now().Day()},
			WeekDays:  utils.WeekDays{},
			StartTime: startTime,
			EndTime:   "",
		}, true
	default:
		return nil, false
	}
}

type AttrGetActionPlan struct {
	ID string
}

func (apierSv1 *APIerSv1) GetActionPlan(attr *AttrGetActionPlan, reply *[]*engine.ActionPlan) error {
	var result []*engine.ActionPlan
	if attr.ID == "" || attr.ID == "*" {
		result = make([]*engine.ActionPlan, 0)
		aplsMap, err := apierSv1.DataManager.GetAllActionPlans()
		if err != nil {
			return err
		}
		for _, apls := range aplsMap {
			result = append(result, apls)
		}
	} else {
		apls, err := apierSv1.DataManager.GetActionPlan(attr.ID, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		result = append(result, apls)
	}
	*reply = result
	return nil
}

func (apierSv1 *APIerSv1) RemoveActionPlan(attr *AttrGetActionPlan, reply *string) (err error) {
	if missing := utils.MissingStructFields(attr, []string{"ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if _, err = guardian.Guardian.Guard(func() (interface{}, error) {
		var prevAccountIDs utils.StringMap
		if prevAP, err := apierSv1.DataManager.GetActionPlan(attr.ID, false, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return 0, err
		} else if prevAP != nil {
			prevAccountIDs = prevAP.AccountIDs
		}
		if err := apierSv1.DataManager.RemoveActionPlan(attr.ID, utils.NonTransactional); err != nil {
			return 0, err
		}
		for acntID := range prevAccountIDs {
			if err := apierSv1.DataManager.RemAccountActionPlans(acntID, []string{attr.ID}); err != nil {
				return 0, utils.NewErrServerError(err)
			}
		}
		if len(prevAccountIDs) != 0 {
			sl := prevAccountIDs.Slice()
			if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
				utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
					ArgsCache: map[string][]string{utils.AccountActionPlanIDs: sl},
				}, reply); err != nil {
				return 0, err
			}
		}
		return 0, nil
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ACTION_PLAN_PREFIX); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// Process dependencies and load a specific AccountActions profile from storDb into dataDb.
func (apierSv1 *APIerSv1) LoadAccountActions(attrs *utils.TPAccountActions, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader, err := engine.NewTpReader(apierSv1.DataManager.DataDB(), apierSv1.StorDb,
		attrs.TPid, apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().SchedulerConns,
		apierSv1.Config.DataDbCfg().DataDbType == utils.INTERNAL)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if _, err := guardian.Guardian.Guard(func() (interface{}, error) {
		return 0, dbReader.LoadAccountActionsFiltered(attrs)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, attrs.LoadId); err != nil {
		return utils.NewErrServerError(err)
	}
	// ToDo: Get the action keys loaded by dbReader so we reload only these in cache
	// Need to do it before scheduler otherwise actions to run will be unknown
	sched := apierSv1.SchedulerService.GetScheduler()
	if sched != nil {
		sched.Reload()
	}
	*reply = utils.OK
	return nil
}

func (apierSv1 *APIerSv1) LoadTariffPlanFromFolder(attrs *utils.AttrLoadTpFromFolder, reply *string) error {
	// verify if FolderPath is present
	if len(attrs.FolderPath) == 0 {
		return fmt.Errorf("%s:%s", utils.ErrMandatoryIeMissing.Error(), "FolderPath")
	}
	// check if exists or is valid
	if fi, err := os.Stat(attrs.FolderPath); err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return utils.ErrInvalidPath
		}
		return utils.NewErrServerError(err)
	} else if !fi.IsDir() {
		return utils.ErrInvalidPath
	}

	// create the TpReader
	loader, err := engine.NewTpReader(apierSv1.DataManager.DataDB(),
		engine.NewFileCSVStorage(utils.CSV_SEP, attrs.FolderPath),
		"", apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().SchedulerConns,
		apierSv1.Config.DataDbCfg().DataDbType == utils.INTERNAL)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	//Load the data
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.DryRun {
		*reply = utils.OK
		return nil // Mission complete, no errors
	}

	if attrs.Validate {
		if !loader.IsValid() {
			return errors.New("invalid data")
		}
	}

	// write data intro Database
	if err := loader.WriteToDatabase(false, false); err != nil {
		return utils.NewErrServerError(err)
	}
	//verify If Caching is present in arguments
	caching := config.CgrConfig().GeneralCfg().DefaultCaching
	if attrs.Caching != nil {
		caching = *attrs.Caching
	}
	// reload cache
	utils.Logger.Info("APIerSv1.LoadTariffPlanFromFolder, reloading cache.")
	if err := loader.ReloadCache(caching, true, attrs.Opts); err != nil {
		return utils.NewErrServerError(err)
	}
	if len(apierSv1.Config.ApierCfg().SchedulerConns) != 0 {
		utils.Logger.Info("APIerSv1.LoadTariffPlanFromFolder, reloading scheduler.")
		if err := loader.ReloadScheduler(true); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	// release the reader with it's structures
	loader.Init()
	*reply = utils.OK
	return nil
}

// RemoveTPFromFolder will load the tarrifplan from folder into TpReader object
// and will delete if from database
func (apierSv1 *APIerSv1) RemoveTPFromFolder(attrs *utils.AttrLoadTpFromFolder, reply *string) error {
	// verify if FolderPath is present
	if len(attrs.FolderPath) == 0 {
		return fmt.Errorf("%s:%s", utils.ErrMandatoryIeMissing.Error(), "FolderPath")
	}
	// check if exists or is valid
	if fi, err := os.Stat(attrs.FolderPath); err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return utils.ErrInvalidPath
		}
		return utils.NewErrServerError(err)
	} else if !fi.IsDir() {
		return utils.ErrInvalidPath
	}

	// create the TpReader
	loader, err := engine.NewTpReader(apierSv1.DataManager.DataDB(),
		engine.NewFileCSVStorage(utils.CSV_SEP, attrs.FolderPath), "", apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().SchedulerConns,
		apierSv1.Config.DataDbCfg().DataDbType == utils.INTERNAL)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	//Load the data
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.DryRun {
		*reply = utils.OK
		return nil // Mission complete, no errors
	}

	if attrs.Validate {
		if !loader.IsValid() {
			return errors.New("invalid data")
		}
	}

	// remove data from Database
	if err := loader.RemoveFromDatabase(false, false); err != nil {
		return utils.NewErrServerError(err)
	}
	//verify If Caching is present in arguments
	caching := config.CgrConfig().GeneralCfg().DefaultCaching
	if attrs.Caching != nil {
		caching = *attrs.Caching
	}
	// reload cache
	utils.Logger.Info("APIerSv1.RemoveTPFromFolder, reloading cache.")
	if err := loader.ReloadCache(caching, true, attrs.Opts); err != nil {
		return utils.NewErrServerError(err)
	}
	if len(apierSv1.Config.ApierCfg().SchedulerConns) != 0 {
		utils.Logger.Info("APIerSv1.RemoveTPFromFolder, reloading scheduler.")
		if err := loader.ReloadScheduler(true); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	// release the reader with it's structures
	loader.Init()
	*reply = utils.OK
	return nil
}

// RemoveTPFromStorDB will load the tarrifplan from StorDB into TpReader object
// and will delete if from database
func (apierSv1 *APIerSv1) RemoveTPFromStorDB(attrs *AttrLoadTpFromStorDb, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader, err := engine.NewTpReader(apierSv1.DataManager.DataDB(), apierSv1.StorDb,
		attrs.TPid, apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().SchedulerConns,
		apierSv1.Config.DataDbCfg().DataDbType == utils.INTERNAL)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := dbReader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.Validate {
		if !dbReader.IsValid() {
			*reply = utils.OK
			return errors.New("invalid data")
		}
	}
	if attrs.DryRun {
		*reply = utils.OK
		return nil // Mission complete, no errors
	}
	// remove data from Database
	if err := dbReader.RemoveFromDatabase(false, false); err != nil {
		return utils.NewErrServerError(err)
	}
	//verify If Caching is present in arguments
	caching := config.CgrConfig().GeneralCfg().DefaultCaching
	if attrs.Caching != nil {
		caching = *attrs.Caching
	}
	// reload cache
	utils.Logger.Info("APIerSv1.RemoveTPFromStorDB, reloading cache.")
	if err := dbReader.ReloadCache(caching, true, attrs.Opts); err != nil {
		return utils.NewErrServerError(err)
	}
	if len(apierSv1.Config.ApierCfg().SchedulerConns) != 0 {
		utils.Logger.Info("APIerSv1.RemoveTPFromStorDB, reloading scheduler.")
		if err := dbReader.ReloadScheduler(true); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	// release the reader with it's structures
	dbReader.Init()
	*reply = utils.OK
	return nil
}

type AttrRemoveRatingProfile struct {
	Tenant   string
	Category string
	Subject  string
}

func (arrp *AttrRemoveRatingProfile) GetId() (result string) {
	result = utils.META_OUT + utils.CONCATENATED_KEY_SEP
	if arrp.Tenant != "" && arrp.Tenant != utils.ANY {
		result += arrp.Tenant + utils.CONCATENATED_KEY_SEP
	} else {
		return
	}

	if arrp.Category != "" && arrp.Category != utils.ANY {
		result += arrp.Category + utils.CONCATENATED_KEY_SEP
	} else {
		return
	}
	if arrp.Subject != "" && arrp.Subject != utils.ANY {
		result += arrp.Subject
	}
	return
}

func (apierSv1 *APIerSv1) RemoveRatingProfile(attr *AttrRemoveRatingProfile, reply *string) error {
	if attr.Tenant == utils.EmptyString {
		attr.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if (attr.Subject != "" && utils.IsSliceMember([]string{attr.Tenant, attr.Category}, "")) ||
		(attr.Category != "" && attr.Tenant == "") {
		return utils.ErrMandatoryIeMissing
	}
	_, err := guardian.Guardian.Guard(func() (interface{}, error) {
		return 0, apierSv1.DataManager.RemoveRatingProfile(attr.GetId(), utils.NonTransactional)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, "RemoveRatingProfile")
	if err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}
	if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
			ArgsCache: map[string][]string{utils.RatingProfileIDs: {attr.GetId()}},
		}, reply); err != nil {
		return err
	}
	//generate a loadID for CacheActionPlans and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRatingProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func (apierSv1 *APIerSv1) GetLoadHistory(attrs *utils.Paginator, reply *[]*utils.LoadInstance) error {
	nrItems := -1
	offset := 0
	if attrs.Offset != nil { // For offset we need full data
		offset = *attrs.Offset
	} else if attrs.Limit != nil {
		nrItems = *attrs.Limit
	}
	loadHist, err := apierSv1.DataManager.DataDB().GetLoadHistory(nrItems, true, utils.NonTransactional)
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

func (apierSv1 *APIerSv1) RemoveActions(attr *AttrRemoveActions, reply *string) error {
	if attr.ActionIDs == nil {
		err := utils.ErrNotFound
		*reply = err.Error()
		return err
	}
	// The check could lead to very long execution time. So we decided to leave it at the user's risck.'
	/*
		stringMap := utils.NewStringMap(attr.ActionIDs...)
		keys, err := apiv1.DataManager.DataDB().GetKeysForPrefix(utils.ACTION_TRIGGER_PREFIX, true)
		if err != nil {
			*reply = err.Error()
			return err
		}
		for _, key := range keys {
			getAttrs, err := apiv1.DataManager.DataDB().GetActionTriggers(key[len(utils.ACTION_TRIGGER_PREFIX):])
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
		allAplsMap, err := apiv1.DataManager.GetAllActionPlans()
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
		if err := apierSv1.DataManager.RemoveActions(aID, utils.NonTransactional); err != nil {
			*reply = err.Error()
			return err
		}
	}
	//CacheReload
	if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithOpts{
			ArgsCache: map[string][]string{utils.ActionIDs: attr.ActionIDs},
		}, reply); err != nil {
		return err
	}
	//generate a loadID for CacheActions and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheActions: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

type ArgsReplyFailedPosts struct {
	FailedRequestsInDir  *string  // if defined it will be our source of requests to be replayed
	FailedRequestsOutDir *string  // if defined it will become our destination for files failing to be replayed, *none to be discarded
	Modules              []string // list of modules for which replay the requests, nil for all
}

// ReplayFailedPosts will repost failed requests found in the FailedRequestsInDir
func (apierSv1 *APIerSv1) ReplayFailedPosts(args *ArgsReplyFailedPosts, reply *string) (err error) {
	failedReqsInDir := apierSv1.Config.GeneralCfg().FailedPostsDir
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
		if len(args.Modules) != 0 {
			var allowedModule bool
			for _, mod := range args.Modules {
				if strings.HasPrefix(file.Name(), mod) {
					allowedModule = true
					break
				}
			}
			if !allowedModule {
				continue // this file is not to be processed due to Modules ACL
			}
		}
		filePath := path.Join(failedReqsInDir, file.Name())
		expEv, err := engine.NewExportEventsFromFile(filePath)
		if err != nil {
			return utils.NewErrServerError(err)
		}

		failoverPath := utils.MetaNone
		if failedReqsOutDir != utils.MetaNone {
			failoverPath = path.Join(failedReqsOutDir, file.Name())
		}

		failedPosts, err := expEv.ReplayFailedPosts(apierSv1.Config.GeneralCfg().PosterAttempts)
		if err != nil && failedReqsOutDir != utils.MetaNone { // Got error from HTTPPoster could be that content was not written, we need to write it ourselves
			if err = failedPosts.WriteToFile(failoverPath); err != nil {
				return utils.NewErrServerError(err)
			}
		}

	}
	*reply = utils.OK
	return nil
}

func (apierSv1 *APIerSv1) GetLoadIDs(args *string, reply *map[string]int64) (err error) {
	var loadIDs map[string]int64
	if loadIDs, err = apierSv1.DataManager.GetItemLoadIDs(*args, false); err != nil {
		return
	}
	*reply = loadIDs
	return
}

type LoadTimeArgs struct {
	Timezone string
	Item     string
}

func (apierSv1 *APIerSv1) GetLoadTimes(args *LoadTimeArgs, reply *map[string]string) (err error) {
	if loadIDs, err := apierSv1.DataManager.GetItemLoadIDs(args.Item, false); err != nil {
		return err
	} else {
		provMp := make(map[string]string)
		for key, val := range loadIDs {
			timeVal, err := utils.ParseTimeDetectLayout(strconv.FormatInt(val, 10), args.Timezone)
			if err != nil {
				return err
			}
			provMp[key] = timeVal.String()
		}
		*reply = provMp
	}
	return
}

func (apierSv1 *APIerSv1) ComputeActionPlanIndexes(_ string, reply *string) (err error) {
	if apierSv1.DataManager.DataDB().GetStorageType() != utils.Redis {
		return utils.ErrNotImplemented
	}
	redisDB, can := apierSv1.DataManager.DataDB().(*engine.RedisStorage)
	if !can {
		return fmt.Errorf("Storage type %s could not be cated to <*engine.RedisStorage>", apierSv1.DataManager.DataDB().GetStorageType())
	}
	if err = redisDB.RebbuildActionPlanKeys(); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// GetActionPlanIDs returns list of ActionPlan IDs registered for a tenant
func (apierSv1 *APIerSv1) GetActionPlanIDs(args *utils.PaginatorWithTenant, attrPrfIDs *[]string) error {
	prfx := utils.ACTION_PLAN_PREFIX
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(utils.ACTION_PLAN_PREFIX)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*attrPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetRatingPlanIDs returns list of RatingPlan IDs registered for a tenant
func (apierSv1 *APIerSv1) GetRatingPlanIDs(args *utils.PaginatorWithTenant, attrPrfIDs *[]string) error {
	prfx := utils.RATING_PLAN_PREFIX
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(utils.RATING_PLAN_PREFIX)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*attrPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// ListenAndServe listen for storbd reload
func (apierSv1 *APIerSv1) ListenAndServe(stopChan chan struct{}) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ApierS))
	for {
		select {
		case <-stopChan:
			return
		case stordb, ok := <-apierSv1.StorDBChan:
			if !ok { // the chanel was closed by the shutdown of stordbService
				return
			}
			apierSv1.CdrDb = stordb
			apierSv1.StorDb = stordb
		}
	}
}

// Ping return pong if the service is active
func (apierSv1 *APIerSv1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

//ExportToFolder export specific items (or all items if items is empty) from DataDB back to CSV
func (apierSv1 *APIerSv1) ExportToFolder(arg *utils.ArgExportToFolder, reply *string) error {
	// if items is empty we need to export all items
	if len(arg.Items) == 0 {
		arg.Items = []string{utils.MetaAttributes, utils.MetaChargers, utils.MetaDispatchers,
			utils.MetaDispatcherHosts, utils.MetaFilters, utils.MetaResources, utils.MetaStats,
			utils.MetaRoutes, utils.MetaThresholds, utils.MetaRateProfiles, utils.MetaActionProfiles}
	}
	if _, err := os.Stat(arg.Path); os.IsNotExist(err) {
		os.Mkdir(arg.Path, os.ModeDir)
	}
	for _, item := range arg.Items {
		switch item {
		case utils.MetaAttributes:
			prfx := utils.AttributeProfilePrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.AttributesCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.AttributeMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				attPrf, err := apierSv1.DataManager.GetAttributeProfile(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelTPAttribute(
					engine.AttributeProfileToAPI(attPrf)) {
					if record, err := engine.CsvDump(model); err != nil {
						return err
					} else if err := csvWriter.Write(record); err != nil {
						return err
					}
				}
			}
			csvWriter.Flush()
		case utils.MetaChargers:
			prfx := utils.ChargerProfilePrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.ChargersCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.ChargerMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				chrPrf, err := apierSv1.DataManager.GetChargerProfile(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelTPCharger(
					engine.ChargerProfileToAPI(chrPrf)) {
					if record, err := engine.CsvDump(model); err != nil {
						return err
					} else if err := csvWriter.Write(record); err != nil {
						return err
					}
				}
			}
			csvWriter.Flush()
		case utils.MetaDispatchers:
			prfx := utils.DispatcherProfilePrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.DispatcherProfilesCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.DispatcherProfileMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				dpsPrf, err := apierSv1.DataManager.GetDispatcherProfile(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelTPDispatcherProfile(
					engine.DispatcherProfileToAPI(dpsPrf)) {
					if record, err := engine.CsvDump(model); err != nil {
						return err
					} else if err := csvWriter.Write(record); err != nil {
						return err
					}
				}
			}
			csvWriter.Flush()
		case utils.MetaDispatcherHosts:
			prfx := utils.DispatcherHostPrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.DispatcherHostsCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.DispatcherHostMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				dpsPrf, err := apierSv1.DataManager.GetDispatcherHost(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				if record, err := engine.CsvDump(engine.APItoModelTPDispatcherHost(
					engine.DispatcherHostToAPI(dpsPrf))); err != nil {
					return err
				} else if err := csvWriter.Write(record); err != nil {
					return err
				}
			}
			csvWriter.Flush()
		case utils.MetaFilters:
			prfx := utils.FilterPrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.FiltersCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.FilterMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				fltr, err := apierSv1.DataManager.GetFilter(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelTPFilter(
					engine.FilterToTPFilter(fltr)) {
					if record, err := engine.CsvDump(model); err != nil {
						return err
					} else if err := csvWriter.Write(record); err != nil {
						return err
					}
				}
			}
			csvWriter.Flush()
		case utils.MetaResources:
			prfx := utils.ResourceProfilesPrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.ResourcesCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.ResourceMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				resPrf, err := apierSv1.DataManager.GetResourceProfile(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelResource(
					engine.ResourceProfileToAPI(resPrf)) {
					if record, err := engine.CsvDump(model); err != nil {
						return err
					} else if err := csvWriter.Write(record); err != nil {
						return err
					}
				}
			}
			csvWriter.Flush()
		case utils.MetaStatS:
			prfx := utils.StatQueueProfilePrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.StatsCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.StatMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				stsPrf, err := apierSv1.DataManager.GetStatQueueProfile(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelStats(
					engine.StatQueueProfileToAPI(stsPrf)) {
					if record, err := engine.CsvDump(model); err != nil {
						return err
					} else if err := csvWriter.Write(record); err != nil {
						return err
					}
				}
			}
			csvWriter.Flush()
		case utils.MetaRoutes:
			prfx := utils.RouteProfilePrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.RoutesCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.RouteMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				spp, err := apierSv1.DataManager.GetRouteProfile(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelTPRoutes(
					engine.RouteProfileToAPI(spp)) {
					if record, err := engine.CsvDump(model); err != nil {
						return err
					} else if err := csvWriter.Write(record); err != nil {
						return err
					}
				}
			}
			csvWriter.Flush()
		case utils.MetaThresholds:
			prfx := utils.ThresholdProfilePrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.ThresholdsCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.ThresholdMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				thPrf, err := apierSv1.DataManager.GetThresholdProfile(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelTPThreshold(
					engine.ThresholdProfileToAPI(thPrf)) {
					if record, err := engine.CsvDump(model); err != nil {
						return err
					} else if err := csvWriter.Write(record); err != nil {
						return err
					}
				}
			}
			csvWriter.Flush()
		case utils.MetaRateProfiles:
			prfx := utils.RateProfilePrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.RateProfilesCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.RateProfileMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				rPrf, err := apierSv1.DataManager.GetRateProfile(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelTPRateProfile(engine.RateProfileToAPI(rPrf)) {
					if record, err := engine.CsvDump(model); err != nil {
						return err
					} else if err := csvWriter.Write(record); err != nil {
						return err
					}
				}
			}
			csvWriter.Flush()

		case utils.MetaActionProfiles:
			prfx := utils.ActionProfilePrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 { // if we don't find items we skip
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.ActionProfilesCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSV_SEP
			//write the header of the file
			if err := csvWriter.Write(engine.ActionProfileMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				rPrf, err := apierSv1.DataManager.GetActionProfile(tntID[0], tntID[1],
					true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelTPActionProfile(engine.ActionProfileToAPI(rPrf)) {
					if record, err := engine.CsvDump(model); err != nil {
						return err
					} else if err := csvWriter.Write(record); err != nil {
						return err
					}
				}
			}
			csvWriter.Flush()

		}

	}
	*reply = utils.OK
	return nil
}

func (apierSv1 *APIerSv1) ExportCDRs(args *utils.ArgExportCDRs, reply *map[string]interface{}) (err error) {
	if len(apierSv1.Config.ApierCfg().EEsConns) == 0 {
		return utils.NewErrNotConnected(utils.EEs)
	}
	cdrsFltr, err := args.RPCCDRsFilter.AsCDRsFilter(apierSv1.Config.GeneralCfg().DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := apierSv1.CdrDb.GetCDRs(cdrsFltr, false)
	if err != nil {
		return err
	} else if len(cdrs) == 0 {
		return utils.ErrNotFound
	}
	withErros := false
	var rplyCdr map[string]map[string]interface{}
	for _, cdr := range cdrs {
		argCdr := &utils.CGREventWithEeIDs{
			EeIDs: args.ExporterIDs,
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: cdr.AsCGREvent(),
				Opts:     make(map[string]interface{}),
			},
		}
		if args.Verbose {
			argCdr.CGREventWithOpts.Opts[utils.OptsEEsVerbose] = struct{}{}
		}
		if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().EEsConns, nil, utils.EeSv1ProcessEvent,
			argCdr, &rplyCdr); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> error: <%s> processing event: <%s> with <%s>",
				utils.ApierS, err.Error(), utils.ToJSON(cdr.AsCGREvent()), utils.EventExporterS))
			withErros = true
		}
	}
	if withErros {
		return utils.ErrPartiallyExecuted
	}
	// we consider only the last reply because it should have the metrics updated
	for exporterID, metrics := range rplyCdr {
		(*reply)[exporterID] = metrics
	}
	return
}
