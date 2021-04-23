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

package config

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// ConfigSCfg config for listening over http
type ConfigSCfg struct {
	Enabled bool
	URL     string
	RootDir string
}

// loadFromJSONCfg loads Database config from JsonCfg
func (cScfg *ConfigSCfg) loadFromJSONCfg(jsnCfg *ConfigSCfgJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		cScfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Url != nil {
		cScfg.URL = *jsnCfg.Url
	}
	if jsnCfg.Root_dir != nil {
		cScfg.RootDir = *jsnCfg.Root_dir
	}
	return
}

// HandlerConfigS handler for httpServer to register the configs
func HandlerConfigS(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	// take out the /configs prefix and use the rest of url as path
	pth := strings.TrimPrefix(r.URL.Path, "/configs")
	pth = path.Join(CgrConfig().ConfigSCfg().RootDir, pth)
	fi, err := os.Stat(pth)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		w.Write([]byte(err.Error()))
		return
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		handleConfigSFolder(pth, w)
	case mode.IsRegular():
		handleConfigSFile(pth, w)
	}
}

func handleConfigSFolder(path string, w http.ResponseWriter) {
	// if the path is a directory, read the directory, construct the config and load it in memory
	cfg, err := newCGRConfigFromPathWithoutEnv(path)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	// convert the config into a json and send it
	if _, err := w.Write([]byte(utils.ToJSON(cfg.AsMapInterface(cfg.generalCfg.RSRSep)))); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to write resonse because: %s",
			utils.ConfigSv1, err))
	}
}

func handleConfigSFile(path string, w http.ResponseWriter) {
	// if the config is a file read the file and send it directly
	dat, err := os.ReadFile(path)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	if _, err := w.Write(dat); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to write resonse because: %s",
			utils.ConfigSv1, err))
	}
}

// AsMapInterface returns the config as a map[string]interface{}
func (cScfg *ConfigSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg: cScfg.Enabled,
		utils.URLCfg:     cScfg.URL,
		utils.RootDirCfg: cScfg.RootDir,
	}
	return
}

// Clone returns a deep copy of ConfigSCfg
func (cScfg *ConfigSCfg) Clone() *ConfigSCfg {
	return &ConfigSCfg{
		Enabled: cScfg.Enabled,
		URL:     cScfg.URL,
		RootDir: cScfg.RootDir,
	}
}

type ConfigSCfgJson struct {
	Enabled  *bool
	Url      *string
	Root_dir *string
}

func diffConfigSCfgJson(d *ConfigSCfgJson, v1, v2 *ConfigSCfg) *ConfigSCfgJson {
	if d == nil {
		d = new(ConfigSCfgJson)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.URL != v2.URL {
		d.Url = utils.StringPointer(v2.URL)
	}
	if v1.RootDir != v2.RootDir {
		d.Root_dir = utils.StringPointer(v2.RootDir)
	}
	return d
}
