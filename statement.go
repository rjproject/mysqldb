package mysqldb

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

type Statement struct {
	adapter   *Adapter
	TableName string
	alias     string
	pk        string
	fields    []string
	join      string
	where     [][]interface{}
	whereRaw  string
	orderBy   string
	groupBy   string
	limit     string
	distinct  string
	operator  map[string]string
}

func (statement *Statement) Table(table string) *Statement {
	if table == "" {
		statement.Error(TABLENAME_ERROR, true)
		return statement
	}
	statement.TableName = strings.TrimSpace(table)
	return statement
}

func (statement *Statement) As(args string) *Statement {
	if args == "" {
		statement.Error(PARAMETER_ERROR, true)
		return statement
	}
	statement.alias = strings.TrimSpace(args)
	return statement
}

func (statement *Statement) SetPk(pk string) *Statement {
	if pk == "" {
		statement.Error(PARAMETER_ERROR, true)
		return statement
	}
	statement.pk = pk
	return statement
}

func (statement *Statement) Fileds(args ...string) *Statement {
	statement.fields = args
	return statement
}

func (statement *Statement) Distinct(args string) *Statement {
	if strings.Trim(args, " ") != "" {
		statement.distinct = "DISTINCT " + args
	} else {
		statement.distinct = ""
	}
	return statement
}

func (statement *Statement) Where(args ...interface{}) *Statement {
	l := len(args)
	if l == 0 {
		statement.Error(PARAMETER_ERROR, true)
		return statement
	}
	if l > 3 {
		args = args[0:3]
	}
	statement.where = append(statement.where, []interface{}{statement.operator["and"], args})
	return statement
}

func (statement *Statement) OrWhere(args ...interface{}) *Statement {
	l := len(args)
	if l == 0 {
		statement.Error(PARAMETER_ERROR, true)
		return statement
	}
	if l > 3 {
		args = args[0:3]
	}
	statement.where = append(statement.where, []interface{}{statement.operator["or"], args})
	return statement
}

func (statement *Statement) WhereRaw(args string) *Statement {
	if args == "" {
		statement.Error(PARAMETER_ERROR)
		return statement
	}
	statement.whereRaw = strings.Trim(args, " ")
	return statement
}

func (statement *Statement) WhereIn(field string, values interface{}) *Statement {
	valuesType := reflect.ValueOf(values).Kind().String()
	if valuesType != "slice" || field == "" {
		statement.Error(PARAMETER_ERROR, true)
		return statement
	}
	args := make([]interface{}, 0)
	args = append(args, field)
	args = append(args, statement.operator["in"])
	args = append(args, values)
	statement.where = append(statement.where, []interface{}{statement.operator["and"], args})
	return statement
}

func (statement *Statement) WhereNotIn(field string, values interface{}) *Statement {
	valuesType := reflect.ValueOf(values).Kind().String()
	if valuesType != "slice" || field == "" {
		statement.Error(PARAMETER_ERROR, true)
		return statement
	}
	args := make([]interface{}, 0)
	args = append(args, field)
	args = append(args, statement.operator["nin"])
	args = append(args, values)
	statement.where = append(statement.where, []interface{}{statement.operator["and"], args})
	return statement
}

func (statement *Statement) Limit(args ...int) *Statement {
	if len(args) == 0 {
		statement.Error(PARAMETER_ERROR, true)
		return statement
	}

	if len(args) > 1 {
		statement.limit = fmt.Sprintf(" Limit %d,%d", args[0], args[1])
		return statement
	}

	statement.limit = fmt.Sprintf(" Limit %d", args[0])

	return statement
}

func (statement *Statement) GroupBy(args string) *Statement {
	if args == "" {
		statement.Error(PARAMETER_ERROR)
		return statement
	}
	statement.orderBy = fmt.Sprintf(" ORDER BY %v", args)
	return statement
}

