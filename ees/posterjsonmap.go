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
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewPosterJSONMapEE(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc *utils.SafeMapStorage) (pstrJSON *PosterJSONMapEE, err error) {
	pstrJSON = &PosterJSONMapEE{
		id:      cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  cfgIdx,
		filterS: filterS,
		dc:      dc,
	}
	switch cgrCfg.EEsCfg().Exporters[cfgIdx].Type {
	case utils.MetaAMQPjsonMap:
		pstrJSON.poster = engine.NewAMQPPoster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts)
	case utils.MetaAMQPV1jsonMap:
		pstrJSON.poster = engine.NewAMQPv1Poster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts)
	case utils.MetaSQSjsonMap:
		pstrJSON.poster = engine.NewSQSPoster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts)
	case utils.MetaKafkajsonMap:
		pstrJSON.poster = engine.NewKafkaPoster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts)
	case utils.MetaS3jsonMap:
		pstrJSON.poster = engine.NewS3Poster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts)
	case utils.MetaNatsjsonMap:
		pstrJSON.poster, err = engine.NewNatsPoster(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts, cgrCfg.EEsCfg().Exporters[cfgIdx].Opts,
			cgrCfg.GeneralCfg().NodeID, cgrCfg.GeneralCfg().ConnectTimeout)
	}
	return
}

// PosterJSONMapEE implements EventExporter interface for .csv files
type PosterJSONMapEE struct {
	id      string
	cgrCfg  *config.CGRConfig
	cfgIdx  int // index of config instance within ERsCfg.Readers
	filterS *engine.FilterS
	poster  engine.Poster
	dc      *utils.SafeMapStorage
}

// ID returns the identificator of this exporter
func (pstrEE *PosterJSONMapEE) ID() string {
	return pstrEE.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (pstrEE *PosterJSONMapEE) OnEvicted(string, interface{}) {
	pstrEE.poster.Close()
}

// ExportEvent implements EventExporter
func (pstrEE *PosterJSONMapEE) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	defer func() {
		updateEEMetrics(pstrEE.dc, cgrEv.ID, cgrEv.Event, err != nil, utils.FirstNonEmpty(pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Timezone,
			pstrEE.cgrCfg.GeneralCfg().DefaultTimezone))
	}()
	pstrEE.dc.Lock()
	pstrEE.dc.MapStorage[utils.NumberOfEvents] = pstrEE.dc.MapStorage[utils.NumberOfEvents].(int64) + 1
	pstrEE.dc.Unlock()

	valMp := make(map[string]interface{})
	if len(pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].ContentFields()) == 0 {
		valMp = cgrEv.Event
	} else {
		oNm := map[string]*utils.OrderedNavigableMap{
			utils.MetaExp: utils.NewOrderedNavigableMap(),
		}
		eeReq := engine.NewExportRequest(map[string]utils.DataStorage{
			utils.MetaReq:  utils.MapStorage(cgrEv.Event),
			utils.MetaDC:   pstrEE.dc,
			utils.MetaOpts: utils.MapStorage(cgrEv.APIOpts),
			utils.MetaCfg:  pstrEE.cgrCfg.GetDataProvider(),
		}, utils.FirstNonEmpty(cgrEv.Tenant, pstrEE.cgrCfg.GeneralCfg().DefaultTenant),
			pstrEE.filterS, oNm)

		if err = eeReq.SetFields(pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].ContentFields()); err != nil {
			return
		}
		for el := eeReq.ExpData[utils.MetaExp].GetFirstElement(); el != nil; el = el.Next() {
			path := el.Value
			nmIt, _ := eeReq.ExpData[utils.MetaExp].Field(path)
			path = path[:len(path)-1] // remove the last index
			valMp[strings.Join(path, utils.NestingSep)] = nmIt.String()
		}
	}

	cgrID := utils.FirstNonEmpty(engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.CGRID), utils.GenUUID())
	runID := utils.FirstNonEmpty(engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.RunID), utils.MetaDefault)
	var body []byte
	if body, err = json.Marshal(valMp); err != nil {
		return
	}
	if err = pstrEE.poster.Post(body, utils.ConcatenatedKey(cgrID, runID)); err != nil &&
		pstrEE.cgrCfg.GeneralCfg().FailedPostsDir != utils.MetaNone {
		engine.AddFailedPost(pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].ExportPath,
			pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Type, utils.EventExporterS, body,
			pstrEE.cgrCfg.EEsCfg().Exporters[pstrEE.cfgIdx].Opts)
	}
	return
}

func (pstrEE *PosterJSONMapEE) GetMetrics() *utils.SafeMapStorage {
	return pstrEE.dc.Clone()
}
