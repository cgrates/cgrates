package config

import (
	"reflect"
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
		Url:     "/randomURL/",
		RootDir: "/randomPath/",
	}
	if cgrCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err := cgrCfg.configSCfg.loadFromJsonCfg(jsonCfgs); err != nil {
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
	eMap := map[string]interface{}{
		utils.EnabledCfg: true,
		utils.UrlCfg:     "",
		utils.RootDirCfg: "/var/spool/cgrates/configs",
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgsJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.configSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestConfigsAsMapInterface2(t *testing.T) {
	cfgsJSONStr := `{
      "configs":{}
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg: false,
		utils.UrlCfg:     "/configs/",
		utils.RootDirCfg: "/var/spool/cgrates/configs",
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgsJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.configSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}