func (statement *Statement) OrderBy(args string) *Statement {
	if args == "" {
		statement.Error(PARAMETER_ERROR)
		return statement
	}
	statement.orderBy = fmt.Sprintf(" ORDER BY %v", args)
	return statement
}

func (statement *Statement) LeftJoin(table, condition string) *Statement {
	if condition == "" {
		statement.Error(PARAMETER_ERROR)
		return statement
	}
	if statement.alias == "" {
		statement.alias = "A"
	}
	statement.join = fmt.Sprintf(" LEFT JOIN %v AS B ON %v", table, condition)
	return statement
}

func (statement *Statement) RightJoin(table, condition string) *Statement {
	if condition == "" {
		statement.Error(PARAMETER_ERROR)
		return statement
	}
	if statement.alias == "" {
		statement.alias = "A"
	}
	statement.join = fmt.Sprintf(" RIGHT JOIN %v AS B ON %v", table, condition)
	return statement
}

func (statement *Statement) Join(table, condition string) *Statement {
	if condition == "" {
		statement.Error(PARAMETER_ERROR)
		return statement
	}
	if statement.alias == "" {
		statement.alias = "A"
	}
	statement.join = fmt.Sprintf(" INNER JOIN %v AS B ON %v", table, condition)
	return statement
}

func (statement *Statement) FullJoin(table, condition string) *Statement {
	if condition == "" {
		statement.Error(PARAMETER_ERROR)
		return statement
	}
	if statement.alias == "" {
		statement.alias = "A"
	}
	statement.join = fmt.Sprintf(" FULL JOIN %v AS B ON %v", table, condition)
	return statement
}

func (statement *Statement) parseField() string {
	if statement.distinct != "" {
		return statement.distinct
	}
	if len(statement.fields) == 0 {
		return "*"
	} else {
		return strings.Join(statement.fields, ",")
	}
}

func (statement *Statement) parseTableName() string {
	if statement.TableName == "" {
		statement.Error(TABLENAME_ERROR, true)
	}
	if statement.alias != "" {
		statement.TableName = statement.TableName + " AS " + statement.alias
	}
	return statement.TableName
}

func (statement *Statement) Init() {
	statement.TableName = ""
	statement.pk = ""
	statement.fields = []string{}
	statement.where = [][]interface{}{}
	statement.whereRaw = ""
	statement.limit = ""
	statement.orderBy = ""
	statement.groupBy = ""
	statement.join = ""
	statement.distinct = ""
	statement.operator = map[string]string{
		"eq":  "=",
		"gt":  ">",
		"lt":  "=",
		"ne":  "!=",
		"ge":  ">=",
		"le":  "<=",
		"in":  "IN",
		"nin": "NOT IN",
		"and": "AND",
		"or":  "OR"}
}

func (statement *Statement) prepareWhere() (string, []interface{}) {
	condition := make([]string, 0)
	params := make([]interface{}, 0)
	for _, v := range statement.where {
		joiner := v[0].(string)
		val := v[1].([]interface{})

		if empty(val[0]) {
			continue
		}

		l := len(val)
		switch l {
		case 1:
			switch v := val[0].(type) {
			case string:
				if strings.Trim(v, " ") == "" {
					continue
				}
				condition = append(condition, joiner+" ("+v+")")
			case map[string]interface{}:
				whereMap := make([]string, 0)
				for key, val := range v {
					whereMap = append(whereMap, key+" = "+singleQuotes(val))
				}
				condition = append(condition, joiner+" ("+strings.Join(whereMap, " AND ")+")")
			}
		default:
			c, b := statement.bindParams(val)
			condition = append(condition, fmt.Sprintf("%s %s", joiner, c))
			params = append(params, b...)
		}
	}

	if statement.whereRaw != "" {
		condition = append(condition, "AND "+statement.whereRaw)
	}

	cond := ""
	if len(condition) > 0 {
		cond = strings.Trim(strings.Join(condition, " "), " ")[4:]
	}

	if cond == "" {
		return cond, params
	}

	return fmt.Sprintf(" WHERE %s", cond), params
}

