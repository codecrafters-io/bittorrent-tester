package bencode

import (
	"fmt"
	"strconv"
	"unicode"
)

func findFirstByte(bencodedString string, b byte) (i int) {
	for i = 0; i < len(bencodedString); i++ {
		if bencodedString[i] == b {
			break
		}
	}
	return i
}

func decodeDictionary(bencodedString string) (map[string]interface{}, int, error) {
	result := make(map[string]interface{})
	var offset int
	var err error

	var key string
	var value interface{}
	newOffset := 1

	for offset = 1; offset < len(bencodedString); offset += newOffset {
		if bencodedString[offset] == 'e' {
			break
		}
		key, newOffset, err = decodeString(bencodedString[offset:])
		if err != nil {
			return result, 0, err
		}
		offset += newOffset
		value, newOffset, err = DecodeBencode(bencodedString[offset:])
		if err != nil {
			return result, 0, err
		}
		result[key] = value
	}
	return result, offset + 1, nil
}

func decodeList(bencodedString string) ([]interface{}, int, error) {
	result := []interface{}{}
	var offset int

	var err error
	var elem interface{}
	newOffset := 1

	for offset = 1; offset < len(bencodedString); offset += newOffset {
		if bencodedString[offset] == 'e' {
			break
		}
		elem, newOffset, err = DecodeBencode(bencodedString[offset:])
		if err != nil {
			return []interface{}{}, 0, err
		}
		result = append(result, elem)
	}

	return result, offset + 1, nil
}

func decodeString(bencodedString string) (string string, offset int, err error) {
	colonIndex := findFirstByte(bencodedString, ':')

	lengthStr := bencodedString[:colonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}

	string = bencodedString[colonIndex+1 : colonIndex+length+1]

	return string, len(strconv.Itoa(length)) + len(string) + 1, nil
}

func decodeInteger(bencodedString string) (integer int, offset int, err error) {
	lastIndex := findFirstByte(bencodedString, 'e')

	numberStr := bencodedString[1:lastIndex]
	number, err := strconv.Atoi(string(numberStr))
	if err != nil {
		return 0, 0, err
	}

	return number, len(strconv.Itoa(number)) + 2, nil
}

func DecodeBencode(bencodedString string) (interface{}, int, error) {
	if rune(bencodedString[0]) == 'l' {
		return decodeList(bencodedString)
	}
	if rune(bencodedString[0]) == 'd' {
		return decodeDictionary(bencodedString)
	}
	if rune(bencodedString[0]) == 'i' {
		return decodeInteger(bencodedString)
	}
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeString(bencodedString)
	} else {
		return "", 0, fmt.Errorf("No decoder")
	}
}
