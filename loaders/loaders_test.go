// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package loaders

import (
	"archive/zip"
	"bytes"
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
)

func TestNewLoaderService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.LoaderCfg()[0].Enabled = true
	cfg.LoaderCfg()[0].RunDelay = -1
	cfg.LoaderCfg()[0].TpInDir = "notAFolder"
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	cM := engine.NewConnManager(cfg)
	cM.SetCache(cacheS)
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM, locker)
	dm.SetCache(cacheS)
	fS := engine.NewFilterS(cfg, cM, dm)
	ld := NewLoaderS(cfg, dm, fS, cM, locker)
	if !ld.Enabled() {
		t.Error("Expected loader to be enabled")
	}

	stop := make(chan struct{})
	close(stop)

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	ld.ListenAndServe(stop)
	runtime.Gosched()
	time.Sleep(time.Nanosecond)
	if expLog, rplyLog := "[ERROR] <LoaderS-*default> error: <no such file or directory>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	cfg.LoaderCfg()[0].Enabled = false
	ld.Reload(dm, fS, cM)
	if ld.Enabled() {
		t.Error("Expected loader to not be enabled")
	}
	ld.ListenAndServe(stop)
}

func TestLoaderServiceV1Run(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: utils.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: utils.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
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

	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	cM := engine.NewConnManager(cfg)
	cM.SetCache(cacheS)
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM, locker)
	dm.SetCache(cacheS)
	fS := engine.NewFilterS(cfg, cM, dm)

	ld := NewLoaderS(cfg, dm, fS, cM, locker)
	var rply string
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		APIOpts: map[string]any{
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
	} else {
		v := &utils.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
		if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
	}
}

