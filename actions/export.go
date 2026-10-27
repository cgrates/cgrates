// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package actions

import (
	"encoding/json"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type actHTTPPost struct {
	config *config.CGRConfig
	aCfg   *utils.APAction
	fltrS  *engine.FilterS

	pstrs []*ees.HTTPjsonMapEE
}

func (aL *actHTTPPost) id() string {
	return aL.aCfg.ID
}

func (aL *actHTTPPost) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actHTTPPost) execute(ctx *context.Context, data utils.MapStorage, _ string) (err error) {
	var body []byte
	if body, err = json.Marshal(data); err != nil {
		return
	}
	var partExec bool
	for _, pstr := range aL.pstrs {
		if async, has := aL.cfg().Opts[utils.MetaAsync]; has && utils.IfaceAsString(async) == utils.TrueStr {
			go ees.ExportWithAttempts(context.Background(), pstr, &ees.HTTPPosterRequest{Body: body, Header: pstr.Headers()}, utils.EmptyString,
				nil, aL.config.GeneralCfg().DefaultTenant, aL.fltrS)
		} else if err = ees.ExportWithAttempts(ctx, pstr, &ees.HTTPPosterRequest{Body: body, Header: pstr.Headers()}, utils.EmptyString,
			nil, aL.config.GeneralCfg().DefaultTenant, aL.fltrS); err != nil {
			if pstr.Cfg().FailedPostsDir != utils.MetaNone {
				err = nil
			} else {
				partExec = true
			}
		}
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
	}
	return
}

type actExport struct {
	tnt     string
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
}

func (aL *actExport) id() string {
	return aL.aCfg.ID
}

func (aL *actExport) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actExport) execute(ctx *context.Context, data utils.MapStorage, _ string) (err error) {
	var exporterIDs []string
	if expIDs, has := aL.cfg().Opts[utils.MetaExporterIDs]; has {
		exporterIDs = strings.Split(utils.IfaceAsString(expIDs), utils.InfieldSep)
	}
	var eesConns []string
	if eesConns, err = engine.GetConnIDs(ctx, aL.config.ActionSCfg().Conns, utils.MetaEEs, aL.tnt, data, nil, aL.fltrS); err != nil {
		return
	}
	var rply map[string]map[string]any
	return aL.connMgr.Call(ctx, eesConns,
		utils.EeSv1ProcessEvent, &utils.CGREventWithEeIDs{
			EeIDs: exporterIDs,
			CGREvent: &utils.CGREvent{
				Tenant:  aL.tnt,
				ID:      utils.GenUUID(),
				Event:   data[utils.MetaReq].(map[string]any),
				APIOpts: data[utils.MetaOpts].(map[string]any),
			},
		}, &rply)
}
