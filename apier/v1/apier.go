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
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type APIerSv1 struct {
	StorDb      engine.LoadStorage // we should consider keeping only one of StorDB type
	CdrDb       engine.CdrStorage
	DataManager *engine.DataManager
	Config      *config.CGRConfig
	FilterS     *engine.FilterS //Used for CDR Exporter
	ConnMgr     *engine.ConnManager

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
				ID:       dstID,
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
					utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithAPIOpts{
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
			utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithAPIOpts{
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
	if err = apierSv1.DataManager.RebuildReverseForPrefix(utils.ReverseDestinationPrefix); err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (apierSv1 *APIerSv1) SetDestination(attrs *utils.AttrSetDestination, reply *string) (err error) {
	if missing := utils.MissingStructFields(attrs, []string{"Id", "Prefixes"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	dest := &engine.Destination{ID: attrs.Id, Prefixes: attrs.Prefixes}
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
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithAPIOpts{
			ArgsCache: map[string][]string{utils.ReverseDestinationIDs: dest.Prefixes,
				utils.DestinationIDs: {attrs.Id}},
		}, reply); err != nil {
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
		apierSv1.Config.ApierCfg().ActionSConns,
		apierSv1.Config.DataDbCfg().Type == utils.Internal)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if loaded, err := dbReader.LoadDestinationsFiltered(attrs.ID); err != nil {
		return utils.NewErrServerError(err)
	} else if !loaded {
		return utils.ErrNotFound
	}
	if err := apierSv1.ConnMgr.Call(apierSv1.Config.ApierCfg().CachesConns, nil,
		utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithAPIOpts{
			ArgsCache: map[string][]string{utils.DestinationIDs: {attrs.ID}},
		}, reply); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrLoadTpFromStorDb struct {
	TPid    string
	DryRun  bool // Only simulate, no write
	APIOpts map[string]interface{}
	Caching *string // Caching strategy
}

// Loads complete data in a TP from storDb
func (apierSv1 *APIerSv1) LoadTariffPlanFromStorDb(attrs *AttrLoadTpFromStorDb, reply *string) error {
	if len(attrs.TPid) == 0 {
		return utils.NewErrMandatoryIeMissing("TPid")
	}
	dbReader, err := engine.NewTpReader(apierSv1.DataManager.DataDB(), apierSv1.StorDb,
		attrs.TPid, apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().ActionSConns,
		apierSv1.Config.DataDbCfg().Type == utils.Internal)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := dbReader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
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
	if err := dbReader.ReloadCache(caching, true, attrs.APIOpts); err != nil {
		return utils.NewErrServerError(err)
	}
	if len(apierSv1.Config.ApierCfg().ActionSConns) != 0 {
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
		StorDB:   apierSv1.StorDb,
		DirPath:  attrs.FolderPath,
		Sep:      rune(attrs.CsvSeparator[0]),
		Verbose:  false,
		ImportID: attrs.RunId,
	}
	if err := csvImporter.Run(); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
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

func verifyFormat(tStr string) bool {
	if tStr == utils.EmptyString ||
		tStr == utils.MetaASAP {
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
		engine.NewFileCSVStorage(utils.CSVSep, attrs.FolderPath),
		"", apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().ActionSConns,
		apierSv1.Config.DataDbCfg().Type == utils.Internal)
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
	if err := loader.ReloadCache(caching, true, attrs.APIOpts); err != nil {
		return utils.NewErrServerError(err)
	}
	if len(apierSv1.Config.ApierCfg().ActionSConns) != 0 {
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
		engine.NewFileCSVStorage(utils.CSVSep, attrs.FolderPath), "", apierSv1.Config.GeneralCfg().DefaultTimezone,
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().ActionSConns,
		apierSv1.Config.DataDbCfg().Type == utils.Internal)
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
	if err := loader.ReloadCache(caching, true, attrs.APIOpts); err != nil {
		return utils.NewErrServerError(err)
	}
	if len(apierSv1.Config.ApierCfg().ActionSConns) != 0 {
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
		apierSv1.Config.ApierCfg().CachesConns, apierSv1.Config.ApierCfg().ActionSConns,
		apierSv1.Config.DataDbCfg().Type == utils.Internal)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := dbReader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
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
	if err := dbReader.ReloadCache(caching, true, attrs.APIOpts); err != nil {
		return utils.NewErrServerError(err)
	}
	if len(apierSv1.Config.ApierCfg().ActionSConns) != 0 {
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
	result = utils.MetaOut + utils.ConcatenatedKeySep
	if arrp.Tenant != "" && arrp.Tenant != utils.MetaAny {
		result += arrp.Tenant + utils.ConcatenatedKeySep
	} else {
		return
	}

	if arrp.Category != "" && arrp.Category != utils.MetaAny {
		result += arrp.Category + utils.ConcatenatedKeySep
	} else {
		return
	}
	if arrp.Subject != "" && arrp.Subject != utils.MetaAny {
		result += arrp.Subject
	}
	return
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
	filesInDir, _ := os.ReadDir(failedReqsInDir)
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
			utils.MetaRoutes, utils.MetaThresholds, utils.MetaRateProfiles, utils.MetaActionProfiles, utils.MetaAccountProfiles}
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
			csvWriter.Comma = utils.CSVSep
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
			csvWriter.Comma = utils.CSVSep
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
			csvWriter.Comma = utils.CSVSep
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
			csvWriter.Comma = utils.CSVSep
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
			csvWriter.Comma = utils.CSVSep
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
			csvWriter.Comma = utils.CSVSep
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
		case utils.MetaStats:
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
			csvWriter.Comma = utils.CSVSep
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
			csvWriter.Comma = utils.CSVSep
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
			csvWriter.Comma = utils.CSVSep
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
			csvWriter.Comma = utils.CSVSep
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
			csvWriter.Comma = utils.CSVSep
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
		case utils.MetaAccountProfiles:
			prfx := utils.AccountProfilePrefix
			keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
			if err != nil {
				return err
			}
			if len(keys) == 0 {
				continue
			}
			f, err := os.Create(path.Join(arg.Path, utils.AccountProfilesCsv))
			if err != nil {
				return err
			}
			defer f.Close()

			csvWriter := csv.NewWriter(f)
			csvWriter.Comma = utils.CSVSep
			//write the header of the file
			if err := csvWriter.Write(engine.AccountProfileMdls{}.CSVHeader()); err != nil {
				return err
			}
			for _, key := range keys {
				tntID := strings.SplitN(key[len(prfx):], utils.InInFieldSep, 2)
				accPrf, err := apierSv1.DataManager.GetAccountProfile(tntID[0], tntID[1])
				if err != nil {
					return err
				}
				for _, model := range engine.APItoModelTPAccountProfile(engine.AccountProfileToAPI(accPrf)) {
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
			EeIDs:    args.ExporterIDs,
			CGREvent: cdr.AsCGREvent(),
		}
		if args.Verbose {
			argCdr.CGREvent.APIOpts[utils.OptsEEsVerbose] = struct{}{}
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
