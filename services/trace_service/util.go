package main

import (
	"math"
	"reflect"
	"strconv"
	"strings"
)

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

func getInt(obj interface{}, path string) int {
	return getObject(obj, path, "int").i
}
func getInt64(obj interface{}, path string) int64 {
	return getObject(obj, path, "int64").i64
}
func getFloat64(obj interface{}, path string) float64 {
	return getObject(obj, path, "float64").f64
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
					if typo == "float64" {
						st.f64 = value.(float64)
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
func toInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}
func toInt(s string) int {
	i, _ := strconv.ParseInt(s, 10, 64)
	return int(i)
}
func chunkSlice[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}
	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
