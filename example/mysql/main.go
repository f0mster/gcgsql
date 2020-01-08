package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

func main() {
	conn, err := sql.Open("mysql", "root:@tcp(localhost:3306)/consys2")
	if err != nil {
		fmt.Println(conn)
		os.Exit(1)
	}
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	res, err := getAllUsers(tx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i := range res {
		fmt.Println(res[i])
	}

}
