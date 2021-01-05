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

package engine

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type HTTPPosterRequest struct {
	Header http.Header
	Body   interface{}
}

// HTTPPostJSON posts without automatic failover
func HTTPPostJSON(url string, content []byte) (respBody []byte, err error) {
	client := &http.Client{Transport: httpPstrTransport}
	var resp *http.Response
	if resp, err = client.Post(url, "application/json", bytes.NewBuffer(content)); err != nil {
		return
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	if resp.StatusCode > 299 {
		err = fmt.Errorf("Unexpected status code received: %d", resp.StatusCode)
	}
	return
}

// NewHTTPPoster return a new HTTP poster
func NewHTTPPoster(replyTimeout time.Duration, addr, contentType string,
	attempts int) (httposter *HTTPPoster, err error) {
	if !utils.SliceHasMember([]string{utils.CONTENT_FORM, utils.ContentJSON, utils.CONTENT_TEXT}, contentType) {
		return nil, fmt.Errorf("unsupported ContentType: %s", contentType)
	}
	return &HTTPPoster{
		httpClient:  &http.Client{Transport: httpPstrTransport, Timeout: replyTimeout},
		addr:        addr,
		contentType: contentType,
		attempts:    attempts,
	}, nil
}

// HTTPPoster used to post cdrs
type HTTPPoster struct {
	httpClient  *http.Client
	addr        string
	contentType string
	attempts    int
}

// PostValues will post the event
func (pstr *HTTPPoster) PostValues(content interface{}, hdr http.Header) (err error) {
	_, err = pstr.GetResponse(content, hdr)
	return
}

// GetResponse will post the event and return the response
func (pstr *HTTPPoster) GetResponse(content interface{}, hdr http.Header) (respBody []byte, err error) {
	fib := utils.Fib()
	for i := 0; i < pstr.attempts; i++ {
		var req *http.Request
		if req, err = pstr.getRequest(content, hdr); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error creating request: <%s>", pstr.addr, err.Error()))
			return
		}
		if respBody, err = pstr.do(req); err != nil {
			if i+1 < pstr.attempts {
				time.Sleep(time.Duration(fib()) * time.Second)
			}
			continue
		}
		return
	}
	return
}

func (pstr *HTTPPoster) do(req *http.Request) (respBody []byte, err error) {
	var resp *http.Response
	if resp, err = pstr.httpClient.Do(req); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error: <%s>", pstr.addr, err.Error()))
		return
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error: <%s>", pstr.addr, err.Error()))
		return
	}
	if resp.StatusCode > 299 {
		err = fmt.Errorf("unexpected status code received: <%d>", resp.StatusCode)
		utils.Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, unexpected status code received: <%d>", pstr.addr, resp.StatusCode))
		return
	}
	return
}

func (pstr *HTTPPoster) getRequest(content interface{}, hdr http.Header) (req *http.Request, err error) {
	var body io.Reader
	if pstr.contentType == utils.CONTENT_FORM {
		body = strings.NewReader(content.(url.Values).Encode())
	} else {
		body = bytes.NewBuffer(content.([]byte))
	}
	contentType := "application/x-www-form-urlencoded"
	if pstr.contentType == utils.ContentJSON {
		contentType = "application/json"
	}
	hdr.Set("Content-Type", contentType)
	if req, err = http.NewRequest(http.MethodPost, pstr.addr, body); err != nil {
		return
	}
	req.Header = hdr
	return
}
