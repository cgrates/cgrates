package utils

import "testing"

func TestCondLoader(t *testing.T) {
	cl := &CondLoader{}
	err := cl.Parse(`{"*or":[{"test":1},{"field":{"*gt":1}},{"best":"coco"}]}`)
	if err != nil || cl.RootElement == nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.RootElement), err)
	}
}
