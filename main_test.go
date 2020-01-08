package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_findArgs(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name             string
		query            string
		outType          rune
		wantOutQuery     string
		wantArgs         []*parsedArg
		wantParams       []*parsedArg
		wantErr          error
		supportReturning bool
	}{
		{
			name:    "negative. missing brackets",
			query:   "select if(a>b, a, b from users",
			outType: '?',
			wantErr: fmt.Errorf("parse error! found unpaired bracket ("),
		},
		{
			name:    "negative. wrong closing bracket",
			query:   "select if(a>b, a], b from users",
			outType: '?',
			wantErr: fmt.Errorf("parse error! wrong bracket! excpected [ but found ( at right of select if(a>b, a"),
		},

		{
			name:    "negative. senseless query inside brackets",
			query:   "select if(a>b, a, a:b) a from users",
			outType: '?',
			wantErr: fmt.Errorf("parse error! use should use proper columnName:type pair for column! only alpha numerics are allowed! near: select if(a>b, a, a:b) a "),
		},

		{
			name:         "positive. senseless query inside brackets",
			query:        "select if(a>b, a, a:b) a:int from users",
			outType:      '?',
			wantOutQuery: "select if(a>b, a, a:b) a from users",
			wantParams:   []*parsedArg{&parsedArg{ArgName: "a", ArgType: "int"}},
			wantArgs:     []*parsedArg{},
		},

		{
			name:    "$ token without name",
			query:   "wdqwq $ $aaas",
			outType: '?',
			wantErr: fmt.Errorf("parse error! no argument name right from wdqwq $"),
		},
		{
			name:    "bad symbol afrer $",
			query:   "wdqwq $( $aaas",
			outType: '?',
			wantErr: fmt.Errorf("parse error! bad token ( after $ at right of wdqwq $"),
		},
		{
			name:    "string after $",
			query:   "wdqwq $aaas wefweffe wef wef",
			outType: '?',
			wantErr: fmt.Errorf("parse error! no argument type right from wdqwq $aaas"),
		},
		{
			name:    "no type",
			query:   "wdqwq $aaas: wefweffe wef wef",
			outType: '?',
			wantErr: fmt.Errorf("parse error! no argument type right from wdqwq $aaas:"),
		},
		{
			name:    "int instead of type",
			query:   "wdqwq $aaas:23 wefweffe wef wef",
			outType: '?',
			wantErr: fmt.Errorf("parse error! no argument type right from wdqwq $aaas:"),
		},

		{
			name:         "positive string after $",
			query:        "wdqwq $aaas:int wefweffe wef wef 2",
			outType:      '?',
			wantOutQuery: "wdqwq ? wefweffe wef wef 2",
			wantArgs:     []*parsedArg{&parsedArg{ArgName: "aaas", ArgType: "int"}},
			wantParams:   []*parsedArg{},
		},

		{
			name:         "positive string after $, outtype $",
			query:        "wdqwq $aaas:int wefweffe wef wef",
			outType:      '$',
			wantOutQuery: "wdqwq $1 wefweffe wef wef",
			wantArgs:     []*parsedArg{&parsedArg{ArgName: "aaas", ArgType: "int"}},
			wantParams:   []*parsedArg{},
		},

		{
			name:    "negative. two arguments with same name and different type",
			query:   "select id:int from messages where from=$userId:int or to=$userId:string",
			outType: '$',
			wantErr: fmt.Errorf("parse error! argument with name userId have different types! int and string"),
		},
		{
			name:    "negative. no params",
			query:   "select * from messages where from=$userId:int or to=$userId:string",
			outType: '$',
			wantErr: fmt.Errorf("parse error! use should use proper columnName:type pair for column! only alpha numerics are allowed! near: select * "),
		},
		{
			name:         "pasitive. with param",
			query:        "select `from`:int, to:int from messages where from=$userId:int or to=$userId:int",
			outType:      '$',
			wantOutQuery: "select `from`, to from messages where from=$1 or to=$1",
			wantArgs:     []*parsedArg{&parsedArg{ArgType: "int", ArgName: "userId"}},
			wantParams: []*parsedArg{
				&parsedArg{ArgName: "`from`", ArgType: "int"},
				&parsedArg{ArgName: "to", ArgType: "int"},
			},
		},

		{
			name:         "positive string after $, outtype $ same arg name. first time with type, second without",
			query:        "select id:int from messages where from=$userId:int or to=$userId",
			outType:      '$',
			wantOutQuery: "select id from messages where from=$1 or to=$1",
			wantArgs: []*parsedArg{
				&parsedArg{ArgName: "userId", ArgType: "int"},
			},
			wantParams: []*parsedArg{&parsedArg{ArgName: "id", ArgType: "int"}},
		},
		{
			name:         "positive string after $, outtype $ same arg name. first time without type, second with",
			query:        "select id:int from messages where from=$userId or to=$userId:int",
			outType:      '$',
			wantOutQuery: "select id from messages where from=$1 or to=$1",
			wantArgs: []*parsedArg{
				&parsedArg{ArgName: "userId", ArgType: "int"},
			},
			wantParams: []*parsedArg{&parsedArg{ArgName: "id", ArgType: "int"}},
		},
		{
			name:         "positive string after $, outtype $ same arg name. first time without type, second with. third with",
			query:        "select id:int from messages where (from=$userId or to=$userId:int) and date>$date:string",
			outType:      '$',
			wantOutQuery: "select id from messages where (from=$1 or to=$1) and date>$2",
			wantArgs: []*parsedArg{
				&parsedArg{ArgName: "userId", ArgType: "int"},
				&parsedArg{ArgName: "date", ArgType: "string"},
			},
			wantParams: []*parsedArg{&parsedArg{ArgName: "id", ArgType: "int"}},
		},

		{
			name:         "positive no arguments",
			query:        "select id:int from messages",
			outType:      '$',
			wantArgs:     []*parsedArg{},
			wantOutQuery: "select id from messages",
			wantParams:   []*parsedArg{&parsedArg{ArgName: "id", ArgType: "int"}},
		},
		{
			name:         "positive no arguments. parameter type with namespace",
			query:        "select id:sql.NullInt, name:string from messages",
			outType:      '$',
			wantArgs:     []*parsedArg{},
			wantOutQuery: "select id, name from messages",
			wantParams: []*parsedArg{
				&parsedArg{ArgName: "id", ArgType: "sql.NullInt"},
				&parsedArg{ArgName: "name", ArgType: "string"},
			},
		},

		{
			name:             "positive. checking returning arguments in delete",
			query:            "delete from messages returning id:int, name:string",
			outType:          '$',
			supportReturning: true,
			wantArgs:         []*parsedArg{},
			wantOutQuery:     "delete from messages returning id, name",
			wantParams: []*parsedArg{
				&parsedArg{ArgName: "id", ArgType: "int"},
				&parsedArg{ArgName: "name", ArgType: "string"},
			},
		},
		{
			name:             "negative. checking returning arguments in delete with supportReturning disabled",
			query:            "delete from messages returning id:int, name:string",
			outType:          '$',
			supportReturning: false,
			wantArgs:         []*parsedArg{},
			wantOutQuery:     "delete from messages returning id:int, name:string",

			wantParams: []*parsedArg{},
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := findArgs(tt.query, tt.outType, tt.supportReturning)
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
				assert.Equal(t, tt.wantParams, res.ReturnParams, "wrong out params")
			}
		})
	}
}
