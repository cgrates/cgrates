package utils

import (
	"reflect"
	"testing"
)

func TestStringMapParse(t *testing.T) {
	sm := ParseStringMap("1;2;3;4")
	if len(sm) != 4 {
		t.Error("Error pasring map: ", sm)
	}
}

func TestStringMapParseNegative(t *testing.T) {
	sm := ParseStringMap("1;2;!3;4")
	if len(sm) != 4 {
		t.Error("Error pasring map: ", sm)
	}
	if sm["3"] != false {
		t.Error("Error parsing negative: ", sm)
	}
}

func TestStringMapCompare(t *testing.T) {
	sm := ParseStringMap("1;2;!3;4")
	if include, found := sm["2"]; include != true && found != true {
		t.Error("Error detecting positive: ", sm)
	}
	if include, found := sm["3"]; include != false && found != true {
		t.Error("Error detecting negative: ", sm)
	}
	if include, found := sm["5"]; include != false && found != false {
		t.Error("Error detecting missing: ", sm)
	}
}

func TestMapMergeMapsStringIface(t *testing.T) {
	mp1 := map[string]interface{}{
		"Hdr1": "Val1",
		"Hdr2": "Val2",
		"Hdr3": "Val3",
	}
	mp2 := map[string]interface{}{
		"Hdr3": "Val4",
		"Hdr4": "Val4",
	}
	eMergedMap := map[string]interface{}{
		"Hdr1": "Val1",
		"Hdr2": "Val2",
		"Hdr3": "Val4",
		"Hdr4": "Val4",
	}
	if mergedMap := MergeMapsStringIface(mp1, mp2); !reflect.DeepEqual(eMergedMap, mergedMap) {
		t.Errorf("Expecting: %+v, received: %+v", eMergedMap, mergedMap)
	}
}