func (statement *Statement) bindParams(args []interface{}) (string, []interface{}) {
	l := len(args)

	operator := []string{"=", ">", "<", "!=", "<>", ">=", "<=", "like", "in", "not in"}

	var params []interface{}

	if l == 2 {
		return fmt.Sprintf("%s = ?", args[0]), []interface{}{args[1]}
	}

	joiner := strings.ToLower(formatString(args[1]))
	if !inSlice(joiner, operator) {
		log.Panicln("bindParams method: operator error")
	}

	params = append(params, args[0])
	params = append(params, args[1])

	placeParams := make([]interface{}, 0)
	placeholder := make([]string, 0)
	if inSlice(joiner, []string{"in", "not in"}) {
		placeParams = iface2Slice(args[2])
		placeholder = make([]string, len(placeParams))
		for i := 0; i < len(placeParams); i++ {
			placeholder[i] = "?"
		}
	}

	switch joiner {
	case "in", "not in":
		return fmt.Sprintf("%s %s (%s)", args[0], args[1], strings.Join(placeholder, ",")), placeParams
	case "like":
		params = append(params, escapeString(formatString(args[2])))
	default:
		params = append(params, args[2])
	}

	return fmt.Sprintf("%s %s ?", args[0], args[1]), []interface{}{args[2]}
}

func (statement *Statement) formatWhere() string {
	condition := make([]string, 0)
	for _, args := range statement.where {
		joiner := args[0].(string)
		params := args[1].([]interface{})

		if empty(params[0]) {
			continue
		}

		l := len(params)
		switch l {
		case 1:
			switch v := params[0].(type) {
			case string:
				condition = append(condition, fmt.Sprintf("%s (%s)", joiner, v))
			case map[string]interface{}:
				whereMap := make([]string, 0)
				for key, val := range v {
					whereMap = append(whereMap, fmt.Sprintf("%s = %s", key, singleQuotes(formatString(val))))
				}
				condition = append(condition, fmt.Sprintf("%s %s", joiner, strings.Join(whereMap, " AND ")))
			}
		case 2:
			condition = append(condition, fmt.Sprintf("%s %s", joiner, statement.formatParams(params)))
		case 3:
			condition = append(condition, fmt.Sprintf("%s %s", joiner, statement.formatParams(params)))
		}
	}

	if statement.whereRaw != "" {
		condition = append(condition, "AND "+statement.whereRaw)
	}

	cond := ""
	if len(condition) > 0 {
		cond = strings.Trim(strings.Join(condition, " "), " ")[4:]
	}

	if cond == "" {
		return cond
	}

	return " WHERE " + cond
}

func (statement *Statement) formatParams(args []interface{}) string {
	l := len(args)

	operator := []string{"=", ">", "<", "!=", "<>", ">=", "<=", "like", "in", "not in"}

	var params []string

	if l == 2 {
		params = append(params, args[0].(string))
		params = append(params, "=")
		params = append(params, singleQuotes(args[1]))
	}

	if l == 3 {
		operStr := strings.ToLower(formatString(args[1]))
		if !inSlice(operStr, operator) {
			panic("FormatParams method: where condition operator error")
		}
		params = append(params, args[0].(string))
		params = append(params, args[1].(string))

		switch operStr {
		case "in":
			params = append(params, "("+implode(args[2], ",")+")")
		case "not in":
			params = append(params, "("+implode(args[2], ",")+")")
		case "like":
			params = append(params, singleQuotes("%"+formatString(args[2])+"%"))
		default:
			params = append(params, singleQuotes(args[2]))
		}
	}

	return strings.Join(params, " ")
}

