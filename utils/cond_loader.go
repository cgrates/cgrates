package utils

/*
When an action is using *conditional_ form before the execution the engine is checking the ExtraParameters field for condition filter, loads it and checks all the balances in the account for one that is satisfying the condition. If one is fond the action is executed, otherwise it will do nothing for this account.

The condition syntax is a json encoded document similar to mongodb query language.

Examples:
- {"Weight":{"*gt":50}} checks for a balance with weight greater than 50
- {"*or":[{"Value":{"*eq":0}},{"Value":{"*gte":100}}] checks for a balance with value equal to 0 or equal or highr than 100

Available operators:
- *eq: equal
- *gt: greater than
- *gte: greater or equal than
- *lt: less then
- *lte: less or equal than
- *exp: expired
- *or: logical or
- *and: logical and
- *has: receives a list of elements and checks that the elements are present in the specified field (also a list)

Equal (*eq) and local and (*and) operators are implicit for shortcuts. In this way:

{"*and":[{"Value":{"*eq":3}},{"Weight":{"*eq":10}}]} is equivalent to: {"Value":3, "Weight":10}.
*/

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
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
	CondHAS = "*has"
)

func NewErrInvalidArgument(arg interface{}) error {
	return fmt.Errorf("INVALID_ARGUMENT: %v", arg)
}

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
	// no conversion
	if ov.operator == CondEQ {
		return ov.value == o, nil
	}
	// StringMap conversion
	if ov.operator == CondHAS {
		var strMap StringMap
		var ok bool
		if strMap, ok = o.(StringMap); !ok {
			return false, NewErrInvalidArgument(o)
		}
		var strSlice []interface{}
		if strSlice, ok = ov.value.([]interface{}); !ok {
			return false, NewErrInvalidArgument(ov.value)
		}
		for _, str := range strSlice {
			if !strMap[str.(string)] {
				return false, nil
			}
		}
		return true, nil
	}
	// date conversion
	if ov.operator == CondEXP {
		var expDate time.Time
		var ok bool
		if expDate, ok = o.(time.Time); !ok {
			return false, NewErrInvalidArgument(o)
		}
		var expired bool
		if expired, ok = ov.value.(bool); !ok {
			return false, NewErrInvalidArgument(ov.value)
		}
		if expired { // check for expiration
			return !expDate.IsZero() && expDate.Before(time.Now()), nil
		} else { // check not expired
			return expDate.IsZero() || expDate.After(time.Now()), nil
		}
	}

	// float conversion
	var of, vf float64
	var ok bool
	if of, ok = o.(float64); !ok {
		return false, NewErrInvalidArgument(o)
	}
	if vf, ok = ov.value.(float64); !ok {
		return false, NewErrInvalidArgument(ov.value)
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

type trueElement struct{}

func (te *trueElement) addChild(condElement) error { return ErrNotImplemented }
func (te *trueElement) checkStruct(o interface{}) (bool, error) {
	return true, nil
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
			if key == CondHAS {
				currentElement = &operatorValue{operator: key, value: t}
			} else {
				currentElement = &operatorSlice{operator: key}
				for _, e := range t {
					cp.load(e.(map[string]interface{}), currentElement)
				}
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
			if len(a) > 1 {
				parentElement = &operatorSlice{operator: CondAND}
				parentElement.addChild(currentElement)
			} else {
				return currentElement, nil
			}
		}
	}
	return parentElement, nil
}

func (cp *CondLoader) Parse(s string) (err error) {
	a := make(map[string]interface{})
	if len(s) != 0 {
		if err := json.Unmarshal([]byte([]byte(s)), &a); err != nil {
			return err
		}
		cp.rootElement, err = cp.load(a, nil)
	} else {
		cp.rootElement = &trueElement{}
	}
	return
}

func (cp *CondLoader) Check(o interface{}) (bool, error) {
	if cp.rootElement == nil {
		return false, ErrParserError
	}
	return cp.rootElement.checkStruct(o)
}
