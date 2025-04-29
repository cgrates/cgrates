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

package actions

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func newActHTTPPost(ctx *context.Context, tnt string, cgrEv *utils.CGREvent,
	fltrS *engine.FilterS, cfg *config.CGRConfig, aCfg *utils.APAction) (aL *actHTTPPost, err error) {
	aL = &actHTTPPost{
		config: cfg,
		aCfg:   aCfg,
		pstrs:  make([]*ees.HTTPjsonMapEE, len(aCfg.Diktats)),
	}
	for i, actD := range aL.cfg().Diktats {
		attempts, err := engine.GetIntOpts(ctx, tnt, cgrEv.AsDataProvider(), nil, fltrS, cfg.ActionSCfg().Opts.PosterAttempts,
			utils.MetaPosterAttempts)
		if err != nil {
			return nil, err
		}
		eeCfg := config.NewEventExporterCfg(aL.id(), "", actD.Path, cfg.EEsCfg().ExporterCfg(utils.MetaDefault).FailedPostsDir,
			attempts, nil)
		aL.pstrs[i], _ = ees.NewHTTPjsonMapEE(eeCfg, cfg, nil, nil)
	}
	return
}

type actHTTPPost struct {
	config *config.CGRConfig
	aCfg   *utils.APAction

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
			go ees.ExportWithAttempts(context.Background(), pstr, &ees.HTTPPosterRequest{Body: body, Header: make(http.Header)}, utils.EmptyString,
				nil, aL.config.GeneralCfg().DefaultTenant)
		} else if err = ees.ExportWithAttempts(ctx, pstr, &ees.HTTPPosterRequest{Body: body, Header: make(http.Header)}, utils.EmptyString,
			nil, aL.config.GeneralCfg().DefaultTenant); err != nil {
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
	var rply map[string]map[string]any
	return aL.connMgr.Call(ctx, aL.config.ActionSCfg().EEsConns,
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
