package utils

import (
	"strconv"

	"pr.net/shared"
)

func StringToInt(val string) int {
	i, err := strconv.Atoi(val)
	shared.CheckErr(err)

	return i
}
