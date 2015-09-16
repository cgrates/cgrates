package engine

import (
	"sort"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// Temporary export AliasService for the ApierV1 to be able to emulate old APIs
func GetAliasService() AliasService {
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

func (al *Alias) GetId() string {
	return utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Context)
}

func (al *Alias) GenerateIds() []string {
	var result []string
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Context))
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, utils.ANY, al.Context))
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, utils.ANY, utils.ANY, al.Context))
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, utils.ANY, utils.ANY, utils.ANY, al.Context))
	result = append(result, utils.ConcatenatedKey(al.Direction, utils.ANY, utils.ANY, utils.ANY, utils.ANY, al.Context))
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
	accountingDb AccountingStorage
	mu           sync.RWMutex
}

func NewAliasHandler(accountingDb AccountingStorage) *AliasHandler {
	return &AliasHandler{
		accountingDb: accountingDb,
	}
}

func (am *AliasHandler) SetAlias(al Alias, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if err := am.accountingDb.SetAlias(&al); err != nil {
		*reply = err.Error()
		return err
	} //add to cache

	aliasesChanged := []string{utils.ALIASES_PREFIX + al.GetId()}
	if err := am.accountingDb.CacheAccountingPrefixValues(map[string][]string{utils.ALIASES_PREFIX: aliasesChanged}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (am *AliasHandler) UpdateAlias(al Alias, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	// get previous value
	oldAlias := &Alias{}
	if err := am.GetAlias(al, oldAlias); err != nil {
		*reply = err.Error()
		return err
	}
	for _, oldValue := range oldAlias.Values {
		found := false
		for _, value := range al.Values {
			if oldValue.Equals(value) {
				found = true
				break
			}
		}
		if !found {
			al.Values = append(al.Values, oldValue)
		}
	}

	if err := am.accountingDb.SetAlias(&al); err != nil {
		*reply = err.Error()
		return err
	} //add to cache

	aliasesChanged := []string{utils.ALIASES_PREFIX + al.GetId()}
	if err := am.accountingDb.CacheAccountingPrefixValues(map[string][]string{utils.ALIASES_PREFIX: aliasesChanged}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (am *AliasHandler) RemoveAlias(al Alias, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	if err := am.accountingDb.RemoveAlias(al.GetId()); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

func (am *AliasHandler) RemoveReverseAlias(attr AttrReverseAlias, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	rKey := utils.REVERSE_ALIASES_PREFIX + attr.Alias + attr.Target + attr.Context
	if x, err := cache2go.Get(rKey); err == nil {
		existingKeys := x.(map[string]bool)
		for key := range existingKeys {
			// get destination id
			elems := strings.Split(key, utils.CONCATENATED_KEY_SEP)
			if len(elems) > 0 {
				key = strings.Join(elems[:len(elems)-1], utils.CONCATENATED_KEY_SEP)
			}
			if err := am.accountingDb.RemoveAlias(key); err != nil {
				*reply = err.Error()
				return err
			}
		}
	}
	*reply = utils.OK
	return nil
}

func (am *AliasHandler) GetAlias(al Alias, result *Alias) error {
	am.mu.RLock()
	defer am.mu.RUnlock()
	variants := al.GenerateIds()
	for _, variant := range variants {
		if r, err := am.accountingDb.GetAlias(variant, false); err == nil {
			*result = *r
			return nil
		}
	}
	return utils.ErrNotFound
}

func (am *AliasHandler) GetReverseAlias(attr AttrReverseAlias, result *map[string][]*Alias) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	aliases := make(map[string][]*Alias)
	rKey := utils.REVERSE_ALIASES_PREFIX + attr.Alias + attr.Target + attr.Context
	if x, err := cache2go.Get(rKey); err == nil {
		existingKeys := x.(map[string]bool)
		for key := range existingKeys {

			// get destination id
			elems := strings.Split(key, utils.CONCATENATED_KEY_SEP)
			var destID string
			if len(elems) > 0 {
				destID = elems[len(elems)-1]
				key = strings.Join(elems[:len(elems)-1], utils.CONCATENATED_KEY_SEP)
			}
			if r, err := am.accountingDb.GetAlias(key, false); err != nil {
				return err
			} else {
				aliases[destID] = append(aliases[destID], r)
			}
		}
	}
	*result = aliases
	return nil
}

func (am *AliasHandler) GetMatchingAlias(attr AttrMatchingAlias, result *string) error {
	response := Alias{}
	if err := aliasService.GetAlias(Alias{
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
		if x, err := cache2go.Get(utils.DESTINATION_PREFIX + p); err == nil {
			destIds := x.(map[interface{}]struct{})
			for _, value := range values {
				for idId := range destIds {
					dId := idId.(string)
					if value.DestinationId == utils.ANY || value.DestinationId == dId {
						if origAliasMap, ok := value.Pairs[attr.Target]; ok {
							if alias, ok := origAliasMap[attr.Original]; ok {
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

type ProxyAliasService struct {
	Client *rpcclient.RpcClient
}

func NewProxyAliasService(addr string, attempts, reconnects int) (*ProxyAliasService, error) {
	client, err := rpcclient.NewRpcClient("tcp", addr, attempts, reconnects, utils.GOB)
	if err != nil {
		return nil, err
	}
	return &ProxyAliasService{Client: client}, nil
}

func (ps *ProxyAliasService) SetAlias(al Alias, reply *string) error {
	return ps.Client.Call("AliasesV1.SetAlias", al, reply)
}

func (ps *ProxyAliasService) UpdateAlias(al Alias, reply *string) error {
	return ps.Client.Call("AliasesV1.UpdateAlias", al, reply)
}

func (ps *ProxyAliasService) RemoveAlias(al Alias, reply *string) error {
	return ps.Client.Call("AliasesV1.RemoveAlias", al, reply)
}

func (ps *ProxyAliasService) GetAlias(al Alias, alias *Alias) error {
	return ps.Client.Call("AliasesV1.GetAlias", al, alias)
}

func (ps *ProxyAliasService) GetMatchingAlias(attr AttrMatchingAlias, alias *string) error {
	return ps.Client.Call("AliasesV1.GetMatchingAlias", attr, alias)
}

func (ps *ProxyAliasService) GetReverseAlias(attr AttrReverseAlias, alias *map[string][]*Alias) error {
	return ps.Client.Call("AliasesV1.GetReverseAlias", attr, alias)
}

func (ps *ProxyAliasService) RemoveReverseAlias(attr AttrReverseAlias, reply *string) error {
	return ps.Client.Call("AliasesV1.RemoveReverseAlias", attr, reply)
}

func (ps *ProxyAliasService) ReloadAliases(in string, reply *string) error {
	return ps.Client.Call("AliasesV1.ReloadAliases", in, reply)
}
