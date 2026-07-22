/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// NewLocker returns a Guardian locker configured for cfg.
func NewLocker(cfg *config.CGRConfig) *guardian.Locker {
	timeout := time.Duration(0)
	if cfg != nil {
		timeout = cfg.GeneralCfg().LockingTimeout
	}
	opts := []guardian.Option{guardian.WithTimeout(timeout)}
	if cfg != nil && cfg.LoggerCfg().Level >= 0 {
		opts = append(opts, guardian.WithLogger(utils.Logger))
	}
	return guardian.New(opts...)
}
