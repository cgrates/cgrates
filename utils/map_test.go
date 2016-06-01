package utils

import "testing"

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
