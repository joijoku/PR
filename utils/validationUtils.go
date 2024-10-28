package utils

import "github.com/gookit/validate"

func SetValidation(mp map[string]interface{}, valMap map[string]string) *validate.Validation {
	v := validate.Map(mp)

	for eachKey, eachVal := range valMap {
		v.StringRule(eachKey, eachVal)
	}

	return v
}
