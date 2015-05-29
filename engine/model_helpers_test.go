package engine

import "testing"

func TestModelHelperCsvLoad(t *testing.T) {
	l := csvLoad(TpDestination{}, []string{"TEST_DEST", "+492"})
	tpd := l.(TpDestination)
	if tpd.Tag != "TEST_DEST" || tpd.Prefix != "+492" {
		t.Errorf("model load failed: %+v", tpd)
	}
}

func TestModelHelperCsvDump(t *testing.T) {
	tpd := &TpDestination{
		Tag:    "TEST_DEST",
		Prefix: "+492"}
	csv, err := csvDump(*tpd, ",")
	if err != nil || csv != "TEST_DEST,+492" {
		t.Errorf("model load failed: %+v", tpd)
	}
}
