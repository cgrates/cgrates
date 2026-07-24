// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"
	"runtime"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreService() *CoreService {
	return &CoreService{}
}

type CoreService struct {
}

// ListenAndServe will initialize the service
func (cS *CoreService) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info("Starting Core service")
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return
}

// Shutdown is called to shutdown the service
func (cS *CoreService) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.CoreS))
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.CoreS))
	return
}

func (cS *CoreService) Status(arg *utils.TenantWithArgDispatcher, reply *map[string]any) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	response := make(map[string]any)
	response[utils.NodeID] = config.CgrConfig().GeneralCfg().NodeID
	response[utils.MemoryUsage] = utils.SizeFmt(float64(memstats.HeapAlloc), "")
	response[utils.ActiveGoroutines] = runtime.NumGoroutine()
	if response[utils.Version], err = utils.GetCGRVersion(); err != nil {
		utils.Logger.Err(err.Error())
		err = nil
	}
	response[utils.RunningSince] = utils.GetStartTime()
	response[utils.GoVersion] = runtime.Version()
	*reply = response
	return
}