func (statement *Statement) buildSelect(args ...bool) (string, []interface{}) {
	if statement.alias == "" && statement.join != "" {
		statement.adapter.logger.Errorf("%s alias is empty", statement.TableName)
		log.Panicln(statement.TableName + " alias is empty")
	}
	sql := ""
	cond, params := statement.prepareWhere()
	if len(args) == 0 {
		sql = fmt.Sprintf(
			"SELECT %v FROM %s%v%v%v%v%v", statement.parseField(), statement.parseTableName(), statement.join, cond, statement.groupBy, statement.orderBy, statement.limit,
		)
	} else {
		sql = fmt.Sprintf(
			"SELECT %v FROM %v%v%v%v%v LIMIT ?", statement.parseField(), statement.parseTableName(), statement.join, cond, statement.groupBy, statement.orderBy,
		)
	}
	return strings.ToLower(sql), params
}

func (statement *Statement) buildCount() (string, []interface{}) {
	if statement.alias == "" && statement.join != "" {
		statement.adapter.logger.Errorf("%s alias is empty", statement.TableName)
		panic(statement.TableName + " alias is empty")
	}

	sql := ""
	cond, params := statement.prepareWhere()

	if statement.distinct == "" {
		sql = fmt.Sprintf(
			"SELECT COUNT(*) AS aggregate FROM %v%v%v%v%v%v", statement.TableName, statement.join, cond, statement.groupBy, statement.orderBy, statement.limit,
		)
	} else {
		sql = fmt.Sprintf(
			"SELECT COUNT(%s) AS aggregate FROM %v%v%v%v%v%v", statement.distinct, statement.TableName, statement.join, cond, statement.groupBy, statement.orderBy, statement.limit,
		)
	}
	return strings.ToLower(sql), params
}

func (statement *Statement) buildInsert(args interface{}) (string, error) {
	fields := make([]string, 0)
	values := make([]string, 0)

	t := reflect.ValueOf(args).Kind().String()

	if t == "ptr" {
		v := reflect.ValueOf(args).Elem()
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			setKey := t.Field(i).Tag.Get("json")
			setVal := v.Field(i)
			if setKey == "" && f.Kind() != reflect.Struct {
				setKey = FormatUpper(t.Field(i).Name)
			}
			if statement.pk != "" {
				if setKey == statement.pk {
					continue
				}
			}

			switch f.Kind() {
			case reflect.Invalid:
				continue
			case reflect.Int, reflect.Int32, reflect.Int64:
				values = append(values, fmt.Sprintf("%v", setVal))
			case reflect.String:
				values = append(values, fmt.Sprintf("%v", escapeString(setVal.String())))
			case reflect.Float32, reflect.Float64:
				values = append(values, strconv.FormatFloat(setVal.Float(), 'f', -1, 64))
			case reflect.Struct, reflect.Ptr, reflect.Map:
				values = append(values, escapeString(json_encode(setVal.Interface())))
			case reflect.Interface:
				if v.IsNil() {
					continue
				} else {
					values = append(values, fmt.Sprintf("%v", v.Interface()))
				}
			}

			fields = append(fields, setKey)
		}
	} else {
		insertData, ok := args.(map[string]interface{})
		if !ok {
			return "", errors.New(PARAMETER_ERROR)
		}
		if statement.pk != "" {
			delete(insertData, statement.pk)
		}
		for key, value := range insertData {
			v := reflect.ValueOf(value)
			switch v.Kind() {
			case reflect.Invalid:
				continue
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				values = append(values, strconv.FormatInt(v.Int(), 10))
			case reflect.String:
				values = append(values, escapeString(v.String()))
			case reflect.Float32:
				values = append(values, strconv.FormatFloat(v.Float(), 'f', -1, 32))
			case reflect.Float64:
				values = append(values, strconv.FormatFloat(v.Float(), 'f', -1, 64))
			case reflect.Struct, reflect.Ptr, reflect.Map:
				b, err := json.Marshal(v.Interface())
				if err != nil {
					values = append(values, "")
				} else {
					values = append(values, string(b))
				}
			case reflect.Interface:
				if v.IsNil() {
					continue
				} else {
					values = append(values, fmt.Sprintf("%v", v.Interface()))
				}
			}
			fields = append(fields, key)
		}
	}

	return fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)", statement.TableName, "`"+strings.Join(fields, "`,`")+"`", "'"+strings.Join(values, "','")+"'",
	), nil
}

