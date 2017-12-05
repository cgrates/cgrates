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
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

type AliasEntry struct {
	FieldName string
	Initial   string
	Alias     string
}

type AliasProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval    // Activation interval
	Aliases            map[string]map[string]string // map[FieldName][InitialValue]AliasValue
	Weight             float64
}

func (tp *AliasProfile) TenantID() string {
	return utils.ConcatenatedKey(tp.Tenant, tp.ID)
}

func NewAliasService(dm *DataManager, filterS *FilterS, indexedFields []string) (*AliasService, error) {
	return &AliasService{dm: dm, filterS: filterS, indexedFields: indexedFields}, nil
}

type AliasService struct {
	dm            *DataManager
	filterS       *FilterS
	indexedFields []string
}

// ListenAndServe will initialize the service
func (alS *AliasService) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info("Starting Alias Service")
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return
}

// Shutdown is called to shutdown the service
func (alS *AliasService) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.AliasS))
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.AliasS))
	return
}
