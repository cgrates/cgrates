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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewHTTPjsonMapEE(cfg *config.EventExporterCfg, cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	dc *utils.SafeMapStorage) (pstrJSON *HTTPjsonMapEE, err error) {
	pstrJSON = &HTTPjsonMapEE{
		cfg:    cfg,
		dc:     dc,
		client: &http.Client{Transport: engine.GetHTTPPstrTransport(), Timeout: cgrCfg.GeneralCfg().ReplyTimeout},
		reqs:   newConcReq(cfg.ConcurrentRequests),
	}
	pstrJSON.hdr, err = pstrJSON.composeHeader(cgrCfg, filterS)
	return
}

// HTTPjsonMapEE implements EventExporter interface for .csv files
type HTTPjsonMapEE struct {
	cfg    *config.EventExporterCfg
	dc     *utils.SafeMapStorage
	client *http.Client
	reqs   *concReq

	hdr http.Header
}

// Compose and cache the header
func (httpEE *HTTPjsonMapEE) composeHeader(cgrCfg *config.CGRConfig, filterS *engine.FilterS) (hdr http.Header, err error) {
	hdr = make(http.Header)
	if len(httpEE.Cfg().HeaderFields()) == 0 {
		return
	}
	var exp *utils.OrderedNavigableMap
	if exp, err = composeHeaderTrailer(utils.MetaHdr, httpEE.Cfg().HeaderFields(), httpEE.dc, cgrCfg, filterS); err != nil {
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

func (httpEE *HTTPjsonMapEE) Cfg() *config.EventExporterCfg { return httpEE.cfg }

func (httpEE *HTTPjsonMapEE) Connect() (_ error) { return }

func (httpEE *HTTPjsonMapEE) ExportEvent(content interface{}, _ string) (err error) {
	httpEE.reqs.get()
	defer httpEE.reqs.done()
	pReq := content.(httpPosterRequest)
	var req *http.Request
	if req, err = prepareRequest(httpEE.Cfg().ExportPath, utils.ContentJSON, pReq.Body, pReq.Header); err != nil {
		return
	}
	_, err = sendHTTPReq(httpEE.client, req)
	return
}

func (httpEE *HTTPjsonMapEE) Close() (_ error) { return }

func (httpEE *HTTPjsonMapEE) GetMetrics() *utils.SafeMapStorage { return httpEE.dc }

func (httpEE *HTTPjsonMapEE) PrepareMap(mp map[string]interface{}) (interface{}, error) {
	body, err := json.Marshal(mp)
	return &httpPosterRequest{
		Header: httpEE.hdr,
		Body:   body,
	}, err
}

func (httpEE *HTTPjsonMapEE) PrepareOrderMap(mp *utils.OrderedNavigableMap) (interface{}, error) {
	valMp := make(map[string]interface{})
	for el := mp.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := mp.Field(path)
		path = path[:len(path)-1] // remove the last index
		valMp[strings.Join(path, utils.NestingSep)] = nmIt.String()
	}
	body, err := json.Marshal(valMp)
	return &httpPosterRequest{
		Header: httpEE.hdr,
		Body:   body,
	}, err
}

func prepareRequest(addr, cType string, content interface{}, hdr http.Header) (req *http.Request, err error) {
	var body io.Reader
	if cType == utils.ContentForm {
		body = strings.NewReader(content.(url.Values).Encode())
	} else {
		body = bytes.NewBuffer(content.([]byte))
	}
	contentType := "application/x-www-form-urlencoded"
	if cType == utils.ContentJSON {
		contentType = "application/json"
	}
	hdr.Set("Content-Type", contentType)
	if req, err = http.NewRequest(http.MethodPost, addr, body); err != nil {
		return
	}
	req.Header = hdr
	return
}

func sendHTTPReq(client *http.Client, req *http.Request) (respBody []byte, err error) {
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return
	}
	respBody, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	if resp.StatusCode > 299 {
		err = fmt.Errorf("unexpected status code received: <%d>", resp.StatusCode)
	}
	return
}
