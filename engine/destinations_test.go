package engine

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/nyaruka/phonenumbers"
)

func TestDestinationStoreRestore(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	s, _ := json.Marshal(nationale)
	d1 := &Destination{Id: "nat"}
	json.Unmarshal(s, d1)
	s1, _ := json.Marshal(d1)
	if string(s1) != string(s) {
		t.Errorf("Expected %q was %q", s, s1)
	}
}

func TestDestinationStorageStore(t *testing.T) {
	nationale := &Destination{Id: "nat",
		Prefixes: []string{"0257", "0256", "0723"}}
	err := dm.SetDestination(nationale, utils.NonTransactional)
	if err != nil {
		t.Error("Error storing destination: ", err)
	}
	result, err := dm.GetDestination(nationale.Id,
		true, true, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}
	if nationale.containsPrefix("0257") == 0 ||
		nationale.containsPrefix("0256") == 0 ||
		nationale.containsPrefix("0723") == 0 {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

func TestDestinationContainsPrefix(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision := nationale.containsPrefix("0256")
	if precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
	}
}

func TestDestinationContainsPrefixLong(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision := nationale.containsPrefix("0256723045")
	if precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
	}
}

func TestDestinationContainsPrefixWrong(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision := nationale.containsPrefix("01234567")
	if precision != 0 {
		t.Error("Should not contain prefix: ", nationale)
	}
}

func TestDestinationGetExists(t *testing.T) {
	d, err := dm.GetDestination("NAT", true, true, utils.NonTransactional)
	if err != nil || d == nil {
		t.Error("Could not get destination: ", d)
	}
}

func TestDestinationReverseGetExistsCache(t *testing.T) {
	dm.GetReverseDestination("0256", true, true, utils.NonTransactional)
	if _, ok := Cache.Get(utils.CacheReverseDestinations, "0256"); !ok {
		t.Error("Destination not cached:", err)
	}
}

func TestDestinationGetNotExists(t *testing.T) {
	if d, ok := Cache.Get(utils.CacheDestinations, "not existing"); ok {
		t.Error("Bad destination cached: ", d)
	}
	d, err := dm.GetDestination("not existing", true, true, utils.NonTransactional)
	if d != nil {
		t.Error("Got false destination: ", d, err)
	}
}

func TestDestinationCachedDestHasPrefix(t *testing.T) {
	if !CachedDestHasPrefix("NAT", "0256") {
		t.Error("Could not find prefix in destination")
	}
}

func TestDestinationCachedDestHasWrongPrefix(t *testing.T) {
	if CachedDestHasPrefix("NAT", "771") {
		t.Error("Prefix should not belong to destination")
	}
}

func TestDestinationNonCachedDestRightPrefix(t *testing.T) {
	if CachedDestHasPrefix("FAKE", "0256") {
		t.Error("Destination should not belong to prefix")
	}
}

func TestDestinationNonCachedDestWrongPrefix(t *testing.T) {
	if CachedDestHasPrefix("FAKE", "771") {
		t.Error("Both arguments should be fake")
	}
}

func TestDestinationcontainsPrefixNilDestination(t *testing.T) {
	var d *Destination
	rcv := d.containsPrefix("prefix")
	if rcv != 0 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", 0, rcv)
	}
}

func TestDestinationString(t *testing.T) {
	d := &Destination{
		Id:       "ID",
		Prefixes: []string{"prefix1", "prefix2", "prefix3"},
	}

	exp := "ID: prefix1, prefix2, prefix3"
	rcv := d.String()

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

/********************************* Benchmarks **********************************/

func BenchmarkDestinationStorageStoreRestore(b *testing.B) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	for i := 0; i < b.N; i++ {
		dm.SetDestination(nationale, utils.NonTransactional)
		dm.GetDestination(nationale.Id, false, true, utils.NonTransactional)
	}
}

