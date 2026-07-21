// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package cdrs

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
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
