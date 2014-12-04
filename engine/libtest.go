/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"github.com/cgrates/cgrates/config"
	"os/exec"
	"time"
)

func InitDataDb(cfg *config.CGRConfig) error {
	ratingDb, err := ConfigureRatingStorage(cfg.RatingDBType, cfg.RatingDBHost, cfg.RatingDBPort, cfg.RatingDBName, cfg.RatingDBUser, cfg.RatingDBPass, cfg.DBDataEncoding)
	if err != nil {
		return err
	}
	accountDb, err := ConfigureAccountingStorage(cfg.AccountDBType, cfg.AccountDBHost, cfg.AccountDBPort, cfg.AccountDBName,
		cfg.AccountDBUser, cfg.AccountDBPass, cfg.DBDataEncoding)
	if err != nil {
		return err
	}
	for _, db := range []Storage{ratingDb, accountDb} {
		if err := db.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func StartEngine(cfgPath string, waitEngine int) error {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		return err
	}
	if err := StopEngine(waitEngine); err != nil {
		return err
	}
	engine := exec.Command(enginePath, "-config", cfgPath)
	if err := engine.Start(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitEngine) * time.Millisecond) // Give time to rater to fire up
	return nil
}

func StopEngine(waitEngine int) error {
	if err := exec.Command("pkill", "cgr-engine").Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitEngine) * time.Millisecond)
	return nil
}
