package sqlstore

import "database/sql"

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=StoreInterface
type StoreInterface interface {
	API() ApiRepositoryInterface
}

type DBInterface interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}
