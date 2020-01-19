package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_funcMap(t *testing.T) {
	data := funcMap["Escape"].(func(s string) string)(`"khk\" h"`)
	assert.Equal(t, "\\\"khk\\\\\\\"\\\" h\\\"", data, "escape not working")
}

/*type a struct {
	data []byte
}

func (b *a) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}
*/
/*func Test_generate(t *testing.T) {
	out := &a{data: []byte{}}
	e := generate(out, yamlFile{SqlDriver:"github.com/go-sql-driver/mysql",QueriesData:map[string]*data{}})
	assert.Equal(t, nil, e, "err not nil")
	assert.Equal(t, "hzche", string(out.data), "no realizied yet")
	out.data = out.data[0:0]
	e = generate(out, yamlFile{SqlDriver:"github.com/lib/pq",QueriesData:map[string]*data{}})
	assert.Equal(t, nil, e, "err not nil")
	assert.Equal(t, "hzche", string(out.data), "no realizied yet")
	out.data = out.data[0:0]
	e = generate(out, yamlFile{SqlDriver:"github.com/lib/pq",QueriesData:map[string]*data{}})
	assert.Equal(t, fmt.Errorf("unsupported db type"), e, "err not nil")
	assert.Equal(t, "hzche", string(out.data), "no realizied yet")
}
*/
