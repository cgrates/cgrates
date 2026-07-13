/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package analyzers

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/blevesearch/bleve/v2/index/scorch"
	"github.com/blevesearch/bleve/v2/index/upsidedown"
	"github.com/blevesearch/bleve/v2/index/upsidedown/store/boltdb"
	"github.com/blevesearch/bleve/v2/index/upsidedown/store/goleveldb"
	"github.com/blevesearch/bleve/v2/index/upsidedown/store/moss"
	"github.com/cgrates/cgrates/utils"
)

func TestGetIndex(t *testing.T) {
	expIdxType, expStore := scorch.Name, scorch.Name
	idxType, store := getIndex(utils.MetaScorch)
	if idxType != expIdxType {
		t.Errorf("Expected index: %q,received:%q", expIdxType, idxType)
	}
	if store != expStore {
		t.Errorf("Expected index: %q,received:%q", expStore, store)
	}

	expIdxType, expStore = upsidedown.Name, boltdb.Name
	idxType, store = getIndex(utils.MetaBoltdb)
	if idxType != expIdxType {
		t.Errorf("Expected index: %q,received:%q", expIdxType, idxType)
	}
	if store != expStore {
		t.Errorf("Expected index: %q,received:%q", expStore, store)
	}

	expIdxType, expStore = upsidedown.Name, goleveldb.Name
	idxType, store = getIndex(utils.MetaLeveldb)
	if idxType != expIdxType {
		t.Errorf("Expected index: %q,received:%q", expIdxType, idxType)
	}
	if store != expStore {
		t.Errorf("Expected index: %q,received:%q", expStore, store)
	}

	expIdxType, expStore = upsidedown.Name, moss.Name
	idxType, store = getIndex(utils.MetaMoss)
	if idxType != expIdxType {
		t.Errorf("Expected index: %q,received:%q", expIdxType, idxType)
	}
	if store != expStore {
		t.Errorf("Expected index: %q,received:%q", expStore, store)
	}
}

func TestNewInfoRPC(t *testing.T) {
	t1 := time.Now()
	expIdx := &InfoRPC{
		RequestDuration:    time.Second,
		RequestStartTime:   t1.UTC().Format(time.RFC3339Nano),
		RequestEncoding:    utils.MetaJSON,
		RequestSource:      "127.0.0.1:5565",
		RequestDestination: "127.0.0.1:2012",
		RequestID:          0,
		RequestMethod:      utils.CoreSv1Status,
		RequestParams:      `"status"`,
		Reply:              `"result"`,
		ReplyError:         "error",
	}
	idx := NewInfoRPC(0, utils.CoreSv1Status, "status", "result", "error",
		utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012", t1, t1.Add(time.Second))
	if !reflect.DeepEqual(expIdx, idx) {
		t.Errorf("Expected:%s, received:%s", utils.ToJSON(expIdx), utils.ToJSON(idx))
	}
	idx = NewInfoRPC(0, utils.CoreSv1Status, "status", "result", errors.New("error"),
		utils.MetaJSON, "127.0.0.1:5565", "127.0.0.1:2012", t1, t1.Add(time.Second))
	if !reflect.DeepEqual(expIdx, idx) {
		t.Errorf("Expected:%s, received:%s", utils.ToJSON(expIdx), utils.ToJSON(idx))
	}

}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want any
	}{
		{name: "empty"},
		{
			name: "trailing whitespace",
			raw:  "{\"number\":94}\n\t ",
			want: map[string]any{"number": json.Number("94")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalJSON(json.RawMessage(tt.raw))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}

	for _, raw := range []string{`true false`, `{} trailing`} {
		if _, err := unmarshalJSON(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}

	var syntaxErr *json.SyntaxError
	if _, err := unmarshalJSON(json.RawMessage(`a`)); !errors.As(err, &syntaxErr) {
		t.Fatalf("got %v, want JSON syntax error", err)
	}
}
