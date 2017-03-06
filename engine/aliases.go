/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package engine

import (
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// Temporary export AliasService for the ApierV1 to be able to emulate old APIs
func GetAliasService() rpcclient.RpcClientConnection {
	return aliasService
}

type Alias struct {
	Direction string
	Tenant    string
	Category  string
	Account   string
	Subject   string
	Context   string
	Values    AliasValues
}

func (al *Alias) ReverseAliasIDs() (rvAl []string) {
	for _, value := range al.Values {
		for target, pairs := range value.Pairs {
			for _, alias := range pairs {
				rvAl = append(rvAl, strings.Join([]string{alias, target, al.Context}, ""))
			}
		}
	}
	return
}

type AliasValue struct {
	DestinationId string
	Pairs         AliasPairs
	Weight        float64
}

func (av *AliasValue) Equals(other *AliasValue) bool {
	return av.DestinationId == other.DestinationId &&
		av.Pairs.Equals(other.Pairs) &&
		av.Weight == other.Weight
}

type AliasPairs map[string]map[string]string

func (ap AliasPairs) Equals(other AliasPairs) bool {
	if len(ap) != len(other) {
		return false
	}

	for attribute, origAlias := range ap {
		otherOrigAlias, ok := other[attribute]
		if !ok || len(origAlias) != len(otherOrigAlias) {
			return false
		}
		for orig := range origAlias {
			if origAlias[orig] != otherOrigAlias[orig] {
				return false
			}
		}
	}
	return true
}

type AliasValues []*AliasValue

func (avs AliasValues) Len() int {
	return len(avs)
}

func (avs AliasValues) Swap(i, j int) {
	avs[i], avs[j] = avs[j], avs[i]
}

func (avs AliasValues) Less(j, i int) bool { // get higher weight in front
	return avs[i].Weight < avs[j].Weight
}

func (avs AliasValues) Sort() {
	sort.Sort(avs)
}

func (avs AliasValues) GetValueByDestId(destID string) *AliasValue {
	for _, value := range avs {
		if value.DestinationId == destID {
			return value
		}
	}
	return nil
}

func (al *Alias) GetId() string {
	return utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Context)
}

func (al *Alias) GenerateIds() []string {
	var result []string
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Context))
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, utils.ANY, al.Context))
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, utils.ANY, utils.ANY, al.Context))
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, utils.ANY, utils.ANY, utils.ANY, al.Context))
	result = append(result, utils.ConcatenatedKey(utils.ANY, utils.ANY, al.Category, utils.ANY, utils.ANY, al.Context))
	result = append(result, utils.ConcatenatedKey(utils.ANY, utils.ANY, utils.ANY, utils.ANY, al.Context))
	return result
}

func (al *Alias) SetId(id string) error {
	vals := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(vals) != 6 {
		return utils.ErrInvalidKey
	}
	al.Direction = vals[0]
	al.Tenant = vals[1]
	al.Category = vals[2]
	al.Account = vals[3]
	al.Subject = vals[4]
	al.Context = vals[5]
	return nil
}

type AttrMatchingAlias struct {
	Destination string
	Direction   string
	Tenant      string
	Category    string
	Account     string
	Subject     string
	Context     string
	Target      string
	Original    string
}

type AttrReverseAlias struct {
	Alias   string
	Target  string
	Context string
}

type AliasService interface {
	SetAlias(Alias, *string) error
	UpdateAlias(Alias, *string) error
	RemoveAlias(Alias, *string) error
	GetAlias(Alias, *Alias) error
	GetMatchingAlias(AttrMatchingAlias, *string) error
	GetReverseAlias(AttrReverseAlias, *map[string][]*Alias) error
	RemoveReverseAlias(AttrReverseAlias, *string) error
}

type AliasHandler struct {
	dataDB DataDB
	mu     sync.RWMutex
}

func NewAliasHandler(dataDB DataDB) *AliasHandler {
	return &AliasHandler{
		dataDB: dataDB,
	}
}

