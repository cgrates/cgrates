// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func newInternalStorDBMigrator(stor engine.StorDB) (iDBMig *internalStorDBMigrator) {
	return &internalStorDBMigrator{
		storDB: &stor,
		iDB:    stor.(*engine.InternalDB),
	}
}

type internalStorDBMigrator struct {
	storDB *engine.StorDB
	iDB    *engine.InternalDB
}

func (iDBMig *internalStorDBMigrator) close() {}

func (iDBMig *internalStorDBMigrator) StorDB() engine.StorDB {
	return *iDBMig.storDB
}

// CDR methods
// get
func (iDBMig *internalStorDBMigrator) getV1CDR() (v1Cdr *v1Cdrs, err error) {
	return nil, utils.ErrNotImplemented
}

// set
func (iDBMig *internalStorDBMigrator) setV1CDR(v1Cdr *v1Cdrs) (err error) {
	return utils.ErrNotImplemented
}

// rem
func (iDBMig *internalStorDBMigrator) remV1CDRs(v1Cdr *v1Cdrs) (err error) {
	return utils.ErrNotImplemented
}

// SMCost methods
// rename
func (iDBMig *internalStorDBMigrator) renameV1SMCosts() (err error) {
	return utils.ErrNotImplemented
}

func (iDBMig *internalStorDBMigrator) createV1SMCosts() (err error) {
	return utils.ErrNotImplemented
}

// get
func (iDBMig *internalStorDBMigrator) getV2SMCost() (v2Cost *v2SessionsCost, err error) {
	return nil, utils.ErrNotImplemented
}

// set
func (iDBMig *internalStorDBMigrator) setV2SMCost(v2Cost *v2SessionsCost) (err error) {
	return utils.ErrNotImplemented
}

// remove
func (iDBMig *internalStorDBMigrator) remV2SMCost(v2Cost *v2SessionsCost) (err error) {
	return utils.ErrNotImplemented
}
