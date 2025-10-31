package common

import (
	"errors"
	"strconv"
)

func ConvertInterfaceToStr(anyValue any) (value string, err error) {

	switch anyValue := anyValue.(type) {
	case float64:
		value = ConvertFloat64ToStr(anyValue)
	case int64:
		value = ConvertInt64ToStr(anyValue)
	case uint32:
		value = strconv.FormatUint(uint64(anyValue), 10)
	case uint64:
		value = strconv.FormatUint(anyValue, 10)
	case string:
		value = anyValue
	default:
		err = errors.New("unsupported type")
	}
	return
}

func ConvertFloat64ToStr(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func ConvertInt64ToStr(value int64) string {
	return strconv.FormatInt(value, 10)
}
