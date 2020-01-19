package main

import (
	"bytes"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type yamlFile struct {
	SqlDriver       string            `yaml:"sqlDriver"`
	OutputFileName  string            `yaml:"outFilePath"`
	Pkg             string            `yaml:"package"`
	Imports         []string          `yaml:"imports"`
	QueriesSlice    map[string]string `yaml:"queries"`
	QueriesData     map[string]*data
	WithContext     bool `yaml:"withContext"`
	WithTransaction bool `yaml:"withTransaction"`
}

type data struct {
	Arguments           slicePA
	ReturnParams        slicePP
	Query               string
	QueryType           string
	HaveRepeatableParts bool
	Name                string
}

var yamlPath = flag.String("yaml", "", "path to yaml")

var arguments = map[string]*data{}
var supportReturning = false

func dieOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func chechMapForEmpty(required map[string]string) {
	var haveErrors = false
	for i := range required {
		if required[i] == "" {
			fmt.Println(i, "argument is empty but required")
			haveErrors = true
		}
	}
	if haveErrors {
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	//*yamlPath = "./example/mysql/sql.yaml"
	//*yamlPath = "./example/postgresql/queries.yaml"
	chechMapForEmpty(map[string]string{"yaml": *yamlPath})
	yamlData := yamlFile{}
	b, e := ioutil.ReadFile(*yamlPath)
	dieOnError(e)
	dieOnError(yaml.Unmarshal(b, &yamlData))

	chechMapForEmpty(map[string]string{"sqlDriver": yamlData.SqlDriver, "package": yamlData.Pkg, "outFilePath": yamlData.OutputFileName})
	switch yamlData.SqlDriver {
	case "github.com/go-sql-driver/mysql":
	case "github.com/lib/pq":
		supportReturning = true
	default:
		fmt.Println("unsupported db type ", yamlData.SqlDriver)
		os.Exit(1)
	}
	for i, v := range yamlData.Imports {
		if strings.TrimSpace(v[len(v)-1:]) != `"` {
			yamlData.Imports[i] = `"` + v + `"`
		}
	}

	for i := range yamlData.QueriesSlice {
		arguments[i], e = findArgs(yamlData.QueriesSlice[i], supportReturning, true)
		dieOnError(e)
		arguments[i].Name = i
	}
	if yamlData.OutputFileName != "/" {
		yamlData.OutputFileName = path.Join(filepath.Dir(*yamlPath), yamlData.OutputFileName)
	}
	yamlData.QueriesData = arguments
	f, e := os.Create(yamlData.OutputFileName)
	dieOnError(e)
	e = generate(f, yamlData)
	dieOnError(e)
	if cmd := exec.Command("go", "fmt", yamlData.OutputFileName); cmd != nil {
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		e = cmd.Run()
		if e != nil {
			dieOnError(fmt.Errorf(`"go fmt" stderr: %s`, errb.String()))
		}
	}
}

type parsedArg struct {
	ArgName string
	ArgType string
	// for repeatable
	Repeatable      bool
	Separator       string
	RepeatedArgs    slicePA
	RepeatedQuery   string
	IsGeneratedName bool
	PlaceHolder     string
}

type parsedParam struct {
	ParamName string
	ParamType string
}

type slicePA []*parsedArg
type slicePP []*parsedParam

func (s slicePA) PrintNamesAndTypes(prefix string, separator string, withType bool) string {
	first := true
	out := ""
	for _, v := range s {
		argType := v.ArgType
		if withType && v.Repeatable {
			argType = "[]" + argType
		}
		out += printNamesAndTypes(prefix, v.ArgName, argType, separator, &first, withType)
	}
	return out
}

func (s slicePA) CountArgs() string {
	total := 0
	out := ""
	for _, v := range s {
		if !v.Repeatable {
			total++
		} else {
			if out != "" {
				out += "+"
			}
			multiplier := 1
			if len(v.RepeatedArgs) > 1 {
				multiplier = len(v.RepeatedArgs)
			}
			out += "len(" + v.ArgName + ")*" + strconv.Itoa(multiplier)
		}
	}
	if total > 0 {
		return out + "+" + strconv.Itoa(total)
	}
	return out
}

func (s slicePP) PrintNamesAndTypes(prefix string, separator string, withType bool) string {
	first := true
	out := ""
	for _, v := range s {
		out += printNamesAndTypes(prefix, v.ParamName, v.ParamType, separator, &first, withType)
	}
	return out
}

func (s *data) ReplacePlaceHoldersInQueryIfNoRepeatable(as rune) string {
	query := s.Query
	if s.HaveRepeatableParts == false {
		for i := range s.Arguments {
			if as == '?' {
				query = strings.Replace(query, s.Arguments[i].PlaceHolder, "?", 1)
			} else {
				query = strings.Replace(query, s.Arguments[i].PlaceHolder, "$"+strconv.Itoa(i+1), 1)
			}
		}
	}
	return query
}
