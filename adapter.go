package mysqldb

import (
	"database/sql"
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

type Adapter struct {
	db     *sql.DB
	logger iLogger
	isLog  bool
}

func (adapter *Adapter) Debug(flag ...bool) {
	if len(flag) == 0 {
		adapter.isLog = true
	} else {
		adapter.isLog = flag[0]
	}
}

func (adapter *Adapter) SetLogLevel(level int) {
	adapter.logger.SetLevel(level)
}

func (adapter *Adapter) SetLogger(logger iLogger) {
	adapter.logger = logger
}

func (adapter *Adapter) NewModel() *Model {
	entity := &Model{adapter: adapter}
	entity.Init()
	return entity
}

func (adapter *Adapter) Table(args string) *Model {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	return entity.Table(args)
}

func (adapter *Adapter) Id(args interface{}) *Model {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	return entity.Id(args)
}

func (adapter *Adapter) T(args string) *Model {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	return entity.Table(args)
}

func (adapter *Adapter) Where(args ...interface{}) *Model {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	return entity.Where(args...)
}

func (adapter *Adapter) WhereIn(field string, values interface{}) *Model {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	return entity.WhereIn(field, values)
}

func (adapter *Adapter) In(field string, values interface{}) *Model {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	return entity.WhereIn(field, values)
}

func (adapter *Adapter) WhereRaw(args string) *Model {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	return entity.WhereRaw(args)
}

func (adapter *Adapter) WhereNotIn(field string, values interface{}) *Model {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	return entity.WhereNotIn(field, values)
}

func (adapter *Adapter) NotIn(field string, values interface{}) *Model {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	return entity.WhereNotIn(field, values)
}

func (adapter *Adapter) Query(sql string, args ...interface{}) ([]map[string]interface{}, error) {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	return entity.Query(sql, args...)
}

func (adapter *Adapter) Exec(sql string, args ...interface{}) (int64, error) {
	entity := adapter.NewModel()
	entity.isAutoCommit = true
	result, err := entity.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (adapter *Adapter) Scan(sour interface{}, dest interface{}) error {
	s := reflect.Indirect(reflect.ValueOf(sour))
	d := reflect.Indirect(reflect.ValueOf(dest))

	if s.Kind() != reflect.Map && s.Kind() != reflect.Slice {
		return errors.New("First parameter must be a map or map slice.")
	}

	if s.Kind() == reflect.Map && d.Kind() != reflect.Ptr {
		return errors.New("First parameter is a map, second parameter must be a struct pointer.")
	}

	if s.Kind() == reflect.Slice && d.Kind() != reflect.Slice {
		return errors.New("First parameter is a map slice, second parameter must be a struct pointer slice.")
	}

	b, err := json.Marshal(sour)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func (adapter *Adapter) DB() *sql.DB {
	return adapter.db
}

func (adapter *Adapter) Close() error {
	return adapter.db.Close()
}

func (adapter *Adapter) SetMaxIdleConns(n int) {
	adapter.db.SetMaxIdleConns(n)
}

func (adapter *Adapter) SetMaxOpenConns(n int) {
	adapter.db.SetMaxOpenConns(n)
}
