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
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestConfigsloadFromJsonCfg(t *testing.T) {
	jsonCfgs := &ConfigSCfgJson{
		Enabled:  utils.BoolPointer(true),
		Url:      utils.StringPointer("/randomURL/"),
		Root_dir: utils.StringPointer("/randomPath/"),
	}
	expectedCfg := &ConfigSCfg{
		Enabled: true,
		URL:     "/randomURL/",
		RootDir: "/randomPath/",
	}
	cgrCfg := NewDefaultCGRConfig()
	if err := cgrCfg.configSCfg.loadFromJSONCfg(jsonCfgs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cgrCfg.configSCfg, expectedCfg) {
		t.Errorf("Expected %+v, received %+v", expectedCfg, cgrCfg.configSCfg)
	}
}

func TestConfigsAsMapInterface(t *testing.T) {
	cfgsJSONStr := `{
      "configs": {
          "enabled": true,
          "url": "",
          "root_dir": "/var/spool/cgrates/configs"
      },
}`
	eMap := map[string]any{
		utils.EnabledCfg: true,
		utils.URLCfg:     "",
		utils.RootDirCfg: "/var/spool/cgrates/configs",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgsJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.configSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestConfigsAsMapInterface2(t *testing.T) {
	cfgsJSONStr := `{
      "configs":{}
}`
	eMap := map[string]any{
		utils.EnabledCfg: false,
		utils.URLCfg:     "/configs/",
		utils.RootDirCfg: "/var/spool/cgrates/configs",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgsJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.configSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestNewCGRConfigFromPathWithoutEnv(t *testing.T) {
	cfgsJSONStr := `{
		"general": {
			"node_id": "*env:NODE_ID",
		},
  }`
	cfg := NewDefaultCGRConfig()

	if err := cfg.loadConfigFromReader(strings.NewReader(cfgsJSONStr), []func(*CgrJsonCfg) error{cfg.loadFromJSONCfg}, true); err != nil {
		t.Fatal(err)
	}
	exp := "*env:NODE_ID"
	if cfg.GeneralCfg().NodeID != exp {
		t.Errorf("Expected %+v, received %+v", exp, cfg.GeneralCfg().NodeID)
	}
}

func TestConfigSCfgClone(t *testing.T) {
	cS := &ConfigSCfg{
		Enabled: true,
		URL:     "/randomURL/",
		RootDir: "/randomPath/",
	}
	rcv := cS.Clone()
	if !reflect.DeepEqual(cS, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cS), utils.ToJSON(rcv))
	}
	if rcv.URL = ""; cS.URL != "/randomURL/" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestHandleConfigSFile(t *testing.T) {

	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "content"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	rr := httptest.NewRecorder()

	handleConfigSFile(tmpFile.Name(), rr)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	if rr.Body.String() != content {
		t.Errorf("Expected response body %q, got %q", content, rr.Body.String())
	}
}

func TestHandleConfigSFileFileReadError(t *testing.T) {

	nonExistentFilePath := "non-existent-file.txt"

	rr := httptest.NewRecorder()

	handleConfigSFile(nonExistentFilePath, rr)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
	}

}

func TestHandleConfigSFolder(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "configtest")

	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	defer os.RemoveAll(tmpDir)
	w := httptest.NewRecorder()
	handleConfigSFolder("/invalid/path", w)
	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, but got %d", resp.StatusCode)
	}

	jsonConfig := `{"Subject": "1001"}`
	configFilePath := tmpDir + "/config.json"
	err = os.WriteFile(configFilePath, []byte(jsonConfig), 0644)

	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	w = httptest.NewRecorder()
	handleConfigSFolder(tmpDir, w)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, but got %d", resp.StatusCode)
	}
}

func TestHandlerConfigS(t *testing.T) {
	tmpRootDir, err := os.MkdirTemp("", "test_config_root")
	if err != nil {
		t.Fatalf("Failed to create temporary root directory: %v", err)
	}

	defer os.RemoveAll(tmpRootDir)
	originalRootDir := CgrConfig().ConfigSCfg().RootDir
	CgrConfig().ConfigSCfg().RootDir = tmpRootDir
	defer func() {
		CgrConfig().ConfigSCfg().RootDir = originalRootDir
	}()
	filePath := path.Join(tmpRootDir, "test_file.txt")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file.Close()
	dirPath := path.Join(tmpRootDir, "test_folder")
	err = os.Mkdir(dirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name         string
		requestPath  string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "File exists",
			requestPath:  "/configs/test_file.txt",
			expectedCode: 200,
			expectedBody: "Handled file",
		},
		{
			name:         "Directory exists",
			requestPath:  "/configs/test_folder",
			expectedCode: 200,
			expectedBody: "Handled directory",
		},
		{
			name:         "File does not exist",
			requestPath:  "/configs/nonexistent_file.txt",
			expectedCode: 404,
			expectedBody: "no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()
			HandlerConfigS(w, req)
		})
	}
}
