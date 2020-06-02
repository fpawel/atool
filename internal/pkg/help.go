package pkg

import (
	"encoding/json"
	"github.com/fpawel/atool/internal/pkg/must"
	"strconv"
	"strings"
)

func FormatFloat(v float64, precision int) string {
	s := strconv.FormatFloat(v, 'f', precision, 64)

	for len(s) > 0 && strings.Contains(s, ".") && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}

	for len(s) > 0 && s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s
}

func MustStructToMap(data interface{}) map[string]interface{} {
	m, err := StructToMap(data)
	must.PanicIf(err)
	return m
}

func StructToMap(data interface{}) (map[string]interface{}, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	mapData := make(map[string]interface{})
	err = json.Unmarshal(dataBytes, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}
