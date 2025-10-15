//go:build integration
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

package general_tests

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestOfflineInternalAPIsDumpDataDB(t *testing.T) {

	if err := os.MkdirAll("/tmp/internal_db/datadb", 0755); err != nil {
		t.Fatal(err)
	}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "offline_internal_apis"),
		GracefulShutdown: true,
		PreserveDataDB:   true,
		DBCfg: engine.DBCfg{
			StorDB: engine.MongoDBCfg.StorDB,
		},
		TpPath:    path.Join(*utils.DataDir, "tariffplans", "testit"),
		LogBuffer: &bytes.Buffer{},
	}
	t.Cleanup(func() {
		if err := os.RemoveAll("/tmp/internal_db"); err != nil {
			t.Error(err)
		}
	})
	client, cfg := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	t.Run("CountDataDBFiles", func(t *testing.T) {
		var totalSize int64
		var dirs, files int
		if err := filepath.Walk(cfg.DataDbCfg().Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				dirs++
			} else {
				totalSize += info.Size() // Add the size of the file
				files++
			}
			return nil
		}); err != nil {
			t.Error(err)
		} else if dirs != 43 {
			t.Errorf("expected <%d> directories, received <%d>", 43, dirs)
		} else if files != 42 {
			t.Errorf("expected 42 files, received <%d>", files)
		}
		if totalSize != 0 {
			t.Errorf("expected folder size <%v>, received <%v>", 0, totalSize)
		}
	})

	t.Run("DumpDataDB", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1DumpDataDB, "", &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(50 * time.Millisecond) // wait for dump to finish
	})

	t.Run("CountDataDBFiles2", func(t *testing.T) {
		var totalSize int64
		var dirs, files int
		if err := filepath.Walk(cfg.DataDbCfg().Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				dirs++
			} else {
				totalSize += info.Size() // Add the size of the file
				files++
			}
			return nil
		}); err != nil {
			t.Error(err)
		} else if dirs != 43 {
			t.Errorf("expected <%d> directories, received <%d>", 43, dirs)
		} else if files != 42 {
			t.Errorf("expected 42 files, received <%d>", files)
		}
		if totalSize < 35600 || totalSize > 35800 {
			t.Errorf("expected folder size to be within range 35500KB to 35700KB, received <%v>KB", totalSize)
		}
	})

}

func TestOfflineInternalAPIsDumpStorDB(t *testing.T) {

	if err := os.MkdirAll("/tmp/internal_db/stordb", 0755); err != nil {
		t.Fatal(err)
	}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "offline_internal_apis"),
		GracefulShutdown: true,
		PreserveStorDB:   true,
		DBCfg: engine.DBCfg{
			DataDB: engine.MongoDBCfg.DataDB,
		},
		TpPath:    path.Join(*utils.DataDir, "tariffplans", "testit"),
		LogBuffer: &bytes.Buffer{},
	}
	t.Cleanup(func() {
		if err := os.RemoveAll("/tmp/internal_db"); err != nil {
			t.Error(err)
		}
	})
	client, cfg := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	t.Run("CountStorDBFiles", func(t *testing.T) {
		var totalSize int64
		var dirs, files int
		if err := filepath.Walk(cfg.StorDbCfg().Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				dirs++
			} else {
				totalSize += info.Size() // Add the size of the file
				files++
			}
			return nil
		}); err != nil {
			t.Error(err)
		} else if dirs != 28 {
			t.Errorf("expected <%d> directories, received <%d>", 28, dirs)
		} else if files != 27 {
			t.Errorf("expected 27 files, received <%d>", files)
		}
		if totalSize != 0 {
			t.Errorf("expected folder size <%v>, received <%v>", 0, totalSize)
		}
	})

	t.Run("DumpStorDB", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.APIerSv1DumpStorDB, "", &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(50 * time.Millisecond) // wait for dump to finish
	})

	t.Run("CountStorDBFiles2", func(t *testing.T) {
		var totalSize int64
		var dirs, files int
		if err := filepath.Walk(cfg.StorDbCfg().Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				dirs++
			} else {
				totalSize += info.Size() // Add the size of the file
				files++
			}
			return nil
		}); err != nil {
			t.Error(err)
		} else if dirs != 28 {
			t.Errorf("expected <%d> directories, received <%d>", 28, dirs)
		} else if files != 27 {
			t.Errorf("expected 27 files, received <%d>", files)
		}
		if totalSize < 500 || totalSize > 1000 {
			t.Errorf("expected folder size to be within range 35500KB to 35700KB, received <%v>KB", totalSize)
		}
	})

}
