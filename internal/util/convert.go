package util

import (
	"fmt"
	"strconv"
)

const (
	NaN         = "NaN"
	INVALID_int = 0
)

func IntAsHexString(i int) (string, error) {
	if i < 0 || i > 15 {
		return NaN, fmt.Errorf("unable to convert to hex string. Value: %d exceeds range 0-15", i)
	}
	h := fmt.Sprintf("%X", i)
	return h, nil
}

func HexAsInt(s string) (int, error) {
	n, err := strconv.ParseInt(s, 16, 8)
	if err != nil {
		return INVALID_int, fmt.Errorf("unable to parse: %s as hex. Detail error: %w", s, err)
	}
	if n < 0 || n > 15 {
		return INVALID_int, fmt.Errorf("unable to convert from hex string. Value: %s exceeds range 0-15", s)
	}
	nu := int(n)
	return nu, nil
}

func BoundTo(i int, min int, max int) int {
	if i < min {
		return min
	}
	if i > max {
		return max
	}
	return i
}
