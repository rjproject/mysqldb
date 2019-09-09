package mysqldb

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"
)

type Model struct {
	db           *sql.DB
	tx           *sql.Tx
	adapter      *Adapter
	statement    Statement
	isAutoCommit bool
	isExecuted   bool
}

func (model *Model) Init() {
	model.statement.Init()
	model.db = model.adapter.db
	model.statement.adapter = model.adapter
	model.isAutoCommit = true
	model.isExecuted = true
}

func (model *Model) reset() {
	model.statement.Init()
}

func (model *Model) Table(args string) *Model {
	model.statement.Table(args)
	return model
}

func (model *Model) As(args string) *Model {
	model.statement.As(args)
	return model
}

func (model *Model) Fields(args ...string) *Model {
	model.statement.fields = args
	return model
}

func (model *Model) WhereRaw(args string) *Model {
	model.statement.WhereRaw(args)
	return model
}

func (model *Model) Where(args ...interface{}) *Model {
	model.statement.Where(args...)
	return model
}

func (model *Model) WhereIn(field string, values interface{}) *Model {
	model.statement.WhereIn(field, values)
	return model
}

func (model *Model) Id(args interface{}) *Model {
	pk := "id"
	if model.statement.pk != "" {
		pk = model.statement.pk
	}
	argsType := reflect.ValueOf(args).Kind().String()
	if argsType == "slice" {
		model.statement.WhereIn(pk, args)
	} else {
		model.statement.Where(pk, args)
	}
	return model
}

func (model *Model) WhereNotIn(field string, values interface{}) *Model {
	model.statement.WhereNotIn(field, values)
	return model
}

func (model *Model) OrWhere(args ...interface{}) *Model {
	model.statement.OrWhere(args...)
	return model
}

func (model *Model) Limit(args ...int) *Model {
	model.statement.Limit(args...)
	return model
}

func (model *Model) SetPk(pk string) *Model {
	model.statement.SetPk(pk)
	return model
}

func (model *Model) GroupBy(args string) *Model {
	model.statement.GroupBy(args)
	return model
}

func (model *Model) OrderBy(args string) *Model {
	model.statement.OrderBy(args)
	return model
}

func (model *Model) LeftJoin(table, condition string) *Model {
	model.statement.LeftJoin(table, condition)
	return model
}

func (model *Model) RightJoin(table, condition string) *Model {
	model.statement.RightJoin(table, condition)
	return model
}

func (model *Model) Join(table, condition string) *Model {
	model.statement.Join(table, condition)
	return model
}

func (model *Model) FullJoin(table, condition string) *Model {
	model.statement.FullJoin(table, condition)
	return model
}

func (model *Model) Distinct(args string) *Model {
	model.statement.Distinct(args)
	return model
}

func (model *Model) Insert(args interface{}) (id int64, err error) {
	if t := reflect.ValueOf(args).Kind().String(); !inSlice(t, []string{"map", "ptr"}) {
		model.statement.CustomError(PARAMETER_ERROR, 2, 2)
		return 0, errors.New(PARAMETER_ERROR)
	}

	var result driver.Result

	sql, e := model.statement.buildInsert(args)
	if e != nil {
		return 0, e
	}

	result, err = model.exec(sql)
	if err != nil {
		return 0, errors.New("Insert error: " + err.Error())
	}

	i, err := result.LastInsertId()
	if err != nil {
		return 0, errors.New("Insert error: " + err.Error())
	}

	return i, err
}

func (model *Model) MultiInsert(args interface{}) (int64, error) {
	v := reflect.ValueOf(args)
	if v.Kind() != reflect.Slice {
		model.statement.CustomError(PARAMETER_ERROR, 2, 2)
		return 0, errors.New(PARAMETER_ERROR)
	}

	t := reflect.TypeOf(args).Elem().Kind().String()
	if t != "map" && t != "ptr" {
		model.statement.CustomError(PARAMETER_ERROR, 2, 2)
		return 0, errors.New(PARAMETER_ERROR)
	}

	if v.Len() == 0 {
		return 0, nil
	}

	var err error

	sql, err := model.statement.buildMultiInsert(args)
	if err != nil {
		return 0, err
	}

	var result driver.Result

	result, err = model.exec(sql)
	if err != nil {
		return 0, errors.New("Insert error: " + err.Error())
	}

	i, err := result.RowsAffected()
	if err != nil {
		return 0, errors.New("Insert error: " + err.Error())
	}

	return i, err
}

func (model *Model) Delete() (num int64, err error) {
	if model.statement.TableName == "" {
		return 0, errors.New(TABLENAME_ERROR)
	}

	if cond := model.statement.formatWhere(); cond != "" {
		result, err := model.exec(fmt.Sprintf("DELETE FROM `%s`%s", model.statement.TableName, cond))
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}

	return 0, errors.New(WHERE_ERROR)

}

