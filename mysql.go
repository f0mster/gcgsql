package main

var templatesMysql = `package sqlQuery

import "database/sql"
{{range $k, $v := .}}{{if gt (len .ReturnType) 0}}
type {{$k}}Struct struct {
{{range $k1, $v1 := .ReturnType}}	{{index $v.ReturnName $k1}} {{$v1}}
{{end}}
}{{end}}

func {{$k}}(tx sql.Tx, {{range $k2, $v2 := $v.Arguments}}v{{$k2}} {{$v2}}{{end}}) ({{if gt (len .ReturnType) 0}}[]*{{$k}}Struct, {{end}}error) {
	{{if gt (len .ReturnType) 0}}res, err := tx.Query("{{.Query| Escape}}"{{PrintCallParams .Arguments "" true false}})
	if err != nil {
		return nil, err
	}
	ret{{$k}}Struct := make([]*{{$k}}Struct, 0)
	for res.Next() {
		retStructRow := {{$k}}Struct{}
		err = res.Scan({{PrintCallParams .ReturnName "&retStructRow." false true}})
		if err != nil {
			return nil, err
		}
		ret{{$k}}Struct = append(ret{{$k}}Struct, &retStructRow)
	}
	res.Close()
	return ret{{$k}}Struct, nil 
	{{else}}res, err := tx.Exec("{{.Query| Escape}}"){{end}}
}
{{end}}
`
