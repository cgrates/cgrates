package engine

import (
	"sort"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type Alias struct {
	Direction string
	Tenant    string
	Category  string
	Account   string
	Subject   string
	Group     string
	Values    AliasValues
}

type AliasValue struct {
	DestinationId string
	Alias         string
	Weight        float64
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

// returns a mapping between aliases and destination ids in a slice sorted by weights
func (avs AliasValues) GetWeightSlice() (result []map[string][]string) {
	avs.Sort()
	prevWeight := -1.0
	for _, value := range avs {
		var m map[string][]string
		if value.Weight != prevWeight {
			// start a new map
			m = make(map[string][]string)
			result = append(result, m)
		} else {
			m = result[len(result)-1]
		}
		m[value.Alias] = append(m[value.Alias], value.DestinationId)
	}
	return
}

func (al *Alias) GetId() string {
	return utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Group)
}

func (al *Alias) GenerateIds() []string {
	var result []string
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Group))
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, utils.ANY, al.Group))
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, utils.ANY, utils.ANY, al.Group))
	result = append(result, utils.ConcatenatedKey(al.Direction, al.Tenant, utils.ANY, utils.ANY, utils.ANY, al.Group))
	result = append(result, utils.ConcatenatedKey(al.Direction, utils.ANY, utils.ANY, utils.ANY, utils.ANY, al.Group))
	result = append(result, utils.ConcatenatedKey(utils.ANY, utils.ANY, utils.ANY, utils.ANY, al.Group))
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
	al.Group = vals[5]
	return nil
}

type AliasService interface {
	SetAlias(Alias, *string) error
	RemoveAlias(Alias, *string) error
	GetAlias(Alias, *Alias) error
	ReloadAliases(string, *string) error
}

type AliasMap struct {
	table        map[string]AliasValues
	accountingDb AccountingStorage
	mu           sync.RWMutex
}

func NewAliasMap(accountingDb AccountingStorage) (*AliasMap, error) {
	um := newAliasMap(accountingDb)
	var reply string
	if err := um.ReloadAliases("", &reply); err != nil {
		return nil, err
	}
	return um, nil
}

func newAliasMap(accountingDb AccountingStorage) *AliasMap {
	return &AliasMap{
		table:        make(map[string]AliasValues),
		accountingDb: accountingDb,
	}
}

func (am *AliasMap) ReloadAliases(in string, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	// backup old data
	oldTable := am.table
	am.table = make(map[string]AliasValues)

	// load from db
	if ups, err := am.accountingDb.GetAliases(); err == nil {
		for _, up := range ups {
			am.table[up.GetId()] = up.Values
		}
	} else {
		// restore old data before return
		am.table = oldTable

		*reply = err.Error()
		return err
	}

	*reply = utils.OK
	return nil
}

func (am *AliasMap) SetAlias(al Alias, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	if err := am.accountingDb.SetAlias(&al); err != nil {
		*reply = err.Error()
		return err
	}
	am.table[al.GetId()] = al.Values
	*reply = utils.OK
	return nil
}

func (am *AliasMap) RemoveAlias(al Alias, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	if err := am.accountingDb.RemoveAlias(al.GetId()); err != nil {
		*reply = err.Error()
		return err
	}
	delete(am.table, al.GetId())
	*reply = utils.OK
	return nil
}

func (am *AliasMap) GetAlias(al Alias, result *Alias) error {
	am.mu.RLock()
	defer am.mu.RUnlock()
	variants := al.GenerateIds()
	//log.Print("TABLE: ", am.table)
	for _, variant := range variants {
		//log.Printf("AL %+v", variant)
		if r, ok := am.table[variant]; ok {
			al.Values = r
			*result = al
			return nil
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
	return ps.Client.Call("AliasV1.SetAlias", al, reply)
}

func (ps *ProxyAliasService) RemoveAlias(al Alias, reply *string) error {
	return ps.Client.Call("AliasV1.RemoveAlias", al, reply)
}

func (ps *ProxyAliasService) GetAlias(al Alias, alias *Alias) error {
	return ps.Client.Call("AliasV1.GetAliases", al, alias)
}

func (ps *ProxyAliasService) ReloadAliases(in string, reply *string) error {
	return ps.Client.Call("AliasV1.ReloadAliases", in, reply)
}

func GetBestAlias(destination, direction, tenant, category, account, subject, group string) (string, error) {
	if aliasService == nil {
		return "", nil
	}
	response := Alias{}
	if err := aliasService.GetAlias(Alias{
		Direction: direction,
		Tenant:    tenant,
		Category:  category,
		Account:   account,
		Subject:   subject,
		Group:     group,
	}, &response); err != nil {
		return "", err
	}
	// sort according to weight
	values := response.Values.GetWeightSlice()
	// check destination ids

	for _, p := range utils.SplitPrefix(destination, MIN_PREFIX_MATCH) {
		if x, err := cache2go.GetCached(utils.DESTINATION_PREFIX + p); err == nil {
			for _, aliasMap := range values {
				for alias, aliasDestIds := range aliasMap {
					destIds := x.(map[interface{}]struct{})
					for idId := range destIds {
						dId := idId.(string)
						for _, aliasDestId := range aliasDestIds {
							if aliasDestId == utils.ANY || aliasDestId == dId {
								return alias, nil
							}
						}
					}
				}
			}
		}
	}
	return "", utils.ErrNotFound
}
