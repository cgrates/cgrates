package engine

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

func csvLoad(s interface{}, values []string) (interface{}, error) {
	fieldValueMap := make(map[string]string)
	st := reflect.TypeOf(s)
	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		re := field.Tag.Get("re")
		index := field.Tag.Get("index")
		if index != "" {
			idx, err := strconv.Atoi(index)
			if err != nil || len(values) <= idx {
				return nil, fmt.Errorf("invalid %v.%v index %v", st.Name(), field.Name, index)
			}
			if re != "" {
				if matched, err := regexp.MatchString(re, values[idx]); !matched || err != nil {
					return nil, fmt.Errorf("invalid %v.%v value %v", st.Name(), field.Name, values[idx])
				}
			}
			fieldValueMap[field.Name] = values[idx]
		}
	}
	elem := reflect.New(st).Elem()
	for fildName, fieldValue := range fieldValueMap {
		field := elem.FieldByName(fildName)
		if field.IsValid() && field.CanSet() {
			switch field.Kind() {
			case reflect.Float64:
				value, err := strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					return nil, fmt.Errorf(`invalid value "%s" for field %s.%s`, fieldValue, st.Name(), fildName)
				}
				field.SetFloat(value)
			case reflect.String:
				field.SetString(fieldValue)
			}
		}
	}
	return elem.Interface(), nil
}

type TpDestinations []*TpDestination

func (tps TpDestinations) GetDestinations() (dsts []*Destination) {
	destinations := make(map[string]*Destination)
	for _, tpDest := range tps {
		var dest *Destination
		var found bool
		if dest, found = destinations[tpDest.Tag]; !found {
			dest = &Destination{Id: tpDest.Tag}
			destinations[tpDest.Tag] = dest
		}
		dest.AddPrefix(tpDest.Prefix)
	}
	for _, dest := range destinations {
		dsts = append(dsts, dest)
	}
	return
}
