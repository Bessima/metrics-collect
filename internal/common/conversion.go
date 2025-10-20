package common

import (
	"errors"
	"strconv"
)

func ConvertInterfaceToStr(anyValue any) (value string, err error) {

	switch anyValue := anyValue.(type) {
	case float64:
		value = strconv.FormatFloat(anyValue, 'f', -1, 64)
	case int64:
		value = strconv.FormatInt(anyValue, 10)
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
