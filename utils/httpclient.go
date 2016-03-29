/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	CONTENT_JSON = "json"
	CONTENT_FORM = "form"
	CONTENT_TEXT = "text"
)

// Converts interface to []byte
func GetBytes(content interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(content)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Post without automatic failover
func HttpJsonPost(url string, skipTlsVerify bool, content interface{}) ([]byte, error) {
	body, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTlsVerify},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		return respBody, fmt.Errorf("Unexpected status code received: %d", resp.StatusCode)
	}
	return respBody, nil
}

// Post with built-in failover
func HttpPoster(addr string, skipTlsVerify bool, content interface{}, contentType string, attempts int, fallbackFilePath string) ([]byte, error) {
	var body []byte
	var urlData url.Values
	var err error
	switch contentType {
	case CONTENT_JSON:
		body = content.([]byte)
	case CONTENT_FORM:
		urlData = content.(url.Values)
	case CONTENT_TEXT:
		body = content.([]byte)
	default:
		err = fmt.Errorf("Unsupported ContentType: %s", contentType)
	}
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTlsVerify},
	}
	client := &http.Client{Transport: tr}
	delay := Fib()
	bodyType := "application/x-www-form-urlencoded"
	if contentType == CONTENT_JSON {
		bodyType = "application/json"
	}
	for i := 0; i < attempts; i++ {
		var resp *http.Response
		if IsSliceMember([]string{CONTENT_JSON, CONTENT_TEXT}, contentType) {
			resp, err = client.Post(addr, bodyType, bytes.NewBuffer(body))
		} else if contentType == CONTENT_FORM {
			resp, err = client.PostForm(addr, urlData)
		}
		if err != nil {
			Logger.Warning(fmt.Sprintf("<HttpPoster> Posting to : <%s>, error: <%s>", addr, err.Error()))
			time.Sleep(delay())
			continue
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Warning(fmt.Sprintf("<HttpPoster> Posting to : <%s>, error: <%s>", addr, err.Error()))
			time.Sleep(delay())
			continue
		}
		if resp.StatusCode > 299 {
			Logger.Warning(fmt.Sprintf("<HttpPoster> Posting to : <%s>, unexpected status code received: <%d>", addr, resp.StatusCode))
			time.Sleep(delay())
			continue
		}
		return respBody, nil
	}
	// If we got that far, post was not possible, write it on disk
	fileOut, err := os.Create(fallbackFilePath)
	if err != nil {
		return nil, err
	}
	defer fileOut.Close()
	if _, err := fileOut.Write(body); err != nil {
		return nil, err
	}
	return nil, nil
}
