package engine

import "github.com/cgrates/cgrates/utils"

func CacheTypeFactory(prefix string) interface{} {
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
