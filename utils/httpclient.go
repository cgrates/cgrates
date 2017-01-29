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
package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	CONTENT_JSON = "json"
	CONTENT_FORM = "form"
	CONTENT_TEXT = "text"
)

// NewFallbackFileNameFronString will revert the meta information in the fallback file name into original data
func NewFallbackFileNameFronString(fileName string) (ffn *FallbackFileName, err error) {
	ffn = new(FallbackFileName)
	moduleIdx := strings.Index(fileName, "_")
	ffn.Module = fileName[:moduleIdx]
	if !IsSliceMember([]string{"cdr"}, ffn.Module) {
		return nil, fmt.Errorf("unsupported module: %s", ffn.Module)
	}
	fileNameWithoutModule := fileName[moduleIdx+1:]
	for _, trspt := range []string{MetaHTTPjsonCDR, MetaHTTPjsonMap, META_HTTP_POST} {
		if strings.HasPrefix(fileNameWithoutModule, trspt) {
			ffn.Transport = trspt
			break
		}
	}
	if ffn.Transport == "" {
		return nil, fmt.Errorf("unsupported transport in fallback file path: %s", fileName)
	}
	fileNameWithoutTransport := fileNameWithoutModule[len(ffn.Transport)+1:]
	reqIDidx := strings.LastIndex(fileNameWithoutTransport, "_")
	if reqIDidx == -1 {
		return nil, fmt.Errorf("cannot find request ID in fallback file path: %s", fileName)
	}
	if ffn.Address, err = url.QueryUnescape(fileNameWithoutTransport[:reqIDidx]); err != nil {
		return nil, err
	}
	fileNameWithoutAddress := fileNameWithoutTransport[reqIDidx+1:]
	for _, suffix := range []string{TxtSuffix, JSNSuffix, FormSuffix} {
		if strings.HasSuffix(fileNameWithoutAddress, suffix) {
			ffn.FileSuffix = suffix
			break
		}
	}
	if ffn.FileSuffix == "" {
		return nil, fmt.Errorf("unsupported suffix in fallback file path: %s", fileName)
	}
	ffn.RequestID = fileNameWithoutAddress[:len(fileNameWithoutAddress)-len(ffn.FileSuffix)]
	return
}

// FallbackFileName is the struct defining the name of a file where CGRateS will dump data which fails to be sent remotely
type FallbackFileName struct {
	Module     string // name of the module writing the file
	Transport  string // transport used to send data remotely
	Address    string // remote address where data should have been sent
	RequestID  string // unique identifier of the request which should make files unique, should not contain _ character
	FileSuffix string // informative file termination suffix
}

func (ffn *FallbackFileName) AsString() string {
	return fmt.Sprintf("%s_%s_%s_%s%s", ffn.Module, ffn.Transport, url.QueryEscape(ffn.Address), ffn.RequestID, ffn.FileSuffix)
}

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
func HttpJsonPost(url string, skipTlsVerify bool, content []byte) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: skipTlsVerify},
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(content))
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

func NewHTTPPoster(skipTLSVerify bool, replyTimeout time.Duration) *HTTPPoster {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
	}
	return &HTTPPoster{httpClient: &http.Client{Transport: tr, Timeout: replyTimeout}}
}

type HTTPPoster struct {
	httpClient *http.Client
}

// Post with built-in failover
// Returns also reference towards client so we can close it's connections when done
func (poster *HTTPPoster) Post(addr string, contentType string, content interface{}, attempts int, fallbackFilePath string) ([]byte, error) {
	if !IsSliceMember([]string{CONTENT_JSON, CONTENT_FORM, CONTENT_TEXT}, contentType) {
		return nil, fmt.Errorf("unsupported ContentType: %s", contentType)
	}
	var body []byte        // Used to write in file and send over http
	var urlVals url.Values // Used when posting form
	if IsSliceMember([]string{CONTENT_JSON, CONTENT_TEXT}, contentType) {
		body = content.([]byte)
	} else if contentType == CONTENT_FORM {
		urlVals = content.(url.Values)
		body = []byte(urlVals.Encode())
	}
	delay := Fib()
	bodyType := "application/x-www-form-urlencoded"
	if contentType == CONTENT_JSON {
		bodyType = "application/json"
	}
	var err error
	for i := 0; i < attempts; i++ {
		var resp *http.Response
		if IsSliceMember([]string{CONTENT_JSON, CONTENT_TEXT}, contentType) {
			resp, err = poster.httpClient.Post(addr, bodyType, bytes.NewBuffer(body))
		} else if contentType == CONTENT_FORM {
			resp, err = poster.httpClient.PostForm(addr, urlVals)
		}
		if err != nil {
			Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error: <%s>", addr, err.Error()))
			time.Sleep(delay())
			continue
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, error: <%s>", addr, err.Error()))
			time.Sleep(delay())
			continue
		}
		if resp.StatusCode > 299 {
			Logger.Warning(fmt.Sprintf("<HTTPPoster> Posting to : <%s>, unexpected status code received: <%d>", addr, resp.StatusCode))
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