type AttrAddAlias struct {
	Alias     *Alias
	Overwrite bool
}

// SetAlias will set/overwrite specified alias
func (am *AliasHandler) SetAlias(attr *AttrAddAlias, reply *string) (err error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	var oldAlias *Alias
	if !attr.Overwrite { // get previous value
		oldAlias, _ = am.dataDB.GetAlias(attr.Alias.GetId(), false, utils.NonTransactional)
	}
	if attr.Overwrite || oldAlias == nil {
		if err = am.dataDB.SetAlias(attr.Alias, utils.NonTransactional); err != nil {
			return err
		}
		if err = am.dataDB.CacheDataFromDB(utils.ALIASES_PREFIX, []string{attr.Alias.GetId()}, true); err != nil {
			return
		}
		if err = am.dataDB.SetReverseAlias(attr.Alias, utils.NonTransactional); err != nil {
			return
		}
		if err = am.dataDB.CacheDataFromDB(utils.REVERSE_ALIASES_PREFIX, attr.Alias.ReverseAliasIDs(), true); err != nil {
			return
		}
	} else {
		for _, value := range attr.Alias.Values {
			found := false
			if value.DestinationId == "" {
				value.DestinationId = utils.ANY
			}
			for _, oldValue := range oldAlias.Values {
				if oldValue.DestinationId == value.DestinationId {
					for target, origAliasMap := range value.Pairs {
						for orig, alias := range origAliasMap {
							if oldValue.Pairs[target] == nil {
								oldValue.Pairs[target] = make(map[string]string)
							}
							oldValue.Pairs[target][orig] = alias
						}
					}
					oldValue.Weight = value.Weight
					found = true
					break
				}
			}
			if !found {
				oldAlias.Values = append(oldAlias.Values, value)
			}
		}
		if err = am.dataDB.SetAlias(oldAlias, utils.NonTransactional); err != nil {
			return
		}
		if err = am.dataDB.CacheDataFromDB(utils.ALIASES_PREFIX, []string{oldAlias.GetId()}, true); err != nil {
			return
		}
		if err = am.dataDB.SetReverseAlias(oldAlias, utils.NonTransactional); err != nil {
			return
		}
		if err = am.dataDB.CacheDataFromDB(utils.REVERSE_ALIASES_PREFIX, attr.Alias.ReverseAliasIDs(), true); err != nil {
			return
		}
	}

	*reply = utils.OK
	return nil
}

func (am *AliasHandler) RemoveAlias(al *Alias, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	if err := am.dataDB.RemoveAlias(al.GetId(), utils.NonTransactional); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

func (am *AliasHandler) GetAlias(al *Alias, result *Alias) error {
	am.mu.RLock()
	defer am.mu.RUnlock()
	variants := al.GenerateIds()
	for _, variant := range variants {
		if r, err := am.dataDB.GetAlias(variant, false, utils.NonTransactional); err == nil && r != nil {
			*result = *r
			return nil
		}
	}
	return utils.ErrNotFound
}

func (am *AliasHandler) GetReverseAlias(attr *AttrReverseAlias, result *map[string][]*Alias) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	aliases := make(map[string][]*Alias)
	rKey := attr.Alias + attr.Target + attr.Context
	if ids, err := am.dataDB.GetReverseAlias(rKey, false, utils.NonTransactional); err == nil {
		for _, key := range ids {
			// get destination id
			elems := strings.Split(key, utils.CONCATENATED_KEY_SEP)
			var destID string
			if len(elems) > 0 {
				destID = elems[len(elems)-1]
				key = strings.Join(elems[:len(elems)-1], utils.CONCATENATED_KEY_SEP)
			}
			if r, err := am.dataDB.GetAlias(key, false, utils.NonTransactional); err != nil {
				return err
			} else {
				aliases[destID] = append(aliases[destID], r)
			}
		}
	}
	*result = aliases
	return nil
}

