package engine

import "testing"

func TestModelHelperCsvLoad(t *testing.T) {
	l, err := csvLoad(TpDestination{}, []string{"TEST_DEST", "+492"})
	tpd := l.(TpDestination)
	if err != nil || tpd.Tag != "TEST_DEST" || tpd.Prefix != "+492" {
		t.Errorf("model load failed: %+v", tpd)
	}
}

func TestModelHelperCsvLoadInt(t *testing.T) {
	l, err := csvLoad(TpCdrStat{}, []string{"CDRST1", "5", "60m", "ASR", "2014-07-29T15:00:00Z;2014-07-29T16:00:00Z", "*voice", "87.139.12.167", "FS_JSON", "*rated", "*out", "cgrates.org", "call", "dan", "dan", "49", "5m;10m", "suppl1", "NORMAL_CLEARING", "default", "rif", "rif", "0;2", "STANDARD_TRIGGERS"})
	tpd := l.(TpCdrStat)
	if err != nil || tpd.QueueLength != 5 {
		t.Errorf("model load failed: %+v", tpd)
	}
}

func TestModelHelperCsvDump(t *testing.T) {
	tpd := TpDestination{
		Tag:    "TEST_DEST",
		Prefix: "+492"}
	csv, err := csvDump(tpd)
	if err != nil || csv[0] != "TEST_DEST" || csv[1] != "+492" {
		t.Errorf("model load failed: %+v", tpd)
	}
}
