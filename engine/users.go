package engine

import (
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type UserProfile struct {
	Tenant   string
	UserName string
	Profile  map[string]string
}

func (ud *UserProfile) GetId() string {
	return utils.ConcatenatedKey(ud.Tenant, ud.UserName)
}

func (ud *UserProfile) SetId(id string) error {
	vals := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(vals) != 2 {
		return utils.ErrInvalidKey
	}
	ud.Tenant = vals[0]
	ud.UserName = vals[1]
	return nil
}

type UserService interface {
	SetUser(UserProfile, *string) error
	RemoveUser(UserProfile, *string) error
	UpdateUser(UserProfile, *string) error
	GetUsers(UserProfile, *[]*UserProfile) error
	AddIndex([]string, *string) error
	GetIndexes(string, *[]string) error
}

type UserMap struct {
	table    map[string]map[string]string
	index    map[string][]string
	ratingDb RatingStorage
}

func NewUserMap(ratingDb RatingStorage) (*UserMap, error) {
	um := newUserMap(ratingDb)
	// load from rating db
	if ups, err := um.ratingDb.GetUsers(); err == nil {
		for _, up := range ups {
			um.table[up.GetId()] = up.Profile
		}
	} else {
		return nil, err
	}
	return um, nil
}

func newUserMap(ratingDb RatingStorage) *UserMap {
	return &UserMap{
		table:    make(map[string]map[string]string),
		index:    make(map[string][]string),
		ratingDb: ratingDb,
	}
}

func (um *UserMap) SetUser(up UserProfile, reply *string) error {
	um.table[up.GetId()] = up.Profile
	if err := um.ratingDb.SetUser(&up); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

func (um *UserMap) RemoveUser(up UserProfile, reply *string) error {
	delete(um.table, up.GetId())
	if err := um.ratingDb.RemoveUser(up.GetId()); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

func (um *UserMap) UpdateUser(up UserProfile, reply *string) error {
	m, found := um.table[up.GetId()]
	if !found {
		*reply = utils.ErrNotFound.Error()
		return utils.ErrNotFound
	}
	if m == nil {
		um.table[up.GetId()] = make(map[string]string, 0)
	}
	for key, value := range up.Profile {
		um.table[up.GetId()][key] = value
	}
	finalUp := &UserProfile{
		Tenant:   up.Tenant,
		UserName: up.UserName,
		Profile:  um.table[up.GetId()],
	}
	if err := um.ratingDb.SetUser(finalUp); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

func (um *UserMap) GetUsers(up UserProfile, results *[]*UserProfile) error {
	table := um.table // no index

	indexUnionKeys := make(map[string]bool)
	// search index
	if up.Tenant != "" {
		if keys, found := um.index[utils.ConcatenatedKey("Tenant", up.Tenant)]; found {
			for _, key := range keys {
				indexUnionKeys[key] = true
			}
		}
	}
	if up.UserName != "" {
		if keys, found := um.index[utils.ConcatenatedKey("UserName", up.UserName)]; found {
			for _, key := range keys {
				indexUnionKeys[key] = true
			}
		}
	}
	for k, v := range up.Profile {
		if keys, found := um.index[utils.ConcatenatedKey(k, v)]; found {
			for _, key := range keys {
				indexUnionKeys[key] = true
			}
		}
	}
	if len(indexUnionKeys) != 0 {
		table = make(map[string]map[string]string)
		for key := range indexUnionKeys {
			table[key] = um.table[key]
		}
	}

	var candidates []*UserProfile
	for key, values := range table {
		if up.Tenant != "" && !strings.HasPrefix(key, up.Tenant+utils.CONCATENATED_KEY_SEP) {
			continue
		}
		if up.UserName != "" && !strings.HasSuffix(key, utils.CONCATENATED_KEY_SEP+up.UserName) {
			continue
		}
		valid := true
		for k, v := range up.Profile {
			if values[k] != v {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		// all filters passed, add to candidates
		nup := &UserProfile{Profile: make(map[string]string)}
		nup.SetId(key)
		for k, v := range values {
			nup.Profile[k] = v
		}
		candidates = append(candidates, nup)
	}
	*results = candidates
	return nil
}

func (um *UserMap) AddIndex(indexes []string, reply *string) error {
	for key, values := range um.table {
		ud := &UserProfile{Profile: values}
		ud.SetId(key)
		for _, index := range indexes {
			if index == "Tenant" {
				if ud.Tenant != "" {
					um.index[utils.ConcatenatedKey(index, ud.Tenant)] = append(um.index[utils.ConcatenatedKey(index, ud.Tenant)], key)
				}
				continue
			}
			if index == "UserName" {
				if ud.UserName != "" {
					um.index[utils.ConcatenatedKey(index, ud.UserName)] = append(um.index[utils.ConcatenatedKey(index, ud.UserName)], key)
				}
				continue
			}

			for k, v := range ud.Profile {
				if k == index && v != "" {
					um.index[utils.ConcatenatedKey(k, v)] = append(um.index[utils.ConcatenatedKey(k, v)], key)
				}
			}
		}
	}
	*reply = utils.OK
	return nil
}

func (um *UserMap) GetIndexes(in string, reply *[]string) error {
	var indexes []string
	for key := range um.index {
		indexes = append(indexes, key)
	}
	*reply = indexes
	return nil
}

type UserProxy struct{}

type ProxyUserService struct {
	Client *rpcclient.RpcClient
}

func NewProxyUserService(addr string, attempts, reconnects int) (*ProxyUserService, error) {
	client, err := rpcclient.NewRpcClient("tcp", addr, attempts, reconnects, utils.GOB)
	if err != nil {
		return nil, err
	}
	return &ProxyUserService{Client: client}, nil
}

func (ps *ProxyUserService) SetUser(ud UserProfile, reply *string) error {
	return ps.Client.Call("UsersV1.SetUser", ud, reply)
}

func (ps *ProxyUserService) RemoveUser(ud UserProfile, reply *string) error {
	return ps.Client.Call("UsersV1.RemoveUser", ud, reply)
}

func (ps *ProxyUserService) UpdateUser(ud UserProfile, reply *string) error {
	return ps.Client.Call("UsersV1.UpdateUser", ud, reply)
}

func (ps *ProxyUserService) GetUsers(ud UserProfile, users *[]*UserProfile) error {
	return ps.Client.Call("UsersV1.GetUsers", ud, users)
}

func (ps *ProxyUserService) AddIndex(indexes []string, reply *string) error {
	return ps.Client.Call("UsersV1.AddIndex", indexes, reply)
}

func (ps *ProxyUserService) GetIndexes(in string, reply *[]string) error {
	return ps.Client.Call("UsersV1.AddIndex", in, reply)
}