func (am *AliasHandler) GetMatchingAlias(attr *AttrMatchingAlias, result *string) error {
	response := Alias{}
	if err := am.GetAlias(&Alias{
		Direction: attr.Direction,
		Tenant:    attr.Tenant,
		Category:  attr.Category,
		Account:   attr.Account,
		Subject:   attr.Subject,
		Context:   attr.Context,
	}, &response); err != nil {
		return err
	}
	// sort according to weight
	values := response.Values
	values.Sort()

	// if destination does not metter get first alias
	if attr.Destination == "" || attr.Destination == utils.ANY {
		for _, value := range values {
			if origAlias, ok := value.Pairs[attr.Target]; ok {
				if alias, ok := origAlias[attr.Original]; ok {
					*result = alias
					return nil
				}
			}
		}
		return utils.ErrNotFound
	}
	// check destination ids
	for _, p := range utils.SplitPrefix(attr.Destination, MIN_PREFIX_MATCH) {
		if destIDs, err := dataStorage.GetReverseDestination(p, false, utils.NonTransactional); err == nil {
			for _, value := range values {
				for _, dId := range destIDs {
					if value.DestinationId == utils.ANY || value.DestinationId == dId {
						if origAliasMap, ok := value.Pairs[attr.Target]; ok {
							if alias, ok := origAliasMap[attr.Original]; ok || attr.Original == "" || attr.Original == utils.ANY {
								*result = alias
								return nil
							}
							if alias, ok := origAliasMap[utils.ANY]; ok {
								*result = alias
								return nil
							}
						}
					}
				}
			}
		}
	}
	return utils.ErrNotFound
}

func (am *AliasHandler) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(am).MethodByName(parts[1])
	if !method.IsValid() {
		return utils.ErrNotImplemented
	}

	// construct the params
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}

	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}

func LoadAlias(attr *AttrMatchingAlias, in interface{}, extraFields string) error {
	if aliasService == nil {
		return nil
	}
	response := Alias{}
	if err := aliasService.Call("AliasesV1.GetAlias", &Alias{
		Direction: attr.Direction,
		Tenant:    attr.Tenant,
		Category:  attr.Category,
		Account:   attr.Account,
		Subject:   attr.Subject,
		Context:   attr.Context,
	}, &response); err != nil {
		return err
	}

	// sort according to weight
	values := response.Values
	values.Sort()

	var rightPairs AliasPairs
	// if destination does not metter get first alias
	if attr.Destination == "" || attr.Destination == utils.ANY {
		rightPairs = values[0].Pairs
	}

	if rightPairs == nil {
		// check destination ids
		for _, p := range utils.SplitPrefix(attr.Destination, MIN_PREFIX_MATCH) {
			if destIDs, err := dataStorage.GetReverseDestination(p, false, utils.NonTransactional); err == nil {
				for _, value := range values {
					for _, dId := range destIDs {
						if value.DestinationId == utils.ANY || value.DestinationId == dId {
							rightPairs = value.Pairs
						}
						if rightPairs != nil {
							break
						}
					}
					if rightPairs != nil {
						break
					}
				}
			}
			if rightPairs != nil {
				break
			}
		}
	}

	if rightPairs != nil {
		// change values in the given object
		v := reflect.ValueOf(in)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		for target, originalAlias := range rightPairs {
			for original, alias := range originalAlias {
				field := v.FieldByName(target)
				if field.IsValid() {
					if field.Kind() == reflect.String {
						if field.CanSet() && (original == "" || original == utils.ANY || field.String() == original) {
							field.SetString(alias)
						}
					}
				}
				if extraFields != "" {
					efField := v.FieldByName(extraFields)
					if efField.IsValid() && efField.Kind() == reflect.Map {
						keys := efField.MapKeys()
						for _, key := range keys {
							if key.Kind() == reflect.String && key.String() == target {
								if original == "" || original == utils.ANY || efField.MapIndex(key).String() == original {
									efField.SetMapIndex(key, reflect.ValueOf(alias))
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}
