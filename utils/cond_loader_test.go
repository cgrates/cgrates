package utils

import (
	"strings"
	"testing"
	"time"
)

func TestCondLoader(t *testing.T) {
	cl := &CondLoader{}
	err := cl.Parse(`{"*or":[{"test":1},{"field":{"*gt":1}},{"best":"coco"}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}

	err = cl.Parse(`{"*has":["NAT","RET","EUR"]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	err = cl.Parse(`{"Field":7, "Other":true}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	err = cl.Parse(``)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
}

func TestCondKeyValue(t *testing.T) {
	o := struct {
		Test    string
		Field   float64
		Other   bool
		ExpDate time.Time
	}{
		Test:    "test",
		Field:   6.0,
		Other:   true,
		ExpDate: time.Date(2016, 1, 19, 20, 47, 0, 0, time.UTC),
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
	err = cl.Parse(`{"Field":6, "Other":true}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":7, "Other":true}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6, "Other":false}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true, "Field":{"*gt":5}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true, "Field":{"*gt":7}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(``)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"ExpDate":{"*exp":true}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
	err = cl.Parse(`{"ExpDate":{"*exp":false}}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); check || err != nil {
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
	root := &operatorValue{operator: CondGT, value: 3.4}
	if check, err := root.checkStruct(3.5); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(root))
	}
	root = &operatorValue{operator: CondEQ, value: 3.4}
	if check, err := root.checkStruct(3.5); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(root))
	}
	root = &operatorValue{operator: CondEQ, value: 3.4}
	if check, err := root.checkStruct(3.4); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(root))
	}
	root = &operatorValue{operator: CondEQ, value: "zinc"}
	if check, err := root.checkStruct("zinc"); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(root))
	}
	root = &operatorValue{operator: CondHAS, value: []interface{}{"NAT", "RET", "EUR"}}
	if check, err := root.checkStruct(StringMap{"WOR": true, "EUR": true, "NAT": true, "RET": true, "ROM": true}); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", !check, err, ToIJSON(root))
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
	if check, err := cl.Check(o); check || !strings.HasPrefix(err.Error(), "INVALID_ARGUMENT") {
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
	if check, err := cl.Check(o); check || !strings.HasPrefix(err.Error(), "INVALID_ARGUMENT") {
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

func TestCondMixed(t *testing.T) {
	o := &struct {
		Test       string
		Field      float64
		Categories StringMap
		Other      bool
	}{
		Test:       "test",
		Field:      6.0,
		Categories: StringMap{"call": true, "data": true, "voice": true},
		Other:      true,
	}
	cl := &CondLoader{}
	err := cl.Parse(`{"*and":[{"Test":"test"},{"Field":{"*gt":5}},{"Other":true},{"Categories":{"*has":["data", "call"]}}]}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, ToIJSON(cl.rootElement))
	}
}

func TestCondBalanceType(t *testing.T) {
	type Balance struct {
		Value float64
	}

	o := &struct {
		BalanceType string
		Balance
	}{
		BalanceType: MONETARY,
		Balance:     Balance{Value: 10},
	}
	cl := &CondLoader{}
	err := cl.Parse(`{"BalanceType":"*monetary","Value":10}`)
	if err != nil {
		t.Errorf("Error loading structure: %+v (%v)", ToIJSON(cl.rootElement), err)
	}
	if check, err := cl.Check(o); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", !check, err, ToIJSON(cl.rootElement))
	}
}
