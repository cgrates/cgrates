/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package actions

import (
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// actLogger will log data to CGRateS logger
type actLog struct {
	aCfg *engine.APAction
}

func (aL *actLog) id() string {
	return aL.aCfg.ID
}

func (aL *actLog) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (actLog) execute(_ *context.Context, data utils.MapStorage, _ string) (err error) {
	utils.Logger.Info(fmt.Sprintf("LOG Event: %s", data.String()))
	return
}

// actCDRLog will log data to CGRateS logger
type actCDRLog struct {
	config  *config.CGRConfig
	filterS *engine.FilterS
	connMgr *engine.ConnManager
	aCfg    *engine.APAction
}

func (aL *actCDRLog) id() string {
	return aL.aCfg.ID
}

func (aL *actCDRLog) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actCDRLog) execute(ctx *context.Context, data utils.MapStorage, _ string) (err error) {
	if len(aL.config.ActionSCfg().CDRsConns) == 0 {
		return fmt.Errorf("no connection with CDR Server")
	}
	template := aL.config.TemplatesCfg()[utils.MetaCdrLog]
	if id, has := aL.cfg().Opts[utils.MetaTemplateID]; has { // if templateID is not present we use default template
		template = aL.config.TemplatesCfg()[utils.IfaceAsString(id)]
	}
	// split the data into Request and Opts to send as parameters to AgentRequest
	reqNm := utils.MapStorage(data[utils.MetaReq].(map[string]interface{})).Clone()
	optsMS := utils.MapStorage(data[utils.MetaOpts].(map[string]interface{})).Clone()
	if optsMS == nil {
		optsMS = utils.MapStorage{}
	}
	optsMS[utils.OptsChargerS] = false // do not try to get the chargers for cdrlog
	oNm := map[string]*utils.OrderedNavigableMap{
		utils.MetaCDR: utils.NewOrderedNavigableMap(),
	}
	// construct an AgentRequest so we can build the reply and send it to CDRServer
	cdrLogReq := engine.NewExportRequest(map[string]utils.DataStorage{
		utils.MetaReq:  reqNm,
		utils.MetaOpts: optsMS,
		utils.MetaCfg:  aL.config.GetDataProvider(),
	}, aL.config.GeneralCfg().DefaultTenant,
		aL.filterS, oNm)

	if err = cdrLogReq.SetFields(ctx, template); err != nil {
		return
	}
	var rply string
	return aL.connMgr.Call(ctx, aL.config.ActionSCfg().CDRsConns,
		utils.CDRsV1ProcessEvent,
		utils.NMAsCGREvent(cdrLogReq.ExpData[utils.MetaCDR], aL.config.GeneralCfg().DefaultTenant,
			utils.NestingSep, optsMS),
		&rply)
}
