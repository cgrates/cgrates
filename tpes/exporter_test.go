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
	cacheS := engine.NewCacheS(cfg, nil, nil, nil)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
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
