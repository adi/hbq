package hbq

import (
	"fmt"
	"net/url"
	"reflect"
	"unicode"
)

func ToSnakeCase(s string) string {
	var res = make([]rune, 0, len(s))
	var p = '_'
	for i, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			res = append(res, '_')
		} else if unicode.IsUpper(r) && i > 0 {
			if unicode.IsLetter(p) && !unicode.IsUpper(p) || unicode.IsDigit(p) {
				res = append(res, '_', unicode.ToLower(r))
			} else {
				res = append(res, unicode.ToLower(r))
			}
		} else {
			res = append(res, unicode.ToLower(r))
		}

		p = r
	}
	return string(res)
}

func HttpBuildQueryRecursive(path string, params url.Values, obj interface{}) {
	vobj := reflect.ValueOf(obj)
	if vobj.Kind() == reflect.Ptr {
		vobj = vobj.Elem()
	}
	if vobj.Kind() == reflect.Struct {
		iobj := reflect.Indirect(vobj)
		for fi := 0; fi < vobj.NumField(); fi++ {
			key := iobj.Type().Field(fi).Name
			tag := iobj.Type().Field(fi).Tag.Get("json")
			if tag == "" {
				tag = ToSnakeCase(key)
			}
			field := vobj.Field(fi)
			if field.CanInterface() {
				val := field.Interface()
				if len(path) > 0 {
					HttpBuildQueryRecursive(path+"["+tag+"]", params, val)
				} else {
					HttpBuildQueryRecursive(tag, params, val)
				}
			}
		}
	} else if vobj.Kind() == reflect.Map {
		iter := vobj.MapRange()
		for iter.Next() {
			tag := fmt.Sprintf("%v", iter.Key().Interface())
			val := iter.Value().Interface()
			if len(path) > 0 {
				HttpBuildQueryRecursive(path+"["+tag+"]", params, val)
			} else {
				HttpBuildQueryRecursive(tag, params, val)
			}
		}
	} else if vobj.Kind() == reflect.Slice || vobj.Kind() == reflect.Array {
		for i := 0; i < vobj.Len(); i++ {
			vobjidx := vobj.Index(i)
			if vobjidx.Kind() == reflect.Ptr {
				vobjidx = vobjidx.Elem()
			}
			val := vobjidx.Interface()
			idx := fmt.Sprintf("[%d]", i)
			if vobjidx.Kind() == reflect.Struct || vobjidx.Kind() == reflect.Map {
				HttpBuildQueryRecursive(path+idx, params, val)
			} else {
				params.Add(path+idx, fmt.Sprintf("%v", val))
			}
		}
	} else {
		params.Add(path, fmt.Sprintf("%v", obj))
	}
}

func HttpBuildQuery(obj interface{}) string {
	params := url.Values{}
	HttpBuildQueryRecursive("", params, obj)
	return params.Encode()
}
