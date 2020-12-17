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

package loaders

import (
	"encoding/csv"
	"errors"
	"flag"
	"io/ioutil"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"testing"

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	waitRater = flag.Int("wait_rater", 200, "Number of miliseconds to wait for rater to start and cache")
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	encoding  = flag.String("rpc", utils.MetaJSON, "what encoding whould be used for rpc comunication")
	dbType    = flag.String("dbtype", utils.MetaInternal, "The type of DataBase (Internal/Mongo/mySql)")
)

var loaderPaths = []string{"/tmp/In", "/tmp/Out", "/tmp/LoaderIn", "/tmp/SubpathWithoutMove",
	"/tmp/SubpathLoaderWithMove", "/tmp/SubpathOut", "/tmp/templateLoaderIn", "/tmp/templateLoaderOut",
	"/tmp/customSepLoaderIn", "/tmp/customSepLoaderOut"}

func newRPCClient(cfg *config.ListenCfg) (c *rpc.Client, err error) {
	switch *encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return rpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

type testMockCacheConn struct {
	calls map[string]func(arg interface{}, rply interface{}) error
}

func (s *testMockCacheConn) Call(method string, arg interface{}, rply interface{}) error {
	if call, has := s.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(arg, rply)
	}
}

func TestProcessContentCalls(t *testing.T) {
	sMock := &testMockCacheConn{
		calls: map[string]func(arg interface{}, rply interface{}) error{
			utils.CacheSv1ReloadCache: func(arg interface{}, rply interface{}) error {
				prply, can := rply.(*string)
				if !can {
					t.Errorf("Wrong argument type : %T", rply)
					return nil
				}
				*prply = utils.OK
				return nil
			},
			utils.CacheSv1Clear: func(arg interface{}, rply interface{}) error {
				prply, can := rply.(*string)
				if !can {
					t.Errorf("Wrong argument type : %T", rply)
					return nil
				}
				*prply = utils.OK
				return nil
			},
		},
	}
	data := engine.NewInternalDB(nil, nil, true)

	internalCacheSChan := make(chan rpcclient.ClientConnector, 1)
	internalCacheSChan <- sMock
	ldr := &Loader{
		ldrID:         "TestProcessContentCalls",
		bufLoaderData: make(map[string][]LoaderData),
		connMgr: engine.NewConnManager(config.CgrConfig(), map[string]chan rpcclient.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): internalCacheSChan,
		}),
		dm:         engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		cacheConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		timezone:   "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaRateProfiles: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "Weight",
				Path:  "Weight",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.2", utils.INFIELD_SEP)},
		},
	}
	thresholdsCsv := `
#Tenant[0],ID[1],Weight[2]
cgrates.org,MOCK_RELOAD_ID,20
`
	rdr := ioutil.NopCloser(strings.NewReader(thresholdsCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaRateProfiles: {
			utils.RateProfilesCsv: &openedCSVFile{
				fileName: utils.RateProfilesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	if err := ldr.processContent(utils.MetaRateProfiles, utils.MetaReload); err != nil {
		t.Error(err)
	}
}
