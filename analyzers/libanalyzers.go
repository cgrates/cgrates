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
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/blevesearch/bleve/v2/index/scorch"
	"github.com/blevesearch/bleve/v2/index/upsidedown"
	"github.com/blevesearch/bleve/v2/index/upsidedown/store/boltdb"
	"github.com/blevesearch/bleve/v2/index/upsidedown/store/goleveldb"
	"github.com/blevesearch/bleve/v2/index/upsidedown/store/moss"
	"github.com/cgrates/cgrates/utils"
)

// NewInfoRPC returns a structure to be indexed
func NewInfoRPC(id uint64, method string,
	params, result, err any,
	enc, from, to string, sTime, eTime time.Time) *InfoRPC {
	var e any
	switch val := err.(type) {
	default:
	case nil:
	case string:
		e = val
	case error:
		e = val.Error()
	}
	return &InfoRPC{
		RequestDuration:  eTime.Sub(sTime),
		RequestStartTime: sTime.UTC().Format(time.RFC3339Nano),
		// EndTime:          eTime,

		RequestEncoding:    enc,
		RequestSource:      from,
		RequestDestination: to,

		RequestID:     id,
		RequestMethod: method,
		RequestParams: utils.ToJSON(params),
		Reply:         utils.ToJSON(result),
		ReplyError:    e,
	}
}

// InfoRPC the structure to be indexed
type InfoRPC struct {
	RequestDuration  time.Duration
	RequestStartTime string
	// EndTime          time.Time

	RequestEncoding    string
	RequestSource      string
	RequestDestination string

	RequestID     uint64
	RequestMethod string
	RequestParams any
	Reply         any
	ReplyError    any
}

type rpcAPI struct {
	ID     uint64 `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
	Error  string `json:"err,omitempty"`

	StartTime time.Time
}

func getIndex(indx string) (indxType, storeType string) {
	switch indx {
	case utils.MetaScorch:
		indxType, storeType = scorch.Name, scorch.Name
	case utils.MetaBoltdb:
		indxType, storeType = upsidedown.Name, boltdb.Name
	case utils.MetaLeveldb:
		indxType, storeType = upsidedown.Name, goleveldb.Name
	case utils.MetaMoss:
		indxType, storeType = upsidedown.Name, moss.Name
	}
	return
}

func unmarshalJSON(jsn json.RawMessage) (any, error) {
	if len(jsn) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(jsn))
	decoder.UseNumber()
	var val any
	if err := decoder.Decode(&val); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, errors.New("multiple JSON values")
		}
		return nil, err
	}
	return val, nil
}

// getDPFromSearchresult will unmarshal the request and reply and populate a DataProvider
// if the req is a map[string]any we will try to put in *opts prefix the Opts field from req
func getDPFromSearchresult(req, rep json.RawMessage, hdr utils.MapStorage) (utils.MapStorage, error) {
	repDP, err := unmarshalJSON(rep)
	if err != nil {
		return nil, err
	}
	reqDP, err := unmarshalJSON(req)
	if err != nil {
		return nil, err
	}
	var opts any
	if reqMp, canCast := reqDP.(map[string]any); canCast {
		opts = reqMp[utils.Opts]
	}
	return utils.MapStorage{
		utils.MetaReq:  reqDP,
		utils.MetaOpts: opts,
		utils.MetaRep:  repDP,
		utils.MetaHdr:  hdr,
	}, nil
}