func TestDynamicDPFieldAsInterface(t *testing.T) {

	dDP := newDynamicDP(nil, nil, nil, nil, nil, "cgrates.org", &Account{})

	if _, err := dDP.fieldAsInterface([]string{"field"}); err == nil {
		t.Error(err)
	}
	if _, err := dDP.fieldAsInterface([]string{utils.MetaAccounts, "field1", "field2"}); err == nil {
		t.Error(err)
	}
}

func TestDPNewLibNumber(t *testing.T) {
	num, err := phonenumbers.ParseAndKeepRawInput("+3554735474", utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	exp := &libphonenumberDP{
		pNumber: num,
		cache:   utils.MapStorage{},
	}
	if val, err := newLibPhoneNumberDP("+3554735474"); err != nil {
		t.Errorf("received <%v>", err)
	} else if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %v,received %v", exp, val)
	}
	expErr := "the phone number supplied is not a number"
	if _, err := newLibPhoneNumberDP("some"); err == nil || err.Error() != expErr {
		t.Errorf("expected %v ,received %v", expErr, err)
	}

}

func TestDMSetDestinationSucces(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.MetaDestinations: {
			Replicate: true,
		},
	}
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	cfg.DataDbCfg().RplFiltered = true
	cfg.DataDbCfg().RplCache = "cache"

	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetDestination: func(ctx *context.Context, args, reply any) error {
				*reply.(*string) = "reply"
				return nil
			},
		},
	}
	connMngr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg): clientConn,
	})

	dest := &Destination{
		Id:       "dest21",
		Prefixes: []string{},
	}
	config.SetCgrConfig(cfg)
	dm := NewDataManager(db, cfg.CacheCfg(), connMngr)

	if err := dm.SetDestination(dest, utils.NonTransactional); err != nil {
		t.Error(err)
	}

}

func TestDMSetAccountSucces(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	/*cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg)}
	cfg.DataDbCfg().RplFiltered = false
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.MetaAccounts: {
			Replicate: true,
		},
	}*/
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetAccount: func(ctx *context.Context, args, reply any) error {
				*reply.(*string) = "reply"
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg): clientConn,
	})
	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}

	acc := &Account{
		ID: "id",
		BalanceMap: map[string]Balances{
			"bal": {
				{
					Uuid:  "uuid",
					ID:    "id",
					Value: 21.3,
				},
			},
		},
		UnitCounters:   UnitCounters{},
		ActionTriggers: ActionTriggers{},
		AllowNegative:  true,
		Disabled:       false,
	}
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	config.SetCgrConfig(cfg)

	if err := dm.SetAccount(acc); err != nil {
		t.Error(err)
	}

}

func TestDMSetReverseDestination(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()

	db, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ReplicatorSv1SetReverseDestination: func(ctx *context.Context, args, reply any) error {
				*reply.(*string) = "reply"
				return nil
			},
		},
	}
	connMngr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicationConnsCfg): clientConn,
	})

	dm := NewDataManager(db, cfg.CacheCfg(), connMngr)
	config.SetCgrConfig(cfg)

	if err := dm.SetReverseDestination("val", []string{"prf"}, utils.NonTransactional); err != nil {
		t.Error(err)
	}

}

func TestDMSetTimingInvalidTime(t *testing.T) {
	dm := NewDataManager(nil, nil, nil)
	expErr := "INVALID_TIME:*any"
	if err := dm.SetTiming(&utils.TPTiming{StartTime: "*any"}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}

	if err := dm.SetTiming(&utils.TPTiming{EndTime: "*any"}); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}
}

