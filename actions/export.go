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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func newActHTTPPost(cfg *config.CGRConfig, aCfg *engine.APAction) (aL *actHTTPPost) {
	aL = &actHTTPPost{
		config: cfg,
		aCfg:   aCfg,
		pstrs:  make([]*ees.HTTPjsonMapEE, len(aCfg.Diktats)),
	}
	for i, actD := range aL.cfg().Diktats {
		aL.pstrs[i], _ = ees.NewHTTPjsonMapEE(&config.EventExporterCfg{
			ID:             aL.id(),
			ExportPath:     actD.Path,
			Attempts:       aL.config.GeneralCfg().PosterAttempts,
			FailedPostsDir: cfg.GeneralCfg().FailedPostsDir,
		}, cfg, nil, nil)
	}
	return
}

type actHTTPPost struct {
	config *config.CGRConfig
	aCfg   *engine.APAction

	pstrs []*ees.HTTPjsonMapEE
}

func (aL *actHTTPPost) id() string {
	return aL.aCfg.ID
}

func (aL *actHTTPPost) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actHTTPPost) execute(_ *context.Context, data utils.MapStorage, _ string) (err error) {
	var body []byte
	if body, err = json.Marshal(data); err != nil {
		return
	}
	var partExec bool
	for _, pstr := range aL.pstrs {
		if async, has := aL.cfg().Opts[utils.MetaAsync]; has && utils.IfaceAsString(async) == utils.TrueStr {
			go ees.ExportWithAttempts(pstr, &ees.HTTPPosterRequest{Body: body, Header: make(http.Header)}, utils.EmptyString)
		} else if err = ees.ExportWithAttempts(pstr, &ees.HTTPPosterRequest{Body: body, Header: make(http.Header)}, utils.EmptyString); err != nil {
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
	aCfg    *engine.APAction
}

func (aL *actExport) id() string {
	return aL.aCfg.ID
}

func (aL *actExport) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actExport) execute(ctx *context.Context, data utils.MapStorage, _ string) (err error) {
	var exporterIDs []string
	if expIDs, has := aL.cfg().Opts[utils.MetaExporterIDs]; has {
		exporterIDs = strings.Split(utils.IfaceAsString(expIDs), utils.InfieldSep)
	}
	var rply map[string]map[string]interface{}
	return aL.connMgr.Call(ctx, aL.config.ActionSCfg().EEsConns,
		utils.EeSv1ProcessEvent, &utils.CGREventWithEeIDs{
			EeIDs: exporterIDs,
			CGREvent: &utils.CGREvent{
				Tenant:  aL.tnt,
				ID:      utils.GenUUID(),
				Event:   data[utils.MetaReq].(map[string]interface{}),
				APIOpts: data[utils.MetaOpts].(map[string]interface{}),
			},
		}, &rply)
}
