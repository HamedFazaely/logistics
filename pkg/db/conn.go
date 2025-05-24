package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"gitlab.com/Hamed1984/logistics/pkg/conf"
)

func NewMysqlConnection(c *conf.Configuration) (*sql.DB, error) {
	db, err := sql.Open("mysql", c.GetDBDSN())
	if err != nil {
		return nil, err
	}
	return db, err
}
