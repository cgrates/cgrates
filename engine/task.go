// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"net"
	"strings"

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

// String implements utils.DataProvider
func (t *Task) String() string {
	return utils.ToJSON(t)
}

// FieldAsInterface implements utils.DataProvider
// ToDo: support Action fields
func (t *Task) FieldAsInterface(fldPath []string) (iface any, err error) {
	return t.FieldAsString(fldPath)
}

// FieldAsInterface implements utils.DataProvider
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

// RemoteHost implements utils.DataProvider
func (t *Task) RemoteHost() (rh net.Addr) {
	return
}
