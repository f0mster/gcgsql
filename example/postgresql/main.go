package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

func main() {
	connstr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", "postgres", "", "localhost", 5432, "test")
	conn, err := sql.Open("postgres", connstr)
	if err != nil {
		fmt.Println(conn)
		os.Exit(1)
	}
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//res, err := getAllUsers(context.Background(), tx)
	res, err := findUsersIdInOrByLoginAndName(context.Background(), tx, []int{1, 2, 3, 4, 6}, []lnStruct{{name: "1", login: "1"}, {login: "2", name: "2"}})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i := range res {
		fmt.Println(res[i])
	}

}

type sometype struct {
	id   int
	name string
}
