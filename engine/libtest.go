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
	"bytes"
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/kr/pty"
	"io"
	"os"
	"os/exec"
	"path"
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
		if err := db.Flush(""); err != nil {
			return err
		}
	}
	ratingDb.CacheRating(nil, nil, nil, nil, nil)
	accountDb.CacheAccounting(nil, nil, nil, nil)
	return nil
}

func InitCdrDb(cfg *config.CGRConfig) error {
	storDb, err := ConfigureLoadStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding,
		cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns)
	if err != nil {
		return err
	}
	if err := storDb.Flush(path.Join(cfg.DataFolderPath, "storage", cfg.StorDBType)); err != nil {
		return err
	}
	return nil
}

// Return reference towards the command started so we can stop it if necessary
func StartEngine(cfgPath string, waitEngine int) (*exec.Cmd, error) {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		return nil, err
	}
	engine := exec.Command(enginePath, "-config_dir", cfgPath)
	if err := engine.Start(); err != nil {
		return nil, err
	}
	time.Sleep(time.Duration(waitEngine) * time.Millisecond) // Give time to rater to fire up
	return engine, nil
}

func KillEngine(waitEngine int) error {
	if err := exec.Command("pkill", "cgr-engine").Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitEngine) * time.Millisecond)
	return nil
}

func StopStartEngine(cfgPath string, waitEngine int) (*exec.Cmd, error) {
	KillEngine(waitEngine)
	return StartEngine(cfgPath, waitEngine)
}

func LoadTariffPlanFromFolder(tpPath string, ratingDb RatingStorage, accountingDb AccountingStorage) error {
	loader := NewFileCSVReader(ratingDb, accountingDb, utils.CSV_SEP,
		path.Join(tpPath, utils.DESTINATIONS_CSV),
		path.Join(tpPath, utils.TIMINGS_CSV),
		path.Join(tpPath, utils.RATES_CSV),
		path.Join(tpPath, utils.DESTINATION_RATES_CSV),
		path.Join(tpPath, utils.RATING_PLANS_CSV),
		path.Join(tpPath, utils.RATING_PROFILES_CSV),
		path.Join(tpPath, utils.SHARED_GROUPS_CSV),
		path.Join(tpPath, utils.LCRS_CSV),
		path.Join(tpPath, utils.ACTIONS_CSV),
		path.Join(tpPath, utils.ACTION_PLANS_CSV),
		path.Join(tpPath, utils.ACTION_TRIGGERS_CSV),
		path.Join(tpPath, utils.ACCOUNT_ACTIONS_CSV),
		path.Join(tpPath, utils.DERIVED_CHARGERS_CSV),
		path.Join(tpPath, utils.CDR_STATS_CSV))
	if err := loader.LoadAll(); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	if err := loader.WriteToDatabase(false, false); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	return nil
}

type PjsuaAccount struct {
	Id, Username, Password, Realm, Registrar string
}

// Returns file reference where we can write to control pjsua in terminal
func StartPjsuaListener(acnts []*PjsuaAccount, waitMs int) (*os.File, error) {
	cmdArgs := []string{"--local-port=5070", "--null-audio", "--auto-answer=200", "--max-calls=32", "--app-log-level=0"}
	for idx, acnt := range acnts {
		if idx != 0 {
			cmdArgs = append(cmdArgs, "--next-account")
		}
		cmdArgs = append(cmdArgs, "--id="+acnt.Id, "--registrar="+acnt.Registrar, "--username="+acnt.Username, "--password="+acnt.Password, "--realm="+acnt.Realm)
	}
	pjsuaPath, err := exec.LookPath("pjsua")
	if err != nil {
		return nil, err
	}
	pjsua := exec.Command(pjsuaPath, cmdArgs...)
	fPty, err := pty.Start(pjsua)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	io.Copy(os.Stdout, buf)                              // Free the content since otherwise pjsua will not start
	time.Sleep(time.Duration(waitMs) * time.Millisecond) // Give time to rater to fire up
	return fPty, nil
}

func PjsuaCallUri(acnt *PjsuaAccount, dstUri, outboundUri string, callDur time.Duration, localPort int) error {
	cmdArgs := []string{"--null-audio", "--app-log-level=0", fmt.Sprintf("--local-port=%d", localPort), fmt.Sprintf("--duration=%d", int(callDur.Seconds())),
		"--outbound=" + outboundUri, "--id=" + acnt.Id, "--username=" + acnt.Username, "--password=" + acnt.Password, "--realm=" + acnt.Realm, dstUri}

	pjsuaPath, err := exec.LookPath("pjsua")
	if err != nil {
		return err
	}
	pjsua := exec.Command(pjsuaPath, cmdArgs...)
	fPty, err := pty.Start(pjsua)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	io.Copy(os.Stdout, buf)
	go func() {
		time.Sleep(callDur + (time.Duration(2) * time.Second))
		fPty.Write([]byte("q\n")) // Destroy the listener
	}()
	return nil
}

func KillFreeSWITCH(waitMs int) error {
	if err := exec.Command("pkill", "freeswitch").Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond)
	return nil
}

func StopFreeSWITCH(scriptPath string, waitStop int) error {
	if err := exec.Command(scriptPath, "stop").Run(); err != nil {
		return err
	}
	return nil
}

func StartFreeSWITCH(scriptPath string, waitMs int) error {
	KillFreeSWITCH(1000)
	if err := exec.Command(scriptPath, "start").Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond) // Give time to rater to fire up
	return nil
}
