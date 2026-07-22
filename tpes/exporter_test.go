// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package tpes

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestExporters(t *testing.T) {
	fileNames := make(map[string]string, len(exporters))
	for exportType, exporter := range exporters {
		if exporter.fileName == "" {
			t.Errorf("%s: missing export filename", exportType)
		}
		if prev, exists := fileNames[exporter.fileName]; exists {
			t.Errorf("%s: duplicate export filename %q, already used by %s", exportType, exporter.fileName, prev)
		}
		fileNames[exporter.fileName] = exportType
		if exporter.keyPrefix == "" {
			t.Errorf("%s: missing key prefix", exportType)
		}
		if exporter.dbItemID == "" {
			t.Errorf("%s: missing DB item ID", exportType)
		}
		if len(exporter.header) == 0 {
			t.Errorf("%s: missing CSV header", exportType)
		}
		if exporter.loadRows == nil {
			t.Errorf("%s: missing row loader", exportType)
		}
	}
}

func TestExportItemsErrors(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewGuardianLocker(cfg)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)

	for exportType, exporter := range exporters {
		t.Run(exportType+"/not found", func(t *testing.T) {
			err := exporter.exportItems(ctx, dm, exportType, io.Discard, utils.CGRateSorg, []string{"missing"})
			want := fmt.Sprintf("cannot find %s with id <missing>: %s",
				strings.TrimPrefix(exportType, utils.Meta), utils.ErrNotFound)
			if err == nil || err.Error() != want {
				t.Fatalf("want %q, got %v", want, err)
			}
			if !errors.Is(err, utils.ErrNotFound) {
				t.Fatalf("want wrapped %v, got %v", utils.ErrNotFound, err)
			}
		})
		t.Run(exportType+"/no database", func(t *testing.T) {
			if err := exporter.exportItems(ctx, nil, exportType, io.Discard, utils.CGRateSorg, []string{"x"}); err != utils.ErrNoDatabaseConn {
				t.Fatalf("want %v, got %v", utils.ErrNoDatabaseConn, err)
			}
		})
	}
}
