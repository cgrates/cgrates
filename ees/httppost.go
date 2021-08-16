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
	"net/http"
	"net/url"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewHTTPPostEE(cfg *config.EventExporterCfg, cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	dc *utils.SafeMapStorage) (httpPost *HTTPPostEE, err error) {
	httpPost = &HTTPPostEE{
		cfg:    cfg,
		dc:     dc,
		client: &http.Client{Transport: engine.GetHTTPPstrTransport(), Timeout: cgrCfg.GeneralCfg().ReplyTimeout},
		reqs:   newConcReq(cfg.ConcurrentRequests),
	}
	httpPost.hdr, err = httpPost.composeHeader(cgrCfg, filterS)
	return
}

// FileCSVee implements EventExporter interface for .csv files
type HTTPPostEE struct {
	cfg    *config.EventExporterCfg
	dc     *utils.SafeMapStorage
	client *http.Client
	reqs   *concReq

	hdr http.Header
}
type httpPosterRequest struct {
	Header http.Header
	Body   interface{}
}

// Compose and cache the header
func (httpPost *HTTPPostEE) composeHeader(cgrCfg *config.CGRConfig, filterS *engine.FilterS) (hdr http.Header, err error) {
	hdr = make(http.Header)
	if len(httpPost.Cfg().HeaderFields()) == 0 {
		return
	}
	var exp *utils.OrderedNavigableMap
	if exp, err = composeHeaderTrailer(utils.MetaHdr, httpPost.Cfg().HeaderFields(), httpPost.dc, cgrCfg, filterS); err != nil {
		return
	}
	for el := exp.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := exp.Field(path) //Safe to ignore error, since the path always exists
		path = path[:len(path)-1]  // remove the last index
		hdr.Set(strings.Join(path, utils.NestingSep), nmIt.String())
	}
	return
}

func (httpPost *HTTPPostEE) Cfg() *config.EventExporterCfg { return httpPost.cfg }

func (httpPost *HTTPPostEE) Connect() (_ error) { return }

func (httpPost *HTTPPostEE) ExportEvent(content interface{}, _ string) (err error) {
	httpPost.reqs.get()
	defer httpPost.reqs.done()
	pReq := content.(*httpPosterRequest)
	var req *http.Request
	if req, err = prepareRequest(httpPost.Cfg().ExportPath, utils.ContentForm, pReq.Body, pReq.Header); err != nil {
		return
	}
	_, err = sendHTTPReq(httpPost.client, req)
	return
}

func (httpPost *HTTPPostEE) Close() (_ error) { return }

func (httpPost *HTTPPostEE) GetMetrics() *utils.SafeMapStorage { return httpPost.dc }

func (httpPost *HTTPPostEE) PrepareMap(mp map[string]interface{}) (interface{}, error) {
	urlVals := url.Values{}
	for k, v := range mp {
		urlVals.Set(k, utils.IfaceAsString(v))
	}
	return &httpPosterRequest{
		Header: httpPost.hdr,
		Body:   urlVals,
	}, nil
}

func (httpPost *HTTPPostEE) PrepareOrderMap(mp *utils.OrderedNavigableMap) (interface{}, error) {
	urlVals := url.Values{}
	for el := mp.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := mp.Field(path)
		path = path[:len(path)-1] // remove the last index
		urlVals.Set(strings.Join(path, utils.NestingSep), nmIt.String())
	}
	return &httpPosterRequest{
		Header: httpPost.hdr,
		Body:   urlVals,
	}, nil
}
