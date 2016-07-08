package engine

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/cgrates/cgrates/utils"
	"github.com/syndtr/goleveldb/leveldb"
)

type cacheDumper struct {
	path        string
	dbMap       map[string]*leveldb.DB
	dbLocker    sync.Mutex
	dataEncoder Marshaler
}

func newCacheDumper(path string) (*cacheDumper, error) {
	if path != "" {
		if err := os.MkdirAll(path, 0766); err != nil {
			return nil, err
		}
	}
	return &cacheDumper{
		path:        path,
		dbMap:       make(map[string]*leveldb.DB),
		dataEncoder: NewCodecMsgpackMarshaler(),
	}, nil
}

func (cd *cacheDumper) getDumpDb(prefix string) (*leveldb.DB, error) {
	if cd == nil || cd.path == "" {
		return nil, nil
	}
	cd.dbLocker.Lock()
	defer cd.dbLocker.Unlock()
	db, found := cd.dbMap[prefix]
	if !found {
		var err error
		db, err = leveldb.OpenFile(filepath.Join(cd.path, prefix+".cache"), nil)
		if err != nil {
			return nil, err
		}
		cd.dbMap[prefix] = db
	}
	return db, nil
}

func (cd *cacheDumper) put(prefix, key string, value interface{}) error {
	db, err := cd.getDumpDb(prefix)
	if err != nil || db == nil {
		return err
	}

	encData, err := cd.dataEncoder.Marshal(value)
	if err != nil {
		return err
	}
	if len(encData) > 1000 {
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		w.Write(encData)
		w.Close()
		encData = buf.Bytes()
	}
	return db.Put([]byte(key), encData, nil)
}

func (cd *cacheDumper) load(prefix string) (map[string]interface{}, error) {
	db, err := cd.getDumpDb(prefix)
	if err != nil || db == nil {
		return nil, err
	}
	val := make(map[string]interface{})
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		k := iter.Key()
		data := iter.Value()
		var encData []byte
		if data[0] == 120 && data[1] == 156 { //zip header
			x := bytes.NewBuffer(data)
			r, err := zlib.NewReader(x)
			if err != nil {
				utils.Logger.Info("<cache decoder>: " + err.Error())
				break
			}
			out, err := ioutil.ReadAll(r)
			if err != nil {
				utils.Logger.Info("<cache decoder>: " + err.Error())
				break
			}
			r.Close()
			encData = out
		} else {
			encData = data
		}
		v := cd.cacheTypeFactory(prefix)
		if err := cd.dataEncoder.Unmarshal(encData, &v); err != nil {
			return nil, err
		}
		val[string(k)] = v
	}
	iter.Release()
	return val, nil
}

func (cd *cacheDumper) delete(prefix, key string) error {
	db, err := cd.getDumpDb(prefix)
	if err != nil || db == nil {
		return err
	}
	return db.Delete([]byte(key), nil)
}

func (cd *cacheDumper) deleteAll(prefix string) error {
	db, err := cd.getDumpDb(prefix)
	if err != nil || db == nil {
		return err
	}
	db.Close()
	delete(cd.dbMap, prefix)
	return os.RemoveAll(filepath.Join(cd.path, prefix+".cache"))
}

func (cd *cacheDumper) cacheTypeFactory(prefix string) interface{} {
	switch prefix {
	case utils.DESTINATION_PREFIX:
		return make(map[string]struct{})
	case utils.RATING_PLAN_PREFIX:
		return &RatingPlan{}
	case utils.RATING_PROFILE_PREFIX:
		return &RatingProfile{}
	case utils.LCR_PREFIX:
		return &LCR{}
	case utils.DERIVEDCHARGERS_PREFIX:
		return &utils.DerivedChargers{}
	case utils.ACTION_PREFIX:
		return Actions{}
	case utils.ACTION_PLAN_PREFIX:
		return &ActionPlan{}
	case utils.SHARED_GROUP_PREFIX:
		return &SharedGroup{}
	case utils.ALIASES_PREFIX:
		return AliasValues{}
	case utils.LOADINST_KEY[:PREFIX_LEN]:
		return make([]*utils.LoadInstance, 0)
	}
	return nil
}
