package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
	"Escape": func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\"`), `"`, `\"`)
	},
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
var arguments = map[string]*data{}
var tmpl = templatesMysql
var argRune = '?'
var supportReturning = false
var queries = map[string]string{}
var haveErrors = false
var required = map[string]*string{"dbtype": dbType, "yaml": yamlPath, "output": outputFileName, "package": pkg}

func dieOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {

	flag.Parse()
	for i := range required {
		if *required[i] == "" {
			fmt.Println(i, "argument is empty but required")
			haveErrors = true
		}
	}
	if haveErrors {
		os.Exit(1)
	}

	b, e := ioutil.ReadFile(*yamlPath)
	dieOnError(e)
	dieOnError(yaml.Unmarshal(b, &queries))

	switch *dbType {
	case "mysql":
	case "postgres":
		argRune = '$'
		supportReturning = true
		tmpl = templatesPostrgeSQL
	default:
		fmt.Println("unsupported db type")
	}
	for i := range queries {
		arguments[i], e = findArgs(queries[i], argRune, supportReturning)
		dieOnError(e)
	}

	t, e := template.New("").Funcs(funcMap).Parse(tmpl)
	dieOnError(e)
	f, e := os.Create(*outputFileName)
	dieOnError(e)
	e = t.Execute(f, struct {
		Queries map[string]*data
		Package string
	}{Queries: arguments, Package: *pkg})
	dieOnError(e)
}

type arg struct {
	argType     string
	argName     string
	posStart    int
	posEnd      int
	isParam     bool
	notFinished bool
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
	argTypeJustSet := false
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
				return nil, fmt.Errorf("wrong bracket! excpected %c but found %c at right of %s", bracketPairs[tok], pair, query[0:s.Offset])
			}
		}
		if len(brackets) == 0 && supportReturning &&
			(queryType == "insert" || queryType == "delete" || queryType == "update") &&
			last6tokens[0].lowerText == "returning" {
			returingStarted = true
		}

		if len(brackets) == 0 && (returingStarted || selectsParams) &&
			(tok == ',' || last6tokens[0].lowerText == "from" || tok == scanner.EOF) {
			if last6tokens[1].token == scanner.Ident && last6tokens[2].token == ':' &&
				isAlphaNumericName.MatchString(last6tokens[3].lowerText) &&
				last6tokens[1].startPos == last6tokens[2].endPos && last6tokens[2].startPos == last6tokens[3].endPos {
				args = append(args, &arg{
					argName:  last6tokens[3].text,
					argType:  last6tokens[1].lowerText,
					posStart: last6tokens[3].startPos,
					posEnd:   last6tokens[1].endPos,
					isParam:  true,
				})
				i++
			} else if last6tokens[1].token == scanner.Ident && last6tokens[2].token == '.' &&
				last6tokens[3].token == scanner.Ident && last6tokens[4].token == ':' &&
				isAlphaNumericName.MatchString(last6tokens[5].lowerText) &&
				last6tokens[1].startPos == last6tokens[2].endPos && last6tokens[2].startPos == last6tokens[3].endPos &&
				last6tokens[3].startPos == last6tokens[4].endPos && last6tokens[4].startPos == last6tokens[5].endPos {
				args = append(args, &arg{
					argName:  last6tokens[5].text,
					argType:  last6tokens[3].text + "." + last6tokens[1].text,
					posStart: last6tokens[5].startPos,
					posEnd:   last6tokens[1].endPos,
					isParam:  true,
				})
				i++
			} else {
				return nil, fmt.Errorf("use should use proper columnName:type pair for column! only alpha numerics are allowed! near: %s", query[0:s.Offset])
			}
		}
		if queryType == "select" && len(brackets) == 0 && last6tokens[0].lowerText == "from" {
			selectsParams = false
		}
		// parse arguments
		if state >= 0 && last6tokens[0].startPos == last6tokens[1].endPos && (tok == '$') {
			return nil, fmt.Errorf("syntax error near: %s", query[0:last6tokens[0].endPos])
		}
		if tok == '$' {
			state = 0
			argTypeJustSet = false
			args = append(args, &arg{posStart: s.Offset, posEnd: s.Offset + 1})
			i++
			continue
		}

		if state < 0 {
			continue
		}

		if last6tokens[0].startPos != last6tokens[1].endPos || tok == scanner.EOF {
			state = -1
			args[i].posEnd = last6tokens[1].endPos
			continue
		}

		if state == 0 {
			if tok != scanner.Ident {
				return nil, fmt.Errorf("bad token %c after $ at right of %s", tok, query[0:s.Offset])
			}
			args[i].argName = s.TokenText()
			state = 1
			if _, ok := argTypeByName[args[i].argName]; !ok {
				argTypeByName[args[i].argName] = &nameAndNumber{argType: "", link: i}
				if outType == '$' {
					argTypeByName[args[i].argName].number = len(argTypeByName)
				}
			}
		} else if state == 1 && tok == ':' {
			state = 2
		} else if state == 2 && tok == scanner.Ident {
			args[i].argType = s.TokenText()
			state = 3
			if argTypeByName[args[i].argName].argType == "" {
				argTypeByName[args[i].argName].argType = args[i].argType
				argTypeJustSet = true
			} else {

			}
		} else if state == 3 && tok == '.' {
			state = 4
			args[i].notFinished = true
		} else if state == 4 && tok == scanner.Ident {
			args[i].argType += "." + s.TokenText()
			args[i].notFinished = false
			state = 5
			if argTypeJustSet {
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
		return nil, fmt.Errorf("found unpaired bracket %c", bracket)
	}
	offset := 0
	res = &data{Arguments: make([]*parsedArg, 0), ReturnParams: make([]*parsedArg, 0)}
	for i := range args {
		if args[i].notFinished {
			return nil, fmt.Errorf("not finished type near %s", query[0:args[i].posEnd])
		}
		if args[i].argName == "" {
			return nil, fmt.Errorf("no argument name right from %s", query[0:args[i].posEnd])
		}
		if args[i].argType == "" && !args[i].isParam {
			args[i].argType = argTypeByName[args[i].argName].argType
		}
		if args[i].argType == "" {
			return nil, fmt.Errorf("no argument type right from %s", query[0:args[i].posEnd])
		}
		if !args[i].isParam && argTypeByName[args[i].argName].argType != args[i].argType {
			return nil, fmt.Errorf("argument with name %s have different types! %s and %s", args[i].argName, argTypeByName[args[i].argName].argType, args[i].argType)
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
			res.Arguments = append(res.Arguments, &parsedArg{
				ArgType: argTypeByName[args[i].argName].argType,
				ArgName: args[i].argName,
			})
		}
		offset = args[i].posEnd
	}
	res.Query += query[offset:]
	res.QueryType = queryType
	return res, nil
}
