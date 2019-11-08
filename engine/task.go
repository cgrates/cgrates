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
	"net"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// Task is a one time action executed by the scheduler
type Task struct {
	Uuid      string
	AccountID string
	ActionsID string
}

func (t *Task) Execute() error {
	return (&ActionTiming{
		Uuid:       t.Uuid,
		ActionsID:  t.ActionsID,
		accountIDs: utils.StringMap{t.AccountID: true},
	}).Execute(nil, nil)
}

// String implements config.DataProvider
func (t *Task) String() string {
	return utils.ToJSON(t)
}

// AsNavigableMap implements config.DataProvider
func (t *Task) AsNavigableMap(_ []*config.FCTemplate) (nm *config.NavigableMap, err error) {
	nm = new(config.NavigableMap)
	nm.Set([]string{utils.UUID}, t.Uuid, false, false)
	nm.Set([]string{utils.AccountID}, t.AccountID, false, false)
	nm.Set([]string{utils.ActionsID}, t.ActionsID, false, false)
	return
}

// FieldAsInterface implements config.DataProvider
// ToDo: support Action fields
func (t *Task) FieldAsInterface(fldPath []string) (iface interface{}, err error) {
	return t.FieldAsString(fldPath)
}

// FieldAsInterface implements config.DataProvider
// ToDo: support Action fields
func (t *Task) FieldAsString(fldPath []string) (s string, err error) {
	if len(fldPath) == 0 {
		return
	}
	if fldPath[0] != utils.MetaAct || len(fldPath) < 2 {
		return "", utils.ErrPrefixNotFound(strings.Join(fldPath, utils.NestingSep))
	}
	switch fldPath[1] {
	case utils.UUID:
		return t.Uuid, nil
	case utils.AccountID:
		return t.AccountID, nil
	case utils.ActionsID:
		return t.ActionsID, nil
	default:
		return "", utils.ErrPrefixNotFound(strings.Join(fldPath, utils.NestingSep))
	}
}

// RemoteHost implements config.DataProvider
func (t *Task) RemoteHost() (rh net.Addr) {
	return
}
