package util

import (
	"encoding/json"
	"strconv"
	"strings"
)

func GetType(value string) string {
	if _, err := strconv.Atoi(value); err == nil {
		return "int"
	}

	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		var obj map[string]interface{}
		if json.Unmarshal([]byte(value), &obj) == nil {
			return "object"
		}
	}

	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		var arr []interface{}
		if json.Unmarshal([]byte(value), &arr) == nil {
			return "array"
		}
	}

	return "string"
}
