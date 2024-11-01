package utils

import (
	"strconv"

	"github.com/joijoku/PR/shared"
)

func StringToInt(val string) int {
	i, err := strconv.Atoi(val)
	shared.CheckErr(err)

	return i
}
