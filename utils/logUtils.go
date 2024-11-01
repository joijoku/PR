package utils

import (
	"encoding/json"
	"log"

	"github.com/joijoku/PR/shared"
)

func ShowDebug(show interface{}) {
	if shared.GetDebug() {
		log.Println(show)
	}
}

func PrettyStruct(data interface{}) string {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return ""
	}
	return string(val)
}
