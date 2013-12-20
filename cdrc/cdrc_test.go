package cdrc

import (
	"testing"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var cgrConfig  *config.CGRConfig
var cdrc *Cdrc

func init() {
	cgrConfig, _ = config.NewDefaultCGRConfig()
	cdrc = &Cdrc{cgrCfg:cgrConfig}
}

func TestParseFieldIndexesFromConfig(t *testing.T) {
	if err := cdrc.parseFieldIndexesFromConfig(); err != nil {
		t.Error("Failed parsing default fieldIndexesFromConfig", err)
	}
}


func TestCdrAsHttpForm(t *testing.T) {
	cdrRow := []string{"firstField", "secondField"}
	_, err := cdrc.cdrAsHttpForm(cdrRow)
	if err == nil {
		t.Error("Failed to corectly detect missing fields from record")
	}
	cdrRow = []string{"acc1", "prepaid", "*out", "cgrates.org", "call", "1001", "1001", "+4986517174963", "2013-02-03 19:54:00", "62", "supplier1", "172.16.1.1"}
	cdrAsForm, err := cdrc.cdrAsHttpForm(cdrRow); 
	if err != nil {
		t.Error("Failed to parse CDR in form", err)
	}
	if cdrAsForm.Get(utils.REQTYPE) != "prepaid" {
		t.Error("Unexpected CDR value received", cdrAsForm.Get(utils.REQTYPE))
	}
	if cdrAsForm.Get("supplier") != "supplier1" {
		t.Error("Unexpected CDR value received", cdrAsForm.Get(utils.REQTYPE))
	}
}
