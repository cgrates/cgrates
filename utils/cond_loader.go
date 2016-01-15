package utils

import (
	"encoding/json"
	"strings"
)

const (
	COND_EQ  = "*eq"
	COND_GT  = "*gt"
	COND_LT  = "*lt"
	COND_EXP = "*exp"
)

type CondElement interface {
	AddChild(CondElement)
	CheckStruct(interface{}) (bool, error)
}

type OperatorSlice struct {
	Operator string
	Slice    []CondElement
}

func (os *OperatorSlice) AddChild(ce CondElement) {
	os.Slice = append(os.Slice, ce)
}
func (os *OperatorSlice) CheckStruct(o interface{}) (bool, error) { return true, nil }

type KeyStruct struct {
	Key    string
	Struct CondElement
}

func (ks *KeyStruct) AddChild(ce CondElement) {
	ks.Struct = ce
}
func (ks *KeyStruct) CheckStruct(o interface{}) (bool, error) { return true, nil }

type OperatorValue struct {
	Operator string
	Value    interface{}
}

func (ov *OperatorValue) AddChild(CondElement)                    {}
func (ov *OperatorValue) CheckStruct(o interface{}) (bool, error) { return true, nil }

type KeyValue struct {
	Key   string
	Value interface{}
}

func (os *KeyValue) AddChild(CondElement)                    {}
func (os *KeyValue) CheckStruct(o interface{}) (bool, error) { return true, nil }

func isOperator(s string) bool {
	return strings.HasPrefix(s, "*")
}

type CondLoader struct {
	RootElement CondElement
}

func (cp *CondLoader) Load(a map[string]interface{}, parentElement CondElement) (CondElement, error) {
	for key, value := range a {
		var currentElement CondElement
		switch t := value.(type) {
		case []interface{}:
			currentElement = &OperatorSlice{Operator: key}
			for _, e := range t {
				cp.Load(e.(map[string]interface{}), currentElement)
			}
		case map[string]interface{}:
			currentElement = &KeyStruct{Key: key}
			//log.Print("map: ", t)
			cp.Load(t, currentElement)
		case interface{}:
			if isOperator(key) {
				currentElement = &OperatorValue{Operator: key, Value: t}
			} else {
				currentElement = &KeyValue{Key: key, Value: t}
			}
			//log.Print("generic interface: ", t)
		default:
			return nil, ErrParserError
		}
		if parentElement != nil {
			parentElement.AddChild(currentElement)
		} else {
			return currentElement, nil
		}
	}
	return nil, nil
}

func (cp *CondLoader) Parse(s string) (err error) {
	a := make(map[string]interface{})
	json.Unmarshal([]byte([]byte(s)), &a)
	cp.RootElement, err = cp.Load(a, nil)
	return
}
