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
	"net/http"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewHTTPjsonMapEE(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc *utils.SafeMapStorage) (pstrJSON *HTTPjsonMapEE, err error) {
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
	dc      *utils.SafeMapStorage
}

// ID returns the identificator of this exporter
func (httpEE *HTTPjsonMapEE) ID() string {
	return httpEE.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (httpEE *HTTPjsonMapEE) OnEvicted(string, interface{}) {}

// ExportEvent implements EventExporter
func (httpEE *HTTPjsonMapEE) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	defer func() {
		updateEEMetrics(httpEE.dc, cgrEv.ID, cgrEv.Event, err != nil, utils.FirstNonEmpty(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].Timezone,
			httpEE.cgrCfg.GeneralCfg().DefaultTimezone))
	}()
	httpEE.dc.Lock()
	httpEE.dc.MapStorage[utils.NumberOfEvents] = httpEE.dc.MapStorage[utils.NumberOfEvents].(int64) + 1
	httpEE.dc.Unlock()

	valMp := make(map[string]interface{})
	hdr := http.Header{}
	if len(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].ContentFields()) == 0 {
		valMp = cgrEv.Event
	} else {
		oNm := map[string]*utils.OrderedNavigableMap{
			utils.MetaExp: utils.NewOrderedNavigableMap(),
		}
		eeReq := engine.NewExportRequest(map[string]utils.DataStorage{
			utils.MetaReq:  utils.MapStorage(cgrEv.Event),
			utils.MetaDC:   httpEE.dc,
			utils.MetaOpts: utils.MapStorage(cgrEv.APIOpts),
			utils.MetaCfg:  httpEE.cgrCfg.GetDataProvider(),
		}, utils.FirstNonEmpty(cgrEv.Tenant, httpEE.cgrCfg.GeneralCfg().DefaultTenant),
			httpEE.filterS, oNm)

		if err = eeReq.SetFields(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].ContentFields()); err != nil {
			return
		}
		for el := eeReq.ExpData[utils.MetaExp].GetFirstElement(); el != nil; el = el.Next() {
			path := el.Value
			nmIt, _ := eeReq.ExpData[utils.MetaExp].Field(path)
			path = path[:len(path)-1] // remove the last index
			valMp[strings.Join(path, utils.NestingSep)] = nmIt.String()
		}
		if hdr, err = httpEE.composeHeader(); err != nil {
			return
		}
	}

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

func (httpEE *HTTPjsonMapEE) GetMetrics() *utils.SafeMapStorage {
	return httpEE.dc.Clone()
}

// Compose and cache the header
func (httpEE *HTTPjsonMapEE) composeHeader() (hdr http.Header, err error) {
	hdr = make(http.Header)
	if len(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].HeaderFields()) == 0 {
		return
	}
	oNm := map[string]*utils.OrderedNavigableMap{
		utils.MetaHdr: utils.NewOrderedNavigableMap(),
	}
	eeReq := engine.NewExportRequest(map[string]utils.DataStorage{
		utils.MetaDC:  httpEE.dc,
		utils.MetaCfg: httpEE.cgrCfg.GetDataProvider(),
	}, httpEE.cgrCfg.GeneralCfg().DefaultTenant,
		httpEE.filterS, oNm)
	if err = eeReq.SetFields(httpEE.cgrCfg.EEsCfg().Exporters[httpEE.cfgIdx].HeaderFields()); err != nil {
		return
	}
	for el := eeReq.ExpData[utils.MetaHdr].GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := eeReq.ExpData[utils.MetaHdr].Field(path) //Safe to ignore error, since the path always exists
		path = path[:len(path)-1]                           // remove the last index
		hdr.Set(strings.Join(path, utils.NestingSep), nmIt.String())
	}
	return
}
