package users

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
}

type UserMap map[string]map[string]string

func (um UserMap) SetUser(up UserProfile, reply *string) error {
	um[up.GetId()] = up.Profile
	*reply = utils.OK
	return nil
}
func (um UserMap) RemoveUser(up UserProfile, reply *string) error {
	delete(um, up.GetId())
	*reply = utils.OK
	return nil
}
func (um UserMap) UpdateUser(up UserProfile, reply *string) error {
	m, found := um[up.GetId()]
	if !found {
		*reply = utils.ErrNotFound.Error()
		return utils.ErrNotFound
	}
	if m == nil {
		um[up.GetId()] = make(map[string]string, 0)
	}
	for key, value := range up.Profile {
		um[up.GetId()][key] = value
	}
	*reply = utils.OK
	return nil
}
func (um UserMap) GetUsers(up UserProfile, results *[]*UserProfile) error {
	// no index
	var candidates []*UserProfile
	for key, values := range um {
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
		*results = candidates
	}
	return nil
}

type UserProxy struct{}

type ProxyUserService struct {
	Client *rpcclient.RpcClient
}

func NewProxyUserService(addr string, reconnects int) (*ProxyUserService, error) {
	client, err := rpcclient.NewRpcClient("tcp", addr, reconnects, utils.GOB)
	if err != nil {
		return nil, err
	}
	return &ProxyUserService{Client: client}, nil
}

func (ps *ProxyUserService) SetUser(ud UserProfile, reply *string) error {
	return ps.Client.Call("UserService.SetUser", ud, reply)
}

func (ps *ProxyUserService) RemoveUser(ud UserProfile, reply *string) error {
	return ps.Client.Call("UserService.RemoveUser", ud, reply)
}

func (ps *ProxyUserService) GetUsers(ud UserProfile, users *[]*UserProfile) error {
	return ps.Client.Call("UserService.GetUsers", ud, users)
}
