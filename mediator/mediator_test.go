package mediator

import (
	"github.com/cgrates/cgrates/config"
	"testing"
)

func TestParseConfig(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	m := &Mediator{cgrCfg: cfg}
	if err := m.parseConfig(); err != nil {
		t.Error(err)
	}
	cfg.MediatorSubjectFields = []string{"subjFieldName1", "subjFieldName2", "subjFieldName3"}
	if err := m.parseConfig(); err == nil {
		t.Error("Failed to detect all fields matching reference one")
	}
	cfg.MediatorRunIds = []string{"run1", "run2"}
	if err := m.parseConfig(); err == nil {
		t.Error("Failed to detect all fields matching reference one")
	}
	cfg.MediatorSubjectFields = []string{"subjFieldName1", "subjFieldName2"}
	cfg.MediatorReqTypeFields = []string{"reqtypeFieldName1", "reqTypeFieldName2"}
	cfg.MediatorDirectionFields = []string{"dirFieldName1", "dirFieldName1"}
	cfg.MediatorTenantFields = []string{"tenantFieldName1", "tenantFieldName2"}
	cfg.MediatorTORFields = []string{"torFieldName1", "torFieldName2"}
	cfg.MediatorAccountFields = []string{"acntFieldName1", "acntFieldName2"}
	cfg.MediatorDestFields = []string{"destFieldName1", "destFieldName2"}
	cfg.MediatorAnswerTimeFields = []string{"answerTimeFieldName1", "answerTimeFieldName1"}
	cfg.MediatorDurationFields = []string{"durFieldName1", "durFieldName2"}
}
