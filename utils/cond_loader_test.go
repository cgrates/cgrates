package utils

import "testing"

func TestCondLoader(t *testing.T) {
	cl := &CondLoader{}
	err := cl.Parse(`{"*or":[{"test":1},{"field":{"*gt":1}},{"best":"coco"}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
}

func TestCondKeyValue(t *testing.T) {
	o := struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &CondLoader{}
	err := cl.Parse(`{"Test":"test"}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
}

func TestCondKeyValuePointer(t *testing.T) {
	o := &struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &CondLoader{}
	err := cl.Parse(`{"Test":"test"}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
}

func TestCondOperatorValue(t *testing.T) {
	root := &operatorValue{operator: "*gt", value: 3.4}
	if check, err := root.checkStruct(3.5); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(root))
	}
	root = &operatorValue{operator: "*eq", value: 3.4}
	if check, err := root.checkStruct(3.5); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(root))
	}
	root = &operatorValue{operator: "*eq", value: 3.4}
	if check, err := root.checkStruct(3.4); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(root))
	}
	root = &operatorValue{operator: "*eq", value: "zinc"}
	if check, err := root.checkStruct("zinc"); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(root))
	}
}

func TestCondKeyStruct(t *testing.T) {
	o := struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &CondLoader{}
	err := cl.Parse(`{"Field":{"*gt": 5}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"*gt": 5}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); check || err != ErrNotNumerical {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*gte": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*lt": 7}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*lte": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*eq": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"*eq": "test"}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
}

func TestCondKeyStructPointer(t *testing.T) {
	o := &struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &CondLoader{}
	err := cl.Parse(`{"Field":{"*gt": 5}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"*gt": 5}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); check || err != ErrNotNumerical {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*gte": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*lt": 7}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*lte": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"*eq": 6}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"*eq": "test"}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
}

func TestCondOperatorSlice(t *testing.T) {
	o := &struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &CondLoader{}
	err := cl.Parse(`{"*or":[{"Test":"test"},{"Field":{"*gt":5}},{"Other":true}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"*or":[{"Test":"test"},{"Field":{"*gt":7}},{"Other":false}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"*and":[{"Test":"test"},{"Field":{"*gt":5}},{"Other":true}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"*and":[{"Test":"test"},{"Field":{"*gt":7}},{"Other":false}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
}
