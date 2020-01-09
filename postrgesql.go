package main

var templatesPostrgeSQL = `// Code generated by gcgSQL. DO NOT EDIT. 
package {{.Package}}

import (
	"context"
	"database/sql"
)
{{range $k, $v := .Queries}}{{if gt (len .ReturnType) 0}}
type {{$k}}Struct struct {
{{range $k1, $v1 := .ReturnType}}	{{index $v.ReturnName $k1}} {{$v1}}
{{end}}
}{{end}}

func {{$k}}(tx *sql.Tx{{range $k2, $v2 := $v.Arguments}},{{if eq (index $v.ArgumentsName $k2) ""}}v{{$k2}}{{else}}{{index $v.ArgumentsName $k2}}{{end}} {{$v2}}{{end}}) ({{if gt (len .ReturnType) 0}}resultRows []*{{$k}}Struct, {{else if eq .QueryType "insert"}}LastInsertedId int64,{{end}}err error) {
	return {{$k}}Context(tx, context.Background(){{range $k2, $v2 := $v.Arguments}},{{if eq (index $v.ArgumentsName $k2) ""}}v{{$k2}}{{else}}{{index $v.ArgumentsName $k2}}{{end}}{{end}})
}

func {{$k}}Context(tx *sql.Tx, ctx context.Context{{range $k2, $v2 := $v.Arguments}},{{if eq (index $v.ArgumentsName $k2) ""}}v{{$k2}}{{else}}{{index $v.ArgumentsName $k2}}{{end}} {{$v2}}{{end}}) ({{if gt (len .ReturnType) 0}}resultRows []*{{$k}}Struct,{{end}} err error) {
	{{if gt (len .ReturnType) 0}}res, err := tx.QueryContext(ctx, "{{.Query| Escape}}"{{PrintCallParams .ArgumentsName "" true true}})
	if err != nil {
		return nil, err
	}
	defer res.Close()
	ret{{$k}}Struct := make([]*{{$k}}Struct, 0)
	for res.Next() {
		retStructRow := {{$k}}Struct{}
		err = res.Scan({{PrintCallParams .ReturnName "&retStructRow." false true}})
		if err != nil {
			return nil, err
		}
		ret{{$k}}Struct = append(ret{{$k}}Struct, &retStructRow)
	}
	return ret{{$k}}Struct, nil 
	{{else}}res, err := tx.ExecContext(ctx, "{{.Query| Escape}}"{{PrintCallParams .ArgumentsName "" true true}})
	if err != nil {
		return err
	}
	return nil

	{{end}}
}
{{end}}
`