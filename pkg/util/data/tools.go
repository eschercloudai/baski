package data

import "fmt"

// GetNestedField enables the ability to step through a map[string]interface{} to get sub keys
func GetNestedField(data map[string]interface{}, keys ...string) (interface{}, error) {
	var result interface{} = data

	for _, key := range keys {
		v, ok := result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("key %s not found or not a map\n", key)
		}

		result = v[key]
	}

	return result, nil
}
