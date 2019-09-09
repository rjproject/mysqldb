package mysqldb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

func Export(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, b, "", "\t")
	if err != nil {
		return ""
	}
	return buf.String()
}

func Json(v interface{}) string {
	return Export(v)
}

func Keys(src map[string]interface{}) []string {
	res := make([]string, 0)
	for k, _ := range src {
		res = append(res, k)
	}
	return res
}

func inSlice(v interface{}, s interface{}) bool {
	val := reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() == reflect.Slice {
		log.Println("inSlice method: first parameter cannot be a slice")
		return false
	}

	sv := reflect.Indirect(reflect.ValueOf(s))
	if sv.Kind() != reflect.Slice {
		log.Println("inSlice method: second parameter Needs a slice")
		return false
	}

	st := reflect.TypeOf(s).Elem().String()
	vt := reflect.TypeOf(v).String()
	if st != vt {
		log.Println("inSlice method: two parameters are not the same type")
		return false
	}

	switch vt {
	case "string":
		for _, vv := range s.([]string) {
			if vv == v {
				return true
			}
		}
	case "int":
		for _, vv := range s.([]int) {
			if vv == v {
				return true
			}
		}
	case "int64":
		for _, vv := range s.([]int64) {
			if vv == v {
				return true
			}
		}
	case "float64":
		for _, vv := range s.([]float64) {
			if vv == v {
				return true
			}
		}
	default:
		log.Println("inSlice method: this type is not supported")
		return false
	}

	return false
}

func implode(data interface{}, sep string) string {
	sv := reflect.Indirect(reflect.ValueOf(data))
	if sv.Kind() != reflect.Slice {
		log.Panicln("implode method: second parameter needs a slice")
		return ""
	}
	var tmp []string
	st := reflect.TypeOf(data).Elem().String()
	switch st {
	case "string":
		for _, vv := range data.([]string) {
			tmp = append(tmp, singleQuotes(escapeString(vv)))
		}
	case "interface {}":
		for _, vv := range data.([]interface{}) {
			tmp = append(tmp, singleQuotes(escapeString(formatString(vv))))
		}
	case "int":
		for _, vv := range data.([]int) {
			tmp = append(tmp, formatString(vv))
		}
		return strings.Join(tmp, sep)
	case "int64":
		for _, vv := range data.([]int64) {
			tmp = append(tmp, formatString(vv))
		}
	case "float64":
		for _, vv := range data.([]float64) {
			tmp = append(tmp, formatString(vv))
		}
	default:
		log.Panicln("implode method: this type is not supported")
		return ""
	}
	return strings.Join(tmp, sep)
}

func iface2Slice(data interface{}) []interface{} {
	res := make([]interface{}, 0)

	sv := reflect.Indirect(reflect.ValueOf(data))
	if sv.Kind() != reflect.Slice {
		return res
	}

	st := reflect.TypeOf(data).Elem().String()
	switch st {
	case "string":
		for _, vv := range data.([]string) {
			res = append(res, vv)
		}
	case "interface {}":
		for _, vv := range data.([]interface{}) {
			res = append(res, vv)
		}
	case "int":
		for _, vv := range data.([]int) {
			res = append(res, vv)
		}
	case "int64":
		for _, vv := range data.([]int64) {
			res = append(res, vv)
		}
	case "float64":
		for _, vv := range data.([]float64) {
			res = append(res, vv)
		}
	default:
		log.Panicln("iface2Slice method: this type is not supported")
	}

	return res
}

func unique(list *[]string) []string {
	r := make([]string, 0)
	temp := map[string]byte{}
	for _, v := range *list {
		l := len(temp)
		temp[v] = 0
		if len(temp) != l {
			r = append(r, v)
		}
	}
	return r
}

func isUint8Slice(arg interface{}) bool {
	switch arg.(type) {
	case []uint8:
		return true
	default:
		return false
	}
}

func singleQuotes(data interface{}) string {
	return "'" + strings.Trim(escapeString(formatString(data)), " ") + "'"
}

func convertInt(v interface{}) (int64, error) {
	switch v.(type) {
	case int:
		return int64(v.(int)), nil
	case int8:
		return int64(v.(int8)), nil
	case int16:
		return int64(v.(int16)), nil
	case int32:
		return int64(v.(int32)), nil
	case int64:
		return v.(int64), nil
	case float32:
		return int64(v.(float32)), nil
	case float64:
		return int64(v.(float64)), nil
	case []byte:
		i, err := strconv.ParseInt(string(v.([]byte)), 10, 64)
		if err != nil {
			return 0, err
		}
		return i, nil
	case string:
		i, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return 0, err
		}
		return i, nil
	}
	return 0, fmt.Errorf("Unsupported type: %v", v)
}

func atoi(v interface{}) int {
	n, err := convertInt(v)
	if err != nil {
		return 0
	}
	return int(n)
}

func startsWith(s, substr string) bool {
	if substr != "" && Substr(s, 0, len([]rune(substr))) == substr {
		return true
	}
	return false
}

