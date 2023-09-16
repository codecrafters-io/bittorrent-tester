package bencode

import (
	"fmt"
	"sort"
)

func encodeDictionary(decoded map[string]interface{}) (string, error) {
	result := "d"

	// Create a slice of keys and sort them
	keys := make([]string, 0, len(decoded))
	for k := range decoded {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Iterate over sorted keys
	for _, key := range keys {
		value := decoded[key]
		encodedKey, err := encodeString(key)
		if err != nil {
			return "", fmt.Errorf("Expected string %#v", key)
		}
		encodedValue, err := EncodeBencode(value)
		if err != nil {
			return "", fmt.Errorf("Unable to encode %#v", value)
		}
		result = result + encodedKey + encodedValue
	}

	return result + "e", nil
}

func encodeString(decoded string) (string, error) {
	length := len(decoded)
	encoded := fmt.Sprintf("%d:%s", length, decoded)
	return encoded, nil
}

func encodeInt(decoded int) (string, error) {
	return fmt.Sprintf("i%de", decoded), nil
}

func encodeList(decoded []interface{}) (string, error) {
	result := "l"
	for _, elem := range decoded {
		encodedElem, err := EncodeBencode(elem)
		if err != nil {
			return "", fmt.Errorf("Unable to encode %#v", elem)
		}
		result += encodedElem
	}
	return result + "e", nil
}

func EncodeBencode(decodedInterface interface{}) (string, error) {
	switch decoded := decodedInterface.(type) {
	case map[string]interface{}:
		return encodeDictionary(decoded)
	case string:
		return encodeString(decoded)
	case int:
		return encodeInt(decoded)
	case []interface{}:
		return encodeList(decoded)
	default:
		return "", fmt.Errorf("No encoder")
	}
}
