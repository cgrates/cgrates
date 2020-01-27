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
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type TestContent struct {
	Var1 string
	Var2 string
}

func TestHttpJsonPoster(t *testing.T) {
	SetFailedPostCacheTTL(1)
	config.CgrConfig().GeneralCfg().FailedPostsDir = "/tmp"
	content := &TestContent{Var1: "Val1", Var2: "Val2"}
	jsn, _ := json.Marshal(content)
	pstr, err := NewHTTPPoster(true, time.Duration(2*time.Second), "http://localhost:8080/invalid", utils.CONTENT_JSON, 3)
	if err != nil {
		t.Error(err)
	}
	if err = pstr.Post(jsn, utils.EmptyString); err == nil {
		t.Error("Expected error")
	}
	addFailedPost("http://localhost:8080/invalid", utils.CONTENT_JSON, "test1", jsn)
	time.Sleep(2)
	fs, err := filepath.Glob("/tmp/test1*")
	if err != nil {
		t.Fatal(err)
	} else if len(fs) == 0 {
		t.Fatal("Expected at least one file")
	}

	ev, err := NewExportEventsFromFile(fs[0])
	if err != nil {
		t.Fatal(err)
	} else if len(ev.Events) == 0 {
		t.Fatal("Expected at least one event")
	}
	if !reflect.DeepEqual(jsn, ev.Events[0]) {
		t.Errorf("Expecting: %q, received: %q", string(jsn), ev.Events[0])
	}
}

func TestHttpBytesPoster(t *testing.T) {
	SetFailedPostCacheTTL(1)
	config.CgrConfig().GeneralCfg().FailedPostsDir = "/tmp"
	content := []byte(`Test
		Test2
		`)
	pstr, err := NewHTTPPoster(true, time.Duration(2*time.Second), "http://localhost:8080/invalid", utils.CONTENT_TEXT, 3)
	if err != nil {
		t.Error(err)
	}
	if err = pstr.Post(content, utils.EmptyString); err == nil {
		t.Error("Expected error")
	}
	addFailedPost("http://localhost:8080/invalid", utils.CONTENT_JSON, "test2", content)
	time.Sleep(2)
	fs, err := filepath.Glob("/tmp/test2*")
	if err != nil {
		t.Fatal(err)
	} else if len(fs) == 0 {
		t.Fatal("Expected at least one file")
	}
	ev, err := NewExportEventsFromFile(fs[0])
	if err != nil {
		t.Fatal(err)
	} else if len(ev.Events) == 0 {
		t.Fatal("Expected at least one event")
	}
	if !reflect.DeepEqual(content, ev.Events[0]) {
		t.Errorf("Expecting: %q, received: %q", string(content), ev.Events[0])
	}
}
