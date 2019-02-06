package genval

import (
	"strings"
)

const (
	exprSep = "."
)

type expression struct {
	path []string
}

func Exp(str string) *expression {
	return &expression{strings.Split(str, exprSep)}
}

func (e expression) Empty() bool {
	return len(e.path) == 0
}

func (e expression) Size() int {
	return len(e.path)
}

func (e expression) GetAt(index int) string {
	if index < 0 || index >= len(e.path) {
		return ""
	}
	return e.path[index]
}

func (e expression) GetPath() []string {
	return e.path
}

func (e expression) String() string {
	return strings.Join(e.path, exprSep)
}
