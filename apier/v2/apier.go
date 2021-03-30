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

package v2

import (
	"fmt"
	"os"
	"strings"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type APIerSv2 struct {
	v1.APIerSv1
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (apiv2 *APIerSv2) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(apiv2, serviceMethod, args, reply)
}

type AttrLoadAccountActions struct {
	TPid             string
	AccountActionsId string
}

func (apiv2 *APIerSv2) LoadTariffPlanFromFolder(attrs *utils.AttrLoadTpFromFolder, reply *utils.LoadInstance) error {
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
	loader, err := engine.NewTpReader(apiv2.DataManager.DataDB(),
		engine.NewFileCSVStorage(utils.CSVSep, attrs.FolderPath), "", apiv2.Config.GeneralCfg().DefaultTimezone,
		apiv2.Config.ApierCfg().CachesConns, apiv2.Config.ApierCfg().ActionConns,
		apiv2.Config.DataDbCfg().Type == utils.INTERNAL)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.DryRun {
		*reply = utils.LoadInstance{RatingLoadID: utils.DryRunCfg, AccountingLoadID: utils.DryRunCfg}
		return nil // Mission complete, no errors
	}

	if err := loader.WriteToDatabase(false, false); err != nil {
		return utils.NewErrServerError(err)
	}

	utils.Logger.Info("APIerSv2.LoadTariffPlanFromFolder, reloading cache.")
	//verify If Caching is present in arguments
	caching := config.CgrConfig().GeneralCfg().DefaultCaching
	if attrs.Caching != nil {
		caching = *attrs.Caching
	}
	if err := loader.ReloadCache(caching, true, attrs.APIOpts); err != nil {
		return utils.NewErrServerError(err)
	}
	if len(apiv2.Config.ApierCfg().ActionConns) != 0 {
		utils.Logger.Info("APIerSv2.LoadTariffPlanFromFolder, reloading scheduler.")
		if err := loader.ReloadScheduler(true); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	// release the reader with it's structures
	loader.Init()
	loadHistList, err := apiv2.DataManager.DataDB().GetLoadHistory(1, true, utils.NonTransactional)
	if err != nil {
		return err
	}
	if len(loadHistList) > 0 {
		*reply = *loadHistList[0]
	}
	return nil
}

type AttrGetActionsCount struct{}

// GetActionsCount sets in reply var the total number of actions registered for the received tenant
// returns ErrNotFound in case of 0 actions
func (apiv2 *APIerSv2) GetActionsCount(attr *AttrGetActionsCount, reply *int) (err error) {
	var actionKeys []string
	if actionKeys, err = apiv2.DataManager.DataDB().GetKeysForPrefix(utils.ActionPrefix); err != nil {
		return err
	}
	*reply = len(actionKeys)
	if len(actionKeys) == 0 {
		return utils.ErrNotFound
	}
	return nil
}

type AttrGetDestinations struct {
	DestinationIDs []string
}

// GetDestinations returns a list of destination based on the destinationIDs given
func (apiv2 *APIerSv2) GetDestinations(attr *AttrGetDestinations, reply *[]*engine.Destination) (err error) {
	if len(attr.DestinationIDs) == 0 {
		// get all destination ids
		if attr.DestinationIDs, err = apiv2.DataManager.DataDB().GetKeysForPrefix(utils.DestinationPrefix); err != nil {
			return
		}
		for i, destID := range attr.DestinationIDs {
			attr.DestinationIDs[i] = destID[len(utils.DestinationPrefix):]
		}
	}
	dests := make([]*engine.Destination, len(attr.DestinationIDs))
	for i, destID := range attr.DestinationIDs {
		if dests[i], err = apiv2.DataManager.GetDestination(destID, true, true, utils.NonTransactional); err != nil {
			return
		}
	}
	*reply = dests
	return
}

// Ping return pong if the service is active
func (apiv2 *APIerSv2) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}
