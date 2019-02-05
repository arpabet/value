package genval_impl

import (
	"github.com/shvid/genval"
	"strings"
)

type Expr struct {
	path []string
}

func ParseExpr(str string) genval.Expression {
	return &Expr{strings.Split(str, ".")}
}

func (e Expr) Empty() bool {
	return len(e.path) == 0
}

func (e Expr) Size() int {
	return len(e.path)
}

func (e Expr) GetAt(index int) string {
	return e.path[index]
}

func (e Expr) GetPath() []string {
	return e.path
}

func (e Expr) String() string {
	return strings.Join(e.path, ".")
}
