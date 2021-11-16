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
	"bytes"
	"log"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestNewLoaderService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.LoaderCfg()[0].Enabled = true
	cfg.LoaderCfg()[0].RunDelay = -1
	cfg.LoaderCfg()[0].TpInDir = "notAFolder"
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := NewLoaderService(cfg, dm, "", fS, cM)
	if exp := (&LoaderS{
		cfg:   cfg,
		cache: cache,
		ldrs: map[string]*loader{
			utils.MetaDefault: {
				cfg:        cfg,
				ldrCfg:     cfg.LoaderCfg()[0],
				dm:         dm,
				filterS:    fS,
				connMgr:    cM,
				dataCache:  cache,
				cacheConns: cfg.LoaderCfg()[0].CacheSConns,
				Locker:     newLocker(cfg.LoaderCfg()[0].GetLockFilePath()),
			},
		},
	}); !reflect.DeepEqual(exp, ld) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(ld))
	}
	if !ld.Enabled() {
		t.Error("Expected loader to be enabled")
	}

	ld.ldrs[utils.MetaDefault].Locker = mockLock{}
	stop := make(chan struct{})
	close(stop)

	buf := bytes.NewBuffer([]byte{})
	log.SetOutput(buf)
	lgr := utils.Logger
	defer func() { utils.Logger = lgr; log.SetOutput(os.Stderr) }()
	utils.Logger, _ = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	utils.Logger.SetLogLevel(7)

	ld.ListenAndServe(stop)
	runtime.Gosched()
	time.Sleep(time.Nanosecond)
	if expLog, rplyLog := "[ERROR] <LoaderS-*default> error: <no such file or directory>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	cfg.LoaderCfg()[0].Enabled = false
	ld.Reload(dm, "", fS, cM)
	if ld.Enabled() {
		t.Error("Expected loader to not be enabled")
	}
	ld.ListenAndServe(stop)
}

func TestLoaderServiceV1Run(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	tmpIn, err := os.MkdirTemp(utils.EmptyString, "TestLoaderServiceV1RunIn")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpIn)
	for _, f := range fc {
		f.ComputePath()
	}
	cfg.LoaderCfg()[0].Enabled = true
	cfg.LoaderCfg()[0].Data = []*config.LoaderDataType{{
		Type:     utils.MetaAttributes,
		Filename: utils.AttributesCsv,
		Fields:   fc,
	}}
	cfg.LoaderCfg()[0].TpInDir = tmpIn
	cfg.LoaderCfg()[0].TpOutDir = utils.EmptyString

	f, err := os.Create(path.Join(tmpIn, utils.AttributesCsv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`cgrates.org,ID`); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	f, err = os.Create(cfg.LoaderCfg()[0].GetLockFilePath())
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)

	ld := NewLoaderService(cfg, dm, "", fS, cM)
	var rply string
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		APIOpts: map[string]interface{}{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   true,
			utils.MetaStopOnError: true,
			utils.MetaForceLock:   true,
		},
	}, &rply); err != nil {
		t.Fatal(err)
	} else if rply != utils.OK {
		t.Errorf("Expected: %q,received: %q", utils.OK, rply)
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID", FilterIDs: []string{}, Attributes: []*engine.Attribute{}}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
}

type mockLock2 struct{}

// lockFolder will attempt to lock the folder by creating the lock file
func (mockLock2) Lock() (_ error)            { return }
func (mockLock2) Unlock() (_ error)          { return utils.ErrExists }
func (mockLock2) Locked() (_ bool, _ error)  { return true, nil }
func (mockLock2) IsLockFile(string) (_ bool) { return }

func TestLoaderServiceV1RunErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fc := []*config.FCTemplate{
		{Filters: []string{"*string"}},
	}
	tmpIn, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessFolderErrorsIn")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpIn)
	for _, f := range fc {
		f.ComputePath()
	}
	cfg.LoaderCfg()[0].Enabled = true
	cfg.LoaderCfg()[0].Data = []*config.LoaderDataType{{
		Type:     utils.MetaAttributes,
		Filename: utils.AttributesCsv,
		Fields:   fc,
	}}
	cfg.LoaderCfg()[0].TpInDir = tmpIn
	cfg.LoaderCfg()[0].TpOutDir = utils.EmptyString

	f, err := os.Create(path.Join(tmpIn, utils.AttributesCsv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`cgrates.org,ID`); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	f, err = os.Create(cfg.LoaderCfg()[0].GetLockFilePath())
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)

	ld := NewLoaderService(cfg, dm, "", fS, cM)
	var rply string

	expErrMsg := "SERVER_ERROR: inline parse error for string: <*string>"
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		APIOpts: map[string]interface{}{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   true,
			utils.MetaStopOnError: true,
			utils.MetaForceLock:   true,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = `strconv.ParseBool: parsing "notfloat": invalid syntax`
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		APIOpts: map[string]interface{}{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   true,
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   true,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		APIOpts: map[string]interface{}{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   "notfloat",
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   true,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	ld.ldrs[utils.MetaDefault].Locker.Lock()
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		APIOpts: map[string]interface{}{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   "notfloat",
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   "notfloat",
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = `ANOTHER_LOADER_RUNNING`
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		APIOpts: map[string]interface{}{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   "notfloat",
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   false,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	ld.ldrs[utils.MetaDefault].Locker = mockLock{}

	expErrMsg = `SERVER_ERROR: EXISTS`
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	ld.ldrs[utils.MetaDefault].Locker = mockLock2{}
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{APIOpts: map[string]interface{}{
		utils.MetaForceLock: true}}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	cfg.LoaderCfg()[0].Enabled = false
	ld.Reload(dm, "", fS, cM)
	expErrMsg = `UNKNOWN_LOADER: *default`
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
}
