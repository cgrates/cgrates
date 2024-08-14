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

package cdrs

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	// dataDir    = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	waitRater = flag.Int("wait_rater", 100, "Number of milliseconds to wait for rater to start and cache")
	encoding  = flag.String("rpc", utils.MetaJSON, "what encoding would be used for rpc communication")
	dbType    = flag.String("dbtype", utils.MetaInternal, "The type of DataBase (Internal/Mongo/mySql)")
)

// initCfg creates a new CGRConfig from the provided configuration content string. It generates a
// temporary directory and file path, writes the content to the configuration file, and returns the
// created CGRConfig, the file path, a cleanup function, and any error encountered.
func initCfg(ctx *context.Context, cfgContent string) (cfg *config.CGRConfig, cfgPath string, cleanFunc func(), err error) {
	if cfgContent == utils.EmptyString {
		return nil, "", func() {}, errors.New("content should not be empty")
	}
	cfgPath = fmt.Sprintf("/tmp/config%d", rand.Int63n(10000))
	err = os.MkdirAll(cfgPath, 0755)
	if err != nil {
		return nil, "", func() {}, err
	}
	removeFunc := func() {
		os.RemoveAll(cfgPath)
	}
	filePath := filepath.Join(cfgPath, "cgrates.json")
	err = os.WriteFile(filePath, []byte(cfgContent), 0644)
	if err != nil {
		return nil, "", removeFunc, err
	}
	cfg, err = config.NewCGRConfigFromPath(ctx, cfgPath)
	if err != nil {
		return nil, "", removeFunc, err
	}

	return cfg, cfgPath, removeFunc, nil
}
