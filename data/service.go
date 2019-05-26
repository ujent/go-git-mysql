//Package data provides methodes to work with MySQL
package data

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type service struct {
	db *sql.DB
}

// New func creates an instance of db service
func New(connectionStr string) (svc Service, err error) {
	db, err := sql.Open("mysql", connectionStr)

	if err != nil {
		return nil, err
	}

	return &service{db: db}, nil
}