func (model *Model) Update(args interface{}) (n int64, err error) {
	argsType := reflect.ValueOf(args).Kind().String()
	if !inSlice(argsType, []string{"map", "ptr"}) {
		model.statement.CustomError(PARAMETER_ERROR, 2, 2)
		return 0, errors.New(PARAMETER_ERROR)
	}

	sql, e := model.statement.buildUpdate(args)
	if e != nil {
		return 0, e
	}

	result, err := model.Exec(sql)
	if err != nil {
		return 0, errors.New("Update Error:" + err.Error())
	}
	return result.RowsAffected()
}

func (model *Model) First(i interface{}) error {
	val := reflect.Indirect(reflect.ValueOf(i))
	if val.Kind() != reflect.Struct {
		return errors.New(PARAMETER_ERROR)
	}
	if model.statement.TableName == "" {
		model.statement.TableName = FormatUpper(val.Type().Name())
	}
	model.statement.fields = ReflectFields(i)
	sql, params := model.statement.buildSelect(true)
	params = append(params, 1)
	list, err := model.Query(sql, params...)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return errors.New(NODATA_ERROR)
	}

	err = map2struct(i, list[0])
	if err != nil {
		return err
	}

	return nil
}

func (model *Model) Find(s interface{}) error {
	sliceValue := reflect.Indirect(reflect.ValueOf(s))
	if sliceValue.Kind() != reflect.Slice {
		return errors.New(SLICEPOINTER_ERROR)
	}

	iType := reflect.TypeOf(s).Elem().Elem().Elem()
	if iType.Kind() != reflect.Struct {
		return errors.New(PARAMETER_ERROR)
	}
	iFace := reflect.New(iType).Interface()

	if model.statement.TableName == "" {
		model.statement.TableName = FormatUpper(iType.Name())
	}
	model.statement.fields = ReflectFields(iFace)

	sql, params := model.statement.buildSelect()
	list, err := model.Query(sql, params...)
	if err != nil {
		return err
	}

	return slice2struct(s, list)
}

func (model *Model) Fetch() (map[string]interface{}, error) {
	sql, params := model.statement.buildSelect(true)
	params = append(params, 1)

	list, err := model.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	if len(list) == 1 {
		return list[0], nil
	}

	return nil, nil
}

func (model *Model) FetchAll() ([]map[string]interface{}, error) {
	sql, params := model.statement.buildSelect()
	return model.Query(sql, params...)
}

func (model *Model) Count() (int64, error) {
	sql, params := model.statement.buildCount()
	result, err := model.Query(sql, params...)
	if err != nil {
		return 0, err
	}
	return convertInt(result[0]["aggregate"])
}

func (model *Model) Query(sql string, args ...interface{}) ([]map[string]interface{}, error) {
	list := make([]map[string]interface{}, 0)
	rows, err := model.query(sql, args...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	return model.resultSet(rows), nil
}

func (model *Model) query(sql string, args ...interface{}) (*sql.Rows, error) {
	defer model.reset()
	if !model.isAutoCommit {
		start := time.Now().UnixNano()
		defer func() {
			model.showSQL(start, sql, args...)
		}()
		rows, err := model.tx.Query(sql, args...)
		if err != nil {
			return nil, err
		}
		return rows, nil
	}

	stmt, err := model.db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	start := time.Now().UnixNano()
	defer func() {
		model.showSQL(start, sql, args...)
	}()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (model *Model) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return model.exec(sql, args...)
}

func (model *Model) exec(sql string, args ...interface{}) (sql.Result, error) {
	defer model.reset()

	if !model.isAutoCommit {
		start := time.Now().UnixNano()
		defer func() {
			model.showSQL(start, sql, args...)
		}()
		return model.tx.Exec(sql, args...)
	}

	stmt, err := model.db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	start := time.Now().UnixNano()
	defer func() {
		model.showSQL(start, sql, args...)
	}()

	result, err := stmt.Exec(args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (model *Model) resultSet(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	count := len(columns)

	result := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)

	scanArgs := make([]interface{}, count)

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		rows.Scan(scanArgs...)
		entry := make(map[string]interface{})
		for i, col := range values {
			if isUint8Slice(col) {
				entry[columns[i]] = string(col.([]byte))
			} else {
				entry[columns[i]] = col
			}
		}
		result = append(result, entry)
	}
	return result
}

func (model *Model) showSQL(start int64, sql string, args ...interface{}) {
	end := time.Now().UnixNano()
	if model.adapter.isLog {
		model.adapter.logger.Debugf("sql# %5dms: %s bind:%v", int64(float64(end-start)/math.Pow(2, 20)), strings.ToLower(sql), args)
	}
}
