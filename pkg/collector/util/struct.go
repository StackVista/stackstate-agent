package util

import "fmt"

// ConvertMapInterfaceToMapString recursively converts map[interface]x (that may appear after YAML decoding)
// to map[string]x that can be encoded to JSON without errors
func ConvertMapInterfaceToMapString(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		newMap := map[string]interface{}{}
		for k, v := range x {
			switch kt := k.(type) {
			case string:
				newMap[kt] = ConvertMapInterfaceToMapString(v)
			default:
				newMap[fmt.Sprintf("%v", kt)] = ConvertMapInterfaceToMapString(v)
			}
		}
		return newMap
	case []interface{}:
		for i, v := range x {
			x[i] = ConvertMapInterfaceToMapString(v)
		}
	}
	return i
}
