package ssb

import (
	"fmt"

	"github.com/pilosa/pdk/leveldb"
	"github.com/pkg/errors"
)

type translator struct {
	lt *leveldb.Translator
}

func newTranslator(storedir string) (*translator, error) {
	lt, err := leveldb.NewTranslator(storedir, []string{"c_city", "c_nation", "c_region", "s_city", "s_nation", "s_region", "p_mfgr", "p_category", "p_brand1"}...)
	if err != nil {
		return nil, err
	}
	return &translator{
		lt: lt,
	}, nil
}

func (t *translator) Get(field string, id uint64) (interface{}, error) {
	switch field {
	case "c_city", "c_nation", "c_region", "s_city", "s_nation", "s_region", "p_mfgr", "p_category", "p_brand1":
		val, err := t.lt.Get(field, id)
		if err != nil {
			return nil, errors.Wrap(err, "string from level translator")
		}
		return string(val.([]byte)), nil
	case "lo_month":
		return monthsSlice[id], nil
	case "lo_weeknum", "lo_year", "lo_quantity_b", "lo_discount_b":
		return id, nil
	default:
		return nil, errors.Errorf("Unimplemented in ssb.Translator.Get field: %v, id: %v", field, id)
	}
}

func (t *translator) GetID(field string, val interface{}) (uint64, error) {
	switch field {
	case "c_city", "c_nation", "c_region", "s_city", "s_nation", "s_region", "p_mfgr", "p_category", "p_brand1":
		return t.lt.GetID(field, []byte(val.(string)))
	case "lo_month":
		valstring := val.(string)
		m, ok := months[valstring]
		if !ok {
			return 0, fmt.Errorf("Val '%s' is not a month", val)
		}
		return m, nil
	case "lo_weeknum", "lo_quantity_b", "lo_discount_b":
		val8, ok := val.(uint8)
		if !ok {
			return 0, fmt.Errorf("Val '%v' is not a valid weeknum/quantity/discount (not uint8)", val)
		}
		return uint64(val8), nil
	case "lo_year":
		val16, ok := val.(uint16)
		if !ok {
			return 0, fmt.Errorf("Val '%v' is not a valid year (not uint16)", val)
		}
		return uint64(val16), nil
	default:
		return 0, fmt.Errorf("Unimplemented in ssb.Translator.GetID field: %v, val: %v", field, val)
	}
}

var months = map[string]uint64{
	"January":   0,
	"February":  1,
	"March":     2,
	"April":     3,
	"May":       4,
	"June":      5,
	"July":      6,
	"Augest":    7,
	"September": 8,
	"Octorber":  9,
	"November":  10,
	"December":  11,
}

var monthsSlice = []string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"Augest",
	"September",
	"Octorber",
	"November",
	"December",
}
