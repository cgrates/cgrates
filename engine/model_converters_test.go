package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestAPItoResourceLimit(t *testing.T) {
	tpRL := &utils.TPResourceLimit{
		TPID: testTPID,
		ID:   "ResGroup1",
		Filters: []*utils.TPRequestFilter{
			&utils.TPRequestFilter{Type: MetaString, FieldName: "Account", Values: []string{"1001", "1002"}},
			&utils.TPRequestFilter{Type: MetaStringPrefix, FieldName: "Destination", Values: []string{"10", "20"}},
			&utils.TPRequestFilter{Type: MetaCDRStats, Values: []string{"CDRST1:*min_ASR:34", "CDRST_1001:*min_ASR:20"}},
			&utils.TPRequestFilter{Type: MetaRSRFields, Values: []string{"Subject(~^1.*1$)", "Destination(1002)"}},
		},
		ActivationTime: "2014-07-29T15:00:00Z",
		Weight:         10,
		Limit:          "2",
	}
	eRL := &ResourceLimit{ID: tpRL.ID, Weight: tpRL.Weight, Filters: make([]*RequestFilter, len(tpRL.Filters))}
	eRL.Filters[0] = &RequestFilter{Type: MetaString, FieldName: "Account", Values: []string{"1001", "1002"}}
	eRL.Filters[1] = &RequestFilter{Type: MetaStringPrefix, FieldName: "Destination", Values: []string{"10", "20"}}
	eRL.Filters[2] = &RequestFilter{Type: MetaCDRStats, Values: []string{"CDRST1:*min_ASR:34", "CDRST_1001:*min_ASR:20"},
		cdrStatSThresholds: []*RFStatSThreshold{
			&RFStatSThreshold{QueueID: "CDRST1", ThresholdType: "*MIN_ASR", ThresholdValue: 34},
			&RFStatSThreshold{QueueID: "CDRST_1001", ThresholdType: "*MIN_ASR", ThresholdValue: 20},
		}}
	eRL.Filters[3] = &RequestFilter{Type: MetaRSRFields, Values: []string{"Subject(~^1.*1$)", "Destination(1002)"},
		rsrFields: utils.ParseRSRFieldsMustCompile("Subject(~^1.*1$);Destination(1002)", utils.INFIELD_SEP),
	}
	eRL.ActivationTime, _ = utils.ParseTimeDetectLayout("2014-07-29T15:00:00Z", "UTC")
	eRL.Limit = 2
	if rl, err := APItoResourceLimit(tpRL, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRL, rl) {
		t.Errorf("Expecting: %+v, received: %+v", eRL, rl)
	}
}
