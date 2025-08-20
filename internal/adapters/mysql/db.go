package mysql

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func Open(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	return db, db.Ping()
}