func (statement *Statement) buildMultiInsert(args interface{}) (string, error) {
	tmp := make([]map[string]string, 0)

	t := reflect.TypeOf(args).Elem().Kind().String()
	v := reflect.ValueOf(args)

	if t == "ptr" {
		for l := 0; l < v.Len(); l++ {
			vv := v.Index(l).Elem()
			tt := vv.Type()
			val := ""
			m := make(map[string]string)
			for i := 0; i < vv.NumField(); i++ {
				f := vv.Field(i)
				setKey := tt.Field(i).Tag.Get("json")
				setVal := vv.Field(i)
				if setKey == "" && f.Kind() != reflect.Struct {
					setKey = FormatUpper(tt.Field(i).Name)
				}
				if statement.pk != "" {
					if setKey == statement.pk {
						continue
					}
				}

				switch f.Kind() {
				case reflect.Invalid:
					val = ""
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					val = strconv.FormatInt(setVal.Int(), 10)
				case reflect.String:
					val = escapeString(setVal.String())
				case reflect.Float32:
					val = strconv.FormatFloat(setVal.Float(), 'f', -1, 32)
				case reflect.Float64:
					val = strconv.FormatFloat(setVal.Float(), 'f', -1, 64)
				case reflect.Struct, reflect.Ptr, reflect.Map:
					b, err := json.Marshal(setVal.Interface())
					if err != nil {
						val = ""
					} else {
						val = string(b)
					}
				case reflect.Interface:
					if v.IsNil() {
						val = ""
					} else {
						val = fmt.Sprintf("%v", setVal.Interface())
					}
				}
				m[setKey] = val
			}
			tmp = append(tmp, m)
		}

	} else {
		var ok bool
		var data []map[string]interface{}

		if data, ok = args.([]map[string]interface{}); !ok {
			return "", errors.New(PARAMETER_ERROR)
		}

		for _, d := range data {
			if statement.pk != "" {
				delete(d, statement.pk)
			}
			val := ""
			m := make(map[string]string)
			for key, value := range d {
				v := reflect.ValueOf(value)
				switch v.Kind() {
				case reflect.Invalid:
					val = ""
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					val = strconv.FormatInt(v.Int(), 10)
				case reflect.String:
					val = escapeString(v.String())
				case reflect.Float32:
					val = strconv.FormatFloat(v.Float(), 'f', -1, 32)
				case reflect.Float64:
					val = strconv.FormatFloat(v.Float(), 'f', -1, 64)
				case reflect.Struct, reflect.Ptr, reflect.Map:
					b, err := json.Marshal(v.Interface())
					if err != nil {
						val = ""
					} else {
						val = string(b)
					}
				case reflect.Interface:
					if v.IsNil() {
						val = ""
					} else {
						val = fmt.Sprintf("%v", v.Interface())
					}
				}
				m[key] = val
			}
			tmp = append(tmp, m)
		}
	}

	var fields []string
	for k, _ := range tmp[0] {
		fields = append(fields, k)
	}

	values := make([][]string, len(tmp))
	for _, v := range fields {
		for i, vv := range tmp {
			values[i] = append(values[i], vv[v])
		}
	}

	vtmp := make([]string, 0)
	for _, v := range values {
		vtmp = append(vtmp, fmt.Sprintf("('%s')", strings.Join(v, "','")))
	}

	sql := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES %s", statement.TableName, "`"+strings.Join(fields, "`,`")+"`", strings.Join(vtmp, ","),
	)

	return sql, nil
}

