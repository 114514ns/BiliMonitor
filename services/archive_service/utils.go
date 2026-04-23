package main

import (
	"encoding/json"
	"math"
	"math/rand"

	"reflect"

	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
)

func RandomM[K comparable, V any](m map[K]V) V {

	if len(m) == 0 {
		var zero V
		return zero
	}

	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(keys))

	return m[keys[randomIndex]]
}

type JsonType struct {
	s     string
	i     int
	i64   int64
	f32   float32
	f64   float64
	array []interface{}
	v     bool
	m     map[string]interface{}
}

func toInt64(s string) int64 {
	i64, _ := strconv.ParseInt(s, 10, 64)
	return i64
}
func toInt(s string) int {
	i64, _ := strconv.ParseInt(s, 10, 64)
	return int(i64)
}
func toFloat64(s string) float64 {
	i64, _ := strconv.ParseFloat(s, 64)
	return (i64)
}

func getInt(obj interface{}, path string) int {
	return getObject(obj, path, "int").i
}
func getInt64(obj interface{}, path string) int64 {
	return getObject(obj, path, "int64").i64
}
func getString(obj interface{}, path string) string {
	return getObject(obj, path, "string").s
}
func getArray(obj interface{}, path string) []interface{} {
	return getObject(obj, path, "array").array
}
func getBool(obj interface{}, path string) bool {
	return getObject(obj, path, "bool").v
}
func getObject(obj interface{}, path string, typo string) JsonType {
	var array = strings.Split(path, ".")
	inner, ok := obj.(map[string]interface{})
	if !ok {
		return JsonType{}
	}
	var st = JsonType{}
	for i, s := range array {
		if i == len(array)-1 {

			value := inner[s]
			if value != nil {
				var t = reflect.TypeOf(value)
				if t.Kind() == reflect.String {
					st.s = value.(string)
				}
				if t.Kind() == reflect.Int {
					st.i, _ = value.(int)
				}
				if t.Kind() == reflect.Int64 {
					if value.(int64) > math.MaxInt {
						st.i64 = value.(int64)
					} else {
						st.i = value.(int)
					}

				}
				if t.Kind() == reflect.Float64 {
					if typo == "int" {
						st.i = int(value.(float64))
					}
					if typo == "int64" {
						st.i64 = int64(value.(float64))
					}
				}
				if t.Kind() == reflect.Slice {
					if typo == "array" {
						st.array = value.([]interface{})
					}
				}
				if t.Kind() == reflect.Bool {
					st.v = value.(bool)
				}
				if t.Kind() == reflect.Map {
					st.m = value.(map[string]interface{})
				}
			}

			return st
		} else {

			if inner[s] == nil {
				return st
			}
			inner = inner[s].(map[string]interface{})
		}
	}
	return st
}
func toString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func prettyJSON(s string) string {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return ""
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

func DifferenceBy[T any, K comparable](
	list1, list2 []T,
	keyFunc func(T) K,
) ([]T, []T) {

	set1 := make(map[K]struct{})
	for _, v := range list1 {
		set1[keyFunc(v)] = struct{}{}
	}

	set2 := make(map[K]struct{})
	for _, v := range list2 {
		set2[keyFunc(v)] = struct{}{}
	}

	left := lo.Filter(list1, func(v T, _ int) bool {
		_, ok := set2[keyFunc(v)]
		return !ok
	})

	right := lo.Filter(list2, func(v T, _ int) bool {
		_, ok := set1[keyFunc(v)]
		return !ok
	})

	return left, right
}
