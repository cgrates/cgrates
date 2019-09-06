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

func (cS *CoreService) Status(arg *utils.TenantWithArgDispatcher, reply *map[string]interface{}) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	response := make(map[string]interface{})
	response[utils.NodeID] = config.CgrConfig().GeneralCfg().NodeID
	response[utils.MemoryUsage] = utils.SizeFmt(float64(memstats.HeapAlloc), "")
	response[utils.ActiveGoroutines] = runtime.NumGoroutine()
	response[utils.Version] = utils.GetCGRVersion()
	response[utils.RunningSince] = utils.GetStartTime()
	response[utils.GoVersion] = runtime.Version()
	// in case we need this info in the future
	// response["Extra MemInfo"] = map[string]interface{}{
	// 	"Alloc":       utils.SizeFmt((float64)(memstats.Alloc), ""),
	// 	"TotalAlloc":  utils.SizeFmt((float64)(memstats.TotalAlloc), ""),
	// 	"Sys":         utils.SizeFmt((float64)(memstats.Sys), ""),
	// 	"Mallocs":     utils.SizeFmt((float64)(memstats.Mallocs), ""),
	// 	"Frees":       utils.SizeFmt((float64)(memstats.Frees), ""),
	// 	"HeapAlloc":   utils.SizeFmt((float64)(memstats.HeapAlloc), ""),
	// 	"HeapSys":     utils.SizeFmt((float64)(memstats.HeapSys), ""),
	// 	"HeapIdle":    utils.SizeFmt((float64)(memstats.HeapIdle), ""),
	// 	"HeapInuse":   utils.SizeFmt((float64)(memstats.HeapInuse), ""),
	// 	"StackInuse":  utils.SizeFmt((float64)(memstats.StackInuse), ""),
	// 	"StackSys":    utils.SizeFmt((float64)(memstats.StackSys), ""),
	// 	"MSpanInuse":  utils.SizeFmt((float64)(memstats.MSpanInuse), ""),
	// 	"MSpanSys":    utils.SizeFmt((float64)(memstats.MSpanSys), ""),
	// 	"MCacheInuse": utils.SizeFmt((float64)(memstats.MCacheInuse), ""),
	// 	"MCacheSys":   utils.SizeFmt((float64)(memstats.MCacheSys), ""),
	// 	"BuckHashSys": utils.SizeFmt((float64)(memstats.BuckHashSys), ""),
	// 	"GCSys":       utils.SizeFmt((float64)(memstats.GCSys), ""),
	// 	"OtherSys":    utils.SizeFmt((float64)(memstats.OtherSys), ""),
	// 	"NextGC":      utils.SizeFmt((float64)(memstats.NextGC), ""),
	// }
	*reply = response
	return
}
