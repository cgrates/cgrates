package utils

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

const (
	CondEQ  = "*eq"
	CondGT  = "*gt"
	CondGTE = "*gte"
	CondLT  = "*lt"
	CondLTE = "*lte"
	CondEXP = "*exp"
	CondOR  = "*or"
	CondAND = "*and"
)

var (
	ErrNotNumerical = errors.New("NOT_NUMERICAL")
)

type condElement interface {
	addChild(condElement) error
	checkStruct(interface{}) (bool, error)
}

type operatorSlice struct {
	operator string
	slice    []condElement
}

func (os *operatorSlice) addChild(ce condElement) error {
	os.slice = append(os.slice, ce)
	return nil
}
func (os *operatorSlice) checkStruct(o interface{}) (bool, error) {
	switch os.operator {
	case CondOR:
		for _, cond := range os.slice {
			check, err := cond.checkStruct(o)
			if err != nil {
				return false, err
			}
			if check {
				return true, nil
			}
		}
	case CondAND:
		accumulator := true
		for _, cond := range os.slice {
			check, err := cond.checkStruct(o)
			if err != nil {
				return false, err
			}
			accumulator = accumulator && check
		}
		return accumulator, nil
	}
	return false, nil
}

type keyStruct struct {
	key  string
	elem condElement
}

func (ks *keyStruct) addChild(ce condElement) error {
	ks.elem = ce
	return nil
}
func (ks *keyStruct) checkStruct(o interface{}) (bool, error) {
	obj := reflect.ValueOf(o)
	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}
	value := obj.FieldByName(ks.key)
	return ks.elem.checkStruct(value.Interface())
}

type operatorValue struct {
	operator string
	value    interface{}
}

func (ov *operatorValue) addChild(condElement) error { return ErrNotImplemented }
func (ov *operatorValue) checkStruct(o interface{}) (bool, error) {
	if ov.operator == CondEQ {
		return ov.value == o, nil
	}
	var of, vf float64
	var ok bool
	if of, ok = o.(float64); !ok {
		return false, ErrNotNumerical
	}
	if vf, ok = ov.value.(float64); !ok {
		return false, ErrNotNumerical
	}
	switch ov.operator {
	case CondGT:
		return of > vf, nil
	case CondGTE:
		return of >= vf, nil
	case CondLT:
		return of < vf, nil
	case CondLTE:
		return of <= vf, nil

	}
	return true, nil
}

type keyValue struct {
	key   string
	value interface{}
}

func (kv *keyValue) addChild(condElement) error { return ErrNotImplemented }
func (kv *keyValue) checkStruct(o interface{}) (bool, error) {
	obj := reflect.ValueOf(o)
	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}
	value := obj.FieldByName(kv.key)
	return value.Interface() == kv.value, nil
}

func isOperator(s string) bool {
	return strings.HasPrefix(s, "*")
}

type CondLoader struct {
	rootElement condElement
}

func (cp *CondLoader) load(a map[string]interface{}, parentElement condElement) (condElement, error) {
	for key, value := range a {
		var currentElement condElement
		switch t := value.(type) {
		case []interface{}:
			currentElement = &operatorSlice{operator: key}
			for _, e := range t {
				cp.load(e.(map[string]interface{}), currentElement)
			}
		case map[string]interface{}:
			currentElement = &keyStruct{key: key}
			//log.Print("map: ", t)
			cp.load(t, currentElement)
		case interface{}:
			if isOperator(key) {
				currentElement = &operatorValue{operator: key, value: t}
			} else {
				currentElement = &keyValue{key: key, value: t}
			}
			//log.Print("generic interface: ", t)
		default:
			return nil, ErrParserError
		}
		if parentElement != nil {
			parentElement.addChild(currentElement)
		} else {
			return currentElement, nil
		}
	}
	return nil, nil
}

func (cp *CondLoader) Parse(s string) (err error) {
	a := make(map[string]interface{})
	if err := json.Unmarshal([]byte([]byte(s)), &a); err != nil {
		return err
	}
	cp.rootElement, err = cp.load(a, nil)
	return
}

func (cp *CondLoader) Check(o interface{}) (bool, error) {
	return cp.rootElement.checkStruct(o)
}
