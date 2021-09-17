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

package main

import (
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/services"
	"github.com/cgrates/cgrates/utils"
)

func RunCGREngine(fs []string) (err error) {
	flags := services.NewCGREngineFlags()
	if err = flags.Parse(fs); err != nil {
		return
	}
	var vers string
	if vers, err = utils.GetCGRVersion(); err != nil {
		return
	}
	if *flags.Version {
		return
	}
	if *flags.PidFile != utils.EmptyString {
		services.CgrWritePid(*flags.PidFile)
	}
	if *flags.Singlecpu {
		runtime.GOMAXPROCS(1) // Having multiple cpus may slow down computing due to CPU management, to be reviewed in future Go releases
	}

	// Init config
	ctx, cancel := context.WithCancel(context.Background())
	var cfg *config.CGRConfig
	if cfg, err = services.InitConfigFromPath(ctx, *flags.CfgPath, *flags.NodeID, *flags.LogLevel); err != nil || *flags.CheckConfig {
		return
	}
	cps := engine.NewCaps(cfg.CoreSCfg().Caps, cfg.CoreSCfg().CapsStrategy)
	server := cores.NewServer(cps)
	cgr := services.NewCGREngine(cfg, engine.NewConnManager(cfg), new(sync.WaitGroup), server, cps)
	defer cgr.Stop(*flags.MemPrfDir, *flags.PidFile)

	if err = cgr.Init(ctx, cancel, flags, vers); err != nil {
		return
	}

	if err = cgr.StartServices(ctx, cancel, *flags.Preload); err != nil {
		return
	}
	<-ctx.Done()
	return
}

func main() {
	if err := RunCGREngine(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
