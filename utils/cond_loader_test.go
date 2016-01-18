package utils

import "testing"

func TestCondLoader(t *testing.T) {
	cl := &CondLoader{}
	root, err := cl.Parse(`{"*or":[{"test":1},{"field":{"*gt":1}},{"best":"coco"}]}`)
	if err != nil || root == nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(root), err)
	}
}
