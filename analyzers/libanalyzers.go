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

package analyzers

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/blevesearch/bleve/index/scorch"
	"github.com/blevesearch/bleve/index/store/boltdb"
	"github.com/blevesearch/bleve/index/store/goleveldb"
	"github.com/blevesearch/bleve/index/store/moss"
	"github.com/blevesearch/bleve/index/upsidedown"
	"github.com/cgrates/cgrates/utils"
)

// NewInfoRPC returns a structure to be indexed
func NewInfoRPC(id uint64, method string,
	params, result, err interface{},
	enc, from, to string, sTime, eTime time.Time) *InfoRPC {
	var e interface{}
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
		RequestStartTime: sTime,
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
	RequestStartTime time.Time
	// EndTime          time.Time

	RequestEncoding    string
	RequestSource      string
	RequestDestination string

	RequestID     uint64
	RequestMethod string
	RequestParams interface{}
	Reply         interface{}
	ReplyError    interface{}
}

type rpcAPI struct {
	ID     uint64      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	Error  string      `json:"err,omitempty"`

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

// unmarshalJSON will transform the message in a map[string]interface{} of []interface{}
// depending of the first character
func unmarshalJSON(jsn json.RawMessage) (interface{}, error) {
	switch {
	case string(jsn) == "null" ||
		len(jsn) == 0: // nil or empty response
		// by default consider nil as an empty map for filtering purposes
		return nil, nil
	case string(jsn) == "true": // booleans
		return true, nil
	case string(jsn) == "false":
		return false, nil
	case jsn[0] == '"': // string
		return string(jsn[1 : len(jsn)-1]), nil
	case jsn[0] >= '0' && jsn[0] <= '9': // float64
		return strconv.ParseFloat(string(jsn), 64)
	case jsn[0] == '[': // slice
		var val []interface{}
		err := json.Unmarshal(jsn, &val)
		return val, err
	case jsn[0] == '{': // map
		var val map[string]interface{}
		err := json.Unmarshal(jsn, &val)
		return val, err
	default:
		return nil, new(json.SyntaxError)
	}
}

// getDPFromSearchresult will unmarshal the request and reply and populate a DataProvider
// if the req is a map[string]interface{} we will try to put in *opts prefix the Opts field from req
func getDPFromSearchresult(req, rep json.RawMessage, hdr utils.MapStorage) (utils.MapStorage, error) {
	repDP, err := unmarshalJSON(rep)
	if err != nil {
		return nil, err
	}
	reqDP, err := unmarshalJSON(req)
	if err != nil {
		return nil, err
	}
	var opts interface{}
	if reqMp, canCast := reqDP.(map[string]interface{}); canCast {
		opts = reqMp[utils.Opts]
	}
	return utils.MapStorage{
		utils.MetaReq:  reqDP,
		utils.MetaOpts: opts,
		utils.MetaRep:  repDP,
		utils.MetaHdr:  hdr,
	}, nil
}
