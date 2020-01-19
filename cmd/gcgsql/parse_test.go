package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_findArgs(t *testing.T) {
	tests := []struct {
		name             string
		query            string
		supportReturning bool
		// response
		wantOutQuery            string
		wantArgs                slicePA
		wantParams              slicePP
		wantErr                 error
		wantHaveRepeatableParts bool
	}{
		{
			name:    "negative. missing close brackets",
			query:   "select if(a>b, a, b from users",
			wantErr: fmt.Errorf("found unpaired bracket ("),
		},
		{
			name:    "negative. missing open brackets",
			query:   "select if(a>b, a, b)) from users where id='1'",
			wantErr: fmt.Errorf("found closing bracket ) but no opening found near select if(a>b, a, b))"),
		},

		{
			name:    "negative. wrong closing bracket",
			query:   "select if(a>b, a], b from users",
			wantErr: fmt.Errorf("wrong bracket! excpected [ but found ( at right of select if(a>b, a"),
		},

		{
			name:    "negative. senseless query inside brackets",
			query:   "select if(a>b, a, a:b) a from users",
			wantErr: fmt.Errorf("use should use proper columnName:type pair for column! only alpha numerics are allowed! near: select if(a>b, a, a:b) a "),
		},

		{
			name:         "positive. senseless query inside brackets",
			query:        "select if(a>b, a, a:b) a:int from users",
			wantOutQuery: "select if(a>b, a, a:b) a from users",
			wantParams: slicePP{
				&parsedParam{ParamName: "a", ParamType: "int"},
			},
			wantArgs: slicePA{},
		},

		{
			name:    "$ token without name",
			query:   "wdqwq $ $aaas",
			wantErr: fmt.Errorf("no argument name right from wdqwq $"),
		},
		{
			name:    "bad symbol afrer $",
			query:   "wdqwq $( $aaas",
			wantErr: fmt.Errorf("bad token ( after $ at right of wdqwq $"),
		},
		{
			name:    "string after $",
			query:   "wdqwq $aaas wefweffe wef wef",
			wantErr: fmt.Errorf("no argument type right from wdqwq $aaas"),
		},
		{
			name:    "no type",
			query:   "wdqwq $aaas: wefweffe wef wef",
			wantErr: fmt.Errorf("no argument type right from wdqwq $aaas:"),
		},
		{
			name:    "int instead of type",
			query:   "wdqwq $aaas:23 wefweffe wef wef",
			wantErr: fmt.Errorf("no argument type right from wdqwq $aaas:"),
		},

		{
			name:         "positive string after $",
			query:        "wdqwq $aaas:int wefweffe wef wef 2",
			wantOutQuery: "wdqwq $0$ wefweffe wef wef 2",
			wantArgs:     slicePA{{ArgName: "aaas", ArgType: "int", PlaceHolder: "$0$"}},
			wantParams:   slicePP{},
		},

		{
			name:         "positive string after $, outtype $",
			query:        "wdqwq $aaas:int wefweffe wef wef",
			wantOutQuery: "wdqwq $0$ wefweffe wef wef",
			wantArgs:     slicePA{{ArgName: "aaas", ArgType: "int", PlaceHolder: "$0$"}},
			wantParams:   slicePP{},
		},

		{
			name:    "negative. two arguments with same name and different type",
			query:   "select id:int from messages where from=$userId:int or to=$userId:sql.NullString",
			wantErr: fmt.Errorf("argument with name \"userId\" have different types! int and sql.NullString"),
		},
		{
			name:    "negative. no params",
			query:   "select * from messages where from=$userId:int or to=$userId:string",
			wantErr: fmt.Errorf("use should use proper columnName:type pair for column! only alpha numerics are allowed! near: select * "),
		},
		{
			name:         "pasitive. with param",
			query:        "select `from`:int, to:int from messages where from=$userId:int or to=$userId:int",
			wantOutQuery: "select `from`, to from messages where from=$0$ or to=$1$",
			wantArgs: slicePA{
				{ArgType: "int", ArgName: "userId", PlaceHolder: "$0$"},
				{ArgType: "int", ArgName: "userId", PlaceHolder: "$1$"},
			},
			wantParams: slicePP{
				&parsedParam{ParamName: "`from`", ParamType: "int"},
				&parsedParam{ParamName: "to", ParamType: "int"},
			},
		},

		{
			name:         "positive string after $, outtype $ same arg name. first time with type, second without",
			query:        "select id:int from messages where from=$userId:int or to=$userId",
			wantOutQuery: "select id from messages where from=$0$ or to=$1$",
			wantArgs: slicePA{
				{ArgName: "userId", ArgType: "int", PlaceHolder: "$0$"},
				{ArgName: "userId", ArgType: "int", PlaceHolder: "$1$"},
			},
			wantParams: slicePP{&parsedParam{ParamName: "id", ParamType: "int"}},
		},
		{
			name:         "positive string after $, outtype $ same arg name. first time without type, second with",
			query:        "select id:int from messages where from=$userId or to=$userId:int",
			wantOutQuery: "select id from messages where from=$0$ or to=$1$",
			wantArgs: slicePA{
				{ArgName: "userId", ArgType: "int", PlaceHolder: "$0$"},
				{ArgName: "userId", ArgType: "int", PlaceHolder: "$1$"},
			},
			wantParams: slicePP{&parsedParam{ParamName: "id", ParamType: "int"}},
		},
		{
			name:         "positive string after $, outtype $ same arg name. first time without type, second with. third with",
			query:        "select id:int from messages where (from=$userId or to=$userId:int) and date>$date:string",
			wantOutQuery: "select id from messages where (from=$0$ or to=$1$) and date>$2$",
			wantArgs: slicePA{
				{ArgName: "userId", ArgType: "int", PlaceHolder: "$0$"},
				{ArgName: "userId", ArgType: "int", PlaceHolder: "$1$"},
				{ArgName: "date", ArgType: "string", PlaceHolder: "$2$"},
			},
			wantParams: slicePP{&parsedParam{ParamName: "id", ParamType: "int"}},
		},

		{
			name:         "positive no arguments",
			query:        "select id:int from messages",
			wantArgs:     slicePA{},
			wantOutQuery: "select id from messages",
			wantParams:   slicePP{&parsedParam{ParamName: "id", ParamType: "int"}},
		},
		{
			name:         "positive no arguments. parameter type with namespace",
			query:        "select id:sql.NullInt, name:string from messages",
			wantArgs:     slicePA{},
			wantOutQuery: "select id, name from messages",
			wantParams: slicePP{
				&parsedParam{ParamName: "id", ParamType: "sql.NullInt"},
				&parsedParam{ParamName: "name", ParamType: "string"},
			},
		},

		{
			name:             "positive. checking returning arguments in delete",
			query:            "delete from messages returning id:int, name:string",
			supportReturning: true,
			wantArgs:         slicePA{},
			wantOutQuery:     "delete from messages returning id, name",
			wantParams: slicePP{
				&parsedParam{ParamName: "id", ParamType: "int"},
				&parsedParam{ParamName: "name", ParamType: "string"},
			},
		},
		{
			name:         "bad sql behavior. checking returning arguments in delete with supportReturning disabled",
			query:        "delete from messages returning id:int, name:string",
			wantArgs:     slicePA{},
			wantOutQuery: "delete from messages returning id:int, name:string",
			wantParams:   slicePP{},
		},

		{
			name:    "negative. $after$",
			query:   "delete from messages where $name:string$aaam:ewew",
			wantErr: fmt.Errorf("syntax error near: delete from messages where $name:string$"),
		},

		{
			name:    "negative. error in argument type with package",
			query:   "delete from messages where $name:sql.",
			wantErr: fmt.Errorf("not finished type near delete from messages where $name:sql."),
		},

		{
			name:         "positive. argument type with package",
			query:        "delete from messages where $name:sql.NullString",
			wantArgs:     slicePA{{ArgName: "name", ArgType: "sql.NullString", PlaceHolder: "$0$"}},
			wantOutQuery: "delete from messages where $0$",
			wantParams:   slicePP{},
		},
		{
			name:         "positive. with first arg is repeatable, second - not repeatable",
			query:        "delete from messages where id in ($id:[]int) and $name:sql.NullString",
			wantOutQuery: "delete from messages where id in ($0$) and $1$",
			wantArgs: slicePA{
				{ArgName: "id", ArgType: "int", PlaceHolder: "$0$", Separator: ",", Repeatable: true},
				{ArgName: "name", ArgType: "sql.NullString", PlaceHolder: "$1$"},
			},
			wantParams:              slicePP{},
			wantHaveRepeatableParts: true,
		},
		{
			name:         "positive. placeholder test",
			query:        "delete from users where name=$name:string and login like '$1$' and email=$email:string ",
			wantOutQuery: "delete from users where name=$0$ and login like '$1$' and email=$2$ ",
			wantArgs: slicePA{
				{ArgName: "name", ArgType: "string", PlaceHolder: "$0$"},
				{ArgName: "email", ArgType: "string", PlaceHolder: "$2$"},
			},
			wantParams: slicePP{},
		},

		{
			name:    "negative. repeatable args, no ending #",
			query:   "delete from messages where $args#(from=$from:int and to=$to:int)",
			wantErr: fmt.Errorf("not finished type near delete from messages where $args#(from=$from:int and to=$to:int)"),
		},
		{
			name:    "negative. repeatable args, no ending # and no separator",
			query:   "delete from messages where $args#(from=$from:int and to=$to:int)",
			wantErr: fmt.Errorf("not finished type near delete from messages where $args#(from=$from:int and to=$to:int)"),
		},

		{
			name:    "negative. repeatable args, error inside repeatable part",
			query:   "delete from messages where $args#(from=$from:int and to=$from:string)#",
			wantErr: fmt.Errorf("error inside repeatable part. argument with name \"from\" have different types! int and string"),
		},

		{
			name:    "negative. repeatable args, have repeatable arg type and type inside repeatable part",
			query:   "delete from messages where $args:sometype#(from=$from:int and to=$to:string)#",
			wantErr: fmt.Errorf("error inside repeatable part. found argument type inside repeatable part, but repeatable already have type. near (from=$from:int"),
		},

		{
			name:         "positive. repeatable args",
			query:        "delete from messages where $args#(`from`=$from:int and to=$to:int)| or #",
			wantOutQuery: "delete from messages where $0$",
			wantArgs: slicePA{
				{
					ArgName:       "args",
					ArgType:       "argsStruct",
					PlaceHolder:   "$0$",
					RepeatedQuery: "(`from`=$0$ and to=$1$)",
					RepeatedArgs: slicePA{
						{ArgType: "int", ArgName: "from", PlaceHolder: "$0$"},
						{ArgType: "int", ArgName: "to", PlaceHolder: "$1$"},
					},
					Separator:       " or ",
					Repeatable:      true,
					IsGeneratedName: true,
				},
			},
			wantParams:              slicePP{},
			wantHaveRepeatableParts: true,
		},
		{
			name:         "positive. repeatable args and have repetable arg type",
			query:        "delete from messages where $args:myType#(`from`=$from and to=$to)| or #",
			wantOutQuery: "delete from messages where $0$",
			wantArgs: slicePA{
				{
					ArgName:       "args",
					PlaceHolder:   "$0$",
					RepeatedQuery: "(`from`=$0$ and to=$1$)",
					RepeatedArgs: slicePA{
						{ArgName: "from", PlaceHolder: "$0$"},
						{ArgName: "to", PlaceHolder: "$1$"},
					},
					Separator:  " or ",
					Repeatable: true,
					ArgType:    "myType",
				},
			},
			wantParams:              slicePP{},
			wantHaveRepeatableParts: true,
		},

		{
			name:         "positive. repeatable args, no separator",
			query:        "delete from messages where $args#(`from`=$from:int and to=$to:int)#",
			wantOutQuery: "delete from messages where $0$",
			wantArgs: slicePA{
				{
					ArgName:       "args",
					ArgType:       "argsStruct",
					PlaceHolder:   "$0$",
					RepeatedQuery: "(`from`=$0$ and to=$1$)",
					RepeatedArgs: slicePA{
						{ArgType: "int", ArgName: "from", PlaceHolder: "$0$"},
						{ArgType: "int", ArgName: "to", PlaceHolder: "$1$"},
					},
					Separator:       ",",
					Repeatable:      true,
					IsGeneratedName: true,
				},
			},
			wantParams:              slicePP{},
			wantHaveRepeatableParts: true,
		},

		{
			name:         "positive. repeatable args with type, no separator",
			query:        "delete from messages where $args:myType#(`from`=$from and to=$to)#",
			wantOutQuery: "delete from messages where $0$",
			wantArgs: slicePA{
				{
					ArgName:       "args",
					PlaceHolder:   "$0$",
					RepeatedQuery: "(`from`=$0$ and to=$1$)",
					RepeatedArgs: slicePA{
						{ArgName: "from", PlaceHolder: "$0$"},
						{ArgName: "to", PlaceHolder: "$1$"},
					},
					Separator:  ",",
					Repeatable: true,
					ArgType:    "myType",
				},
			},
			wantParams:              slicePP{},
			wantHaveRepeatableParts: true,
		},
		{
			name:    "negative. repeatable args with type, no separator",
			query:   "delete from messages where $args:[]myType#(`from`=$from:int and to=$to:int)#",
			wantErr: fmt.Errorf("found [] type before # in delete from messages where $args:[]myType#"),
		},
		{
			name:                    "positive. big query with 2 repeatable arguments and one non repeatable",
			query:                   "select id:int, name:string, info:string, login:string from users where id in ($userId:[]int) $ln# or (login=$login:string and name=$name:string)|# or admin=$isAdmin:int",
			wantOutQuery:            "select id, name, info, login from users where id in ($0$) $1$ or admin=$2$",
			wantHaveRepeatableParts: true,
			wantParams: slicePP{
				{ParamName: "id", ParamType: "int"},
				{ParamName: "name", ParamType: "string"},
				{ParamName: "info", ParamType: "string"},
				{ParamName: "login", ParamType: "string"},
			},
			wantArgs: slicePA{
				{
					ArgName:         "userId",
					ArgType:         "int",
					Repeatable:      true,
					Separator:       ",",
					RepeatedArgs:    nil,
					RepeatedQuery:   "",
					IsGeneratedName: false,
					PlaceHolder:     "$0$",
				},
				{
					ArgName:    "ln",
					ArgType:    "lnStruct",
					Repeatable: true,
					Separator:  "",
					RepeatedArgs: slicePA{
						{
							ArgName:         "login",
							ArgType:         "string",
							Repeatable:      false,
							Separator:       "",
							RepeatedArgs:    nil,
							RepeatedQuery:   "",
							IsGeneratedName: false,
							PlaceHolder:     "$0$",
						},
						{
							ArgName:         "name",
							ArgType:         "string",
							Repeatable:      false,
							Separator:       "",
							RepeatedArgs:    nil,
							RepeatedQuery:   "",
							IsGeneratedName: false,
							PlaceHolder:     "$1$",
						},
					},
					RepeatedQuery:   " or (login=$0$ and name=$1$)",
					IsGeneratedName: true,
					PlaceHolder:     "$1$",
				},
				{
					ArgName:         "isAdmin",
					ArgType:         "int",
					Repeatable:      false,
					Separator:       "",
					RepeatedArgs:    nil,
					RepeatedQuery:   "",
					IsGeneratedName: false,
					PlaceHolder:     "$2$",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := findArgs(tt.query, tt.supportReturning, true)
			assert.Equal(t, tt.wantErr, err, "wrong error")
			if err != nil {
				if res != nil {
					t.Error("got error and result!")
				}
			} else if res == nil {
				t.Error("result is nil with empty error!")
			} else {
				assert.Equal(t, tt.wantArgs, res.Arguments, "wrong out args")
				assert.Equal(t, tt.wantOutQuery, res.Query, "wrong out query")
				assert.Equal(t, tt.wantHaveRepeatableParts, res.HaveRepeatableParts, "wrong HaveRepeatableParts")
				assert.Equal(t, tt.wantParams, res.ReturnParams, "wrong out params")
			}
		})
	}
}
