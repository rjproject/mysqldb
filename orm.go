package mysqldb

import (
	"database/sql"
	"os"

	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Options struct {
	User         string
	Password     string
	Host         string
	Port         int
	Database     string
	Charset      string
	MaxIdleConns int
	MaxOpenConns int
	Debug        bool
}

func New(options *Options) (*Adapter, error) {
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s",
		options.User,
		options.Password,
		options.Host,
		options.Port,
		options.Database,
		options.Charset,
	))
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(options.MaxIdleConns)
	db.SetMaxOpenConns(options.MaxOpenConns)

	adapter := &Adapter{
		db:    db,
		isLog: false,
	}

	logger := InitLogger(os.Stdout)

	logger.SetLevel(LOG_DEBUG)

	adapter.SetLogger(logger)
	adapter.Debug(options.Debug)

	return adapter, nil
}
