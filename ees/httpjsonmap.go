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

package ees

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewHTTPjsonMapEE(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (pstrJSON *HTTPjsonMapEE, err error) {
	pstrJSON = &HTTPjsonMapEE{
		id:      cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  cfgIdx,
		filterS: filterS,
		dc:      dc,
	}

	pstrJSON.pstr, err = engine.NewHTTPPoster(cgrCfg.GeneralCfg().ReplyTimeout,
		cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
		utils.PosterTransportContentTypes[cgrCfg.EEsCfg().Exporters[cfgIdx].Type],
		cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts)
	return
}

// HTTPjsonMapEE implements EventExporter interface for .csv files
type HTTPjsonMapEE struct {
	id      string
	cgrCfg  *config.CGRConfig
	cfgIdx  int // index of config instance within ERsCfg.Readers
	filterS *engine.FilterS
	pstr    *engine.HTTPPoster
	dc      utils.MapStorage
	sync.RWMutex
}

// ID returns the identificator of this exporter
func (httpEE *HTTPjsonMapEE) ID() string {
	return httpEE.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (httpEE *HTTPjsonMapEE) OnEvicted(string, interface{}) {
	return
}

// ExportEvent implements EventExporter
func (httpEE *HTTPjsonMapEE) ExportEvent(cgrEv *utils.CGREventWithOpts) (err error) {
	httpEE.Lock()
	defer func() {
		if err != nil {
			httpEE.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		} else {
			httpEE.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
		}
		httpEE.Unlock()
	}()

	httpEE.dc[utils.NumberOfEvents] = httpEE.dc[utils.NumberOfEvents].(int64) + 1

	valMp := make(map[string]interface{})
	hdr := http.Header{}
	if len(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].ContentFields()) == 0 {
		valMp = cgrEv.Event
	} else {
		eeReq := NewEventExporterRequest(utils.MapStorage(cgrEv.Event), httpEE.dc, cgrEv.Opts,
			httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].Tenant,
			httpEE.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].Timezone,
				httpEE.cgrCfg.GeneralCfg().DefaultTimezone), httpEE.filterS)

		if err = eeReq.SetFields(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].ContentFields()); err != nil {
			return
		}
		for el := eeReq.cnt.GetFirstElement(); el != nil; el = el.Next() {
			var nmIt utils.NMInterface
			if nmIt, err = eeReq.cnt.Field(el.Value); err != nil {
				return
			}
			itm, isNMItem := nmIt.(*config.NMItem)
			if !isNMItem {
				err = fmt.Errorf("cannot encode reply value: %s, err: not NMItems", utils.ToJSON(el.Value))
				return
			}
			if itm == nil {
				continue // all attributes, not writable to diameter packet
			}
			valMp[strings.Join(itm.Path, utils.NestingSep)] = utils.IfaceAsString(itm.Data)
		}
		if hdr, err = httpEE.composeHeader(); err != nil {
			return
		}
	}
	updateEEMetrics(httpEE.dc, cgrEv.Event, utils.FirstNonEmpty(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].Timezone,
		httpEE.cgrCfg.GeneralCfg().DefaultTimezone))
	var body []byte
	if body, err = json.Marshal(valMp); err != nil {
		return
	}
	if err = httpEE.pstr.PostValues(body, hdr); err != nil &&
		httpEE.cgrCfg.GeneralCfg().FailedPostsDir != utils.MetaNone {
		engine.AddFailedPost(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].ExportPath,
			httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].Type, utils.EventExporterS,
			&engine.HTTPPosterRequest{Header: hdr, Body: body},
			httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].Opts)
	}
	return
}

func (httpEE *HTTPjsonMapEE) GetMetrics() utils.MapStorage {
	return httpEE.dc.Clone()
}

// Compose and cache the header
func (httpEE *HTTPjsonMapEE) composeHeader() (hdr http.Header, err error) {
	hdr = make(http.Header)
	if len(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].HeaderFields()) == 0 {
		return
	}
	eeReq := NewEventExporterRequest(nil, httpEE.dc, nil,
		httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].Tenant,
		httpEE.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].Timezone,
			httpEE.cgrCfg.GeneralCfg().DefaultTimezone),
		httpEE.filterS)
	if err = eeReq.SetFields(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].HeaderFields()); err != nil {
		return
	}
	for el := eeReq.hdr.GetFirstElement(); el != nil; el = el.Next() {
		var nmIt utils.NMInterface
		if nmIt, err = eeReq.hdr.Field(el.Value); err != nil {
			return
		}
		itm, isNMItem := nmIt.(*config.NMItem)
		if !isNMItem {
			err = fmt.Errorf("cannot encode reply value: %s, err: not NMItems", utils.ToJSON(el.Value))
			return
		}
		if itm == nil {
			continue // all attributes, not writable to diameter packet
		}
		hdr.Set(strings.Join(itm.Path, utils.NestingSep), utils.IfaceAsString(itm.Data))
	}
	return
}