func TestDestinationClone(t *testing.T) {
	t.Run("nil destination", func(t *testing.T) {
		var d *Destination
		clone := d.Clone()
		if clone != nil {
			t.Errorf("Expected nil clone for nil destination, got %v", clone)
		}
	})

	t.Run("empty destination", func(t *testing.T) {
		d := &Destination{}
		clone := d.Clone()

		if clone == nil {
			t.Fatal("Clone should not be nil")
		}

		if clone == d {
			t.Error("Clone should be a different instance")
		}

		if clone.Id != d.Id {
			t.Errorf("Expected empty Id, got %q", clone.Id)
		}

		if clone.Prefixes != nil {
			t.Errorf("Expected nil Prefixes, got %v", clone.Prefixes)
		}
	})

	t.Run("destination with data", func(t *testing.T) {
		d := &Destination{
			Id:       "ID1",
			Prefixes: []string{"prefix1", "prefix2", "prefix3"},
		}

		clone := d.Clone()

		if clone == nil {
			t.Fatal("Clone should not be nil")
		}

		if clone == d {
			t.Error("Clone should be a different instance")
		}

		if clone.Id != d.Id {
			t.Errorf("Expected Id %q, got %q", d.Id, clone.Id)
		}

		if !reflect.DeepEqual(clone.Prefixes, d.Prefixes) {
			t.Errorf("Expected Prefixes %v, got %v", d.Prefixes, clone.Prefixes)
		}

		d.Id = "changed-id"
		d.Prefixes[0] = "changed-prefix"

		if clone.Id == d.Id {
			t.Error("Changing original Id should not affect clone")
		}

		if clone.Prefixes[0] == d.Prefixes[0] {
			t.Error("Changing original Prefixes should not affect clone")
		}
	})

	t.Run("destination with empty prefixes", func(t *testing.T) {
		d := &Destination{
			Id:       "dest-456",
			Prefixes: []string{},
		}

		clone := d.Clone()

		if clone == nil {
			t.Fatal("Clone should not be nil")
		}

		if clone.Prefixes == nil {
			t.Error("Clone should have an empty slice, not nil")
		}

		if len(clone.Prefixes) != 0 {
			t.Errorf("Expected empty Prefixes slice, got %v", clone.Prefixes)
		}
	})
}

func TestDestinationCacheClone(t *testing.T) {
	tests := []struct {
		name  string
		input *Destination
	}{
		{
			name:  "nil destination",
			input: nil,
		},
		{
			name:  "empty destination",
			input: &Destination{},
		},
		{
			name: "destination with data",
			input: &Destination{
				Id:       "ID1",
				Prefixes: []string{"prefix1", "prefix2", "prefix3"},
			},
		},
		{
			name: "destination with empty prefixes",
			input: &Destination{
				Id:       "dest-456",
				Prefixes: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloneAny := tt.input.CacheClone()

			if tt.input == nil {
				if cloneAny != nil {
					clone, ok := cloneAny.(*Destination)
					if !ok {
						t.Fatalf("CacheClone returned %T, expected *Destination", cloneAny)
					}
					if clone != nil {
						t.Errorf("Expected nil clone for nil destination, got %v", clone)
					}
				}
				return
			}

			if cloneAny == nil {
				t.Fatal("CacheClone returned nil for non-nil destination")
			}

			clone, ok := cloneAny.(*Destination)
			if !ok {
				t.Fatalf("CacheClone returned %T, expected *Destination", cloneAny)
			}

			if clone == tt.input {
				t.Error("Clone should be a different instance")
			}

			if !reflect.DeepEqual(clone, tt.input) {
				t.Errorf("Clone doesn't match original value.\nGot: %+v\nExpected: %+v", clone, tt.input)
			}

			if tt.input.Id != "" {
				originalId := tt.input.Id
				tt.input.Id = "modified-id"

				if clone.Id != originalId {
					t.Error("Modifying original Id should not affect clone")
				}
			}

			if tt.input.Prefixes != nil && len(tt.input.Prefixes) > 0 {
				originalPrefix := tt.input.Prefixes[0]
				tt.input.Prefixes[0] = "modified-prefix"

				if clone.Prefixes[0] != originalPrefix {
					t.Error("Modifying original Prefixes should not affect clone")
				}
			}
		})
	}
}
