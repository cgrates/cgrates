package engine

import "github.com/cgrates/cgrates/utils"

type cacheKeyValue interface {
	Key() string
	Value() interface{}
}

type mapKeyValue struct {
	K string
	V map[string]struct{}
}

func (mkv *mapKeyValue) Key() string {
	return mkv.K
}

func (mkv *mapKeyValue) Value() interface{} {
	return mkv.V
}

type rpKeyValue struct {
	K string
	V *RatingPlan
}

func (mkv *rpKeyValue) Key() string {
	return mkv.K
}

func (mkv *rpKeyValue) Value() interface{} {
	return mkv.V
}

type rpfKeyValue struct {
	K string
	V *RatingProfile
}

func (mkv *rpfKeyValue) Key() string {
	return mkv.K
}

func (mkv *rpfKeyValue) Value() interface{} {
	return mkv.V
}

type lcrKeyValue struct {
	K string
	V *LCR
}

func (mkv *lcrKeyValue) Key() string {
	return mkv.K
}

func (mkv *lcrKeyValue) Value() interface{} {
	return mkv.V
}

type dcKeyValue struct {
	K string
	V *utils.DerivedChargers
}

func (mkv *dcKeyValue) Key() string {
	return mkv.K
}

func (mkv *dcKeyValue) Value() interface{} {
	return mkv.V
}

type acsKeyValue struct {
	K string
	V Actions
}

func (mkv *acsKeyValue) Key() string {
	return mkv.K
}

func (mkv *acsKeyValue) Value() interface{} {
	return mkv.V
}

type aplKeyValue struct {
	K string
	V *ActionPlan
}

func (mkv *aplKeyValue) Key() string {
	return mkv.K
}

func (mkv *aplKeyValue) Value() interface{} {
	return mkv.V
}

type sgKeyValue struct {
	K string
	V *SharedGroup
}

func (mkv *sgKeyValue) Key() string {
	return mkv.K
}

func (mkv *sgKeyValue) Value() interface{} {
	return mkv.V
}

type alsKeyValue struct {
	K string
	V AliasValues
}

func (mkv *alsKeyValue) Key() string {
	return mkv.K
}

func (mkv *alsKeyValue) Value() interface{} {
	return mkv.V
}

type loadKeyValue struct {
	K string
	V []*utils.LoadInstance
}

func (mkv *loadKeyValue) Key() string {
	return mkv.K
}

func (mkv *loadKeyValue) Value() interface{} {
	return mkv.V
}

func CacheTypeFactory(prefix string, key string, value interface{}) cacheKeyValue {
	switch prefix {
	case utils.DESTINATION_PREFIX:
		if value != nil {
			return &mapKeyValue{key, value.(map[string]struct{})}
		}
		return &mapKeyValue{"", make(map[string]struct{})}
	case utils.RATING_PLAN_PREFIX:
		if value != nil {
			return &rpKeyValue{key, value.(*RatingPlan)}
		}
		return &rpfKeyValue{"", &RatingProfile{}}
	case utils.RATING_PROFILE_PREFIX:
		if value != nil {
			return &rpfKeyValue{key, value.(*RatingProfile)}
		}
		return &rpfKeyValue{"", &RatingProfile{}}
	case utils.LCR_PREFIX:
		if value != nil {
			return &lcrKeyValue{key, value.(*LCR)}
		}
		return &lcrKeyValue{"", &LCR{}}
	case utils.DERIVEDCHARGERS_PREFIX:
		if value != nil {
			return &dcKeyValue{key, value.(*utils.DerivedChargers)}
		}
		return &dcKeyValue{"", &utils.DerivedChargers{}}
	case utils.ACTION_PREFIX:
		if value != nil {
			return &acsKeyValue{key, value.(Actions)}
		}
		return &acsKeyValue{"", Actions{}}
	case utils.ACTION_PLAN_PREFIX:
		if value != nil {
			return &aplKeyValue{key, value.(*ActionPlan)}
		}
		return &aplKeyValue{"", &ActionPlan{}}
	case utils.SHARED_GROUP_PREFIX:
		if value != nil {
			return &sgKeyValue{key, value.(*SharedGroup)}
		}
		return &sgKeyValue{"", &SharedGroup{}}
	case utils.ALIASES_PREFIX:
		if value != nil {
			return &alsKeyValue{key, value.(AliasValues)}
		}
		return &alsKeyValue{"", AliasValues{}}
	case utils.LOADINST_KEY[:PREFIX_LEN]:
		if value != nil {
			return &loadKeyValue{key, value.([]*utils.LoadInstance)}
		}
		return &loadKeyValue{"", make([]*utils.LoadInstance, 0)}
	}
	return nil
}
