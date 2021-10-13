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
	"io"
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

func TestProcessContentCallsRemoveItems(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})

	sMock := &testMockCacheConn{
		calls: map[string]func(arg interface{}, rply interface{}) error{
			utils.CacheSv1RemoveItems: func(arg interface{}, rply interface{}) error {
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
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)

	internalCacheSChan := make(chan rpcclient.ClientConnector, 1)
	internalCacheSChan <- sMock
	ldr := &Loader{
		ldrID:         "TestProcessContentCallsRemoveItems",
		bufLoaderData: make(map[string][]LoaderData),
		connMgr: engine.NewConnManager(config.CgrConfig(), map[string]chan rpcclient.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): internalCacheSChan,
		}),
		dm:         engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		cacheConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		timezone:   "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	attributeCsv := `
#Tenant[0],ID[1]
cgrates.org,MOCK_RELOAD_ID
`
	rdr := io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	if err := ldr.processContent(utils.MetaAttributes, utils.MetaRemove); err != nil {
		t.Error(err)
	}

	// Calling the method again while cacheConnsID is not valid
	ldr.cacheConns = []string{utils.MetaInternal}
	rdr = io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err := ldr.processContent(utils.MetaAttributes, utils.MetaRemove); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	// Calling the method again while caching method is invalid
	ldr.cacheConns = []string{utils.MetaInternal}
	rdr = io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expected = "UNSUPPORTED_SERVICE_METHOD"
	if err := ldr.processContent(utils.MetaAttributes, "invalid_caching_api"); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestProcessContentCallsClear(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})

	sMock := &testMockCacheConn{
		calls: map[string]func(arg interface{}, rply interface{}) error{
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
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)

	internalCacheSChan := make(chan rpcclient.ClientConnector, 1)
	internalCacheSChan <- sMock
	ldr := &Loader{
		ldrID:         "TestProcessContentCallsClear",
		bufLoaderData: make(map[string][]LoaderData),
		connMgr: engine.NewConnManager(config.CgrConfig(), map[string]chan rpcclient.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): internalCacheSChan,
		}),
		dm:         engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		cacheConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		timezone:   "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	attributeCsv := `
#Tenant[0],ID[1]
cgrates.org,MOCK_RELOAD_ID
`
	rdr := io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	if err := ldr.processContent(utils.MetaAttributes, utils.MetaClear); err != nil {
		t.Error(err)
	}

	//inexisting method(*none) of cache and reinitialized the reader will do nothing
	rdr = io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	if err := ldr.processContent(utils.MetaAttributes, utils.MetaNone); err != nil {
		t.Error(err)
	}

	// Calling the method again while cacheConnsID is not valid
	ldr.cacheConns = []string{utils.MetaInternal}
	rdr = io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err := ldr.processContent(utils.MetaAttributes, utils.MetaClear); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveContentCallsReload(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})

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
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)

	internalCacheSChan := make(chan rpcclient.ClientConnector, 1)
	internalCacheSChan <- sMock
	ldr := &Loader{
		ldrID:         "TestRemoveContentCallsReload",
		bufLoaderData: make(map[string][]LoaderData),
		connMgr: engine.NewConnManager(config.CgrConfig(), map[string]chan rpcclient.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): internalCacheSChan,
		}),
		cacheConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		dm:         engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:   "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	attributeCsv := `
#Tenant[0],ID[1]
cgrates.org,MOCK_RELOAD_2
`
	rdr := io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	attrPrf := &engine.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "MOCK_RELOAD_2",
	}
	if err := ldr.dm.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	if err := ldr.removeContent(utils.MetaAttributes, utils.MetaReload); err != nil {
		t.Error(err)
	}

	//Calling the method again while cacheConnsID is not valid
	ldr.cacheConns = []string{utils.MetaInternal}
	rdr = io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}

	//set and remove again from database
	if err := ldr.dm.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err := ldr.removeContent(utils.MetaAttributes, utils.MetaReload); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveContentCallsLoad(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})

	sMock := &testMockCacheConn{
		calls: map[string]func(arg interface{}, rply interface{}) error{
			utils.CacheSv1LoadCache: func(arg interface{}, rply interface{}) error {
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
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)

	internalCacheSChan := make(chan rpcclient.ClientConnector, 1)
	internalCacheSChan <- sMock
	ldr := &Loader{
		ldrID:         "TestRemoveContentCallsReload",
		bufLoaderData: make(map[string][]LoaderData),
		connMgr: engine.NewConnManager(config.CgrConfig(), map[string]chan rpcclient.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): internalCacheSChan,
		}),
		cacheConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		dm:         engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:   "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	attributeCsv := `
#Tenant[0],ID[1]
cgrates.org,MOCK_RELOAD_3
`
	rdr := io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	attrPrf := &engine.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "MOCK_RELOAD_3",
	}
	if err := ldr.dm.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	if err := ldr.removeContent(utils.MetaAttributes, utils.MetaLoad); err != nil {
		t.Error(err)
	}

	//Calling the method again while cacheConnsID is not valid
	ldr.cacheConns = []string{utils.MetaInternal}
	rdr = io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}

	//set and remove again from database
	if err := ldr.dm.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err := ldr.removeContent(utils.MetaAttributes, utils.MetaLoad); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestRemoveContentCallsRemove(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})

	sMock := &testMockCacheConn{
		calls: map[string]func(arg interface{}, rply interface{}) error{
			utils.CacheSv1RemoveItems: func(arg interface{}, rply interface{}) error {
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
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)

	internalCacheSChan := make(chan rpcclient.ClientConnector, 1)
	internalCacheSChan <- sMock
	ldr := &Loader{
		ldrID:         "TestRemoveContentCallsReload",
		bufLoaderData: make(map[string][]LoaderData),
		connMgr: engine.NewConnManager(config.CgrConfig(), map[string]chan rpcclient.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): internalCacheSChan,
		}),
		cacheConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		dm:         engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:   "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	attributeCsv := `
#Tenant[0],ID[1]
cgrates.org,MOCK_RELOAD_4
`
	rdr := io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	attrPrf := &engine.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "MOCK_RELOAD_4",
	}
	if err := ldr.dm.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	if err := ldr.removeContent(utils.MetaAttributes, utils.MetaRemove); err != nil {
		t.Error(err)
	}

	//Calling the method again while cacheConnsID is not valid
	ldr.cacheConns = []string{utils.MetaInternal}
	rdr = io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}

	//set and remove again from database
	if err := ldr.dm.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err := ldr.removeContent(utils.MetaAttributes, utils.MetaRemove); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	//inexisting method(*none) of cache and reinitialized the reader will do nothing
	rdr = io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	if err := ldr.dm.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	if err := ldr.removeContent(utils.MetaAttributes, utils.MetaNone); err != nil {
		t.Error(err)
	}
}

func TestRemoveContentCallsClear(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})

	sMock := &testMockCacheConn{
		calls: map[string]func(arg interface{}, rply interface{}) error{
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
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)

	internalCacheSChan := make(chan rpcclient.ClientConnector, 1)
	internalCacheSChan <- sMock
	ldr := &Loader{
		ldrID:         "TestRemoveContentCallsReload",
		bufLoaderData: make(map[string][]LoaderData),
		connMgr: engine.NewConnManager(config.CgrConfig(), map[string]chan rpcclient.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): internalCacheSChan,
		}),
		cacheConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)},
		dm:         engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil),
		timezone:   "UTC",
	}
	ldr.dataTpls = map[string][]*config.FCTemplate{
		utils.MetaAttributes: {
			{Tag: "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true},
			{Tag: "ProfileID",
				Path:      "ID",
				Type:      utils.MetaComposed,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
				Mandatory: true},
		},
	}
	attributeCsv := `
#Tenant[0],ID[1]
cgrates.org,MOCK_RELOAD_3
`
	rdr := io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv := csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	attrPrf := &engine.AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "MOCK_RELOAD_3",
	}
	if err := ldr.dm.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	if err := ldr.removeContent(utils.MetaAttributes, utils.MetaClear); err != nil {
		t.Error(err)
	}

	//Calling the method again while cacheConnsID is not valid
	ldr.cacheConns = []string{utils.MetaInternal}
	rdr = io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}

	//set and remove again from database
	if err := ldr.dm.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err := ldr.removeContent(utils.MetaAttributes, utils.MetaClear); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	// Calling the method again while caching method is invalid
	rdr = io.NopCloser(strings.NewReader(attributeCsv))
	rdrCsv = csv.NewReader(rdr)
	rdrCsv.Comment = '#'
	ldr.rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaAttributes: {
			utils.AttributesCsv: &openedCSVFile{
				fileName: utils.AttributesCsv,
				rdr:      rdr,
				csvRdr:   rdrCsv,
			},
		},
	}
	if err := ldr.dm.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	expected = "UNSUPPORTED_SERVICE_METHOD"
	if err := ldr.removeContent(utils.MetaAttributes, "invalid_caching_api"); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}