func endsWith(s, substr string) bool {
	if Substr(s, -len([]rune(substr)), len(s)) == substr {
		return true
	}
	return false
}

func Substr(s string, start, length int) string {
	strlist := []rune(s)
	l := len(strlist)
	end := 0

	if start < 0 {
		start = l + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}

	if start > l {
		start = l
	}

	if end < 0 {
		end = 0
	}

	if end > l {
		end = l
	}

	return string(strlist[start:end])
}

// Escape strings
func escapeString(s string) string {
	str := strconv.Quote(s)
	str = strings.Replace(str, "'", "\\'", -1)
	strlist := []rune(str)
	l := len(strlist)
	return Substr(str, 1, l-2)
}

// Interface{} to strings
func formatString(iface interface{}) string {
	switch val := iface.(type) {
	case []byte:
		return string(val)
	}
	v := reflect.ValueOf(iface)
	switch v.Kind() {
	case reflect.Invalid:
		return ""
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Ptr:
		b, err := json.Marshal(v.Interface())
		if err != nil {
			return "nil"
		}
		return string(b)
	}
	return fmt.Sprintf("%v", iface)
}

func FormatUpper(str string) string {
	upperIndex := []int{}
	for i, v := range str {
		if v >= 65 && v <= 90 {
			upperIndex = append(upperIndex, i)
		}
	}
	s := strings.ToLower(str)
	name := []string{}

	if len(upperIndex) == 0 {
		return s
	}

	for i, v := range upperIndex {
		if i == 0 && v != 0 {
			name = append(name, s[:v])
			if i+1 > len(upperIndex)-1 {
				name = append(name, s[v:])
			} else {
				name = append(name, s[v:upperIndex[i+1]])
			}

			continue
		}
		if i == len(upperIndex)-1 && i != len(s)-1 {
			name = append(name, s[v:])
			continue
		}
		if i+1 > len(upperIndex)-1 {
			name = append(name, s[v:])
		} else {
			name = append(name, s[v:upperIndex[i+1]])
		}

	}
	return strings.Join(name, "_")
}

func json_encode(i interface{}) string {
	//json := jsoniter.ConfigCompatibleWithStandardLibrary
	jsonByte, err := json.Marshal(i)
	if err != nil {
		return ""
	}
	return string(jsonByte)
}

func map2struct(i interface{}, m map[string]interface{}) error {
	//json := jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, i)
}

func slice2struct(i interface{}, s []map[string]interface{}) error {
	//json := jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, i)
}

func empty(arg interface{}) bool {
	if arg == nil {
		return true
	}
	v := reflect.ValueOf(arg)
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		if v.Len() > 0 {
			return false
		}
		return true
	case reflect.Ptr:
		v = v.Elem()
	}
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

func ReflectFields(iface interface{}) []string {
	fields := make([]string, 0)
	v := reflect.ValueOf(iface).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		iKey := t.Field(i).Tag.Get("json")
		if iKey == "" && f.Kind() != reflect.Struct {
			iKey = FormatUpper(t.Field(i).Name)
		}
		fields = append(fields, iKey)
	}
	return fields
}

func reflectValue(iface interface{}, source map[string]interface{}) (err error) {
	v := reflect.ValueOf(iface).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		iKey := t.Field(i).Tag.Get("json")
		if iKey == "" && f.Kind() != reflect.Struct {
			iKey = FormatUpper(t.Field(i).Name)
		}
		iVal := source[iKey]
		/* iVal, ok := source[iKey]
		if !ok {
			continue
		} */
		switch f.Kind() {
		case reflect.Bool:
			val, e := strconv.ParseBool(formatString(iVal))
			if e != nil {
				err = e
			}
			f.SetBool(val)
		case reflect.String:
			f.SetString(formatString(iVal))
		case reflect.Float32, reflect.Float64:
			val, e := strconv.ParseFloat(formatString(iVal), 10)
			if e != nil {
				err = e
			}
			f.SetFloat(val)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, e := strconv.ParseInt(formatString(iVal), 10, 64)
			if e != nil {
				val = 0
				err = e
			}
			f.SetInt(val)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			val, e := strconv.ParseUint(formatString(iVal), 10, 64)
			if e != nil {
				val = 0
				err = e
			}
			f.SetUint(val)
		case reflect.Struct:
			if f.CanSet() {
				if iKey == "" {
					iVal = source
				}
				childFace := reflect.New(f.Type())
				switch iVal.(type) {
				case map[string]interface{}:
					mVal, ok := iVal.(map[string]interface{})
					if !ok {
						err = errors.New(iKey + " is not map[string]interface{}")
					}
					reflectValue(childFace.Interface(), mVal)
					f.Set(childFace.Elem())
				case string:
					e := json.Unmarshal([]byte(iVal.(string)), childFace.Interface())
					if e != nil {
						err = e
					} else {
						f.Set(childFace.Elem())
					}
				}
			}
		}
	}
	return err
}
