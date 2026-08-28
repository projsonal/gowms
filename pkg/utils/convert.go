package utils

import "strconv"

func UintToString(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}
