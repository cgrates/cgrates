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
package migrator

import (
	"fmt"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

type v1Redis struct {
	dbPool   *pool.Pool
	ms       engine.Marshaler
	dataKeys []string
	qryIdx   *int
}

func newv1RedisStorage(address string, db int, pass, mrshlerStr string) (*v1Redis, error) {
	df := func(network, addr string) (*redis.Client, error) {
		client, err := redis.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		if len(pass) != 0 {
			if err = client.Cmd("AUTH", pass).Err; err != nil {
				client.Close()
				return nil, err
			}
		}
		if db != 0 {
			if err = client.Cmd("SELECT", db).Err; err != nil {
				client.Close()
				return nil, err
			}
		}
		return client, nil
	}
	p, err := pool.NewCustom("tcp", address, 1, df)
	if err != nil {
		return nil, err
	}
	var mrshler engine.Marshaler
	if mrshlerStr == utils.MSGPACK {
		mrshler = engine.NewCodecMsgpackMarshaler()
	} else if mrshlerStr == utils.JSON {
		mrshler = new(engine.JSONMarshaler)
	} else {
		return nil, fmt.Errorf("Unsupported marshaler: %v", mrshlerStr)
	}
	return &v1Redis{dbPool: p, ms: mrshler}, nil
}

// This CMD function get a connection from the pool.
// Handles automatic failover in case of network disconnects
func (v1rs *v1Redis) cmd(cmd string, args ...interface{}) *redis.Resp {
	c1, err := v1rs.dbPool.Get()
	if err != nil {
		return redis.NewResp(err)
	}
	result := c1.Cmd(cmd, args...)
	if result.IsType(redis.IOErr) { // Failover mecanism
		utils.Logger.Warning(fmt.Sprintf("<RedisStorage> error <%s>, attempting failover.", result.Err.Error()))
		c2, err := v1rs.dbPool.Get()
		if err == nil {
			if result2 := c2.Cmd(cmd, args...); !result2.IsType(redis.IOErr) {
				v1rs.dbPool.Put(c2)
				return result2
			}
		}
	} else {
		v1rs.dbPool.Put(c1)
	}
	return result
}

func (v1rs *v1Redis) getKeysForPrefix(prefix string) ([]string, error) {
	r := v1rs.cmd("KEYS", prefix+"*")
	if r.Err != nil {
		return nil, r.Err
	}
	return r.List()
}

//Account methods
//get
func (v1rs *v1Redis) getv1Account() (v1Acnt *v1Account, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.getKeysForPrefix(v1AccountDBPrefix)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		v1Acnt = &v1Account{Id: v1rs.dataKeys[*v1rs.qryIdx]}
		if err := v1rs.ms.Unmarshal(strVal, v1Acnt); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1Acnt, nil
}

//set
func (v1rs *v1Redis) setV1Account(x *v1Account) (err error) {
	key := v1AccountDBPrefix + x.Id
	bit, err := v1rs.ms.Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//ActionPlans methods
//get
func (v1rs *v1Redis) getV1ActionPlans() (v1aps *v1ActionPlans, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.getKeysForPrefix(utils.ACTION_PLAN_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.ms.Unmarshal(strVal, &v1aps); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1aps, nil
}

//set
func (v1rs *v1Redis) setV1ActionPlans(x *v1ActionPlans) (err error) {
	key := utils.ACTION_PLAN_PREFIX + (*x)[0].Id
	bit, err := v1rs.ms.Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//Actions methods
//get
func (v1rs *v1Redis) getV1Actions() (v1acs *v1Actions, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.getKeysForPrefix(utils.ACTION_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.ms.Unmarshal(strVal, &v1acs); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1acs, nil
}

//set
func (v1rs *v1Redis) setV1Actions(x *v1Actions) (err error) {
	key := utils.ACTION_PREFIX + (*x)[0].Id
	bit, err := v1rs.ms.Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//ActionTriggers methods
//get
func (v1rs *v1Redis) getV1ActionTriggers() (v1acts *v1ActionTriggers, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.getKeysForPrefix(utils.ACTION_TRIGGER_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.ms.Unmarshal(strVal, &v1acts); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1acts, nil
}

//set
func (v1rs *v1Redis) setV1ActionTriggers(x *v1ActionTriggers) (err error) {
	key := utils.ACTION_TRIGGER_PREFIX + (*x)[0].Id
	bit, err := v1rs.ms.Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}

//SharedGroup methods
//get
func (v1rs *v1Redis) getV1SharedGroup() (v1sg *v1SharedGroup, err error) {
	if v1rs.qryIdx == nil {
		v1rs.dataKeys, err = v1rs.getKeysForPrefix(utils.SHARED_GROUP_PREFIX)
		if err != nil {
			return
		} else if len(v1rs.dataKeys) == 0 {
			return nil, utils.ErrNotFound
		}
		v1rs.qryIdx = utils.IntPointer(0)
	}
	if *v1rs.qryIdx <= len(v1rs.dataKeys)-1 {
		strVal, err := v1rs.cmd("GET", v1rs.dataKeys[*v1rs.qryIdx]).Bytes()
		if err != nil {
			return nil, err
		}
		if err := v1rs.ms.Unmarshal(strVal, &v1sg); err != nil {
			return nil, err
		}
		*v1rs.qryIdx = *v1rs.qryIdx + 1
	} else {
		v1rs.qryIdx = nil
		return nil, utils.ErrNoMoreData
	}
	return v1sg, nil
}

//set
func (v1rs *v1Redis) setV1SharedGroup(x *v1SharedGroup) (err error) {
	key := utils.SHARED_GROUP_PREFIX + x.Id
	bit, err := v1rs.ms.Marshal(x)
	if err != nil {
		return err
	}
	if err = v1rs.cmd("SET", key, bit).Err; err != nil {
		return err
	}
	return
}
