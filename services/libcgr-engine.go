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

package services

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewCGREngineFlags() *CGREngineFlags {
	fs := flag.NewFlagSet(utils.CgrEngine, flag.ExitOnError)
	return &CGREngineFlags{
		FlagSet:           fs,
		CfgPath:           fs.String(utils.CfgPathCgr, utils.ConfigPath, "Configuration directory path"),
		Version:           fs.Bool(utils.VersionCgr, false, "Print application version and exit"),
		PidFile:           fs.String(utils.PidCgr, utils.EmptyString, "Path to write the PID file"),
		CpuPrfDir:         fs.String(utils.CpuProfDirCgr, utils.EmptyString, "Directory for CPU profiles"),
		MemPrfDir:         fs.String(utils.MemProfDirCgr, utils.EmptyString, "Directory for memory profiles"),
		MemPrfInterval:    fs.Duration(utils.MemProfIntervalCgr, 15*time.Second, "Interval between memory profile saves"),
		MemPrfMaxF:        fs.Int(utils.MemProfMaxFilesCgr, 1, "Number of memory profiles to keep (most recent)"),
		MemPrfTS:          fs.Bool(utils.MemProfTimestampCgr, false, "Add timestamp to memory profile files"),
		ScheduledShutdown: fs.String(utils.ScheduledShutdownCgr, utils.EmptyString, "Shutdown the engine after the specified duration"),
		SingleCPU:         fs.Bool(utils.SingleCpuCgr, false, "Run on a single CPU core"),
		Logger:            fs.String(utils.LoggerCfg, utils.EmptyString, "Logger type <*syslog|*stdout|*kafkaLog>"),
		NodeID:            fs.String(utils.NodeIDCfg, utils.EmptyString, "Node ID of the engine"),
		LogLevel:          fs.Int(utils.LogLevelCfg, -1, "Log level (0=emergency to 7=debug)"),
		Preload:           fs.String(utils.PreloadCgr, utils.EmptyString, "Loader IDs used to load data before engine starts"),
		CheckConfig:       fs.Bool(utils.CheckCfgCgr, false, "Verify the config without starting the engine"),
		SetVersions:       fs.Bool(utils.SetVersionsCgr, false, "Overwrite database versions (equivalent to cgr-migrator -exec=*set_versions)"),
	}
}

type CGREngineFlags struct {
	*flag.FlagSet

	CfgPath           *string
	Version           *bool
	PidFile           *string
	CpuPrfDir         *string
	MemPrfDir         *string
	MemPrfInterval    *time.Duration
	MemPrfMaxF        *int
	MemPrfTS          *bool
	ScheduledShutdown *string
	SingleCPU         *bool
	Logger            *string
	NodeID            *string
	LogLevel          *int
	Preload           *string
	CheckConfig       *bool
	SetVersions       *bool
}

func CgrWritePid(pidFile string) (err error) {
	var f *os.File
	if f, err = os.Create(pidFile); err != nil {
		err = fmt.Errorf("could not create pid file: %s", err)
		return
	}
	if _, err = f.WriteString(strconv.Itoa(os.Getpid())); err != nil {
		f.Close()
		err = fmt.Errorf("could not write pid file: %s", err)
		return
	}
	if err = f.Close(); err != nil {
		err = fmt.Errorf("could not close pid file: %s", err)
	}
	return
}

func waitForFilterS(ctx *context.Context, fsCh chan *engine.FilterS) (filterS *engine.FilterS, err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case filterS = <-fsCh:
		fsCh <- filterS
	}
	return
}

func InitConfigFromPath(ctx *context.Context, path, nodeID, logType string, logLevel int) (cfg *config.CGRConfig, err error) {
	// Init config
	if cfg, err = config.NewCGRConfigFromPath(ctx, path); err != nil {
		err = fmt.Errorf("could not parse config: <%s>", err)
		return
	}
	if cfg.ConfigDBCfg().Type != utils.MetaInternal {
		var d config.ConfigDB
		if d, err = engine.NewDataDBConn(cfg.ConfigDBCfg().Type,
			cfg.ConfigDBCfg().Host, cfg.ConfigDBCfg().Port,
			cfg.ConfigDBCfg().Name, cfg.ConfigDBCfg().User,
			cfg.ConfigDBCfg().Password, cfg.GeneralCfg().DBDataEncoding,
			cfg.ConfigDBCfg().Opts, nil); err != nil { // Cannot configure getter database, show stopper
			err = fmt.Errorf("could not configure configDB: <%s>", err)
			return
		}
		if err = cfg.LoadFromDB(ctx, d); err != nil {
			err = fmt.Errorf("could not parse config from DB: <%s>", err)
			return
		}
	}
	if nodeID != utils.EmptyString {
		cfg.GeneralCfg().NodeID = nodeID
	}
	if logLevel != -1 { // Modify the log level if provided by command arguments
		cfg.LoggerCfg().Level = logLevel
	}
	if logType != utils.EmptyString {
		cfg.LoggerCfg().Type = logType
	}
	if utils.ConcurrentReqsLimit != 0 { // used as shared variable
		cfg.CoreSCfg().Caps = utils.ConcurrentReqsLimit
	}
	if len(utils.ConcurrentReqsStrategy) != 0 {
		cfg.CoreSCfg().CapsStrategy = utils.ConcurrentReqsStrategy
	}
	config.SetCgrConfig(cfg) // Share the config object
	return
}
