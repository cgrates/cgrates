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

//CDR methods
//get
func (v1ms *mongoMigrator) getV1CDR() (v1Cdr *v1Cdrs, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.session.DB(v1ms.db).C(engine.ColCDRs).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v1Cdr)

	if v1Cdr == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v1Cdr, nil
}

//set
func (v1ms *mongoMigrator) setV1CDR(v1Cdr *v1Cdrs) (err error) {
	if err = v1ms.session.DB(v1ms.db).C(engine.ColCDRs).Insert(v1Cdr); err != nil {
		return err
	}
	return
}

//SMCost methods
//get
func (v1ms *mongoMigrator) getSMCost() (v2Cost *v2SessionsCost, err error) {
	if v1ms.qryIter == nil {
		v1ms.qryIter = v1ms.session.DB(v1ms.db).C(utils.SessionsCostsTBL).Find(nil).Iter()
	}
	v1ms.qryIter.Next(&v2Cost)

	if v2Cost == nil {
		v1ms.qryIter = nil
		return nil, utils.ErrNoMoreData

	}
	return v2Cost, nil
}

//set
func (v1ms *mongoMigrator) setSMCost(v2Cost *v2SessionsCost) (err error) {
	if err = v1ms.session.DB(v1ms.db).C(utils.SessionsCostsTBL).Insert(v2Cost); err != nil {
		return err
	}
	return
}

//remove
func (v1ms *mongoMigrator) remSMCost(v2Cost *v2SessionsCost) (err error) {
	if err = v1ms.session.DB(v1ms.db).C(utils.SessionsCostsTBL).Remove(nil); err != nil {
		return err
	}
	return
}
