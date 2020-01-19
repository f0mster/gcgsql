package main

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

func generate(writer io.Writer, yamlData yamlFile) error {
	var tmpl = templatesMysql
	switch yamlData.SqlDriver {
	case "github.com/go-sql-driver/mysql":
	case "github.com/lib/pq":
		supportReturning = true
		tmpl = templatesPostrgeSQL
	default:
		return fmt.Errorf("unsupported db type")
	}

	t, e := template.New("").Funcs(funcMap).Parse(tmpl)
	if e != nil {
		return e
	}
	//	t, e = t.New("").Parse()
	e = t.Execute(writer, yamlData)
	return e
}

var funcMap = template.FuncMap{
	"Escape": func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\"`), `"`, `\"`)
	},
	"Title":   strings.Title,
	"Replace": strings.Replace,
	"MakeMapStringBool": func() map[string]bool {
		return map[string]bool{}
	},
	"SetMapStringBoolValue": func(data map[string]bool, key string, value bool) bool {
		data[key] = value
		return true
	},
}

func printNamesAndTypes(elprefix, elname, eltype, separator string, first *bool, withType bool) string {
	out := ""
	if *first {
		*first = false
	} else {
		out += separator
	}
	out += elprefix + elname
	if withType {
		out += " " + eltype
	}
	return out
}