func (statement *Statement) buildUpdate(args interface{}) (string, error) {
	var values []string

	argsType := reflect.ValueOf(args).Kind().String()
	/*if !inSlice(argsType, []string{"map", "ptr"}) {
		return "", errors.New(PARAMETER_ERROR)
	}*/
	if !inSlice(argsType, []string{"map"}) {
		return "", errors.New(PARAMETER_ERROR)
	}

	cond := statement.formatWhere()
	if cond == "" {
		return "", errors.New(WHERE_ERROR)
	}

	if argsType == "ptr" {
		v := reflect.ValueOf(args).Elem()
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			setKey := t.Field(i).Tag.Get("json")
			setVal := v.Field(i)
			if setKey == "" && f.Kind() != reflect.Struct {
				setKey = FormatUpper(t.Field(i).Name)
			}
			if statement.pk != "" {
				if setKey == statement.pk {
					continue
				}
			}
			switch f.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64:
				setParam := fmt.Sprintf("%v = %v", setKey, setVal)
				values = append(values, setParam)
			case reflect.String:
				setParam := fmt.Sprintf("%v = '%v'", setKey, escapeString(setVal.String()))
				values = append(values, setParam)
			case reflect.Float32, reflect.Float64:
				setParam := fmt.Sprintf("%v = '%v'", setKey, strconv.FormatFloat(setVal.Float(), 'f', -1, 64))
				values = append(values, setParam)
			case reflect.Ptr:
				setParam := fmt.Sprintf("%v = '%v'", setKey, escapeString(json_encode(setVal.Interface())))
				values = append(values, setParam)
			}

		}
	} else {
		insertData, ok := args.(map[string]interface{})
		if !ok {
			return "", errors.New(PARAMETER_ERROR)
		}
		if statement.pk != "" {
			delete(insertData, statement.pk)
		}

		for key, value := range insertData {
			switch value.(type) {
			case int, int64, int32:
				setParam := fmt.Sprintf("%v = %v", key, value)
				values = append(values, setParam)
			case string:
				setParam := fmt.Sprintf("%v = '%v'", key, escapeString(value.(string)))
				values = append(values, setParam)
			case float32, float64:
				setParam := fmt.Sprintf("%v = '%v'", key, strconv.FormatFloat(value.(float64), 'f', -1, 64))
				values = append(values, setParam)
			case map[string]interface{}:
				setParam := fmt.Sprintf("%v = '%v'", key, escapeString(json_encode(value)))
				values = append(values, setParam)
			}
		}
	}

	return fmt.Sprintf(
		"UPDATE `%s` SET %s%s", statement.TableName, strings.Join(values, ","), cond,
	), nil
}

func (statement *Statement) Trace() (file string, line int, function string) {
	pc, file, line, _ := runtime.Caller(4)
	fname := runtime.FuncForPC(pc)
	return file, line, fname.Name()
}

func (statement *Statement) __function__(l ...int) string {
	n := 0
	if len(l) > 0 {
		n = l[0]
	}
	pc, _, _, _ := runtime.Caller(n)
	name := runtime.FuncForPC(pc).Name()
	list := strings.Split(name, ".")
	if len(list) == 0 {
		return ""
	}
	i := len(list) - 1

	return list[i]
}

func (statement *Statement) Error(msg string, flag ...bool) {
	exit := false

	if len(flag) > 0 {
		if flag[0] {
			exit = true
		}
	}

	execName := statement.__function__(2)

	defer func() {
		if exit {
			log.Panicf("%s method: %s", execName, msg)
		}
	}()

	pc, file, line, _ := runtime.Caller(4)

	name := runtime.FuncForPC(pc).Name()

	statement.adapter.logger.Errorf("%s method: %s %s %d %s", execName, msg, file, line, name)

	return
}

func (statement *Statement) CustomError(msg string, f, c int, flag ...bool) {
	exit := false

	if len(flag) > 0 {
		if flag[0] {
			exit = true
		}
	}

	execName := statement.__function__(f)

	defer func() {
		if exit {
			log.Panicf("%s method: %s", execName, msg)
		}
	}()

	pc, file, line, _ := runtime.Caller(c)

	name := runtime.FuncForPC(pc).Name()

	statement.adapter.logger.Errorf("%s method: %s %s %d %s", execName, msg, file, line, name)

	return
}
