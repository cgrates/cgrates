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
	"bytes"
	"fmt"
	"io"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/kr/pty"
)

func InitDataDb(cfg *config.CGRConfig) error {
	dm, err := ConfigureDataStorage(cfg.DataDbCfg().DataDbType,
		cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort,
		cfg.DataDbCfg().DataDbName, cfg.DataDbCfg().DataDbUser,
		cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
		cfg.CacheCfg(), cfg.DataDbCfg().DataDbSentinelName)
	if err != nil {
		return err
	}
	if err := dm.DataDB().Flush(""); err != nil {
		return err
	}
	//dm.LoadDataDBCache(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	//	Write version before starting
	if err := SetDBVersions(dm.dataDB); err != nil {
		return err
	}
	return nil
}

func InitStorDb(cfg *config.CGRConfig) error {
	storDb, err := ConfigureLoadStorage(cfg.StorDbCfg().StorDBType,
		cfg.StorDbCfg().StorDBHost, cfg.StorDbCfg().StorDBPort,
		cfg.StorDbCfg().StorDBName, cfg.StorDbCfg().StorDBUser,
		cfg.StorDbCfg().StorDBPass, cfg.GeneralCfg().DBDataEncoding,
		cfg.StorDbCfg().StorDBMaxOpenConns, cfg.StorDbCfg().StorDBMaxIdleConns,
		cfg.StorDbCfg().StorDBConnMaxLifetime, cfg.StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		return err
	}
	if err := storDb.Flush(path.Join(cfg.DataFolderPath, "storage",
		cfg.StorDbCfg().StorDBType)); err != nil {
		return err
	}
	if utils.IsSliceMember([]string{utils.MYSQL, utils.POSTGRES, utils.MONGO},
		cfg.StorDbCfg().StorDBType) {
		if err := SetDBVersions(storDb); err != nil {
			return err
		}
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
	cfg, err := config.NewCGRConfigFromFolder(cfgPath)
	if err != nil {
		return nil, err
	}
	fib := utils.Fib()
	var connected bool
	for i := 0; i < 200; i++ {
		time.Sleep(time.Duration(fib()) * time.Millisecond)
		if _, err := jsonrpc.Dial("tcp", cfg.ListenCfg().RPCJSONListen); err != nil {
			utils.Logger.Warning(fmt.Sprintf("Error <%s> when opening test connection to: <%s>",
				err.Error(), cfg.ListenCfg().RPCJSONListen))
		} else {
			connected = true
			break
		}
	}
	if !connected {
		return nil, fmt.Errorf("engine did not open port <%s>", cfg.ListenCfg().RPCJSONListen)
	}
	time.Sleep(time.Duration(waitEngine) * time.Millisecond) // wait for rater to register all subsistems
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

func LoadTariffPlanFromFolder(tpPath, timezone string, dm *DataManager, disable_reverse bool) error {
	loader := NewTpReader(dm.dataDB, NewFileCSVStorage(utils.CSV_SEP,
		path.Join(tpPath, utils.DESTINATIONS_CSV),
		path.Join(tpPath, utils.TIMINGS_CSV),
		path.Join(tpPath, utils.RATES_CSV),
		path.Join(tpPath, utils.DESTINATION_RATES_CSV),
		path.Join(tpPath, utils.RATING_PLANS_CSV),
		path.Join(tpPath, utils.RATING_PROFILES_CSV),
		path.Join(tpPath, utils.SHARED_GROUPS_CSV),
		path.Join(tpPath, utils.ACTIONS_CSV),
		path.Join(tpPath, utils.ACTION_PLANS_CSV),
		path.Join(tpPath, utils.ACTION_TRIGGERS_CSV),
		path.Join(tpPath, utils.ACCOUNT_ACTIONS_CSV),
		path.Join(tpPath, utils.DERIVED_CHARGERS_CSV),

		path.Join(tpPath, utils.USERS_CSV),
		path.Join(tpPath, utils.ALIASES_CSV),
		path.Join(tpPath, utils.ResourcesCsv),
		path.Join(tpPath, utils.StatsCsv),
		path.Join(tpPath, utils.ThresholdsCsv),
		path.Join(tpPath, utils.FiltersCsv),
		path.Join(tpPath, utils.SuppliersCsv),
		path.Join(tpPath, utils.AttributesCsv),
		path.Join(tpPath, utils.ChargersCsv),
	), "", timezone)
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if err := loader.WriteToDatabase(false, false, disable_reverse); err != nil {
		return utils.NewErrServerError(err)
	}
	return nil
}

type PjsuaAccount struct {
	Id, Username, Password, Realm, Registrar string
}

// Returns file reference where we can write to control pjsua in terminal
func StartPjsuaListener(acnts []*PjsuaAccount, localPort, waitDur time.Duration) (*os.File, error) {
	cmdArgs := []string{fmt.Sprintf("--local-port=%d", localPort), "--null-audio", "--auto-answer=200", "--max-calls=32", "--app-log-level=0"}
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
	io.Copy(os.Stdout, buf) // Free the content since otherwise pjsua will not start
	time.Sleep(waitDur)     // Give time to rater to fire up
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

func KillProcName(procName string, waitMs int) error {
	if err := exec.Command("pkill", procName).Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond)
	return nil
}

func ForceKillProcName(procName string, waitMs int) error {
	if err := exec.Command("pkill", "-9", procName).Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond)
	return nil
}

func CallScript(scriptPath string, subcommand string, waitMs int) error {
	if err := exec.Command(scriptPath, subcommand).Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond) // Give time to rater to fire up
	return nil
}
