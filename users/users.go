package users

import (
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type UserData struct {
	Tenant   string
	UserName string
	Data     map[string]string
}

func (ud *UserData) GetId() string {
	return ud.Tenant + utils.CONCATENATED_KEY_SEP + ud.UserName
}

func (ud *UserData) SetId(id string) error {
	vals := strings.Split(id, utils.CONCATENATED_KEY_SEP)
	if len(vals) != 2 {
		return utils.ErrInvalidKey
	}
	ud.Tenant = vals[0]
	ud.UserName = vals[1]
	return nil
}

type UserService interface {
	SetUser(UserData, *string) error
	RemoveUser(UserData, *string) error
	UpdateUser(UserData, *string) error
	GetUsers(UserData, *[]UserData) error
}

type UserMap map[string]map[string]string

func NewUserMap() UserMap {
	return make(UserMap, 0)
}

func (ud *UserData) SetUser(UserData, *string) error {

	return nil
}
func (ud *UserData) RemoveUser(UserData, *string) error   { return nil }
func (ud *UserData) UpdateUser(UserData, *string) error   { return nil }
func (ud *UserData) GetUsers(UserData, *[]UserData) error { return nil }

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

func (ps *ProxyUserService) SetUser(ud UserData, reply *string) error {
	return ps.Client.Call("UserService.SetUser", ud, reply)
}

func (ps *ProxyUserService) RemoveUser(ud UserData, reply *string) error {
	return ps.Client.Call("UserService.RemoveUser", ud, reply)
}

func (ps *ProxyUserService) GetUsers(ud UserData, users *[]UserData) error {
	return ps.Client.Call("UserService.GetUsers", ud, users)
}
