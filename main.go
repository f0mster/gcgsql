package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type data struct {
	Arguments  []string
	ReturnType []string
	ReturnName []string
	Query      string
}

var dbType = flag.String("dbtype", "mysql", "package source directory, useful for vendored code")
var dbString = flag.String("dbconn", "", "db connection string e.g. root:@tcp(localhost:3306)/dbname")
var yamlPath = flag.String("yaml", "", "path to yaml")
var outputFileName = flag.String("output", "", "where should i put created file")
var conn *sql.DB

func connect() *sql.DB {
	if conn != nil {
		return conn
	}
	var err error
	conn, err = sql.Open(*dbType, *dbString)
	if err != nil {
		os.Exit(1)
	}
	return conn
}

func main() {
	/*	flag.

	 */
	flag.Parse()
	arguments := map[string]*data{}

	querys := map[string]string{"getUsers": "select * from users where id=?int", "getGroups": "select * from groups where id=?int"}
	if b, e := ioutil.ReadFile(*yamlPath); e != nil {
		fmt.Println(e)
		return
	} else {
		if e := yaml.Unmarshal(b, &querys); e != nil {
			fmt.Println(e)
			return
		}
	}

	for funcName, query := range querys {
		arguments[funcName] = &data{Arguments: make([]string, 0)}
		argumentsRE := regexp.MustCompile(`\?([^\s]*)(\b|\s|$)`)
		result := argumentsRE.FindAllSubmatch([]byte(query), -1)
		arguments[funcName].Query = argumentsRE.ReplaceAllString(query, "?")
		sqlargs := make([]string, 0)
		selectArgs := []interface{}{}
		for _, val := range result {
			if string(val[1]) != "int" && string(val[1]) != "string" && string(val[1]) != "int64" && string(val[1]) != "uint" {
				fmt.Println("wrong argument \"?" + string(val[1]) + "\"")
				os.Exit(1)
			}
			sqlargs = append(sqlargs, string(val[1]))
			if string(val[1]) == "string" {
				selectArgs = append(selectArgs, "")
			} else {
				selectArgs = append(selectArgs, 0)
			}
		}
		arguments[funcName].Arguments = sqlargs
		if strings.Index(strings.TrimSpace(query), "select") >= 0 || strings.Index(strings.TrimSpace(query), "show") >= 0 {
			res, err := connect().Query(arguments[funcName].Query, selectArgs ...)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(res.Columns())
			tmp, err := res.ColumnTypes()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			arguments[funcName].ReturnName = make([]string,len(tmp))
			arguments[funcName].ReturnType = make([]string,len(tmp))
			for i := range tmp {
				arguments[funcName].ReturnType[i] = tmp[i].ScanType().String()
				arguments[funcName].ReturnName[i] = tmp[i].Name()
			}
		} else if strings.Index(strings.TrimSpace(query), "insert") < 0 &&
			strings.Index(strings.TrimSpace(query), "update") < 0 &&
			strings.Index(strings.TrimSpace(query), "delete") < 0 {
			fmt.Println("unsupported query type", query)
			os.Exit(1)
		}

	}

	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"Escape":  func(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\"`), `"`, `\"`) },
		"Title":   strings.Title,
		"ToLower": strings.ToLower,
		"PrintCallParams": func(params []string, prefix string, firstComma bool, useNamesFromParams bool) string {
			first := 0
			if firstComma {
				first = 1
			}
			out := ""
			for i := range params {
				if first == 0 {
					first = 1
				} else {
					out += ", "
				}
				if useNamesFromParams {
					out += prefix + params[i]
				} else {
					out += prefix + "v" + strconv.Itoa(i)
				}
			}
			return out
		},
	}

	tmpl := ""
	switch *dbType {
	case "mysql":
		tmpl = templatesMysql
	}

	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		log.Print(err)
		return
	}
	f, err := os.Create(*outputFileName)
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = t.Execute(f, arguments)
	if err != nil {
		log.Print("execute: ", err)
		return
	}
}
