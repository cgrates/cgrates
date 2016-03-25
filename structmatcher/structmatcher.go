package structmatcher

/*
The condition syntax is a json encoded string similar to mongodb query language.

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
- *not: logical not
- *has: receives a list of elements and checks that the elements are present in the specified field (StringMap type)
- *rsr: will apply a rsr check to the field (see utils/rsrfield.go)

Equal (*eq) and local and (*and) operators are implicit for shortcuts. In this way:

{"*and":[{"Value":{"*eq":3}},{"Weight":{"*eq":10}}]} is equivalent to: {"Value":3, "Weight":10}.
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
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
	CondNOT = "*not"
	CondHAS = "*has"
	CondRSR = "*rsr"
)

func NewErrInvalidArgument(arg interface{}) error {
	return fmt.Errorf("INVALID_ARGUMENT: %v", arg)
}

type StringMap map[string]bool

var (
	ErrParserError = errors.New("PARSER_ERROR")

	operatorMap = map[string]func(field, value interface{}) (bool, error){
		CondEQ: func(field, value interface{}) (bool, error) {
			return value == field, nil
		},
		CondGT: func(field, value interface{}) (bool, error) {
			var of, vf float64
			var ok bool
			if of, ok = field.(float64); !ok {
				return false, NewErrInvalidArgument(field)
			}
			if vf, ok = value.(float64); !ok {
				return false, NewErrInvalidArgument(value)
			}
			return of > vf, nil
		},
		CondGTE: func(field, value interface{}) (bool, error) {
			var of, vf float64
			var ok bool
			if of, ok = field.(float64); !ok {
				return false, NewErrInvalidArgument(field)
			}
			if vf, ok = value.(float64); !ok {
				return false, NewErrInvalidArgument(value)
			}
			return of >= vf, nil
		},
		CondLT: func(field, value interface{}) (bool, error) {
			var of, vf float64
			var ok bool
			if of, ok = field.(float64); !ok {
				return false, NewErrInvalidArgument(field)
			}
			if vf, ok = value.(float64); !ok {
				return false, NewErrInvalidArgument(value)
			}
			return of < vf, nil
		},
		CondLTE: func(field, value interface{}) (bool, error) {
			var of, vf float64
			var ok bool
			if of, ok = field.(float64); !ok {
				return false, NewErrInvalidArgument(field)
			}
			if vf, ok = value.(float64); !ok {
				return false, NewErrInvalidArgument(value)
			}
			return of <= vf, nil
		},
		CondEXP: func(field, value interface{}) (bool, error) {
			var expDate time.Time
			var ok bool
			if expDate, ok = field.(time.Time); !ok {
				return false, NewErrInvalidArgument(field)
			}
			var expired bool
			if expired, ok = value.(bool); !ok {
				return false, NewErrInvalidArgument(value)
			}
			if expired { // check for expiration
				return !expDate.IsZero() && expDate.Before(time.Now()), nil
			} else { // check not expired
				return expDate.IsZero() || expDate.After(time.Now()), nil
			}
		},
		CondHAS: func(field, value interface{}) (bool, error) {
			var strMap StringMap
			var ok bool
			if strMap, ok = field.(StringMap); !ok {
				return false, NewErrInvalidArgument(field)
			}
			var strSlice []interface{}
			if strSlice, ok = value.([]interface{}); !ok {
				return false, NewErrInvalidArgument(value)
			}
			for _, str := range strSlice {
				if !strMap[str.(string)] {
					return false, nil
				}
			}
			return true, nil
		},
		CondRSR: func(field, value interface{}) (bool, error) {
			fltr, err := utils.NewRSRFilter(value.(string))
			if err != nil {
				return false, err
			}
			return fltr.Pass(fmt.Sprintf("%v", field)), nil
		},
	}
)

type compositeElement interface {
	element
	addChild(element) error
}

type element interface {
	checkStruct(interface{}) (bool, error)
}

type operatorSlice struct {
	operator string
	slice    []element
}

func (os *operatorSlice) addChild(ce element) error {
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
	case CondAND, CondNOT:
		accumulator := true
		for _, cond := range os.slice {
			check, err := cond.checkStruct(o)
			if err != nil {
				return false, err
			}
			accumulator = accumulator && check
		}
		if os.operator == CondAND {
			return accumulator, nil
		} else {
			return !accumulator, nil
		}
	}
	return false, nil
}

type keyStruct struct {
	key  string
	elem element
}

func (ks *keyStruct) addChild(ce element) error {
	ks.elem = ce
	return nil
}
func (ks *keyStruct) checkStruct(o interface{}) (bool, error) {
	obj := reflect.ValueOf(o)
	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}
	value := obj.FieldByName(ks.key)
	if !value.IsValid() {
		return false, NewErrInvalidArgument(ks.key)
	}
	return ks.elem.checkStruct(value.Interface())
}

type operatorValue struct {
	operator string
	value    interface{}
}

func (ov *operatorValue) checkStruct(o interface{}) (bool, error) {
	if f, ok := operatorMap[ov.operator]; ok {
		return f(o, ov.value)
	}
	return false, nil
}

type keyValue struct {
	key   string
	value interface{}
}

func (kv *keyValue) checkStruct(o interface{}) (bool, error) {
	obj := reflect.ValueOf(o)
	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}
	value := obj.FieldByName(kv.key)
	if !value.IsValid() {
		return false, NewErrInvalidArgument(kv.key)
	}
	return value.Interface() == kv.value, nil
}

type trueElement struct{}

func (te *trueElement) checkStruct(o interface{}) (bool, error) {
	return true, nil
}

func isOperator(s string) bool {
	return strings.HasPrefix(s, "*")
}

func notEmpty(x interface{}) bool {
	return !reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

type StructMatcher struct {
	rootElement element
}

func NewStructMatcher(q string) (sm *StructMatcher, err error) {
	sm = &StructMatcher{}
	err = sm.Parse(q)
	return
}

func (sm *StructMatcher) load(a map[string]interface{}, parentElement compositeElement) (element, error) {
	for key, value := range a {
		var currentElement element
		switch t := value.(type) {
		case []interface{}:
			if key == CondHAS {
				currentElement = &operatorValue{operator: key, value: t}
			} else {
				currentElement = &operatorSlice{operator: key}
				for _, e := range t {
					sm.load(e.(map[string]interface{}), currentElement.(compositeElement))
				}
			}
		case map[string]interface{}:
			currentElement = &keyStruct{key: key}
			//log.Print("map: ", t)
			sm.load(t, currentElement.(compositeElement))
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
		if parentElement != nil { // normal recurrent action
			parentElement.addChild(currentElement)
		} else {
			if len(a) > 1 { // we have more keys in the map
				parentElement = &operatorSlice{operator: CondAND}
				parentElement.addChild(currentElement)
			} else { // it was only one key value
				return currentElement, nil
			}
		}
	}
	return parentElement, nil
}

func (sm *StructMatcher) Parse(s string) (err error) {
	a := make(map[string]interface{})
	if len(s) != 0 {
		if err := json.Unmarshal([]byte([]byte(s)), &a); err != nil {
			return err
		}
		sm.rootElement, err = sm.load(a, nil)
	} else {
		sm.rootElement = &trueElement{}
	}
	return
}

func (sm *StructMatcher) Match(o interface{}) (bool, error) {
	if sm.rootElement == nil {
		return false, ErrParserError
	}
	return sm.rootElement.checkStruct(o)
}
