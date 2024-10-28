package utils

func CheckInterface(myInterface interface{}) string {
	result := ""

	switch myInterface.(type) {
	case int:
		result = "integer"
	case float64:
		result = "float64"
	case string:
		result = "string"
	case map[string]interface{}:
		result = "map"
	case []interface{}:
		result = "list"
	case bool:
		result = "bool"
	default:
		// And here I'm feeling dumb. ;)
		result = "none"
	}

	return result
}

func InterfaceToString(myInterface interface{}) string {
	return myInterface.(string)
}

func InterfaceToInteger(myInterface interface{}) int {
	return myInterface.(int)
}

func InterfaceToFloat64(myInterface interface{}) float64 {
	return myInterface.(float64)
}

func InterfaceToBool(myInterface interface{}) bool {
	return myInterface.(bool)
}
