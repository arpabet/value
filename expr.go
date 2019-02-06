package genval

import (
	"strings"
)

const (
	exprSep = "."
)

type Expression struct {
	path []string
}

func ParseExpr(str string) *Expression {
	return &Expression{strings.Split(str, exprSep)}
}

func (e Expression) Empty() bool {
	return len(e.path) == 0
}

func (e Expression) Size() int {
	return len(e.path)
}

func (e Expression) GetAt(index int) string {
	return e.path[index]
}

func (e Expression) GetPath() []string {
	return e.path
}

func (e Expression) String() string {
	return strings.Join(e.path, exprSep)
}
