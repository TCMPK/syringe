package main

import (
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func SetStructMemberFromEnvVariables(c *ResolverConfiguration) {
	e := reflect.ValueOf(c).Elem()
	env_var_prefix := "syringe"

	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		varNameClean := strings.ToUpper(CleanString(varName))
		varType := e.Type().Field(i).Type
		for _, env := range os.Environ() {
			parts := strings.Split(env, "=")
			envKey := parts[0]
			envKeyClean := strings.ToUpper(CleanString(strings.Replace(strings.ToLower(envKey), strings.ToLower(env_var_prefix)+"_", "", 1)))
			envValue := strings.Join(parts[1:], "=")
			if varNameClean == envKeyClean {
				e.Field(i).Set(reflect.ValueOf(ConvertToNumberIfNumeric(envValue, varType)))
			}
		}
	}
}

func CleanString(s string) string {
	re := regexp.MustCompile(`[^\w]`)
	return re.ReplaceAllString(s, "")
}

func ConvertToNumberIfNumeric(data string, t reflect.Type) any {
	num, err := strconv.Atoi(data)

	if err == nil {
		if reflect.Uint64 == t.Kind() {
			return uint64(num)
		}
		if reflect.Uint32 == t.Kind() {
			return uint32(num)
		}
		if reflect.Int32 == t.Kind() {
			return int32(num)
		}
		if reflect.Int == t.Kind() {
			return int(num)
		}
		if reflect.Int64 == t.Kind() {
			return int64(num)
		}
	}
	return data
}
