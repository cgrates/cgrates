package engine

import (
	"sort"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type Alias struct {
	Direction     string
	Tenant        string
	Category      string
	Account       string
	Subject       string
	DestinationId string
	Group         string
	Alias         string
	Weight        float64
}

type Aliases []*Alias

func (ups Aliases) Len() int {
	return len(ups)
}

func (ups Aliases) Swap(i, j int) {
	ups[i], ups[j] = ups[j], ups[i]
}

func (ups Aliases) Less(j, i int) bool { // get higher ponder in front
	return ups[i].ponder < ups[j].ponder
}

func (ups Aliases) Sort() {
	sort.Sort(ups)
}

func (al *Alias) GetId() string {
	return utils.ConcatenatedKey(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.DestinationId, al.Group)
}

func (al *Alias) SetId(id string) error {
	vals := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(vals) != 7 {
		return utils.ErrInvalidKey
	}
	al.Direction = vals[0]
	al.Tenant = vals[1]
	al.Category = vals[2]
	al.Account = vals[3]
	al.Subject = vals[4]
	al.DestinationId = vals[5]
	al.Group = vals[6]
	return nil
}

type AliasService interface {
	SetAlias(Alias, *string) error
	AliasAlias(Alias, *string) error
	GetAliases(Alias, *Aliases) error
	ReloadAliases(string, *string) error
}

type AliasMap struct {
	table        map[string]string
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
		table:        make(map[string]string),
		accountingDb: accountingDb,
	}
}

func (am *AliasMap) ReloadAliases(in string, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// backup old data
	oldTable := am.table
	am.table = make(map[string]string)

	// load from rating db
	if ups, err := am.accountingDb.GetAliases(); err == nil {
		for _, up := range ups {
			am.table[up.GetId()] = up.Profile
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
	am.table[al.GetId()] = al.Alias
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
	am.deleteIndex(&al)
	*reply = utils.OK
	return nil
}

func (am *AliasMap) GetAliases(al Alias, results *Aliases) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	*results = am.table[al.GetId()]
	return nil
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

func (ps *ProxyAliasService) GetAliases(al Alias, aliases *Aliases) error {
	return ps.Client.Call("AliasV1.GetAliases", al, aliases)
}

func (ps *ProxyAliasService) ReloadAliases(in string, reply *string) error {
	return ps.Client.Call("AliasV1.ReloadAliases", in, reply)
}
