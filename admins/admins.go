// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package admins

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

func NewAdminS(cfg *config.CGRConfig, dm *engine.DataManager, connMgr *engine.ConnManager, fltrS *engine.FilterS) *AdminS {
	return &AdminS{
		cfg:     cfg,
		dm:      dm,
		connMgr: connMgr,
		fltrS:   fltrS,
	}
}

type AdminS struct {
	cfg     *config.CGRConfig
	dm      *engine.DataManager
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
}
