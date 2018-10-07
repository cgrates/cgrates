// +build integration

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
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type TestContent struct {
	Var1 string
	Var2 string
}

func TestHttpJsonPoster(t *testing.T) {

	content := &TestContent{Var1: "Val1", Var2: "Val2"}
	jsn, _ := json.Marshal(content)
	filePath := "/tmp/cgr_test_http_poster.json"
	if _, err := NewHTTPPoster(true, time.Duration(2*time.Second)).Post("http://localhost:8080/invalid", utils.CONTENT_JSON, jsn, 3, filePath); err != nil {
		t.Error(err)
	}
	if readBytes, err := ioutil.ReadFile(filePath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsn, readBytes) {
		t.Errorf("Expecting: %q, received: %q", string(jsn), string(readBytes))
	}
	if err := os.Remove(filePath); err != nil {
		t.Error("Failed removing file: ", filePath)
	}
}

func TestHttpBytesPoster(t *testing.T) {

	content := []byte(`Test
		Test2
		`)
	filePath := "/tmp/test_http_poster.http"
	if _, err := NewHTTPPoster(true, time.Duration(2*time.Second)).Post("http://localhost:8080/invalid", utils.CONTENT_TEXT, content, 3, filePath); err != nil {
		t.Error(err)
	}
	if readBytes, err := ioutil.ReadFile(filePath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(content, readBytes) {
		t.Errorf("Expecting: %q, received: %q", string(content), string(readBytes))
	}
	if err := os.Remove(filePath); err != nil {
		t.Error("Failed removing file: ", filePath)
	}
}
