// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package actions

import (
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// actLogger will log data to CGRateS logger
type actLog struct {
	aCfg *utils.APAction
}

func (aL *actLog) id() string {
	return aL.aCfg.ID
}

func (aL *actLog) cfg() *utils.APAction {
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
	cache   *engine.CacheS
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager
	aCfg    *utils.APAction
}

func (aL *actCDRLog) id() string {
	return aL.aCfg.ID
}

func (aL *actCDRLog) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actCDRLog) execute(ctx *context.Context, data utils.MapStorage, tnt string) (err error) {
	cdrsConns, err := engine.GetConnIDs(ctx, aL.config.ActionSCfg().Conns, utils.MetaCDRs, tnt, data, nil, aL.fltrS)
	if err != nil {
		return
	}
	if len(cdrsConns) == 0 {
		return fmt.Errorf("no connection with CDR Server")
	}
	template := aL.config.TemplatesCfg()[utils.MetaCdrLog]
	if id, has := aL.cfg().Opts[utils.MetaTemplateID]; has { // if templateID is not present we use default template
		template = aL.config.TemplatesCfg()[utils.IfaceAsString(id)]
	}
	// split the data into Request and Opts to send as parameters to AgentRequest
	reqNm := utils.MapStorage(data[utils.MetaReq].(map[string]any)).Clone()
	optsMS := utils.MapStorage(data[utils.MetaOpts].(map[string]any)).Clone()
	if optsMS == nil {
		optsMS = utils.MapStorage{}
	}
	optsMS[utils.MetaChargers] = false // do not try to get the chargers for cdrlog
	oNm := map[string]*utils.OrderedNavigableMap{
		utils.MetaCDR: utils.NewOrderedNavigableMap(),
	}
	// construct an AgentRequest so we can build the reply and send it to CDRServer
	cdrLogReq := ees.NewExportRequest(map[string]utils.DataStorage{
		utils.MetaReq:  reqNm,
		utils.MetaOpts: optsMS,
		utils.MetaCfg:  aL.config.GetDataProvider(),
	}, aL.config.GeneralCfg().DefaultTenant,
		aL.cache, aL.fltrS, oNm,
		aL.config.GeneralCfg().RoundingDecimals, aL.config.GeneralCfg().DefaultTimezone)

	if err = cdrLogReq.SetFields(ctx, template); err != nil {
		return
	}
	var rply string
	return aL.connMgr.Call(ctx, cdrsConns,
		utils.CDRsV1ProcessEvent,
		utils.NMAsCGREvent(cdrLogReq.ExpData[utils.MetaCDR], aL.config.GeneralCfg().DefaultTenant,
			utils.NestingSep, optsMS),
		&rply)
}
