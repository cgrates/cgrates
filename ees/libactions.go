// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"encoding/gob"
	"encoding/json"
	"net/http"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	gob.Register(new(HTTPPosterRequest))
	gob.Register(new(sqlPosterRequest))

	engine.RegisterActionFunc(utils.MetaHTTPPost, callURL)
	engine.RegisterActionFunc(utils.HttpPostAsync, callURLAsync)
	engine.RegisterActionFunc(utils.MetaPostEvent, postEvent)
}

func getOneData(ub *engine.Account, extraData any) ([]byte, error) {
	switch {
	case ub != nil:
		return json.Marshal(ub)
	case extraData != nil:
		return json.Marshal(extraData)
	}
	return nil, nil
}

func callURL(ub *engine.Account, a *engine.Action, _ engine.Actions, _ *engine.FilterS, extraData any,
	_ engine.SharedActionsData, _ engine.ActionConnCfg) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	eeCfg := config.NewEventExporterCfg(a.Id, utils.MetaHTTPjsonMap, a.ExtraParameters, config.CgrConfig().EEsCfg().FailedPosts.Dir,
		config.CgrConfig().GeneralCfg().PosterAttempts, false, nil)
	pstr, err := NewHTTPjsonMapEE(eeCfg, config.CgrConfig(), nil, nil)
	if err != nil {
		return err
	}
	err = ExportWithAttempts(pstr, &HTTPPosterRequest{Body: body, Header: make(http.Header)}, "")
	if config.CgrConfig().EEsCfg().FailedPosts.Dir != utils.MetaNone {
		err = nil
	}
	return err
}

// Does not block for posts, no error reports
func callURLAsync(ub *engine.Account, a *engine.Action, _ engine.Actions, _ *engine.FilterS, extraData any,
	_ engine.SharedActionsData, _ engine.ActionConnCfg) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	eeCfg := config.NewEventExporterCfg(a.Id, utils.MetaHTTPjsonMap, a.ExtraParameters, config.CgrConfig().EEsCfg().FailedPosts.Dir,
		config.CgrConfig().GeneralCfg().PosterAttempts, false, nil)
	pstr, err := NewHTTPjsonMapEE(eeCfg, config.CgrConfig(), nil, nil)
	if err != nil {
		return err
	}
	go ExportWithAttempts(pstr, &HTTPPosterRequest{Body: body, Header: make(http.Header)}, "")
	return nil
}

func postEvent(_ *engine.Account, a *engine.Action, _ engine.Actions, _ *engine.FilterS, extraData any,
	_ engine.SharedActionsData, _ engine.ActionConnCfg) error {
	body, err := json.Marshal(extraData)
	if err != nil {
		return err
	}
	eeCfg := config.NewEventExporterCfg(a.Id, utils.MetaHTTPjsonMap, a.ExtraParameters, config.CgrConfig().EEsCfg().FailedPosts.Dir,
		config.CgrConfig().GeneralCfg().PosterAttempts, false, nil)
	pstr, err := NewHTTPjsonMapEE(eeCfg, config.CgrConfig(), nil, nil)
	if err != nil {
		return err
	}
	err = ExportWithAttempts(pstr, &HTTPPosterRequest{Body: body, Header: make(http.Header)}, "")
	if config.CgrConfig().EEsCfg().FailedPosts.Dir != utils.MetaNone {
		err = nil
	}
	return err
}
