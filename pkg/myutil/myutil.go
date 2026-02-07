package myutil

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"time"
)

func MapToJson(param map[string]any) (string, error) {
	dataType, err := json.Marshal(param)
	if err != nil {
		return "", err
	}
	dataString := string(dataType)
	return dataString, nil
}

func JsonToMap(str string) (map[string]any, error) {
	var tempMap map[string]any
	err := json.Unmarshal([]byte(str), &tempMap)
	if err != nil {
		return nil, err
	}
	if tempMap == nil {
		tempMap = make(map[string]any)
	}
	return tempMap, nil
}

func IsMember[T int | float64 | string, Slice1 []T](elem T, elems Slice1) bool {
	for _, v := range elems {
		if v == elem {
			return true
		}
	}
	return false
}

func StructToMap(obj any) map[string]any {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Struct {
		return nil
	}

	objType := objValue.Type()
	result := make(map[string]any)

	for i := 0; i < objValue.NumField(); i++ {
		fieldName := objType.Field(i).Name
		fieldValue := objValue.Field(i).Interface()
		result[fieldName] = fieldValue
	}

	return result
}

func RandomString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // 设置随机种子（局部 RNG，避免影响全局）
	const charset = "0123456789"                    // 可以包含的字符集
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}
