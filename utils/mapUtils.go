package utils

import (
	"encoding/json"

	"pr.net/shared"
)

func ListOrMapToJsonString(mp interface{}) (string, error) {
	var err error
	var result []byte
	shared.Block{
		Try: func() {
			result, err = json.Marshal(mp)
			shared.CheckErr(err)
		},
		Catch: func(e shared.Exception) {
			result = []byte("")
			err = e.(error)
		},
	}.Do()

	return string(result), err
}

func IsListContains(list []string, find string) (bool, int) {
	result := false
	var idx int
	for i, each := range list {
		if find == each {
			result = true
			idx = i
		}
	}

	return result, idx
}

func IsKeyExistOnMap(maps map[string]interface{}, list []string) bool {
	result := false
	lsResult := make([]bool, 0)

	for key, x := range maps {
		ShowDebug((x))
		for _, each := range list {
			if key == each {
				lsResult = append(lsResult, true)
			}
		}
	}

	if len(lsResult) > 0 {
		result = true
		for _, each := range lsResult {
			result = result && each
		}
	}

	return result
}