func TestLoaderServiceV1RunErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
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

	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	cM := engine.NewConnManager(cfg)
	cM.SetCache(cacheS)
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM, locker)
	dm.SetCache(cacheS)
	fS := engine.NewFilterS(cfg, cM, dm)

	ld := NewLoaderS(cfg, dm, fS, cM, locker)
	var rply string

	expErrMsg := "SERVER_ERROR: inline parse error for string: <*string>"
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		APIOpts: map[string]any{
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
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   true,
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   true,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   "notfloat",
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   true,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if _, err := ld.ldrs[utils.MetaDefault].locker.lock(); err != nil {
		t.Fatal(err)
	}
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		APIOpts: map[string]any{
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
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   "notfloat",
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   false,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	cfg.LoaderCfg()[0].Enabled = false
	ld.Reload(dm, fS, cM)
	expErrMsg = `UNKNOWN_LOADER: *default`
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
}

func TestLoaderServiceV1ImportZip(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: utils.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: utils.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	cfg.LoaderCfg()[0].Enabled = true
	cfg.LoaderCfg()[0].Data = []*config.LoaderDataType{{
		Type:     utils.MetaAttributes,
		Filename: utils.AttributesCsv,
		Fields:   fc,
	}}
	cfg.LoaderCfg()[0].LockFilePath = utils.MetaMemory
	cfg.LoaderCfg()[0].TpInDir = utils.EmptyString
	cfg.LoaderCfg()[0].TpOutDir = utils.EmptyString

	buf := new(bytes.Buffer)
	wr := zip.NewWriter(buf)
	f, err := wr.Create(utils.AttributesCsv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(`cgrates.org,ID`)); err != nil {
		t.Fatal(err)
	}
	if err := wr.Close(); err != nil {
		t.Fatal(err)
	}

	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	cM := engine.NewConnManager(cfg)
	cM.SetCache(cacheS)
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM, locker)
	dm.SetCache(cacheS)
	fS := engine.NewFilterS(cfg, cM, dm)

	ld := NewLoaderS(cfg, dm, fS, cM, locker)
	var rply string
	if err := ld.V1ImportZip(context.Background(), &ArgsProcessZip{
		Data: buf.Bytes(),
		APIOpts: map[string]any{
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
	} else {
		v := &utils.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
		if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
	}
}

func TestLoaderServiceV1ImportZipErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	fc := []*config.FCTemplate{
		{Filters: []string{"*string"}},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	cfg.LoaderCfg()[0].Enabled = true
	cfg.LoaderCfg()[0].Data = []*config.LoaderDataType{{
		Type:     utils.MetaAttributes,
		Filename: utils.AttributesCsv,
		Fields:   fc,
	}}
	cfg.LoaderCfg()[0].TpInDir = utils.EmptyString
	cfg.LoaderCfg()[0].TpOutDir = utils.EmptyString
	defer os.Remove(cfg.LoaderCfg()[0].LockFilePath)
	buf := new(bytes.Buffer)
	wr := zip.NewWriter(buf)
	f, err := wr.Create(utils.AttributesCsv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(`cgrates.org,ID`)); err != nil {
		t.Fatal(err)
	}
	if err := wr.Close(); err != nil {
		t.Fatal(err)
	}

	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	cM := engine.NewConnManager(cfg)
	cM.SetCache(cacheS)
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM, locker)
	dm.SetCache(cacheS)
	fS := engine.NewFilterS(cfg, cM, dm)

	ld := NewLoaderS(cfg, dm, fS, cM, locker)
	var rply string

	expErrMsg := "SERVER_ERROR: inline parse error for string: <*string>"
	if err := ld.V1ImportZip(context.Background(), &ArgsProcessZip{
		Data: buf.Bytes(),
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   true,
			utils.MetaStopOnError: true,
			utils.MetaForceLock:   true,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = "zip: not a valid zip file"
	if err := ld.V1ImportZip(context.Background(), &ArgsProcessZip{
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   true,
			utils.MetaStopOnError: true,
			utils.MetaForceLock:   true,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = `strconv.ParseBool: parsing "notfloat": invalid syntax`
	if err := ld.V1ImportZip(context.Background(), &ArgsProcessZip{
		Data: buf.Bytes(),
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   true,
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   true,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := ld.V1ImportZip(context.Background(), &ArgsProcessZip{
		Data: buf.Bytes(),
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   "notfloat",
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   true,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if _, err := ld.ldrs[utils.MetaDefault].locker.lock(); err != nil {
		t.Fatal(err)
	}
	if err := ld.V1ImportZip(context.Background(), &ArgsProcessZip{
		Data: buf.Bytes(),
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   "notfloat",
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   "notfloat",
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = `ANOTHER_LOADER_RUNNING`
	if err := ld.V1ImportZip(context.Background(), &ArgsProcessZip{
		Data: buf.Bytes(),
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   "notfloat",
			utils.MetaStopOnError: "notfloat",
			utils.MetaForceLock:   false,
		},
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	cfg.LoaderCfg()[0].Enabled = false
	ld.Reload(dm, fS, cM)
	expErrMsg = `UNKNOWN_LOADER: *default`
	if err := ld.V1ImportZip(context.Background(), &ArgsProcessZip{
		Data: buf.Bytes(),
	}, &rply); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
}

func waitForLoaderCall(t *testing.T, unlock func(), call func() error) error {
	t.Helper()
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		close(started)
		done <- call()
	}()
	<-started
	select {
	case err := <-done:
		unlock()
		t.Fatalf("loader call did not wait: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	unlock()
	select {
	case err := <-done:
		return err
	case <-time.After(time.Second):
		t.Fatal("loader call remained blocked")
		return nil
	}
}

func newMemoryLoaderS(t *testing.T) *LoaderS {
	t.Helper()
	cfg := config.NewDefaultCGRConfig()
	cfg.LoaderCfg()[0].Enabled = true
	cfg.LoaderCfg()[0].LockFilePath = utils.MetaMemory
	cfg.LoaderCfg()[0].TpInDir = t.TempDir()
	cfg.LoaderCfg()[0].TpOutDir = ""
	cfg.LoaderCfg()[0].Data = nil
	cfg.LoaderCfg()[0].Conns = nil
	cfg.LoaderCfg()[0].RunDelay = 0
	locker := engine.NewLocker(cfg)
	return NewLoaderS(cfg, nil, nil, nil, locker)
}

func TestLoaderMemoryLockCoordinatesWork(t *testing.T) {
	t.Run("RunForceLock", func(t *testing.T) {
		ldrS := newMemoryLoaderS(t)
		ldr := ldrS.ldrs[utils.MetaDefault]
		unlock, err := ldr.locker.lock()
		if err != nil {
			t.Fatal(err)
		}
		var reply string
		err = waitForLoaderCall(t, unlock, func() error {
			return ldrS.V1Run(context.Background(), &ArgsProcessFolder{
				APIOpts: map[string]any{utils.MetaForceLock: true},
			}, &reply)
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ImportZip", func(t *testing.T) {
		ldrS := newMemoryLoaderS(t)
		ldr := ldrS.ldrs[utils.MetaDefault]
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)
		if err := zipWriter.Close(); err != nil {
			t.Fatal(err)
		}
		unlock, err := ldr.locker.lock()
		if err != nil {
			t.Fatal(err)
		}
		var reply string
		err = waitForLoaderCall(t, unlock, func() error {
			return ldrS.V1ImportZip(context.Background(), &ArgsProcessZip{
				Data: buf.Bytes(),
			}, &reply)
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Periodic", func(t *testing.T) {
		ldrS := newMemoryLoaderS(t)
		ldr := ldrS.ldrs[utils.MetaDefault]
		unlock, err := ldr.locker.lock()
		if err != nil {
			t.Fatal(err)
		}
		if err := waitForLoaderCall(t, unlock, func() error {
			return ldr.processFolder(context.Background(), "", nil, false, false)
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("FSNotify", func(t *testing.T) {
		ldrS := newMemoryLoaderS(t)
		ldr := ldrS.ldrs[utils.MetaDefault]
		ldr.ldrCfg.Data = []*config.LoaderDataType{{Filename: "missing.csv"}}
		unlock, err := ldr.locker.lock()
		if err != nil {
			t.Fatal(err)
		}
		if err := waitForLoaderCall(t, unlock, func() error {
			return ldr.processIFile("missing.csv")
		}); err == nil {
			t.Fatal("expected missing input file error")
		}
	})

	t.Run("Reload", func(t *testing.T) {
		ldrS := newMemoryLoaderS(t)
		oldLoader := ldrS.ldrs[utils.MetaDefault]
		unlock, err := oldLoader.locker.lock()
		if err != nil {
			t.Fatal(err)
		}
		ldrS.Reload(nil, nil, nil)
		var reply string
		err = waitForLoaderCall(t, unlock, func() error {
			return ldrS.V1Run(context.Background(), &ArgsProcessFolder{}, &reply)
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestLoaderServiceV1RunPath(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: utils.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: utils.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	tmpPath, err := os.MkdirTemp(utils.EmptyString, "TestLoaderServiceV1RunPath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpPath)
	for _, f := range fc {
		f.ComputePath()
	}
	cfg.LoaderCfg()[0].Enabled = true
	cfg.LoaderCfg()[0].Data = []*config.LoaderDataType{{
		Type:     utils.MetaAttributes,
		Filename: utils.AttributesCsv,
		Fields:   fc,
	}}
	cfg.LoaderCfg()[0].TpOutDir = utils.EmptyString

	f, err := os.Create(path.Join(tmpPath, utils.AttributesCsv))
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

	inPath, err := os.MkdirTemp(utils.EmptyString, "TestLoaderServiceV1RunTpIn")
	if err != nil {
		t.Fatal(err)
	}
	cfg.LoaderCfg()[0].TpInDir = inPath
	defer os.RemoveAll(inPath)

	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	cM := engine.NewConnManager(cfg)
	cM.SetCache(cacheS)
	idb, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: idb}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, cM, locker)
	dm.SetCache(cacheS)
	fS := engine.NewFilterS(cfg, cM, dm)

	ld := NewLoaderS(cfg, dm, fS, cM, locker)
	var rply string
	if err := ld.V1Run(context.Background(), &ArgsProcessFolder{
		Path: tmpPath,
		APIOpts: map[string]any{
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
	} else {
		v := &utils.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
		if !reflect.DeepEqual(v, prf) {
			t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
		}
	}
}
