package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"
	"text/template"
)

type yamlFile struct {
	DbType  string
	Queries map[string]string `yaml:"queries"`
}

type data struct {
	Arguments    []*parsedArg
	ReturnParams []*parsedArg
	Query        string
	QueryType    string
}

var dbType = flag.String("dbtype", "", "package source directory, useful for vendored code")
var yamlPath = flag.String("yaml", "", "path to yaml")
var outputFileName = flag.String("output", "", "where should i put created file")
var pkg = flag.String("package", "", "output package name")

var isAlphaNumericName = regexp.MustCompile("^['`\"]?[_a-zA-Z][_a-zA-Z0-9]*['`\"]?$")

var funcMap = template.FuncMap{
	"Escape": func(s string) string { return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\"`), `"`, `\"`) },
	"PrintCallParams": func(params []*parsedArg, prefix string, firstComma bool) string {
		out := ""
		for i := range params {
			if !firstComma {
				firstComma = true
			} else {
				out += ", "
			}
			out += prefix + params[i].ArgName
		}
		return out
	},
}

func main() {
	flag.Parse()
	required := map[string]*string{"dbtype": dbType, "yaml": yamlPath, "output": outputFileName, "package": pkg}
	haveErrors := false
	for i := range required {
		if *required[i] == "" {
			fmt.Println(i, "argument is empty but required")
			haveErrors = true
		}
	}
	if haveErrors {
		os.Exit(1)
	}
	var queries = map[string]string{}
	if b, e := ioutil.ReadFile(*yamlPath); e != nil {
		fmt.Println(e)
		return
	} else {
		err := yaml.Unmarshal(b, &queries)
		if err != nil {
			fmt.Println("error", err)
			return
		}
	}
	var arguments = map[string]*data{}
	tmpl := ""
	var err error
	switch *dbType {
	case "mysql":
		for i := range queries {
			arguments[i], err = findArgs(queries[i], '?', false)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		tmpl = templatesMysql
	case "postgres":
		fmt.Println("pg")
		for i := range queries {
			arguments[i], err = findArgs(queries[i], '$', true)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		tmpl = templatesPostrgeSQL
	default:
		fmt.Println("unsupported db type")
	}

	t, err := template.New("").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		log.Print(err)
		return
	}
	f, err := os.Create(*outputFileName)
	log.Println("create file: ")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = t.Execute(f, struct {
		Queries map[string]*data
		Package string
	}{Queries: arguments, Package: *pkg})
	if err != nil {
		log.Print("execute: ", err)
		return
	}
}

type arg struct {
	argType  string
	argName  string
	posStart int
	posEnd   int
	isParam  bool
}

type parsedArg struct {
	ArgName string
	ArgType string
}

type nameAndNumber struct {
	argType string
	number  int
	link    int
}

type token struct {
	token     rune
	text      string
	lowerText string
	startPos  int
	endPos    int
}

var bracketPairs = map[rune]rune{
	')': '(',
	']': '[',
	'}': '{',
}

func findArgs(query string, outType rune, supportReturning bool) (res *data, err error) {
	var s = scanner.Scanner{Error: func(*scanner.Scanner, string) {}}
	s.Init(strings.NewReader(query))
	args := make([]*arg, 0)
	i := -1
	lastTokenPos := -1
	state := -1
	argTypeByName := map[string]*nameAndNumber{}
	last6tokens := [6]token{}
	brackets := make(stack, 0)
	tok := s.Scan()
	var pair rune
	queryType := strings.ToLower(s.TokenText())
	returingStarted := false
	selectsParams := false
	if queryType == "select" {
		selectsParams = true
	}
	for tok = s.Scan(); last6tokens[0].token != scanner.EOF; tok = s.Scan() {
		// parse return values
		last6tokens[5], last6tokens[4], last6tokens[3], last6tokens[2], last6tokens[1] = last6tokens[4], last6tokens[3], last6tokens[2], last6tokens[1], last6tokens[0]
		last6tokens[0] = token{
			token:     tok,
			text:      s.TokenText(),
			lowerText: strings.ToLower(s.TokenText()),
			startPos:  s.Offset,
			endPos:    s.Offset + len(s.TokenText()),
		}
		if tok == '{' || tok == '(' || tok == '[' {
			brackets = brackets.Push(tok)
		}
		if bracketPairs[tok] != 0 {
			brackets, pair = brackets.Pop()
			if pair != bracketPairs[tok] {
				return nil, fmt.Errorf("parse error! wrong bracket! excpected %c but found %c at right of %s", bracketPairs[tok], pair, query[0:s.Offset])
			}
		}
		if len(brackets) == 0 && supportReturning &&
			(queryType == "insert" || queryType == "delete" || queryType == "update") &&
			last6tokens[0].lowerText == "returning" {
			returingStarted = true
		}

		if len(brackets) == 0 && (returingStarted || selectsParams) && (tok == ',' || last6tokens[0].lowerText == "from" || tok == scanner.EOF) {
			if last6tokens[1].token == scanner.Ident && last6tokens[2].token == ':' &&
				isAlphaNumericName.MatchString(last6tokens[3].lowerText) &&
				last6tokens[1].startPos == last6tokens[2].endPos && last6tokens[2].startPos == last6tokens[3].endPos {
				args = append(args, &arg{argName: last6tokens[3].text, argType: last6tokens[1].lowerText, posStart: last6tokens[3].startPos, posEnd: last6tokens[1].endPos, isParam: true})
				i++
			} else if last6tokens[1].token == scanner.Ident && last6tokens[2].token == '.' && last6tokens[3].token == scanner.Ident && last6tokens[4].token == ':' &&
				isAlphaNumericName.MatchString(last6tokens[5].lowerText) &&
				last6tokens[1].startPos == last6tokens[2].endPos && last6tokens[2].startPos == last6tokens[3].endPos && last6tokens[3].startPos == last6tokens[4].endPos && last6tokens[4].startPos == last6tokens[5].endPos {
				args = append(args, &arg{argName: last6tokens[5].text, argType: last6tokens[3].text + "." + last6tokens[1].text, posStart: last6tokens[5].startPos, posEnd: last6tokens[1].endPos, isParam: true})
				i++
			} else {
				return nil, fmt.Errorf("parse error! use should use proper columnName:type pair for column! only alpha numerics are allowed! near: %s", query[0:s.Offset])
			}
		}
		if queryType == "select" && len(brackets) == 0 && last6tokens[0].lowerText == "from" {
			selectsParams = false
		}
		// parse params
		if tok == '$' {
			lastTokenPos = s.Offset
			state = 0
			args = append(args, &arg{posStart: s.Offset, posEnd: s.Offset + 1})
			i++
		} else if state < 0 {
			continue
		} else if s.Offset-1 != lastTokenPos {
			state = -1
			args[i].posEnd = s.Offset - 1
		} else if state == 0 {
			if tok != scanner.Ident {
				return nil, fmt.Errorf("parse error! bad token %c after $ at right of %s", tok, query[0:s.Offset])
			}
			args[i].argName = s.TokenText()
			lastTokenPos = s.Offset - 1 + len(s.TokenText())
			state = 1
			args[i].posEnd = s.Offset + len(s.TokenText())
			if _, ok := argTypeByName[args[i].argName]; !ok {
				argTypeByName[args[i].argName] = &nameAndNumber{argType: "", link: i}
				if outType == '$' {
					argTypeByName[args[i].argName].number = len(argTypeByName)
				}
			}
		} else if state == 1 && tok == ':' {
			state = 2
			lastTokenPos = s.Offset
		} else if state == 2 && tok == scanner.Ident {
			args[i].argType = s.TokenText()
			lastTokenPos = s.Offset + len(s.TokenText())
			state = -1
			args[i].posEnd = s.Offset + len(s.TokenText())
			if argTypeByName[args[i].argName].argType == "" {
				argTypeByName[args[i].argName].argType = args[i].argType
			} else {

			}
		} else {
			state = -1
			args[i].posEnd = s.Offset
		}
	}
	if len(brackets) > 0 {
		_, bracket := brackets.Pop()
		return nil, fmt.Errorf("parse error! found unpaired bracket %c", bracket)
	}
	offset := 0
	res = &data{Arguments: make([]*parsedArg, 0), ReturnParams: make([]*parsedArg, 0)}
	for i := range args {
		if args[i].argName == "" {
			return nil, fmt.Errorf("parse error! no argument name right from %s", query[0:args[i].posEnd])
		}
		if args[i].argType == "" && !args[i].isParam {
			args[i].argType = argTypeByName[args[i].argName].argType
		}
		if args[i].argType == "" {
			return nil, fmt.Errorf("parse error! no argument type right from %s", query[0:args[i].posEnd])
		}
		if !args[i].isParam && argTypeByName[args[i].argName].argType != args[i].argType {
			return nil, fmt.Errorf("parse error! argument with name %s have different types! %s and %s", args[i].argName, argTypeByName[args[i].argName].argType, args[i].argType)
		}
		if args[i].isParam {
			res.Query += query[offset:args[i].posStart] + args[i].argName
		} else if outType == '?' {
			res.Query += query[offset:args[i].posStart] + "?"
		} else {
			res.Query += query[offset:args[i].posStart] + "$" + strconv.Itoa(argTypeByName[args[i].argName].number)
		}
		if args[i].isParam {
			res.ReturnParams = append(res.ReturnParams, &parsedArg{ArgType: args[i].argType, ArgName: args[i].argName})
		} else if outType == '?' || argTypeByName[args[i].argName].link == i {
			res.Arguments = append(res.Arguments, &parsedArg{ArgType: argTypeByName[args[i].argName].argType, ArgName: args[i].argName})
		}
		offset = args[i].posEnd
	}
	res.Query += query[offset:]
	res.QueryType = queryType
	return res, nil
}
