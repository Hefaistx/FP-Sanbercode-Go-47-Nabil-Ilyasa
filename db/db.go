package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func init() {
	// DB Connection
	var err error
	Db, err = sql.Open("mysql", "root:root123@tcp(34.128.105.170)/finalprojectdb")
	if err != nil {
		panic(err)
	}

	err = Db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MySQL database!")
}
