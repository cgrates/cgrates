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

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func newMapStorDBMigrator(stor engine.StorDB) (mpMig *mapStorDBMigrator) {
	return &mapStorDBMigrator{
		storDB: &stor,
		mp:     stor.(*engine.MapStorage),
	}
}

type mapStorDBMigrator struct {
	storDB   *engine.StorDB
	mp       *engine.MapStorage
	dataKeys []string
	qryIdx   *int
}

func (mpMig *mapStorDBMigrator) StorDB() engine.StorDB {
	return *mpMig.storDB
}

//CDR methods
//get
func (mpMig *mapStorDBMigrator) getV1CDR() (v1Cdr *v1Cdrs, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mpMig *mapStorDBMigrator) setV1CDR(v1Cdr *v1Cdrs) (err error) {
	return utils.ErrNotImplemented
}

//SMCost methods
//rename
func (mpMig *mapStorDBMigrator) renameV1SMCosts() (err error) {
	return utils.ErrNotImplemented
}

func (mpMig *mapStorDBMigrator) createV1SMCosts() (err error) {
	return utils.ErrNotImplemented
}

//get
func (mpMig *mapStorDBMigrator) getV2SMCost() (v2Cost *v2SessionsCost, err error) {
	return nil, utils.ErrNotImplemented
}

//set
func (mpMig *mapStorDBMigrator) setV2SMCost(v2Cost *v2SessionsCost) (err error) {
	return utils.ErrNotImplemented
}

//remove
func (mpMig *mapStorDBMigrator) remV2SMCost(v2Cost *v2SessionsCost) (err error) {
	return utils.ErrNotImplemented
}
