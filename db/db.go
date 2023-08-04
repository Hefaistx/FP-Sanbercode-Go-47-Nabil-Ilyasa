package db

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
)

var Db *sql.DB

func init() {
	// DB Connection
	var err error
	Db, err = sql.Open("mysql", "root:@tcp(localhost:3306)/finalprojectdb")
	if err != nil {
		panic(err)
	}

	err = Db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to MySQL database!")
}

func RootHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Connected to MySQL database!")
}
