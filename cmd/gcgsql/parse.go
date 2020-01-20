package main

import (
	"fmt"
	"regexp"
	"strings"
	"text/scanner"
)

type arg struct {
	argType   string
	argName   string
	separator string

	repeatedArgs    []*parsedArg
	repeatedQuery   string
	repeatable      bool
	isGeneratedName bool

	posStart    int
	posEnd      int
	isParam     bool
	notFinished bool
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

var scannerError = func(*scanner.Scanner, string) {}

// typeRequired if have repeatable part with type -> we must omit type check inside
// repeatable part and return error if type found.
func findArgs(query string, supportReturning bool, typeRequired bool) (res *data, err error) {
	var isAlphaNumericName = regexp.MustCompile("^['`\"]?[_a-zA-Z][_a-zA-Z0-9]*['`\"]?$")
	var s scanner.Scanner
	s.Init(strings.NewReader(query))
	s.Error = scannerError
	args := make([]*arg, 0)
	i := -1
	state := -1
	argTypeByName := map[string]*nameAndNumber{}
	last6tokens := [6]token{}
	brackets := make(stack, 0)

	var pair rune
	queryType := ""
	returningStarted := false
	selectsParams := false
	argTypeJustSet := false
	RepeatedPlaceHolder := int64(0)
	reperableStartPos := 0
	res = &data{Arguments: make([]*parsedArg, 0), ReturnParams: make([]*parsedParam, 0)}
	for tok := s.Scan(); last6tokens[0].token != scanner.EOF; tok = s.Scan() {
		if s.Offset == 0 {
			queryType = strings.ToLower(s.TokenText())
			if queryType == "select" {
				selectsParams = true
			}
		}
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
			if len(brackets) == 0 {
				return nil, fmt.Errorf("found closing bracket %c but no opening found near %s", tok, query[0:s.Offset+1])
			}
			brackets, pair = brackets.Pop()
			if pair != bracketPairs[tok] {
				return nil, fmt.Errorf("wrong bracket! excpected %c but found %c at right of %s", bracketPairs[tok], pair, query[0:s.Offset])
			}
		}
		if len(brackets) == 0 && supportReturning &&
			(queryType == "insert" || queryType == "delete" || queryType == "update") &&
			last6tokens[0].lowerText == "returning" {
			returningStarted = true
		}

		if len(brackets) == 0 && (returningStarted || selectsParams) &&
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
			} else if last6tokens[1].token == scanner.Ident && last6tokens[1].text == "byte" &&
				last6tokens[2].token == ']' &&
				last6tokens[3].token == '[' &&
				last6tokens[4].token == ':' &&
				isAlphaNumericName.MatchString(last6tokens[5].lowerText) &&
				last6tokens[1].startPos == last6tokens[2].endPos && last6tokens[2].startPos == last6tokens[3].endPos &&
				last6tokens[3].startPos == last6tokens[4].endPos && last6tokens[4].startPos == last6tokens[5].endPos {
				args = append(args, &arg{
					argName:  last6tokens[5].text,
					argType:  "[]byte",
					posStart: last6tokens[5].startPos,
					posEnd:   last6tokens[1].endPos,
					isParam:  true,
				})
				i++

			} else {
				return nil, fmt.Errorf("use should use proper columnName:type pair for column! only alpha numerics are allowed! near: %s", query[0:s.Offset])
			}
		}
		if selectsParams && len(brackets) == 0 && last6tokens[0].lowerText == "from" {
			selectsParams = false
		}
		// parse arguments
		if state >= 0 && state < 20 && last6tokens[0].startPos == last6tokens[1].endPos && tok == '$' {
			return nil, fmt.Errorf("syntax error near: %s", query[0:last6tokens[0].endPos])
		}
		if tok == '$' && state < 20 {
			state = 0
			argTypeJustSet = false
			args = append(args, &arg{posStart: s.Offset, posEnd: s.Offset + 1})
			i++
			continue
		}

		if state < 0 {
			continue
		}

		if (state < 20 && last6tokens[0].startPos != last6tokens[1].endPos) || tok == scanner.EOF {
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
			}
		} else if (state == 1 || state == 5 || state == 7) && tok == '#' {
			reperableStartPos = last6tokens[0].endPos
			res.HaveRepeatableParts = true
			if args[i].repeatable {
				return nil, fmt.Errorf("found [] type before # in %s", query[0:last6tokens[0].endPos])
			}
			state = 20
			args[i].notFinished = true
			args[i].repeatable = true
		} else if state == 20 {
			if tok == '#' {
				//args[i].repeatedQuery
				args[i].notFinished = false
				args[i].repeatedQuery = query[reperableStartPos:last6tokens[0].startPos]
				args[i].separator = ","
				state = 25
			} else if tok == '|' {
				args[i].repeatedQuery = query[reperableStartPos:last6tokens[0].startPos]
				reperableStartPos = last6tokens[0].endPos
				state = 21
			}
		} else if state == 21 {
			if tok == '#' {
				args[i].notFinished = false
				args[i].separator = query[reperableStartPos:last6tokens[0].startPos]
				state = 25
			}
		} else if state == 1 && tok == ':' {
			state = 2
		} else if state == 2 && tok == '[' {
			state = 3
		} else if state == 3 && tok == ']' {
			state = 4
			args[i].repeatable = true
			args[i].separator = ","
			res.HaveRepeatableParts = true
		} else if (state == 4 || state == 2) && tok == scanner.Ident {
			args[i].argType = s.TokenText()
			state = 5
			if argTypeByName[args[i].argName].argType == "" {
				argTypeByName[args[i].argName].argType = args[i].argType
				argTypeJustSet = true
			} else {

			}
		} else if state == 5 && tok == '.' {
			state = 6
			args[i].notFinished = true
		} else if state == 6 && tok == scanner.Ident {
			args[i].argType += "." + s.TokenText()
			args[i].notFinished = false
			state = 7
			if argTypeJustSet {
				argTypeByName[args[i].argName].argType = args[i].argType
			} else {

			}
		} else {
			state = -1
			args[i].posEnd = last6tokens[1].endPos
		}
	}
	if len(brackets) > 0 {
		_, bracket := brackets.Pop()
		return nil, fmt.Errorf("found unpaired bracket %c", bracket)
	}
	offset := 0
	placeHolder := ""
	for i := range args {
		if !args[i].isParam {
			for {
				placeHolder = fmt.Sprintf("$%d$", RepeatedPlaceHolder)
				RepeatedPlaceHolder++
				if strings.Index(query, placeHolder) < 0 {
					break
				}
			}
		}
		if args[i].notFinished {
			return nil, fmt.Errorf("not finished type near %s", query[0:args[i].posEnd])
		}
		if args[i].argName == "" {
			return nil, fmt.Errorf("no argument name right from %s", query[0:args[i].posEnd])
		}
		if !args[i].isParam {
			if args[i].argType == "" {
				args[i].argType = argTypeByName[args[i].argName].argType
			}

			if typeRequired && args[i].argType == "" && !args[i].repeatable {
				return nil, fmt.Errorf("no argument type right from %s", query[0:args[i].posEnd])
			}
			if !typeRequired && args[i].argType != "" {
				return nil, fmt.Errorf("found argument type inside repeatable part, but repeatable already have type. near %s", query[0:args[i].posEnd])
			}

			if argTypeByName[args[i].argName].argType != args[i].argType {
				return nil, fmt.Errorf("argument with name \"%s\" have different types! %s and %s", args[i].argName, argTypeByName[args[i].argName].argType, args[i].argType)
			}
		}

		if args[i].isParam {
			res.Query += query[offset:args[i].posStart] + args[i].argName
		} else {
			res.Query += query[offset:args[i].posStart] + placeHolder
		}
		if args[i].repeatable && args[i].repeatedQuery != "" {
			tmp, err := findArgs(args[i].repeatedQuery, false, args[i].argType == "")
			if err != nil {
				return nil, fmt.Errorf("error inside repeatable part. %s", err.Error())
			}
			args[i].repeatedArgs = tmp.Arguments
			args[i].repeatedQuery = tmp.Query
			if args[i].argType == "" {
				args[i].argType = args[i].argName + "Struct"
				args[i].isGeneratedName = true
				if argTypeByName[args[i].argName].argType == "" {
					argTypeByName[args[i].argName].argType = args[i].argType
				}
			}
		}
		if args[i].isParam {
			res.ReturnParams = append(res.ReturnParams, &parsedParam{
				ParamType: args[i].argType,
				ParamName: args[i].argName,
			})
		} else {
			res.Arguments = append(res.Arguments, &parsedArg{
				ArgType:         argTypeByName[args[i].argName].argType,
				ArgName:         args[i].argName,
				PlaceHolder:     placeHolder,
				RepeatedArgs:    args[i].repeatedArgs,
				RepeatedQuery:   args[i].repeatedQuery,
				Separator:       args[i].separator,
				Repeatable:      args[i].repeatable,
				IsGeneratedName: args[i].isGeneratedName,
			})
		}
		offset = args[i].posEnd
	}

	res.Query += query[offset:]
	res.QueryType = queryType
	return res, nil
}
